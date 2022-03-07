package api_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/internal/api"
)

const (
	id                 = "5ea8cf595a13466876a10215"
	imageName          = "nginx-116"
	marketPlaceOrg     = "redhat-marketplace"
	packageName        = "amq-streams"
	redHatOrg          = "redhat-operators"
	repository         = "rhel8"
	unKnownRepository  = "wrong_repo"
	unKnownImageName   = "wrong_id"
	unknownPackageName = "unknownPackage"
	version            = "4.8"
	jsonResponseFound  = `{
	"data": [{
		"_id": "5ea8cf595a13466876a10215",
		"_links": {
			"certification_project": {
				"href": "/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-116/projects/certification"
			},
			"images": {
				"href": "/v1/repositories/registry/registry.access.redhat.com/repository/rhel8/nginx-116/images"
			},
			"vendor": {
				"href": "/v1/vendors/label/redhat"
			}
		},
		"application_categories": [
			"Web Services"
		]
	}]
}`
	jsonResponseNotFound = `{
				  "detail": "The requested URL was not found on the server. If you entered the URL manually please check your spelling and try again.",
				  "status": 404,
				  "title": "Not Found",
				  "type": "about:blank"
					}`
)

var (
	client = api.CertAPIClient{}
	// GetDoFunc fetches the mock client's `Do` func
	GetDoFunc func(req *http.Request) (*http.Response, error)
)

// MockClient is the mock client
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

//nolint:gochecknoinits
func init() {
	client.Client = &MockClient{}
}

var (
	containerTestCases = []struct {
		repository     string
		name           string
		id             string
		expectedError  error
		expectedResult bool
		responseData   string
		responseStatus int
	}{
		{repository: repository, name: imageName, expectedError: nil, id: "", expectedResult: true,
			responseData: jsonResponseFound, responseStatus: http.StatusAccepted},
		{repository: unKnownRepository, name: unKnownImageName, expectedError: api.GetContainer404Error(), id: "", expectedResult: false,
			responseData: jsonResponseNotFound, responseStatus: http.StatusNotFound},
	}

	operatorTestCases = []struct {
		packageName         string
		org                 string
		id                  string
		expectedErrorString string
		expectedResult      bool
		responseData        string
		responseStatus      int
		version             string
	}{
		{packageName: packageName, org: redHatOrg, expectedErrorString: "", id: "", expectedResult: true,
			responseData: jsonResponseFound, responseStatus: http.StatusAccepted, version: version},
		{packageName: unknownPackageName, org: marketPlaceOrg, expectedErrorString: api.GetContainer404Error().Error(), id: "", expectedResult: false,
			responseData: jsonResponseNotFound, responseStatus: http.StatusNotFound, version: version},
	}
)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return GetDoFunc(req)
}

func getDoFunc(data string, status int) func(req *http.Request) (*http.Response, error) {
	response := io.NopCloser(bytes.NewReader([]byte(data)))
	defer response.Close()
	return func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       response,
		}, nil
	}
}
func TestApiClient_IsContainerCertified(t *testing.T) {
	for _, c := range containerTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result := client.IsContainerCertified(c.repository, c.name)
		assert.Equal(t, c.expectedResult, result)
	}
}

func TestApiClient_IsOperatorCertified(t *testing.T) {
	for _, c := range operatorTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result, err := client.IsOperatorCertified(c.org, c.packageName, c.version)
		assert.Equal(t, c.expectedResult, result, err)
	}
}

func TestApiClient_GetImageById(t *testing.T) {
	containerTestCases[0].id = id
	for _, c := range containerTestCases {
		GetDoFunc = getDoFunc(c.responseData, c.responseStatus) //nolint:bodyclose
		result, err := client.GetImageByID(c.id)
		assert.Equal(t, c.expectedError, err)
		if err == nil {
			assert.True(t, len(result) > 0)
		}
	}
}

func TestCertApiClient_Find(t *testing.T) {
	testData := []struct {
		data map[string]interface{}
	}{
		{data: map[string]interface{}{"_id": id}},
		{data: map[string]interface{}{"index": "index.html",
			"specs": map[string]interface{}{
				"_id":  id,
				"edit": "edit.html",
			}}},
	}
	for _, c := range testData {
		val, found := client.Find(c.data, "_id")
		assert.True(t, found)
		assert.Equal(t, id, fmt.Sprintf("%v", val))
	}
}
