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

// I dont Know if this works. Check it later
// I copied From : https://gosamples.dev/zip-file/
// CHEECHK THIISS LAAATTERRR
// Running This Function Twice Makes a Zip File Whose Size keeps increasing until the Entire Disk
// is filled
// Dont USE THISSSSS
func zippToInfiniteSize(workspace_path string) (string, error) {
	// 2024-01-20.zip
	zipFileName := strings.Split(time.Now().String(), " ")[0] + ".zip"

	file, err := os.Create(workspace_path + "\\.PKr\\" + zipFileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	writer := zip.NewWriter(file)

	// This Might Break in Linux...
	return zipFileName, filepath.Walk(workspace_path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			header.Method = zip.Deflate

			relPath, err := filepath.Rel(filepath.Dir(workspace_path), path)
			if err != nil {
				return err
			}
			header.Name = relPath

			if info.IsDir() {
				header.Name += "/"
			}

			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}

			defer f.Close()

			_, err = io.Copy(headerWriter, f)
			return err
		})
}

func addFilesToZip(writer *zip.Writer, dirpath string, relativepath string) error {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		// log.Println(err)
		return err
	}

	for _, file := range files {
		// Comment This Later ... Only For Debugging
		// config.AddUsersLogEntry(log.Sprintf("File: %s", file.Name()))
		// ..........
		if file.Name() == ".PKr" || file.Name() == "PKr-Base.exe" || file.Name() == "PKr-Cli.exe" || file.Name() == "tmp" || file.Name() == "PKr-Base" || file.Name() == "PKr-Cli" {
			continue
		} else if !file.IsDir() {
			content, err := os.ReadFile(filepath.Join(dirpath, file.Name()))

			if err != nil {
				// log.Println(err)
				return err
			}

			file, err := writer.Create(filepath.Join(relativepath, file.Name()))
			if err != nil {
				// log.Println(err)
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

	addFilesToZip(writer, workspace_path+string(filepath.Separator), "")

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
		} else {
			dir, _ := filepath.Split(file.Name)
			if dir != "" {
				if err := os.MkdirAll(dir, 0777); err != nil {
					return err
				}
			}
			unzipfile, err := os.Create(file.Name)
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
			log.Printf("%d] File: %s\n", count, unzipfile.Name())
		}
	}
	log.Printf("\nTotal Files Recieved: %d\n", totalfiles)
	return nil
}

func ZipUpdates(updates config.Updates, workspace_path string, hash string) (err error) {
	files := []string{}
	for _, changes := range updates.Changes {
		if changes.Type == "Updated" {
			files = append(files, changes.FilePath)
		}
	}

	// Create current change cache folder -> in Changes
	storeFolderPath := filepath.Join(workspace_path, ".PKr", "Files", "Changes", hash)
	if err = os.Mkdir(storeFolderPath, 0777); err != nil {
		log.Println("Could not Create the Dir: ")
		log.Println("Error: ", err)
		log.Println("Source: ZipHashUpdates()")
		return err
	}

	// Create Zip File
	zipFilePath := filepath.Join(storeFolderPath, hash+".zip")
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		log.Printf("Error Could Not Create File %v: %v\n", zipFilePath, err)
		log.Println("Source: ZipHashUpdates()")
		return err
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	for _, relPath := range files {
		filePath := filepath.Join(workspace_path, relPath)

		// Open source file
		file, err := os.Open(filePath)
		if err != nil {
			log.Println("Could Not Open source file")
			log.Println("Source: ZipHashUpdates()")
			return err
		}
		defer file.Close()

		info, err := file.Stat()
		if err != nil {
			log.Println("Could Not see File Stats")
			log.Println("Source: ZipHashUpdates()")
			return err
		}

		// Create zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Println("Could Not Create zip header")
			log.Println("Source: ZipHashUpdates()")
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		// Create writer for file
		writerPart, err := writer.CreateHeader(header)
		if err != nil {
			log.Println("Could Not Create writer for file")
			log.Println("Source: ZipHashUpdates()")

			return err
		}

		// Copy file contents to zip
		_, err = io.Copy(writerPart, file)
		if err != nil {
			log.Println("Could Not Create writer for file")
			log.Println("Source: ZipHashUpdates()")
			return err
		}
	}

	return nil
}
