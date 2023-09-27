// Copyright (C) 2020-2023 Red Hat, Inc.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program; if not, write to the Free Software Foundation, Inc.,
// 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
package onlinecheck

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/certdb/config"
	"github.com/test-network-function/cnf-certification-test/internal/certdb/offlinecheck"
	yaml "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/release"
)

// Endpoints document can be found here
// https://docs.engineering.redhat.com/pages/viewpage.action?spaceKey=EXD&title=Pyxis
// There are external and internal endpoints. External does not need authentication
// Here we are using only External endpoint to collect published containers and operator information

const filterCertifiedOperatorsOrg = "organization==certified-operators"
const certifiedOperatorsCatalogURL = "https://catalog.redhat.com/api/containers/v1/operators/bundles?page_size=100&page=0&filter=csv_name==%s;%s"
const certifiedContainerCatalogURL = "https://catalog.redhat.com/api/containers/v1/repositories/registry/%s/repository/%s/images?"
const certifiedContainerCatalogDigestURL = "https://catalog.redhat.com/api/containers/v1/images?filter=image_id==%s"
const certifiedContainerCatalogTagURL = "https://catalog.redhat.com/api/containers/v1/repositories/registry/%s/repository/%s/tag/%s"
const redhatCatalogPingURL = "https://catalog.redhat.com/api/containers/v1/ping"
const redhatCatalogPingMongoDBURL = "https://catalog.redhat.com/api/containers/v1/status/mongo"

var (
	dataKey           = "data"
	errorContainer404 = fmt.Errorf("error code 404: A container/operator with the specified identifier was not found")
	idKey             = "_id"
)

// GetContainer404Error return error object with 404 error string
func GetContainer404Error() error {
	return errorContainer404
}

// HTTPClient Client interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// CertAPIClient is http client to handle `pyxis` rest api
type OnlineValidator struct {
	Client HTTPClient
}

// NewOnlineValidator return an online implementation of the CertificationValidator interface
func NewOnlineValidator() OnlineValidator {
	return OnlineValidator{Client: &http.Client{}}
}

// IsServiceReachable check if redhat catalog is reachable and its database is available to query
func (validator OnlineValidator) IsServiceReachable() bool {
	if _, err := validator.GetRequest(redhatCatalogPingURL); err != nil {
		return false
	}
	if _, err := validator.GetRequest(redhatCatalogPingMongoDBURL); err != nil {
		return false
	}
	return true
}

// GetImageByID get container image Id using the digest.
// return imageID if entry exists,
// return empty string if entry does not exist
func (validator OnlineValidator) getImageByDigest(digest string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogDigestURL, digest)
	log.Trace(url)
	if responseData, err = validator.GetRequest(url); err != nil || len(responseData) == 0 {
		return imageID, nil
	}
	containerEntries := offlinecheck.ContainerPageCatalog{}
	err = json.Unmarshal(responseData, &containerEntries)
	if err != nil {
		log.Error("Cannot marshall binary data", err)
		return
	}
	if len(containerEntries.Data) > 0 {
		for _, repo := range containerEntries.Data[0].Repositories {
			if config.IsRegistryRedhatOnlyImages(repo.Registry) {
				log.Trace("This image is a Redhat provided image and is certified by default")
				return containerEntries.Data[0].ID, nil
			}
		}
		if containerEntries.Data[0].Certified {
			return containerEntries.Data[0].ID, nil
		}
	}

	return imageID, errors.New("certified image not found")
}

// GetImageByID get container image Id using the tag.
// return imageID if entry exists,
// return empty string if entry does not exist
func (validator OnlineValidator) getImageByTag(registry, repository, tag string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogTagURL, registry, repository, tag)
	log.Trace(url)
	if responseData, err = validator.GetRequest(url); err != nil || len(responseData) == 0 {
		return imageID, err
	}
	db := make(map[string]*offlinecheck.ContainerCatalogEntry)
	_, err = offlinecheck.LoadBinary(responseData, db)
	if err != nil {
		return "", fmt.Errorf("failed to load binary data: %w", err)
	}

	if len(db) == 0 {
		return imageID, errors.New("certified image not found")
	}
	for _, v := range db {
		for _, repo := range v.Repositories {
			if config.IsRegistryRedhatOnlyImages(repo.Registry) {
				log.Trace("This image is a Redhat provided image and is certified by default")
				return v.ID, nil
			}
		}
		if v.Certified {
			return v.ID, nil
		}
	}
	return imageID, errors.New("certified image not found")
}

// GetImageByID get container image Id using the tag.
// return imageID if any of containers with same registry and repository is certified
// return empty string if entry does not exist
func (validator OnlineValidator) getImageByRepository(registry, repository string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogURL, registry, repository)
	log.Trace(url)
	if responseData, err = validator.GetRequest(url); err == nil {
		fmt.Println(string(responseData))
		imageID, _ = validator.getIDFromResponse(responseData)
	}
	db := make(map[string]*offlinecheck.ContainerCatalogEntry)
	_, err = offlinecheck.LoadBinary(responseData, db)
	if err != nil {
		return "", fmt.Errorf("failed to loadbinary data: %w", err)
	}

	if len(db) == 0 {
		return imageID, errors.New("certified image not found")
	}
	for _, v := range db {
		for _, repo := range v.Repositories {
			if config.IsRegistryRedhatOnlyImages(repo.Registry) {
				log.Trace("This image is a Redhat provided image and is certified by default")
				return v.ID, nil
			}
		}
		if v.Certified {
			return v.ID, nil
		}
	}
	return imageID, errors.New("certified image not found")
}

// IsContainerCertified get container image info by registry/repository [tag|digest]
// returns true if the container is present and is certified.
// returns false otherwise
func (validator OnlineValidator) IsContainerCertified(registry, repository, tag, digest string) bool {
	// overwrite registry value with hardcoded one due to Pyxis implementation
	value, ok := config.HardcodedRegistryMapping[registry]
	if ok {
		registry = value
	}
	if digest != "" {
		if imageID, err := validator.getImageByDigest(digest); err != nil || imageID == "" {
			return false
		}
		return true
	}
	if tag != "" {
		if imageID, err := validator.getImageByTag(registry, repository, tag); err != nil || imageID == "" {
			return false
		}
		return true
	}
	if imageID, err := validator.getImageByRepository(registry, repository); err != nil || imageID == "" {
		return false
	}
	return true
}

// IsOperatorCertified get operator bundle by csv name from the certified-operators org
// If present then returns `true` if channel and ocp version match.
func (validator OnlineValidator) IsOperatorCertified(csvName, ocpVersion, channel string) bool {
	log.Tracef("Searching csv %s (channel %s) for ocp %q", csvName, channel, ocpVersion)
	_, operatorVersion := offlinecheck.ExtractNameVersionFromName(csvName)
	var responseData []byte
	var err error
	url := fmt.Sprintf(certifiedOperatorsCatalogURL, csvName, filterCertifiedOperatorsOrg)
	log.Trace(url)
	if responseData, err = validator.GetRequest(url); err != nil || len(responseData) == 0 {
		return false
	}
	operatorEntries := offlinecheck.OperatorCatalog{}
	err = json.Unmarshal(responseData, &operatorEntries)
	if err != nil {
		log.Error("Cannot marshall binary data", err)
		return false
	}
	for _, operator := range operatorEntries.Data {
		_, opVersion := offlinecheck.ExtractNameVersionFromName(operator.CsvName)
		if (opVersion == operatorVersion) && (operator.OcpVersion == ocpVersion || ocpVersion == "") && operator.Channel == channel {
			return true
		}
	}
	return false
}

func (validator OnlineValidator) IsHelmChartCertified(helm *release.Release, ourKubeVersion string) bool {
	charts, err := validator.GetCertifiedCharts()
	if err != nil {
		return false
	}
	offlinecheck.LoadHelmCharts(charts)
	return offlinecheck.OfflineValidator{}.IsHelmChartCertified(helm, ourKubeVersion)
}

// getRequest a http call to rest api, returns byte array or error
func (validator OnlineValidator) GetRequest(url string) (response []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := validator.Client.Do(req.WithContext(context.TODO()))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		err = GetContainer404Error()
		return
	}
	if response, err = io.ReadAll(resp.Body); err != nil {
		err = GetContainer404Error()
		return
	}
	return
}

// getIDFromResponse searches for first occurrence of id and return
func (validator OnlineValidator) getIDFromResponse(response []byte) (id string, err error) {
	var data interface{}
	if err = json.Unmarshal(response, &data); err != nil {
		log.Errorf("Error calling API Request %v", err.Error())
		err = GetContainer404Error()
		return
	}
	m := data.(map[string]interface{})
	for k, v := range m {
		if k == dataKey {
			// if the value is an array, search recursively
			// from each element
			if va, ok := v.([]interface{}); ok {
				for _, a := range va {
					if res, ok := validator.Find(a, idKey); ok {
						id = fmt.Sprintf("%v", res)
						break
					}
				}
			}
		}
	}

	return
}

// Find key in interface (recursively) and return value as interface
func (validator OnlineValidator) Find(obj interface{}, key string) (interface{}, bool) {
	// if the argument is not a map, ignore it
	mobj, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false
	}
	for k, v := range mobj {
		// key match, return value
		if k == key {
			return v, true
		}
		// if the value is a map, search recursively
		if m, ok := v.(map[string]interface{}); ok {
			if res, ok := validator.Find(m, key); ok {
				return res, true
			}
		}
		// if the value is an array, search recursively
		// from each element
		if va, ok := v.([]interface{}); ok {
			for _, a := range va {
				if res, ok := validator.Find(a, key); ok {
					return res, true
				}
			}
		}
	}
	// element not found
	return nil, false
}
func (validator OnlineValidator) GetCertifiedCharts() (offlinecheck.ChartStruct, error) {
	url := ("https://charts.openshift.io/index.yaml")
	responseData, err := validator.GetRequest(url)
	var charts offlinecheck.ChartStruct
	if err != nil {
		log.Error("error reading the helm certification list ", err)
		return charts, err
	}
	if err = yaml.Unmarshal(responseData, &charts); err != nil {
		log.Error("error while parsing the yaml file of the helm certification list ", err)
	}
	return charts, err
}
