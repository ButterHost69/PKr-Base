package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/encrypt"
)

var TREE_REL_PATH = filepath.Join(".PKr", "file_tree.json")

type FileTree struct {
	Nodes []Node
}

type Node struct {
	FilePath string `json:"file_path"`
	Hash     string `json:"hash"`
}

func CreateFileTreeIfNotExits(workspace_file_path string) error {
	tree_file_path := filepath.Join(workspace_file_path, TREE_REL_PATH)
	if _, err := os.Stat(tree_file_path); os.IsExist(err) {
		fmt.Println("~ tree_file already Exists")
		return err
	}

	fileTree, err := GetNewTree(workspace_file_path)
	if err != nil {
		log.Println("Error while Getting New Tree: ", err)
		log.Println("Source: CreateFileTreeIfNotExits()")
		return err
	}

	err = WriteToFileTree(workspace_file_path, fileTree)
	if err != nil {
		log.Println("Error while Writing File Tree: ", err)
		log.Println("Source: CreateFileTreeIfNotExits()")
		return err
	}

	return nil
}

func GetNewTree(workspace_file_path string) (FileTree, error) {
	var tree FileTree

	err := filepath.Walk(workspace_file_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // skip files we can't read
		}

		if info.IsDir() && (info.Name() == ".PKr" || info.Name() == "tmp") {
			return filepath.SkipDir
		} else if !info.IsDir() {
			if info.Name() == "PKr-Base.exe" || info.Name() == "PKr-Cli.exe" {
				return nil
			}
			relPath, err := filepath.Rel(workspace_file_path, path)
			if err != nil {
				log.Println("Could Not Generate RelPath")
				log.Printf("Path: %s | Error: %v\n", workspace_file_path, err)
				log.Println("Source: GetNewTree()")
				return err
			}

			hash, err := encrypt.GenerateHashWithFilePath(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to hash %s: %v\n", path, err)
				return nil
			}

			tree.Nodes = append(tree.Nodes, Node{
				FilePath: relPath,
				Hash:     hash,
			})
		}
		return nil
	})

	if err != nil {
		log.Println("Error walking the path:", err)
		return tree, nil
	}

	log.Println(tree)
	return tree, nil
}

func ReadFromTreeFile(workspace_tree_path string) (FileTree, error) {
	file, err := os.Open(filepath.Join(workspace_tree_path, TREE_REL_PATH))
	if err != nil {
		// AddUsersLogEntry("error in opening PKR config file.... pls check if .PKr/workspaceConfig.json available ")
		log.Println("error in opening PKR config file.... pls check if .PKr/workspaceConfig.json available ")
		return FileTree{}, err
	}
	defer file.Close()

	var fileTree FileTree
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fileTree)
	if err != nil {
		log.Println("error in decoding json data")
		// AddUsersLogEntry("error in decoding json data")
		return FileTree{}, err
	}

	// fmt.Println(pkrConfig)
	return fileTree, nil
}

func WriteToFileTree(workspace_tree_path string, FileTree FileTree) error {
	jsonData, err := json.MarshalIndent(FileTree, "", "	")
	// fmt.Println(jsonData)
	if err != nil {
		log.Println("error occured in Marshalling the data to JSON")
		log.Println(err)
		return err
	}

	// fmt.Println(string(jsonData))
	err = os.WriteFile(filepath.Join(workspace_tree_path, TREE_REL_PATH), jsonData, 0777)
	if err != nil {
		log.Println("error occured in storing data in userconfig file")
		log.Println(err)
		return err
	}

	return nil
}

func CompareTrees(oldTree, newTree FileTree, new_hash string) Updates {
	// Build lookup maps
	oldMap := make(map[string]string)
	newMap := make(map[string]string)

	for _, node := range oldTree.Nodes {
		oldMap[node.FilePath] = node.Hash
	}

	for _, node := range newTree.Nodes {
		newMap[node.FilePath] = node.Hash
	}

	var changes []FileChange

	// Detect created or updated
	for path, newHash := range newMap {
		oldHash, exists := oldMap[path]
		if !exists {
			// New file
			changes = append(changes, FileChange{
				FilePath: path,
				FileHash: newHash,
				Type:     "Updated",
			})
		} else if newHash != oldHash {
			// Updated file
			changes = append(changes, FileChange{
				FilePath: path,
				FileHash: newHash,
				Type:     "Updated",
			})
		}
	}

	// Detect removed
	for path, oldHash := range oldMap {
		if _, exists := newMap[path]; !exists {
			changes = append(changes, FileChange{
				FilePath: path,
				FileHash: oldHash,
				Type:     "Removed",
			})
		}
	}

	return Updates{
		Hash:    new_hash,
		Changes: changes,
	}
}
