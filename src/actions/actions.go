package actions

import (
	"fmt"
	"path/filepath"
	"stash/src/constants"
	"stash/src/parser"
	"stash/src/types"
	filesystem "stash/src/utils"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/rodaine/table"
)

var version = "0.0.0"

func PrintVersion() {
	fmt.Println("Version: " + version)
}

func GetDefaultManifest() types.Manifest {
	return types.Manifest{
		Id: "",
		Message: "No message",
		Files: []types.Config{},
		Packages: []types.Package{},
	}

}

func GetRestoreList()  []string {
	files := filesystem.ListFiles(constants.RESTORE_PATH)
	var filePaths []string
	
	for i := 0; i < len(files); i++ {
		filePaths[i] = filesystem.Filename(files[i].Path)
	}

	// sort.Slice(files, func(i, j int) bool {
    //     return natsort.Compare(files[i], files[j])
    // })

	return filePaths
}

func GetLastRestoreIndex() int{
	if !filesystem.FileExists(constants.RESTORE_PATH) {
		return -1
	}

	files := GetRestoreList()


	if len(files) == 0 {
		return  -1
	}


	index := len(files) - 1
	// latestFilename := files[index]
	// lastestFilePath := filepath.Join(constants.RESTORE_PATH, latestFilename)

	// defaultManifest := GetDefaultManifest()
	// content := filesystem.ReadJSON[types.Manifest](lastestFilePath,defaultManifest)

	// fmt.Println("last file", latestFilename)
	return index
}

func GetLastManifest() (types.Manifest, bool) {
	if !filesystem.FileExists(constants.MANIFEST_PATH) {
		return types.Manifest{}, false
	}

	manifest := filesystem.ReadJSON[types.Manifest](constants.MANIFEST_PATH, GetDefaultManifest())
	return manifest, true

	// files := filesystem.ListFiles(constants.RESTORE_PATH)
	// if len(files) == 0 {
	// 	return types.Manifest{}, false
	// }

	// latestFilePath := files[len(files) - 1]
	// defaultManifest := GetDefaultManifest()
	// content := filesystem.ReadJSON[types.Manifest](latestFilePath,defaultManifest)

	// fmt.Println("last file", latestFilePath)
	// return content, true
}

func Commit(Manifest types.Manifest) {
	oldManifest := filesystem.ReadJSON[types.Manifest](constants.MANIFEST_PATH,GetDefaultManifest())
	diffed, configs  := parser.DiffFiles(oldManifest.Files, Manifest.Files)

	if len(diffed) == 0 {
		fmt.Println("No file changes found.")
		return
	}

	for i := 0; i < len(diffed); i++ {
		diff := diffed[i]
		value := configs[i]
		homeDir := filesystem.GetHomeDir()
	
		originalPath := strings.ReplaceAll(value.FilePath, "~", homeDir)
		originalTreePath := constants.ORIGINAL_TREE_PATH + value.FilePath
		resolvedPath := constants.RESOLVED_TREE_PATH + value.FilePath

		if diff.Action == types.DIFF_REMOVE {
			if (filesystem.FileExists(originalPath)) {
				if filesystem.FileExists(originalTreePath) {
					// replace with original
					originalFileContent := filesystem.ReadFile(originalTreePath)
					filesystem.CreateFile(originalPath, originalFileContent)
					
				} else {
					// removing file
					filesystem.RemoveFile(originalPath)
				}
			}

			if filesystem.FileExists(resolvedPath) {
				filesystem.RemoveFile(resolvedPath)
			}

			fmt.Println("  Removing file: ", originalPath)

			continue
		}

		if diff.Action == types.DIFF_MODIFY {
			fmt.Println("  Modifying file: ", originalPath)
		} 

		if diff.Action == types.DIFF_CREATE {
			fmt.Println("  Creating file: ", originalPath)
		}

		var content string

		if (!filesystem.FileExists(originalTreePath)) {
			if (filesystem.FileExists(originalPath)) {
				content = filesystem.ReadFile(originalPath)
				// copy file to originalTree
				filesystem.CreateFile(originalTreePath, content)
			}
		} else {
			content = filesystem.ReadFile(originalTreePath)
		}


		if(content != "") {
			// file at target path doesn't exist
			resolvedContent := value.Body
			if (value.MergeType == "append") {
				resolvedContent = string(content) + "\n" + value.Body
			} else  {
				resolvedContent = value.Body
			}
			
			filesystem.CreateFile(resolvedPath, resolvedContent)
		} else {
			// file at target path doesn't exist
			filesystem.CreateFile(resolvedPath, value.Body)
		}

		targetPath := strings.ReplaceAll(value.FilePath, "~", homeDir)
		if (filesystem.FileExists(targetPath)) {
			// TODO: try to see if the link already exists and prevent removing
			filesystem.RemoveFile(targetPath)
		}
		
		// resolvedConfigs = append(resolvedConfigs, resolved)
		filesystem.CreateSymlink(resolvedPath, targetPath)
	
	}

	_, foundPrevManifest := GetLastManifest()
	Manifest.Id = uuid.New().String()

	if(foundPrevManifest && filesystem.FileExists(constants.MANIFEST_PATH)) {
		if(!filesystem.FileExists(constants.RESTORE_PATH)) {
			filesystem.CreateFolder(constants.RESTORE_PATH)
		}

		timestamp := int(time.Now().Unix())
		newPath := filepath.Join(constants.RESTORE_PATH, strconv.Itoa(timestamp)  + ".json")
		filesystem.Move(constants.MANIFEST_PATH, newPath)
		fmt.Println("Create restore point", newPath)
	}

	filesystem.WriteJSON(constants.MANIFEST_PATH, Manifest)
}

func Apply() {
	fileString := filesystem.ReadFile(constants.CONFIG_PATH)
	Manifest := parser.Parse(fileString)
	Commit(Manifest)
}

func GetRestorePoints() {
	files := filesystem.ListFiles(constants.RESTORE_PATH)

	if len(files) == 0 {
		fmt.Println("No return points found!")
		return
	}

	fmt.Println("Restore points\n")

	tbl := table.New("Index", "Id", "Message")

	headerFmt := color.New(color.FgHiMagenta, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for i := len(files) - 1; i >= 0; i-- {
		file := files[i]
		content := filesystem.ReadJSON[types.Manifest](file.Path, GetDefaultManifest())
		tbl.AddRow(strconv.Itoa(i), content.Id, content.Message)
		// tbl.AddRow(file.Path, content.Message)
	}

	tbl.Print()
}

func Restore(index int) {
	files := filesystem.ListFiles(constants.RESTORE_PATH)

	file := files[index]	
	Manifest := filesystem.ReadJSON[types.Manifest](file.Path, GetDefaultManifest())

	// replace restore point with current manifest
	// apply

	Commit(Manifest)

	fmt.Println("Not implemented")
}

func Revert() {
	fmt.Println("Not implemented")
}