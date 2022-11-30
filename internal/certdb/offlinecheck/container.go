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
package offlinecheck

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	containerdb            = make(map[string]*ContainerCatalogEntry)
	containersRelativePath = "%s/data/containers/containers.db"
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

func loadContainersCatalog(offlineDBPath string) error {
	if containersLoaded {
		return nil
	}
	containersLoaded = true
	filename := fmt.Sprintf(containersRelativePath, offlineDBPath)
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open file %s, err: %v", filename, err)
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("cannot read file %s, err: %v", filename, err)
	}
	err = json.Unmarshal(bytes, &containerdb)
	if err != nil {
		return fmt.Errorf("cannot marshall file %s, err: %v", filename, err)
	}

	return nil
}

func LoadBinary(bytes []byte, db map[string]*ContainerCatalogEntry) (entries int, err error) {
	aCatalog := ContainerPageCatalog{}
	err = json.Unmarshal(bytes, &aCatalog)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshall binary data: %w, data: %s", err, string(bytes))
	}

	entries = len(aCatalog.Data)
	for i := 0; i < entries; i++ {
		c := aCatalog.Data[i]
		db[c.DockerImageDigest] = &c
	}

	return entries, nil
}

func (validator OfflineValidator) IsContainerCertified(registry, repository, tag, digest string) bool {
	const tagLatest = "latest"

	if digest != "" {
		if _, ok := containerdb[digest]; ok {
			logrus.Trace("container is certified based on digest", digest)
			return true
		}
		return false
	}

	// When tag is not provided, we'll default it to 'latest'
	if tag == "" {
		tag = tagLatest
	}

	// This is a non optimized code to process the certified containers
	// The reason behind it is users do not necessarily use image digest
	// in deployment file. The code runs under 100 us: not an issue in our case.
	for _, c := range containerdb {
		for _, repo := range c.Repositories {
			if repo.Registry == registry && repo.Repository == repository {
				for _, t := range repo.Tags {
					if t.Name == tag {
						logrus.Trace(fmt.Sprintf("container is not certified %s/%s:%s %s", registry, repository, tag, digest))
						return true
					}
				}
			}
		}
	}
	logrus.Error(fmt.Sprintf("container is not certified %s/%s:%s %s", registry, repository, tag, digest))
	return false
}
