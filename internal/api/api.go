package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-yaml/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/pkg/configuration"
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

type ContainerCatalogEntry struct {
	ID string `json:"_id"`
	/*Links struct {
		RpmManifest struct {
			Href string `json:"href"`
		} `json:"rpm_manifest"`
		Vulnerabilities struct {
			Href string `json:"href"`
		} `json:"vulnerabilities"`
	} `json:"_links"`
	Architecture string `json:"architecture"`
	Brew         struct {
		Build          string    `json:"build"`
		CompletionDate time.Time `json:"completion_date"`
		Nvra           string    `json:"nvra"`
		Package        string    `json:"package"`
	} `json:"brew"`
	Certified       bool      `json:"certified"`
	ContentSets     []string  `json:"content_sets"`
	CpeIds          []string  `json:"cpe_ids"`
	CreationDate    time.Time `json:"creation_date"`
	DockerImageID   string    `json:"docker_image_id"`*/
	FreshnessGrades []ContainerImageFreshnessGrade `json:"freshness_grades"`
	/*
		ImageID        string    `json:"image_id"`
		LastUpdateDate time.Time `json:"last_update_date"`
		ObjectType     string    `json:"object_type"`
		ParsedData     struct {
			Architecture  string    `json:"architecture"`
			Command       string    `json:"command"`
			Comment       string    `json:"comment"`
			Created       time.Time `json:"created"`
			DockerVersion string    `json:"docker_version"`
			EnvVariables  []string  `json:"env_variables"`
			Labels        []struct {
				Name  string `json:"name"`
				Value string `json:"value"`
			} `json:"labels"`
			Layers                 []string `json:"layers"`
			Os                     string   `json:"os"`
			Size                   int      `json:"size"`
			UncompressedLayerSizes []struct {
				LayerID   string `json:"layer_id"`
				SizeBytes int    `json:"size_bytes"`
			} `json:"uncompressed_layer_sizes"`
			UncompressedSizeBytes int    `json:"uncompressed_size_bytes"`
			User                  string `json:"user"`
		} `json:"parsed_data"`
		Repositories []struct {
			Links struct {
				ImageAdvisory struct {
					Href string `json:"href"`
				} `json:"image_advisory"`
				Repository struct {
					Href string `json:"href"`
				} `json:"repository"`
			} `json:"_links"`
			Comparison struct {
				AdvisoryRpmMapping []struct {
					AdvisoryIds []string `json:"advisory_ids"`
					Nvra        string   `json:"nvra"`
				} `json:"advisory_rpm_mapping"`
				Reason     string `json:"reason"`
				ReasonText string `json:"reason_text"`
				Rpms       struct {
					Downgrade []interface{} `json:"downgrade"`
					New       []string      `json:"new"`
					Remove    []string      `json:"remove"`
					Upgrade   []string      `json:"upgrade"`
				} `json:"rpms"`
				WithNvr string `json:"with_nvr"`
			} `json:"comparison"`
			ContentAdvisoryIds    []string  `json:"content_advisory_ids"`
			ImageAdvisoryID       string    `json:"image_advisory_id"`
			ManifestListDigest    string    `json:"manifest_list_digest"`
			ManifestSchema2Digest string    `json:"manifest_schema2_digest"`
			Published             bool      `json:"published"`
			PublishedDate         time.Time `json:"published_date"`
			PushDate              time.Time `json:"push_date"`
			Registry              string    `json:"registry"`
			Repository            string    `json:"repository"`
			Signatures            []struct {
				KeyLongID string   `json:"key_long_id"`
				Tags      []string `json:"tags"`
			} `json:"signatures"`
			Tags []struct {
				Links struct {
					TagHistory struct {
						Href string `json:"href"`
					} `json:"tag_history"`
				} `json:"_links"`
				AddedDate time.Time `json:"added_date"`
				Name      string    `json:"name"`
			} `json:"tags"`
		} `json:"repositories"`
		SumLayerSizeBytes      int    `json:"sum_layer_size_bytes"`
		TopLayerID             string `json:"top_layer_id"`
		UncompressedTopLayerID string `json:"uncompressed_top_layer_id"`*/
}
type ChartStruct struct {
	Entries map[string][]struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		KubeVersion string `yaml:"kubeVersion"`
	} `yaml:"entries"`
}

type catalogQueryResponse struct {
	Page     uint `json:"page"`
	PageSize uint `json:"page_size"`
	Total    uint `json:"total"`
}

type ContainerImageFreshnessGrade struct {
	// CreationDate time.Time `json:"creation_date"`
	Grade string `json:"grade"`
	// StartDate    time.Time `json:"start_date"`
}

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
func (api CertAPIClient) IsOperatorCertified(org, packageName, version string) (bool, error) {
	imageID, err := api.GetOperatorBundleIDByPackageName(org, packageName, version)
	if err == nil {
		if imageID == "" {
			return false, nil
		}
		return true, nil
	}
	return false, err
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

type containerCatalogQueryResponse struct {
	catalogQueryResponse
	Data []ContainerCatalogEntry `json:"data"`
}

// GetContainerCatalogEntry gets the container image entry with highest freshness grade
func (api CertAPIClient) GetContainerCatalogEntry(id configuration.ContainerImageIdentifier) (*ContainerCatalogEntry, error) {
	responseData, err := api.getRequest(CreateContainerCatalogQueryURL(id))
	if err == nil {
		var response containerCatalogQueryResponse
		err = json.Unmarshal(responseData, &response)
		if err == nil && len(response.Data) > 0 {
			return &response.Data[0], nil
		}
	}
	return nil, err
}

// GetOperatorBundleIDByPackageName get published operator bundle Id by organization and package name.
// Returns (ImageID, error).
func (api CertAPIClient) GetOperatorBundleIDByPackageName(org, name, vsersion string) (string, error) {
	var imageID string
	url := ""
	if vsersion != "" {
		url = fmt.Sprintf("%s/bundles?page_size=1&filter=organization==%s;csv_name==%s;ocp_version==%s", apiOperatorCatalogExternalBaseEndPoint, org, name, vsersion)
	} else {
		url = fmt.Sprintf("%s/bundles?page_size=1&filter=organization==%s;csv_name==%s", apiOperatorCatalogExternalBaseEndPoint, org, name)
	}

	responseData, err := api.getRequest(url)
	if err == nil {
		imageID, err = api.getIDFromResponse(responseData)
	}

	return imageID, err
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
func (api CertAPIClient) GetYamlFile() (ChartStruct, error) {
	url := ("https://charts.openshift.io/index.yaml")
	responseData, err := api.getRequest(url)
	var charts ChartStruct
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

func (e ContainerCatalogEntry) GetBestFreshnessGrade() string {
	grade := "F"
	for _, g := range e.FreshnessGrades {
		if g.Grade < grade {
			grade = g.Grade
		}
	}
	return grade
}
