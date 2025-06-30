package filetracker

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/config"
)

func addFilesToZip(writer *zip.Writer, dirpath string, relativepath string) error {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		fmt.Println("Error while Reading Dir:", err)
		fmt.Println("Source: addFilesToZip()")
		return err
	}

	for _, file := range files {
		if file.Name() == ".PKr" || file.Name() == "PKr-Base.exe" || file.Name() == "PKr-Cli.exe" || file.Name() == "tmp" || file.Name() == "PKr-Base" || file.Name() == "PKr-Cli" {
			continue
		} else if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(dirpath, file.Name()))

			if err != nil {
				fmt.Println("Error while Reading File:", err)
				fmt.Println("Source: addFilesToZip()")
				return err
			}

			file, err := writer.Create(filepath.Join(relativepath, file.Name()))
			if err != nil {
				fmt.Println("Error while Creating Entry in Zip File:", err)
				fmt.Println("Source: addFilesToZip()")
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

func ZipData(workspace_path string, destination_path string, zip_file_name string) error {
	zipFileName := zip_file_name + ".zip"
	fullZipPath := filepath.Join(destination_path, zipFileName)

	// Ensure the destination directory exists
	err := os.MkdirAll(destination_path, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating destination directory:", err)
		fmt.Println("Source: ZipData()")
		return err
	}

	zip_file, err := os.Create(fullZipPath)
	if err != nil {
		fmt.Println("Error while Creating Zip File:", err)
		fmt.Println("Source: ZipData()")
		return err
	}

	writer := zip.NewWriter(zip_file)
	addFilesToZip(writer, workspace_path, "")

	if err = writer.Close(); err != nil {
		fmt.Println("Error while Closing zip writer:", err)
		fmt.Println("Source: ZipData()")
		return err
	}
	zip_file.Close()
	return nil
}

func UnzipData(src, dest string) error {
	fmt.Printf("Unzipping Files: %s\n\t to %s\n", src, dest)
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
			if err := os.MkdirAll(dir, 0700); err != nil {
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
		fmt.Printf("%d] File: %s\n", count, file.Name)
	}
	fmt.Printf("\nTotal Files Recieved: %d\n", totalfiles)
	return nil
}

func returnZipFileObj(zip_file_reader *zip.ReadCloser, search_file_name string) *zip.File {
	for _, file := range zip_file_reader.File {
		if file.Name == search_file_name {
			return file
		}
	}
	return nil
}

func ZipUpdates(changes []config.FileChange, src_path string, dst_path string) (err error) {
	dst_dir, _ := filepath.Split(dst_path)
	if err = os.Mkdir(dst_dir, 0700); err != nil {
		fmt.Println("Error Could not Create the Dir:", err)
		fmt.Println("Source: ZipUpdates()")
		return err
	}

	// Open Src Zip File
	src_zip_file, err := zip.OpenReader(src_path)
	if err != nil {
		fmt.Println("Error while Opening Source Zip File:", err)
		fmt.Println("Source: ZipUpdates()")
		return err
	}
	defer src_zip_file.Close()

	// Create Dest Zip File
	dst_zip_file, err := os.Create(dst_path)
	if err != nil {
		fmt.Printf("Error Could Not Create File %v: %v\n", dst_path, err)
		fmt.Println("Source: ZipUpdates()")
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

		zip_file_obj := returnZipFileObj(src_zip_file, change.FilePath)
		if zip_file_obj == nil {
			fmt.Println(filepath.Join(src_path, change.FilePath), "is nil")
			return
		}

		zip_file_obj_reader, err := zip_file_obj.Open()
		if err != nil {
			return err
		}
		defer zip_file_obj_reader.Close()

		new_file, err := writer.Create(zip_file_obj.Name)
		if err != nil {
			return err
		}

		// Copy the contents
		_, err = io.Copy(new_file, zip_file_obj_reader)
		if err != nil {
			return err
		}

	}
	return nil
}
