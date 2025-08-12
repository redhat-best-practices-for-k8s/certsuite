package resultsspreadsheet

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

func createDriveFolder(srv *drive.Service, folderName, parentFolderID string) (*drive.File, error) {
	driveFolder := &drive.File{
		Name:     folderName,
		Parents:  []string{parentFolderID},
		MimeType: "application/vnd.google-apps.folder",
	}

	// Search for an existing folder with the same name
	q := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false", folderName, parentFolderID)
	call := srv.Files.List().Q(q).Fields("files(id, name)")

	files, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to list files: %v", err)
	}

	if len(files.Files) > 0 {
		return nil, fmt.Errorf("folder %s already exists in %s folder ID", folderName, parentFolderID)
	}

	createdFolder, err := srv.Files.Create(driveFolder).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create folder: %v", err)
	}

	return createdFolder, nil
}

func MoveSpreadSheetToFolder(srv *drive.Service, folder *drive.File, spreadsheet *sheets.Spreadsheet) error {
	file, err := srv.Files.Get(spreadsheet.SpreadsheetId).Fields("parents").Do()
	if err != nil {
		log.Fatalf("Unable to get file: %v", err)
	}

	// Collect the current parent IDs to remove (if needed)
	oldParents := append([]string{}, file.Parents...)

	updateCall := srv.Files.Update(spreadsheet.SpreadsheetId, nil)
	updateCall.AddParents(folder.Id)

	// Remove the file from its old parents
	if len(oldParents) > 0 {
		for _, parent := range oldParents {
			updateCall.RemoveParents(parent)
		}
	}

	_, err = updateCall.Do()
	if err != nil {
		log.Fatalf("Unable change file location: %v", err)
	}

	return nil
}

func extractFolderIDFromURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	pathSegments := strings.Split(parsedURL.Path, "/")

	// The folder ID is the last segment in the path
	return pathSegments[len(pathSegments)-1], nil
}
