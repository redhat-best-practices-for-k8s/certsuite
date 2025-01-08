package catalogsource

import (
	"context"
	"fmt"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	olmpkgv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/containerd/containerd"
)

func SkipPMBasedOnChannel(channels []olmpkgv1.PackageChannel, csvName string) bool {
	// This logic is in place because it is possible for an operator to pull from a multiple package manifests.
	skipPMBasedOnChannel := true
	for c := range channels {
		log.Debug("Comparing channel currentCSV %q with current CSV %q", channels[c].CurrentCSV, csvName)
		log.Debug("Number of channel entries %d", len(channels[c].Entries))
		for _, entry := range channels[c].Entries {
			log.Debug("Comparing entry name %q with current CSV %q", entry.Name, csvName)

			if entry.Name == csvName {
				log.Debug("Skipping package manifest based on channel entry %q", entry.Name)
				skipPMBasedOnChannel = false
				break
			}
		}

		if !skipPMBasedOnChannel {
			break
		}
	}

	return skipPMBasedOnChannel
}

// TODO: Get all catalog sources of application operators
func getOperatorCatalogSources() {

}

// In Porgress
func getOperatorCatalogAllImageSize() {

	const sizeLimit = 200 * 1024 * 1024 * 1024 // 200GB in bytes

	// Call getOperatorCatalogSources()

	catalogSize, err := getCatalogSize("a", "b")
	if err != nil {
		log.Fatal("Error calculating catalog image size: %v", err)
	}

	operatorImages, err := getOperatorImages(nil, "operatorNamespace")
	if err != nil {
		log.Fatal("Failed to retrieve operator images: %v", err)
	}

	operatorImagesSize, err := calculateImagesSize(operatorImages)
	if err != nil {
		log.Fatal("Error calculating operator image sizes: %v", err)
	}

	// Step 6: Check if total size exceeds the limit
	totalSize := catalogSize + operatorImagesSize
	fmt.Printf("Catalog Image Size: %.2f GB\n", float64(catalogSize)/1024/1024/1024)
	fmt.Printf("Operator Images Size: %.2f GB\n", float64(operatorImagesSize)/1024/1024/1024)
	fmt.Printf("Total Size: %.2f GB\n", float64(totalSize)/1024/1024/1024)

	if totalSize > sizeLimit {
		log.Fatal("Total size exceeds the 200GB limit!")
	} else {
		fmt.Println("Total size is within the 200GB limit.")
	}
}

// getOperatorImages retrieves images associated with the operator from its ClusterServiceVersion (CSV)
func getOperatorImages(olmClient client.Client, namespace string) ([]string, error) {
	// Retrieve the ClusterServiceVersion (CSV) for the operator
	csvList := &v1alpha1.ClusterServiceVersionList{}
	err := olmClient.List(context.TODO(), csvList, client.InNamespace(namespace))
	if err != nil {
		return nil, fmt.Errorf("failed to list ClusterServiceVersions: %w", err)
	}

	var images []string
	for i := range csvList.Items {
		csv := &csvList.Items[i]
		for j := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			dep := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[j]
			for k := range dep.Spec.Template.Spec.Containers {
				container := &dep.Spec.Template.Spec.Containers[k]
				images = append(images, container.Image)
			}
		}
	}
	return images, nil
}

// getCatalogSize calculates the size of the catalog image
func getCatalogSize(catalogNamespace, catalogSourceName string) (int64, error) {
	oc := clientsholder.GetClientsHolder()
	catalogSource, err := oc.OlmClient.OperatorsV1alpha1().CatalogSources(catalogNamespace).Get(context.TODO(), catalogSourceName, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve catalog source: %w", err)
	}

	catalogSourceSpec := catalogSource.Spec
	image := catalogSourceSpec.Image

	return getImageSize(image)
}

// calculateImagesSize calculates the total size of a list of images
func calculateImagesSize(images []string) (int64, error) {
	var totalSize int64
	for _, image := range images {
		size, err := getImageSize(image)
		if err != nil {
			return 0, fmt.Errorf("error calculating size for image %s: %w", image, err)
		}
		totalSize += size
	}
	return totalSize, nil
}

// getImageSize calculates the size of a single image
func getImageSize(imageName string) (int64, error) {
	// Initialize containerd client to check image sizes
	containerdClient, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		log.Fatal("Failed to connect to containerd: %v", err)
	}
	defer containerdClient.Close()
	image, err := containerdClient.GetImage(context.Background(), imageName)
	if err != nil {
		return 0, fmt.Errorf("failed to get image: %w", err)
	}

	// Use the Size method to directly get the image size
	size, err := image.Size(context.Background())
	if err != nil {
		return 0, fmt.Errorf("failed to get image size: %w", err)
	}

	return size, nil
}
