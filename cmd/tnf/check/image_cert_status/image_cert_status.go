// Copyright (C) 2020-2023 Red Hat, Inc.
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

package imagecert

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/test-network-function/cnf-certification-test/internal/certdb"
)

// generateCmd is the root of the "catalog generate" CLI program.
var checkImageCertStatusCmd = &cobra.Command{
	Use:   "image-cert-status",
	Short: "Verifies the container's image certification status",
	RunE:  checkImageCertStatus,
}

func checkImageCertStatus(cmd *cobra.Command, args []string) error {
	imageName, _ := cmd.Flags().GetString("name")
	imageRegistry, _ := cmd.Flags().GetString("registry")
	imageTag, _ := cmd.Flags().GetString("tag")
	offlineDb, _ := cmd.Flags().GetString("offline-db")

	validator, err := certdb.GetValidator(offlineDb)
	if err != nil {
		return fmt.Errorf("could not get a validator for container images, error: %v", err)
	}

	fmt.Printf("Image name: %s\nImage registry: %s\nImage tag: %s\n", imageName, imageRegistry, imageTag)

	if validator.IsContainerCertified(imageRegistry, imageName, imageTag, "") {
		fmt.Println("Image is certified")
	} else {
		fmt.Println("Image is not certified")
	}

	return nil
}

// Execute executes the "catalog" CLI.
func NewCommand() *cobra.Command {
	checkImageCertStatusCmd.PersistentFlags().String("name", "", "name of the image to verify")

	err := checkImageCertStatusCmd.MarkPersistentFlagRequired("name")
	if err != nil {
		log.Fatalf("failed to mark name flag as required:  :%v", err)
		return nil
	}
	checkImageCertStatusCmd.PersistentFlags().String("registry", "", "registry where the image is stored")

	err = checkImageCertStatusCmd.MarkPersistentFlagRequired("registry")
	if err != nil {
		log.Fatalf("failed to mark registry flag as required:  :%v", err)
		return nil
	}

	checkImageCertStatusCmd.PersistentFlags().String("tag", "latest", "image tag to be fetched")

	checkImageCertStatusCmd.PersistentFlags().String("offline-db", "", "path to the offline db (for disconnected evironments)")

	return checkImageCertStatusCmd
}
