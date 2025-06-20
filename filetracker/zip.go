package filetracker

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
)

func addFilesToZip(writer *zip.Writer, dirpath string, relativepath string) error {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() == ".PKr" || file.Name() == "PKr-Base.exe" || file.Name() == "PKr-Cli.exe" || file.Name() == "tmp" || file.Name() == "PKr-Base" || file.Name() == "PKr-Cli" {
			continue
		} else if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(dirpath, file.Name()))

			if err != nil {
				return err
			}

			file, err := writer.Create(filepath.Join(relativepath, file.Name()))
			if err != nil {
				return err
			}
			file.Write(content)
		} else if file.IsDir() {
			newDirPath := filepath.Join(dirpath, file.Name()) + string(os.PathSeparator)
			newRelativePath := filepath.Join(relativepath, file.Name()) + string(os.PathSeparator)

			addFilesToZip(writer, newDirPath, newRelativePath)
		}
	}

	return nil
}

func ZipData(workspace_path string, destination_path string) (string, error) {
	zipFileName := strings.Split(time.Now().String(), " ")[0] + ".zip"
	fullZipPath := filepath.Join(destination_path, zipFileName)

	zip_file, err := os.Create(fullZipPath)
	if err != nil {
		// config.AddLogEntry(workspace_name, err)
		return "", err
	}

	writer := zip.NewWriter(zip_file)

	addFilesToZip(writer, workspace_path, "")

	if err = writer.Close(); err != nil {
		return "", err
	}

	hashFileName, err := encrypt.GenerateHashWithFileIO(zip_file)
	// hashFileName, err := encrypt.GenerateHashWithFilePath(fullZipPath)
	if err != nil {
		return "", err
	}

	zip_file.Close()

	hashFileName = hashFileName + ".zip"
	fullHashFilePath := destination_path + filepath.Join(destination_path, hashFileName)

	workspace_split := strings.Split(workspace_path, string(filepath.Separator))
	workspace_name := workspace_split[len(workspace_split)-1]

	if err = os.Rename(fullZipPath, fullHashFilePath); err != nil {
		logdata := fmt.Sprintf("could rename zip file to new hash name: %s | zipped file path: %s.\nError: %v", fullHashFilePath, fullZipPath, err)
		config.AddLogEntry(workspace_name, true, logdata)
		return "", err
	}

	return hashFileName, nil
}

func UnzipData(src, dest string) error {
	log.Printf("Unzipping Files: %s\n\t to %s\n", src, dest)
	zipper, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer zipper.Close()
	totalfiles := 0
	for count, file := range zipper.File {
		if file.FileInfo().IsDir() {
			continue
		}
		abs_path := filepath.Join(dest, file.Name)
		dir, _ := filepath.Split(abs_path)
		if dir != "" {
			if err := os.MkdirAll(dir, 0777); err != nil {
				return err
			}
		}
		unzipfile, err := os.Create(abs_path)
		if err != nil {
			return err
		}
		defer unzipfile.Close()

		content, err := file.Open()
		if err != nil {
			return err
		}
		defer content.Close()

		_, err = io.Copy(unzipfile, content)
		if err != nil {
			return err
		}
		totalfiles += 1
		log.Printf("%d] File: %s\n", count, file.Name)
	}
	log.Printf("\nTotal Files Recieved: %d\n", totalfiles)
	return nil
}

func returnZipFileObj(zip_file_reader *zip.ReadCloser, search_file_name string) *zip.File {
	for _, file := range zip_file_reader.File {
		log.Println("Zip File:", file.Name)
		if file.Name == search_file_name {
			return file
		}
	}
	return nil
}

func ZipUpdates(changes []config.FileChange, src_path string, dst_path string) (err error) {
	dst_dir, _ := filepath.Split(dst_path)
	if err = os.Mkdir(dst_dir, 0600); err != nil {
		log.Println("Could not Create the Dir: ")
		log.Println("Error: ", err)
		log.Println("Source: ZipUpdates()")
		return err
	}

	log.Println("Zipping Updates ...")
	// Open Src Zip File
	src_zip_file, err := zip.OpenReader(src_path)
	if err != nil {
		log.Println("Error while Opening Source Zip File:", err)
		log.Println("Source: ZipUpdates()")
		return err
	}
	defer src_zip_file.Close()

	// Create Dest Zip File
	dst_zip_file, err := os.Create(dst_path)
	if err != nil {
		log.Printf("Error Could Not Create File %v: %v\n", dst_path, err)
		log.Println("Source: ZipUpdates()")
		return err
	}
	defer dst_zip_file.Close()

	// Dest Zip Writer
	writer := zip.NewWriter(dst_zip_file)
	defer writer.Close()

	for _, change := range changes {
		if change.Type != "Updated" {
			continue
		}
		log.Println("Change.FilePath:", change.FilePath)

		zip_file_obj := returnZipFileObj(src_zip_file, filepath.Join(src_path, change.FilePath))
		err = writer.Copy(zip_file_obj)
		if err != nil {
			log.Println("Error while Copying Zip.File obj into New Dest Zip Writer:", err)
			log.Println("Source: ZipUpdates()")
			return err
		}
	}
	return nil
}
