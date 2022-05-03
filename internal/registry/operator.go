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
package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	CSVNAMELENGTH = 2
)

var (
	operatorLoaded    = false
	operatorsFilePath = "%s/../cmd/tnf/fetch/data/operators/"
	operatordb        = make(map[string][]OperatorOcpVersionMatch)
)

type OperatorCatalog struct {
	Page     uint           `json:"page"`
	PageSize uint           `json:"page_size"`
	Total    uint           `json:"total"`
	Data     []OperatorData `json:"data"`
	NodeName string         `json:"nodeName"`
}

type OperatorData struct {
	CsvName    string `json:"csv_name"`
	OcpVersion string `json:"ocp_version"`
}

type OperatorOcpVersionMatch struct {
	ocpVersion      string
	operatorVersion string
}

func buildOperatorKey(op *OperatorData) (operatorName, operatorVersion, ocpVersion string, err error) {
	if len(strings.Split(op.CsvName, ".")) < CSVNAMELENGTH {
		return "", "", "", errors.New("non valid operator")
	}
	operatorName = strings.Split(op.CsvName, ".")[0]
	operatorVersion = strings.Split(op.CsvName, operatorName+".")[1]
	return operatorName, operatorVersion, op.OcpVersion, nil
}

func extractNameVersionFromName(operatorName string) (name, version string) {
	if len(strings.Split(operatorName, ".")) < CSVNAMELENGTH {
		return operatorName, ""
	}
	name = strings.Split(operatorName, ".")[0]
	version = strings.Split(operatorName, name+".")[1]
	return name, version
}

//nolint:funlen
func loadOperatorsCatalog(pathToRoot string) {
	if operatorLoaded {
		log.Trace("operator catalog already loaded, return")
		return
	}
	var fullCatalog OperatorCatalog
	operatorLoaded = true
	path := fmt.Sprintf(operatorsFilePath, pathToRoot)
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		filePath := fmt.Sprintf("%s/%s", path, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			log.Error("can't process file", file.Name(), err, " trying to proceed")
			f.Close()
			continue
		}
		log.Trace("load fine ", filePath)
		bytes, err := io.ReadAll(f)
		if err != nil {
			f.Close()
			log.Error("can't process file", file.Name(), err, " trying to proceed")
		}
		err = json.Unmarshal(bytes, &fullCatalog)
		if err != nil {
			log.Error("can't unmarshal file", file.Name(), err, " trying to proceed")
			f.Close()
			continue
		}

		for i := 0; i < len(fullCatalog.Data); i++ {
			if opName, opV, ocpV, err := buildOperatorKey(&fullCatalog.Data[i]); err == nil {
				operatordb[opName] = append(operatordb[opName], OperatorOcpVersionMatch{ocpVersion: ocpV, operatorVersion: opV})
			}
		}
		f.Close()
	}
}

// isOperatorCertified check the presence of operator name in certified operators db
// the operator name is the csv
// ocpVersion is Major.Minor OCP version
func IsOperatorCertified(operatorName, ocpVersion string) bool {
	name, operatorVersion := extractNameVersionFromName(operatorName)
	if v, ok := operatordb[name]; ok {
		for _, version := range v {
			if version.operatorVersion == operatorVersion && (ocpVersion == "" || version.ocpVersion == ocpVersion) {
				log.Trace("operator ", name, " found in db")
				return true
			}
		}
		log.Info("operator found on db, but not this particular version")
	}
	return false
}
