package filetracker

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/encrypt"
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

func FolderTree(folder_path string) (map[string]string, error){
		result := make(map[string]string)

	err := filepath.Walk(folder_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".PKr" || info.Name() == "PKr-Base.exe" || info.Name() == "PKr-Cli.exe" || info.Name() == "tmp" {
			return filepath.SkipDir
		}
			if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		hash, err := encrypt.GenerateHashWithFileIO(f)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(folder_path, path)
		if err != nil {
			return err
		}

		result[relPath] = hash
		return nil
	})

	return result, err
}
