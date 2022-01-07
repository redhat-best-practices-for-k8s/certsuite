package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// Endpoints document can be found here
// https://docs.engineering.redhat.com/pages/viewpage.action?spaceKey=EXD&title=Pyxis
// There are external and internal endpoints. External doesn't need authentication
// Here we are using only External endpoint to collect published containers and operator information

const apiContainerCatalogExternalBaseEndPoint = "https://catalog.redhat.com/api/containers/v1"
const apiOperatorCatalogExternalBaseEndPoint = "https://catalog.redhat.com/api/containers/v1/operators"
const apiCatalogByRepositoriesBaseEndPoint = "https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.access.redhat.com/repository"

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
type CertAPIClient struct {
	Client HTTPClient
}

// NewHTTPClient return new http client
func NewHTTPClient() CertAPIClient {
	return CertAPIClient{Client: &http.Client{}}
}

// IsContainerCertified get container image info by repo/name and checks if container details is present
// If present then returns `true` as certified operators.
func (api CertAPIClient) IsContainerCertified(repository, imageName string) bool {
	if imageID, err := api.GetImageIDByRepository(repository, imageName); err != nil || imageID == "" {
		return false
	}
	return true
}

// IsOperatorCertified get operator bundle by package name and check if package details is present
// If present then returns `true` as certified operators.
func (api CertAPIClient) IsOperatorCertified(org, packageName string) bool {
	if imageID, err := api.GetOperatorBundleIDByPackageName(org, packageName); err != nil || imageID == "" {
		return false
	}
	return true
}

// GetImageByID get container image data for the given container Id
func (api CertAPIClient) GetImageByID(id string) (response string, err error) {
	var responseData []byte
	url := fmt.Sprintf("%s/images/id/%s", apiContainerCatalogExternalBaseEndPoint, id)
	if responseData, err = api.getRequest(url); err == nil {
		response = string(responseData)
	}
	return
}

// GetImageIDByRepository get container image data for the given container Id
func (api CertAPIClient) GetImageIDByRepository(repository, imageName string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf("%s/%s/%s/images?page_size=1", apiCatalogByRepositoriesBaseEndPoint, repository, imageName)
	if responseData, err = api.getRequest(url); err == nil {
		imageID, err = api.getIDFromResponse(responseData)
	}
	return
}

// GetOperatorBundleIDByPackageName get published operator bundle Id by organization and package name
func (api CertAPIClient) GetOperatorBundleIDByPackageName(org, name string) (imageID string, err error) {
	var responseData []byte
	url := fmt.Sprintf("%s/bundles?page_size=1&organization=%s&package=%s", apiOperatorCatalogExternalBaseEndPoint, org, name)
	if responseData, err = api.getRequest(url); err == nil {
		imageID, err = api.getIDFromResponse(responseData)
	}
	return
}

// getRequest a http call to rest api, returns byte array or error
func (api CertAPIClient) getRequest(url string) (response []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody) //nolint:noctx
	if err != nil {
		return nil, err
	}
	resp, err := api.Client.Do(req)
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
func (api CertAPIClient) getIDFromResponse(response []byte) (id string, err error) {
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
					if res, ok := api.Find(a, idKey); ok {
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
func (api CertAPIClient) Find(obj interface{}, key string) (interface{}, bool) {
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
			if res, ok := api.Find(m, key); ok {
				return res, true
			}
		}
		// if the value is an array, search recursively
		// from each element
		if va, ok := v.([]interface{}); ok {
			for _, a := range va {
				if res, ok := api.Find(a, key); ok {
					return res, true
				}
			}
		}
	}
	// element not found
	return nil, false
}
