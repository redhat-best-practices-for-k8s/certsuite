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
	"github.com/test-network-function/cnf-certification-test/internal/registry"
	"gopkg.in/yaml.v2"
)

var (
	filterCsi              = "&filter=csv_description=iregex=CSI;organization==certified-operators"
	containercatalogURL    = "https://catalog.redhat.com/api/containers/v1/images?page_size=%d&page=%d&filter=certified==true"
	operatorcatalogURL     = "https://catalog.redhat.com/api/containers/v1/operators/bundles?"
	helmcatalogURL         = "https://charts.openshift.io/index.yaml"
	containersRelativePath = "%s/cmd/tnf/fetch/data/containers/containers.db"
	operatorsRelativePath  = "%s/cmd/tnf/fetch/data/operators/"
	helmRelativePath       = "%s/cmd/tnf/fetch/data/helm/helm.db"
	certifiedcatalogdata   = "%s/cmd/tnf/fetch/data/archive.db"
	operatorFileFormat     = "operator_catalog_page_%d_%d.db"
)

const (
	containerCatalogPageSize = 500
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
	data := getCertifiedNumbers()
	log.Info(data)
	b, err := cmd.PersistentFlags().GetBool(operatorFlag)
	if err != nil {
		log.Error("Can't process the flag, ", operatorFlag)
		return err
	} else if b {
		getOperatorCatalog(&data)
	}
	b, err = cmd.PersistentFlags().GetBool(containerFlag)
	if err != nil {
		return err
	} else if b {
		getContainerCatalog(&data)
	}
	b, err = cmd.PersistentFlags().GetBool(helmFlag)
	if err != nil {
		return err
	} else if b {
		getHelmCatalog()
	}
	log.Info(data)
	serializeData(data)
	return nil
}

// getHTTPBody helper function to get binary data from URL
func getHTTPBody(url string) []uint8 {
	//nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		log.Errorf("Http request failed with error:%s", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Error("Error reading body ", err)
	}
	return body
}

func getCertifiedNumbers() CertifiedCatalog {
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
func getOperatorCatalogSize() (size, pagesize uint) {
	body := getHTTPBody(fmt.Sprintf("%spage=%d%s", operatorcatalogURL, 0, filterCsi))
	var aCatalog registry.OperatorCatalog
	err := json.Unmarshal(body, &aCatalog)
	if err != nil {
		log.Fatalf("Error in unmarshaling body: %v", err)
	}
	return aCatalog.Total, aCatalog.PageSize
}

func getOperatorCatalogPage(page, size uint) {
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	url := fmt.Sprintf("%spage=%d%s", operatorcatalogURL, page, filterCsi)
	body := getHTTPBody(url)
	filename := fmt.Sprintf(operatorsRelativePath+"/"+operatorFileFormat, path, page, size)

	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("couldn't open file ", err)
	}
	defer f.Close()
	_, err = f.Write(body)
	if err != nil {
		log.Error("can't write to file ", filename, err)
	}
}

func getOperatorCatalog(data *CertifiedCatalog) {
	start := time.Now()
	total, pageSize := getOperatorCatalogSize()
	if total == uint(data.Operators) {
		log.Info("no new certified operator found")
		return
	}
	removeOperatorsDB()
	log.Info("we should fetch new data", total, data.Operators)
	pages := total / pageSize
	remaining := total - pages*pageSize
	for page := uint(0); page < pages; page++ {
		getOperatorCatalogPage(page, pageSize)
	}
	if remaining != 0 {
		getOperatorCatalogPage(pages, remaining)
	}
	data.Operators = int(total)
	log.Info("time to process all the operators=", time.Since(start))
}

func getContainerCatalogSize() (total, pagesize uint) {
	url := fmt.Sprintf(containercatalogURL, 1, 1)
	body := getHTTPBody(url)
	var aCatalog registry.ContainerPageCatalog
	err := json.Unmarshal(body, &aCatalog)
	if err != nil {
		log.Fatalf("Error in unmarshaling body: %v", err)
	}
	return aCatalog.Total, uint(containerCatalogPageSize)
}

func getContainerCatalogPage(page, size uint, db map[string]*registry.ContainerCatalogEntry) {
	start := time.Now()
	log.Info("start fetching data of page ", page)
	url := fmt.Sprintf(containercatalogURL, size, page)
	body := getHTTPBody(url)
	log.Info("time to fetch binary data ", time.Since(start))
	start = time.Now()
	registry.LoadBinary(body, db)
	log.Info("time to load the data", time.Since(start))
}

func serializeContainersDB(db map[string]*registry.ContainerCatalogEntry) {
	start := time.Now()
	log.Info("start serializing container catalog")
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	filename := fmt.Sprintf(containersRelativePath, path)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Couldn't open file")
	}
	log.Trace("marshall container db into file=", f.Name())
	defer f.Close()
	bytes, _ := json.Marshal(db)
	_, err = f.Write(bytes)
	if err != nil {
		log.Error(err)
	}
	log.Info("serialization time", time.Since(start))
}

func getContainerCatalog(data *CertifiedCatalog) {
	start := time.Now()
	db := make(map[string]*registry.ContainerCatalogEntry)
	total, pageSize := getContainerCatalogSize()
	if total == uint(data.Containers) {
		log.Info("no new certified container found")
		return
	}
	removeContainersDB()
	pages := total / pageSize
	remaining := total - pages*pageSize
	for page := uint(0); page < pages; page++ {
		log.Info("getting data from page=", page, (pages - page), " pages to go")
		getContainerCatalogPage(page, pageSize, db)
	}
	if remaining != 0 {
		getContainerCatalogPage(pages, remaining, db)
	}
	serializeContainersDB(db)
	log.Info("time to serialize all the container=", time.Since(start))
	data.Containers = int(total)
	log.Info("time to process all the container=", time.Since(start))
}

func getHelmCatalog() {
	start := time.Now()
	removeHelmDB()
	body := getHTTPBody(helmcatalogURL)
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	filename := fmt.Sprintf(helmRelativePath, path)
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal("Couldn't open file")
	}
	_, err = f.Write(body)
	if err != nil {
		log.Error(err)
	}
	log.Info("time to process all the charts=", time.Since(start))
}

func removeContainersDB() {
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	filename := fmt.Sprintf(containersRelativePath, path)
	err = os.Remove(filename)
	if err != nil {
		log.Error("can't remove file", err)
	}
}
func removeHelmDB() {
	path, err := os.Getwd()
	if err != nil {
		log.Error("can't get current working dir", err)
		return
	}
	filename := fmt.Sprintf(helmRelativePath, path)
	err = os.Remove(filename)
	if err != nil {
		log.Error("can't remove file", err)
	}
}
func removeOperatorsDB() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	path = fmt.Sprintf(operatorsRelativePath, path)
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", path, file.Name())
		if err = os.Remove(filePath); err != nil {
			log.Error("can't remove file ", filePath)
		}
	}
}
