package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ButterHost69/PKr-Base/encrypt"
)

var TREE_REL_PATH = filepath.Join(".PKr", "file-tree.json")

type FileTree struct {
	Nodes []Node
}

type Node struct {
	FilePath string `json:"file_path"`
	Hash     string `json:"hash"`
}

func CreateFileTreeIfNotExits(workspace_path string) error {
	tree_file_path := filepath.Join(workspace_path, TREE_REL_PATH)
	if _, err := os.Stat(tree_file_path); os.IsExist(err) {
		fmt.Println("File Tree Already Exists")
		return nil
	}

	fileTree, err := GetNewTree(workspace_path)
	if err != nil {
		fmt.Println("Error while Getting New Tree: ", err)
		fmt.Println("Source: CreateFileTreeIfNotExits()")
		return err
	}

	err = WriteToFileTree(workspace_path, fileTree)
	if err != nil {
		fmt.Println("Error while Writing in File Tree: ", err)
		fmt.Println("Source: CreateFileTreeIfNotExits()")
		return err
	}
	return nil
}

func GetNewTree(workspace_path string) (FileTree, error) {
	var tree FileTree

	err := filepath.Walk(workspace_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // skip files we can't read
		}

		if info.IsDir() && (info.Name() == ".PKr" || info.Name() == "tmp") {
			return filepath.SkipDir
		} else if !info.IsDir() {
			if info.Name() == "PKr-Base.exe" || info.Name() == "PKr-Cli.exe" {
				return nil
			}
			relPath, err := filepath.Rel(workspace_path, path)
			if err != nil {
				fmt.Println("Error while Getting Relative Path:", err)
				fmt.Println("Source: GetNewTree()")
				return err
			}

			hash, err := encrypt.GenerateHashWithFilePath(path)
			if err != nil {
				fmt.Println("Error while Hashing with File Path:", err)
				fmt.Println("Source: GetNewTree()")
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
		fmt.Println("Error walking the path:", err)
		fmt.Println("Source: GetNewTree()")
		return tree, nil
	}
	return tree, nil
}

func ReadFromTreeFile(workspace_tree_path string) (FileTree, error) {
	file, err := os.Open(filepath.Join(workspace_tree_path, TREE_REL_PATH))
	if err != nil {
		fmt.Println("Error while opening tree file:", err)
		fmt.Println("Source: ReadFromTreeFile()")
		return FileTree{}, err
	}
	defer file.Close()

	var fileTree FileTree
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fileTree)
	if err != nil {
		fmt.Println("Error while Decoding JSON Data from tree file:", err)
		fmt.Println("Source: ReadFromTreeFile()")
		return FileTree{}, err
	}
	return fileTree, nil
}

func WriteToFileTree(workspace_tree_path string, FileTree FileTree) error {
	jsonData, err := json.MarshalIndent(FileTree, "", "	")
	if err != nil {
		fmt.Println("Error while Marshalling the file-tree to JSON:", err)
		fmt.Println("Source: WriteToFileTree()")
		return err
	}

	err = os.WriteFile(filepath.Join(workspace_tree_path, TREE_REL_PATH), jsonData, 0600)
	if err != nil {
		fmt.Println("Error while writing data in file-tree:", err)
		fmt.Println("Source: WriteToFileTree()")
		return err
	}
	return nil
}

func CompareTrees(oldTree, newTree FileTree) []FileChange {
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
	return changes
}
