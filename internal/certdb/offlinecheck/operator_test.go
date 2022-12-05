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
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const operatorDBJSON = `{
	"data":[
	   {
		  "_id":"5f8f29d33b6621a763342f7c",
		  "alm_examples":[
			 
		  ],
		  "annotations":{
			 "infrastructure_features":[
				
			 ],
			 "valid_subscription":[
				
			 ]
		  },
		  "architectures":[
			 
		  ],
		  "bundle_path":"registry.connect.redhat.com/ibm/ibm-spectrum-scale-csi-operator-bundle@sha256:70f310cb36f6f58221377ac77bfacce3ea80d11811b407131d9723507eaada42",
		  "bundle_path_digest":"sha256:70f310cb36f6f58221377ac77bfacce3ea80d11811b407131d9723507eaada42",
		  "capabilities":[
			 ""
		  ],
		  "channel_name":"stable",
		  "creation_date":"2020-10-20T18:17:55.675000+00:00",
		  "csv_description":"",
		  "csv_display_name":"",
		  "csv_metadata_description":"",
		  "csv_name":"ibm-spectrum-scale-csi-operator.v2.0.0",
		  "in_index_img":true,
		  "install_modes":[
			 
		  ],
		  "is_default_channel":true,
		  "last_update_date":"2022-10-24T11:20:28.738000+00:00",
		  "latest_in_channel":false,
		  "ocp_version":"4.6",
		  "organization":"certified-operators",
		  "package":"ibm-spectrum-scale-csi",
		  "provided_apis":[
			 {
				"group":"csi.ibm.com",
				"kind":"CSIScaleOperator",
				"plural":"csiscaleoperators",
				"version":"v1"
			 }
		  ],
		  "provider":"",
		  "related_images":[
			 
		  ],
		  "replaces":null,
		  "skip_range":null,
		  "skips":[
			 
		  ],
		  "source_index_container_path":"registry.redhat.io/redhat/certified-operator-index:v4.6",
		  "version":"2.0.0",
		  "version_original":"2.0.0"
	   }
	]
 }`

func loadOperatorsDB() error {
	var fullCatalog OperatorCatalog
	bytes, err := io.ReadAll(strings.NewReader(operatorDBJSON))
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &fullCatalog)
	if err != nil {
		return err
	}
	for i := 0; i < len(fullCatalog.Data); i++ {
		if opName, opV, ocpV, channel, err := buildOperatorKey(&fullCatalog.Data[i]); err == nil {
			operatordb[opName] = append(operatordb[opName], OperatorOcpVersionMatch{ocpVersion: ocpV, operatorVersion: opV, channel: channel})
		}
	}

	return nil
}

func TestIsOperatorCertified(t *testing.T) {
	validator := OfflineValidator{}

	assert.NoError(t, loadOperatorsDB())

	name := "ibm-spectrum-scale-csi-operator.v2.0.0"
	ocpversion := "4.6"
	channel := "stable"

	assert.True(t, validator.IsOperatorCertified(name, ocpversion, channel))

	name = "falcon-alpha"
	assert.False(t, validator.IsOperatorCertified(name, ocpversion, channel))
}
