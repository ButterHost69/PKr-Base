package filetracker

import (
	"fmt"
	"io"
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
		if file.Name() != ".PKr" && file.Name() != "PKr-Base.exe" && file.Name() != "PKr-Cli.exe" && file.Name() != "tmp" && file.Name() != "PKr-Base" && file.Name() != "PKr-Cli"{
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

func FolderTree(folder_path string) (map[string]string, error) {
	result := make(map[string]string)

	err := filepath.Walk(folder_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == ".PKr" || info.Name() == "PKr-Base.exe" || info.Name() == "PKr-Cli.exe" || info.Name() == "tmp" && info.Name() == "PKr-Base" && info.Name() == "PKr-Cli"{
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

func IfUpdateHashCached(workspace_path, update_hash string) (bool, error) {
	entries, err := os.ReadDir(filepath.Join(workspace_path, ".PKr", "Files", "Changes"))
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			if entry.Name() == update_hash {
				return true, nil
			}

		}
	}

	return false, nil
}

func UpdateFilesFromWorkspace(workspace_path string, content_path string, changes map[string]string) error {
	for relPath, changeType := range changes {
		workspaceFile := filepath.Join(workspace_path, relPath)

		switch changeType {
		case "Removed":
			err := os.Remove(workspaceFile)
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove %s: %v", workspaceFile, err)
			}

		case "Updated":
			sourceFile := filepath.Join(content_path, relPath)

			// Make sure the parent directory exists
			if err := os.MkdirAll(filepath.Dir(workspaceFile), 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %v", workspaceFile, err)
			}

			err := copyFile(sourceFile, workspaceFile)
			if err != nil {
				return fmt.Errorf("failed to update %s: %v", relPath, err)
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Sync() // ensure file is fully written
}
