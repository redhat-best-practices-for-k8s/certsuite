package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/api/offlinecheck"
	"gopkg.in/yaml.v3"
)

var (
	containersCatalogSizeURL = "https://catalog.redhat.com/api/containers/v1/images?filter=certified==true&page=0&include=total,page_size"
	containersCatalogPageURL = "https://catalog.redhat.com/api/containers/v1/images?filter=certified==true&page_size=%d&page=%d&include=data.repositories,data.docker_image_digest,data.architecture"
	operatorsCatalogSizeURL  = "https://catalog.redhat.com/api/containers/v1/operators/bundles?filter=organization==certified-operators"
	operatorsCatalogPageURL  = "https://catalog.redhat.com/api/containers/v1/operators/bundles?filter=organization==certified-operators&page_size=%d&page=%d"
	helmCatalogURL           = "https://charts.openshift.io/index.yaml"
	containersRelativePath   = "%s/cmd/tnf/fetch/data/containers/containers.db"
	operatorsRelativePath    = "%s/cmd/tnf/fetch/data/operators/"
	helmRelativePath         = "%s/cmd/tnf/fetch/data/helm/helm.db"
	certifiedcatalogdata     = "%s/cmd/tnf/fetch/data/archive.json"
	operatorFileFormat       = "operator_catalog_page_%d_%d.db"
)

var (
	command = &cobra.Command{
		Use:   "fetch",
		Short: "fetch the list of certified operators and containers.",
		RunE:  RunCommand,
	}
	operatorFlag  = "operator"
	containerFlag = "container"
	helmFlag      = "helm"
)

type CertifiedCatalog struct {
	Containers int `json:"containers"`
	Operators  int `json:"operators"`
	Charts     int `json:"charts"`
}

func NewCommand() *cobra.Command {
	command.PersistentFlags().BoolP(operatorFlag, "o", false,
		"if specified, the operators DB will be updated")
	command.PersistentFlags().BoolP(containerFlag, "c", false,
		"if specified, the certified containers DB will be updated")
	command.PersistentFlags().BoolP(helmFlag, "m", false,
		"if specified, the helm file will be updated")
	return command
}

// RunCommand execute the fetch subcommands
func RunCommand(cmd *cobra.Command, args []string) error {
	data := getCertifiedCatalogOnDisk()
	log.Infof("Current offline artifacts: %+v", data)
	b, err := cmd.PersistentFlags().GetBool(operatorFlag)
	if err != nil {
		log.Error("Can't process the flag, ", operatorFlag)
		return err
	} else if b {
		err = getOperatorCatalog(&data)
		if err != nil {
			log.Fatalf("fetching operators failed: %v", err)
		}
	}
	b, err = cmd.PersistentFlags().GetBool(containerFlag)
	if err != nil {
		return err
	} else if b {
		err = getContainerCatalog(&data)
		if err != nil {
			log.Fatalf("fetching containers failed: %v", err)
		}
	}
	b, err = cmd.PersistentFlags().GetBool(helmFlag)
	if err != nil {
		return err
	} else if b {
		err = getHelmCatalog()
		if err != nil {
			log.Fatalf("fetching helm charts failed: %v", err)
		}
	}

	log.Info(data)
	serializeData(data)
	return nil
}

// getHTTPBody helper function to get binary data from URL
func getHTTPBody(url string) ([]uint8, error) {
	//nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http request %s failed with error: %w", url, err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body from %s: %w, body: %s", url, err, string(body))
	}
	return body, nil
}

func getCertifiedCatalogOnDisk() CertifiedCatalog {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	filePath := fmt.Sprintf(certifiedcatalogdata, path)
	if _, err = os.Stat(filePath); err != nil {
		return CertifiedCatalog{0, 0, 0}
	}
	f, err := os.Open(filePath)
	if err != nil {
		log.Error("can't process file", err, " trying to proceed")
		return CertifiedCatalog{0, 0, 0}
	}
	defer f.Close()
	bytes, err := io.ReadAll(f)
	if err != nil {
		log.Error("can't process file", err, " trying to proceed")
	}
	var data CertifiedCatalog
	if err = yaml.Unmarshal(bytes, &data); err != nil {
		log.Error("error when parsing the data", err)
	}
	return data
}

func serializeData(data CertifiedCatalog) {
	start := time.Now()
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	filename := fmt.Sprintf(certifiedcatalogdata, path)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Couldn't open file")
	}
	log.Trace("marshall container db into file=", f.Name())
	defer f.Close()
	bytes, _ := json.Marshal(data)
	_, err = f.Write(bytes)
	if err != nil {
		log.Error(err)
	}
	log.Info("serialization time", time.Since(start))
}

func getOperatorCatalogSize() (size, pagesize uint, err error) {
	log.Infof("Getting operators catalog size, url: %s", operatorsCatalogSizeURL)

	body, err := getHTTPBody(operatorsCatalogSizeURL)
	if err != nil {
		return 0, 0, err
	}

	var aCatalog offlinecheck.OperatorCatalog
	err = json.Unmarshal(body, &aCatalog)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to unmarshall response from %s: %w, body: %s",
			operatorsCatalogSizeURL, err, string(body))
	}

	return aCatalog.Total, aCatalog.PageSize, nil
}

func getOperatorCatalogPage(page, size uint, isLastPage bool) error {
	const (
		excludeFilter            = "&exclude=page,total,page_size"
		excludeFilterForLastPage = "&exclude=page,page_size"
	)

	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	url := fmt.Sprintf(operatorsCatalogPageURL, size, page)
	// Add "total" count only in the last page.
	if isLastPage {
		url += excludeFilterForLastPage
	} else {
		url += excludeFilter
	}

	log.Infof("Getting operators catalog page %d, url: %s", page, url)

	body, err := getHTTPBody(url)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf(operatorsRelativePath+"/"+operatorFileFormat, path, page, size)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}
	return nil
}

//nolint:funlen
func getOperatorCatalog(data *CertifiedCatalog) error {
	start := time.Now()
	total, pageSize, err := getOperatorCatalogSize()
	if err != nil {
		return fmt.Errorf("failed to get operators catalog size: %w", err)
	}

	log.Infof("Certified operators in the online catalog: %d, page size: %d", total, pageSize)
	if total == uint(data.Operators) {
		log.Info("No new certified operator found")
		return nil
	}

	err = removeOperatorsDB()
	if err != nil {
		return fmt.Errorf("failed to remove operators db: %w", err)
	}

	pages := total / pageSize
	remaining := total - pages*pageSize
	log.Infof("Downloading %d pages of size %d plus another page for the %d remaining entries.",
		pages, pageSize, remaining)

	for page := uint(0); page < pages; page++ {
		isLastPage := remaining == 0 && page == (pages-1)
		err = getOperatorCatalogPage(page, pageSize, isLastPage)
		if err != nil {
			return fmt.Errorf("failed to get operators page %d (total %d)", page, total)
		}
	}
	if remaining != 0 {
		err = getOperatorCatalogPage(pages, remaining, true)
		if err != nil {
			return fmt.Errorf("failed to get remaining operators page %d (total %d)", pages, total)
		}
	}

	data.Operators = int(total)

	log.Info("Time to process all the operators: ", time.Since(start))
	return nil
}

func getContainerCatalogSize() (total, pagesize uint, err error) {
	log.Infof("Getting containers catalog size, url: %s", containersCatalogSizeURL)
	body, err := getHTTPBody(containersCatalogSizeURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get url %s response body: %w", containersCatalogSizeURL, err)
	}

	var aCatalog offlinecheck.ContainerPageCatalog
	err = json.Unmarshal(body, &aCatalog)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to unmarshall body from url %s: %w, body: %s", containersCatalogSizeURL, err, string(body))
	}
	return aCatalog.Total, aCatalog.PageSize, nil
}

func getContainerCatalogPage(page, size uint, db map[string]*offlinecheck.ContainerCatalogEntry) error {
	start := time.Now()

	url := fmt.Sprintf(containersCatalogPageURL, size, page)
	log.Infof("Getting containers catalog page %d, url: %s", page, url)

	body, err := getHTTPBody(url)
	if err != nil {
		return fmt.Errorf("failed to get containers page %s: %w", url, err)
	}

	log.Info("Time to fetch binary data: ", time.Since(start))

	start = time.Now()
	err = offlinecheck.LoadBinary(body, db)
	if err != nil {
		return fmt.Errorf("failed to load binary data: %w", err)
	}

	log.Info("Time to load the data: ", time.Since(start))
	return nil
}

func serializeContainersDB(db map[string]*offlinecheck.ContainerCatalogEntry) error {
	start := time.Now()
	log.Info("start serializing container catalog")
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	filename := fmt.Sprintf(containersRelativePath, path)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed go create file %s: %w", filename, err)
	}

	log.Trace("marshall container db into file=", f.Name())
	defer f.Close()
	bytes, _ := json.Marshal(db)
	_, err = f.Write(bytes)
	if err != nil {
		return fmt.Errorf("failed to write into file %s: %w", filename, err)
	}

	log.Info("serialization time", time.Since(start))
	return nil
}

//nolint:funlen
func getContainerCatalog(data *CertifiedCatalog) error {
	start := time.Now()
	db := make(map[string]*offlinecheck.ContainerCatalogEntry)
	total, pageSize, err := getContainerCatalogSize()
	if err != nil {
		return fmt.Errorf("failed to get first page: %w", err)
	}

	log.Infof("Certified containers in the online catalog: %d, page size: %d", total, pageSize)
	if total == uint(data.Containers) {
		log.Info("No new certified container found")
		return nil
	}

	err = removeContainersDB()
	if err != nil {
		return fmt.Errorf("failed to remove containers db: %w", err)
	}

	pages := total / pageSize
	remaining := total - pages*pageSize
	log.Infof("Downloading %d pages of size %d plus another page for the %d remaining entries.",
		pages, pageSize, remaining)

	for page := uint(0); page < pages; page++ {
		err = getContainerCatalogPage(page, pageSize, db)
		if err != nil {
			return fmt.Errorf("failed to get containers page %d (total %d): %w", pages, total, err)
		}
	}
	if remaining != 0 {
		err = getContainerCatalogPage(pages, remaining, db)
		if err != nil {
			return fmt.Errorf("failed to get remaining containers page %d (total %d): %w", pages, total, err)
		}
	}

	serializeStart := time.Now()
	err = serializeContainersDB(db)
	if err != nil {
		return fmt.Errorf("failed to serialize containers db: %w", err)
	}

	data.Containers = int(total)

	log.Info("Time to serialize all the container: ", time.Since(serializeStart))
	log.Info("Time to process all the container: ", time.Since(start))

	return nil
}

func getHelmCatalog() error {
	start := time.Now()
	err := removeHelmDB()
	if err != nil {
		return err
	}

	log.Infof("Getting helm charts catalog page, url: %s", helmCatalogURL)
	body, err := getHTTPBody(helmCatalogURL)
	if err != nil {
		return err
	}

	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	filename := fmt.Sprintf(helmRelativePath, path)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	_, err = f.Write(body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filename, err)
	}

	log.Info("Time to process all the charts: ", time.Since(start))
	return nil
}

func removeContainersDB() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	filename := fmt.Sprintf(containersRelativePath, path)
	err = os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file %s: %w", filename, err)
	}

	return nil
}
func removeHelmDB() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	filename := fmt.Sprintf(helmRelativePath, path)
	err = os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove file %s: %w", filename, err)
	}

	return nil
}
func removeOperatorsDB() error {
	path, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	path = fmt.Sprintf(operatorsRelativePath, path)
	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read %s files: %w", path, err)
	}
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", path, file.Name())
		if err = os.Remove(filePath); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", filePath, err)
		}
	}

	return nil
}
