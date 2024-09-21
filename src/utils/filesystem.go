package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func GetAbsPath(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		log.Panicf("Failed to read the file: %v", err)
	}

	return absolutePath
}

func ReadFile(filePath string) string {
	// Read the file content as a byte slice
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Panic if there's an error
		log.Panicf("Failed to read the file: %v", err)
	}

	// Return the file content as a string
	return string(content)
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	// os.IsNotExist checks if the error indicates that the file or directory does not exist
	if os.IsNotExist(err) {
		return false
	}
	// If err is nil, the file/folder exists. Other errors should be handled as needed.
	return err == nil
}

func RemoveFolderRecursively(folderPath string) error {
	err := os.RemoveAll(folderPath)
	if err != nil {
		return fmt.Errorf("error removing directory: %w", err)
	}
	return nil
}

func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("error moving folder: %w", err)
	}
	return nil
}


func GetHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("Error finding home directory:", err)
	}

	return homeDir
}

func RemoveFile(path string) error {
	// Check if the file or directory exists
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	} else if err != nil {
		return fmt.Errorf("error checking file: %w", err)
	}

	// Remove the file or directory
	err = os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

func Filename(path string) string {
    return filepath.Base(path)
}


func FilenameWithoutExt(path string) string {
    filename := filepath.Base(path)
    ext := filepath.Ext(filename)
    return strings.TrimSuffix(filename, ext)
}

type FileInfo struct {
	Path    string
	ModTime time.Time
}

func ListFiles(dirPath string) []FileInfo {
	files := []FileInfo{}

	dir, filePathErr := filepath.Abs(dirPath)
	if (filePathErr != nil) {
		return files
	}

	// Walk through the directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		if !info.IsDir() {
			files = append(files, FileInfo{
				Path:    path,
				ModTime: info.ModTime(),
			})
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
		return []FileInfo{}
	}

	// Sort files by modification date
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	return files
}

func CreateFolder(dirPath string) error {
    err := os.MkdirAll(dirPath, 0755)
    if err != nil {
        return err
    }
    return nil
}


func CreateFile(filePath string, content string) {
	parts := strings.Split(filePath, "/")
	dir :=  strings.Join( parts[0:len(parts) - 1], "/") 

	
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return
	}

	// /Create a new file (or overwrite if it already exists)
	fmt.Println("Created file: " + filePath)
	file, err := os.Create(filePath)
	if err != nil {
		log.Panicf("Error creating file: %v", err)
		return
	}
	// Ensure the file is closed after we're done writing to it
	defer file.Close()

	// Write some data to the file
	_, err = file.WriteString(content)
	if err != nil {
		log.Panicf("Error writing to file: %v", err)
		return
	}

	fmt.Println("File created and data written successfully!")
}

func CreateSymlink(src string, dest string) error {
	// Get the directory part of the destination path
	destDir := filepath.Dir(dest)

	// Create directories leading up to the destination if they don't exist
	err := os.MkdirAll(destDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	if(FileExists(dest)) {
		fmt.Println("Deleting existing file at: ",dest)
		RemoveFile(dest)
	}

	fmt.Println("Creating symlink From: ", src, " To:", dest)
	fmt.Println(destDir)


	// Create the symbolic link
	err = os.Symlink(GetAbsPath(src), GetAbsPath(dest))
	if err != nil {
		return fmt.Errorf("failed to create symbolic link: %w", err)
	}

	return nil
}

// Read from a file and deserialize its contents to a generic type (for JSON)
func ReadJSON[T any](filePath string, defaultValue T) T {
	// content := ReadFile(filePath)

	content, readErr := os.ReadFile(filePath)
	if readErr != nil {
		// Panic if there's an error
		// log.Panicf("Failed to read the file: %v", err)
		return defaultValue
	}


	var v T
	unmarshallErr := json.Unmarshal([]byte(content), &v)
	if unmarshallErr != nil {
		// log.Panicf("Error deserializing JSON: %v", err)
		return defaultValue
	}

	return v
}

func WriteJSON[T any](filePath string, data T) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Panicf("Error serializing to JSON: %v", err)
	}

	CreateFile(filePath, string(jsonData))
}

func WriteJSONPretty[T any](filePath string, data T) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Panicf("Error serializing to JSON: %v", err)
	}

	CreateFile(filePath, string(jsonData))
}

