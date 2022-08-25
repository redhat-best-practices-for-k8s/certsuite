// Copyright (C) 2020-2022 Red Hat, Inc.
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

	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/api/offlinecheck"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
	"helm.sh/helm/v3/pkg/release"
)

// Endpoints document can be found here
// https://docs.engineering.redhat.com/pages/viewpage.action?spaceKey=EXD&title=Pyxis
// There are external and internal endpoints. External doesn't need authentication
// Here we are using only External endpoint to collect published containers and operator information

const apiCatalogByRepositoriesBaseEndPoint = "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository"
const filterCertifiedOperatorsOrg = "organization==certified-operators"
const certifiedOperatorsCatalogURL = "https://catalog.redhat.com/api/containers/v1/operators/bundles?page_size=100&page=0&filter=csv_name==%s;%s"
const certifiedContainerCatalogURL = "https://catalog.redhat.com/api/containers/v1/repositories/registry/%s/repository/%s/images?"
const certifiedContainerCatalogDigestURL = "https://catalog.redhat.com/api/containers/v1/images/registry/%s/repository/%s/manifest_digest/%s"
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
func (checker OnlineValidator) IsServiceReachable() bool {
	if _, err := checker.GetRequest(redhatCatalogPingURL); err != nil {
		return false
	}
	if _, err := checker.GetRequest(redhatCatalogPingMongoDBURL); err != nil {
		return false
	}
	return true
}

// GetImageByID get container image Id using the digest.
// return imageID if entry exists,
// return empty string if entry does not exist
func (checker OnlineValidator) getImageByDigest(registry, repository, digest string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogDigestURL, registry, repository, digest)
	log.Trace(url)
	if responseData, err = checker.GetRequest(url); err != nil || len(responseData) == 0 {
		return imageID, nil
	}
	containerEntry := offlinecheck.ContainerCatalogEntry{}
	err = json.Unmarshal(responseData, &containerEntry)
	if err != nil {
		log.Error("Cannot marshall binary data", err)
		return
	}
	if containerEntry.Certified {
		return containerEntry.ID, nil
	}
	return imageID, errors.New("certified image not found")
}

// GetImageByID get container image Id using the tag.
// return imageID if entry exists,
// return empty string if entry does not exist
func (checker OnlineValidator) getImageByTag(registry, repository, tag string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogTagURL, registry, repository, tag)
	log.Trace(url)
	if responseData, err = checker.GetRequest(url); err != nil || len(responseData) == 0 {
		return imageID, err
	}
	db := make(map[string]*offlinecheck.ContainerCatalogEntry)
	err = offlinecheck.LoadBinary(responseData, db)
	if err != nil {
		return "", fmt.Errorf("failed to load binary data: %w", err)
	}

	if len(db) == 0 {
		return imageID, errors.New("certified image not found")
	}
	for _, v := range db {
		if v.Certified {
			return v.ID, nil
		}
	}
	return imageID, errors.New("certified image not found")
}

// GetImageByID get container image Id using the tag.
// return imageID if any of containers with same registry and repository is certified
// return empty string if entry does not exist
func (checker OnlineValidator) getImageByRepository(registry, repository string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf(certifiedContainerCatalogURL, registry, repository)
	log.Trace(url)
	if responseData, err = checker.GetRequest(url); err == nil {
		fmt.Println(string(responseData))
		imageID, _ = checker.getIDFromResponse(responseData)
	}
	db := make(map[string]*offlinecheck.ContainerCatalogEntry)
	err = offlinecheck.LoadBinary(responseData, db)
	if err != nil {
		return "", fmt.Errorf("failed to loadbinary data: %w", err)
	}

	if len(db) == 0 {
		return imageID, errors.New("certified image not found")
	}
	for _, v := range db {
		if v.Certified {
			return v.ID, nil
		}
	}
	return imageID, errors.New("certified image not found")
}

// IsContainerCertified get container image info by registry/repository [tag|digest]
// returns true if the container is present and is certified.
// returns false otherwise
func (checker OnlineValidator) IsContainerCertified(registry, repository, tag, digest string, justDigest bool) bool {
	if justDigest {
		if digest == "" {
			return false
		}
	}
	if digest != "" {
		if imageID, err := checker.getImageByDigest(registry, repository, digest); err != nil || imageID == "" {
			return false
		}
		return true
	}
	if tag != "" {
		if imageID, err := checker.getImageByTag(registry, repository, tag); err != nil || imageID == "" {
			return false
		}
		return true
	}
	if imageID, err := checker.getImageByRepository(registry, repository); err != nil || imageID == "" {
		return false
	}
	return true
}

// IsOperatorCertified get operator bundle by csv name from the certified-operators org
// If present then returns `true` if channel and ocp version match.
func (checker OnlineValidator) IsOperatorCertified(csvName, ocpVersion, channel string) bool {
	log.Tracef("Searching csv %s (channel %s) for ocp %q", csvName, channel, ocpVersion)
	_, operatorVersion := offlinecheck.ExtractNameVersionFromName(csvName)
	var responseData []byte
	var err error
	url := fmt.Sprintf(certifiedOperatorsCatalogURL, csvName, filterCertifiedOperatorsOrg)
	log.Trace(url)
	if responseData, err = checker.GetRequest(url); err != nil || len(responseData) == 0 {
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

func (checker OnlineValidator) IsReleaseCertified(helm *release.Release, ourKubeVersion string) bool {
	charts, err := checker.GetCertifiedCharts()
	if err != nil {
		return false
	}
	offlinecheck.LoadHelmCharts(charts)
	return offlinecheck.OfflineChecker{}.IsReleaseCertified(helm, ourKubeVersion)
}

// getRequest a http call to rest api, returns byte array or error
func (checker OnlineValidator) GetRequest(url string) (response []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := checker.Client.Do(req.WithContext(context.TODO()))
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
func (checker OnlineValidator) getIDFromResponse(response []byte) (id string, err error) {
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
					if res, ok := checker.Find(a, idKey); ok {
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
func (checker OnlineValidator) Find(obj interface{}, key string) (interface{}, bool) {
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
			if res, ok := checker.Find(m, key); ok {
				return res, true
			}
		}
		// if the value is an array, search recursively
		// from each element
		if va, ok := v.([]interface{}); ok {
			for _, a := range va {
				if res, ok := checker.Find(a, key); ok {
					return res, true
				}
			}
		}
	}
	// element not found
	return nil, false
}
func (checker OnlineValidator) GetCertifiedCharts() (offlinecheck.ChartStruct, error) {
	url := ("https://charts.openshift.io/index.yaml")
	responseData, err := checker.GetRequest(url)
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
func CreateContainerCatalogQueryURL(id configuration.ContainerImageIdentifier) string {
	var url string
	const defaultTag = "latest"
	const arch = "amd64"
	if id.Digest == "" {
		if id.Tag == "" {
			id.Tag = defaultTag
		}
		url = fmt.Sprintf("%s/%s/%s/images?filter=architecture==%s;repositories.repository==%s/%s;repositories.tags.name==%s",
			apiCatalogByRepositoriesBaseEndPoint, id.Repository, id.Name, arch, id.Repository, id.Name, id.Tag)
	} else {
		url = fmt.Sprintf("%s/%s/%s/images?filter=architecture==%s;image_id==%s", apiCatalogByRepositoriesBaseEndPoint, id.Repository, id.Name, arch, id.Digest)
	}
	return url
}
