package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	containerdb            = make(map[string]*ContainerCatalogEntry)
	containersRelativePath = "%s/../cmd/tnf/fetch/data/containers/containers.db"
	containersLoaded       = false
)

type Tag struct {
	Name string `json:"name"`
}
type Repository struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tags       []Tag  `json:"tags"`
}

type ContainerCatalogEntry struct {
	ID                string       `json:"_id"`
	Architecture      string       `json:"architecture"`
	Certified         bool         `json:"certified"`
	DockerImageDigest string       `json:"docker_image_digest"`
	DockerImageID     string       `json:"docker_image_id"` // image digest
	Repositories      []Repository `json:"repositories"`
}
type ContainerPageCatalog struct {
	Page     uint                    `json:"page"`
	PageSize uint                    `json:"page_size"`
	Total    uint                    `json:"total"`
	Data     []ContainerCatalogEntry `json:"data"`
}

func loadContainersCatalog(pathToRoot string) {
	if containersLoaded {
		return
	}
	containersLoaded = true
	filename := fmt.Sprintf(containersRelativePath, pathToRoot)
	f, err := os.Open(filename)
	if err != nil {
		logrus.Errorln("can't open file", filename, err)
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		logrus.Error("can't process file", f.Name(), err, " trying to proceed")
	}
	err = json.Unmarshal(bytes, &containerdb)
	if err != nil {
		logrus.Error("can't marshall file", filename, err, " trying to proceed")
	}
}

func LoadBinary(bytes []byte, db map[string]*ContainerCatalogEntry) {
	aCatalog := ContainerPageCatalog{}
	err := json.Unmarshal(bytes, &aCatalog)
	if err != nil {
		logrus.Error("can't marshall binary data", err)
		return
	}
	for i := 0; i < len(aCatalog.Data); i++ {
		c := aCatalog.Data[i]
		if c.Certified {
			db[c.DockerImageDigest] = &c
		}
	}
}

func IsCertified(registry, repository, tag, digest string) bool {
	if digest != "" {
		if _, ok := containerdb[digest]; ok {
			logrus.Trace("container is certified based on digest", digest)
			return true
		}
		return false
	}
	// This is a non optimized code to process
	// the certified containers
	// The reason behind it is users don't necessarily use image digest
	// in deployment file.
	// The code runs under 100 us. Not an issue in our case
	for _, c := range containerdb {
		for _, repo := range c.Repositories {
			if repo.Registry == registry && repo.Repository == repository {
				for _, t := range repo.Tags {
					if t.Name == tag {
						logrus.Trace("container is certified :", repo.Registry, repo.Repository, tag)
						return true
					}
				}
			}
		}
	}
	return false
}
