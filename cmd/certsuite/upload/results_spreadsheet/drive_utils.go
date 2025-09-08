package resultsspreadsheet

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// createDriveFolder creates a new folder in Google Drive
//
// The function builds a folder metadata object with the specified name, parent
// ID, and MIME type for folders. It first checks if a folder with that name
// already exists under the given parent by querying the Drive API; if found it
// returns an error to avoid duplication. If no existing folder is detected, it
// calls the API to create the folder and returns the resulting file object or
// any creation errors.
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

// MoveSpreadSheetToFolder Moves a spreadsheet into a specified Google Drive folder
//
// This function retrieves the current parent folders of the given spreadsheet
// using the Drive service, then updates the file to add the target folder as a
// new parent while removing any existing parents. It performs these operations
// via the Drive API's Update call and logs fatal errors if any step fails. On
// success it returns nil, indicating the spreadsheet has been relocated.
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

// extractFolderIDFromURL extracts the final path segment from a URL
//
// This routine parses an input string as a URL, splits its path into
// components, and returns the last component which represents a folder
// identifier. If parsing fails it propagates the error; otherwise it provides
// the ID and no error.
func extractFolderIDFromURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	pathSegments := strings.Split(parsedURL.Path, "/")

	// The folder ID is the last segment in the path
	return pathSegments[len(pathSegments)-1], nil
}
