package registry

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func LoadCatalogs() {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	loadContainersCatalog(path)
	loadHelmCatalog(path)
	loadOperatorsCatalog(path)
}
