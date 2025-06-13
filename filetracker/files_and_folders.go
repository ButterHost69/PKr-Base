package filetracker

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

// Delete files and folders in the Workspace Except: /.PKr , PKr-base.exe, PKr-cli.exe, /tmp
func CleanFilesFromWorkspace(workspace_path string) error {
	files, err := ioutil.ReadDir(workspace_path)
	if err != nil {
		return err
	}

	log.Printf("Deleting All Files at: %s\n\n", workspace_path)
	for _, file := range files {
		if file.Name() != ".PKr" && file.Name() != "PKr-Base.exe" && file.Name() != "PKr-Cli.exe" && file.Name() != "tmp" {
			if err = os.RemoveAll(path.Join([]string{workspace_path, file.Name()}...)); err != nil {
				return err
			}
		}
	}

	return nil
}

// Create New File of name `dest`.
// Save Data to the File
func SaveDataToFile(data []byte, dest string) error {
	zippedfile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer zippedfile.Close()

	zippedfile.Write(data)

	return nil
}
