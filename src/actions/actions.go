package actions

import (
	"fmt"
	"os"
	"stash/src/constants"
	"stash/src/parser"
	"stash/src/types"
	filesystem "stash/src/utils"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/rodaine/table"
)

var version = "0.0.0"

func PrintVersion() {
	fmt.Println("Version: " + version)
}

func GetHistory() types.History {
	files := filesystem.ReadJSON[types.History](constants.HISTORY_PATH, types.History{})
	return files
}

func GetManifest() types.Manifest {
	files := filesystem.ReadJSON[types.Manifest](constants.MANIFEST_PATH, types.Manifest{})
	return files
}

func GetDefaultConfig() types.Config {
	return types.Config{
		Id: "",
		Message: "No message",
		Files: []types.ConfigFile{},
		Packages: []types.Package{},
	}

}

func getLastConfig() (types.Config, bool) {
	history := GetHistory() 

	if len(history) == 0 {
		return types.Config{}, false
	}

	return history[0], true
}


func getSelectedConfig() (types.Config, bool) {
	history := GetHistory() 
	manifest := GetManifest()

	if len(history) == 0 {
		return types.Config{}, false
	}

	var config types.Config
	var found bool = false


	for _, c := range history {
		if (manifest.SelectedConfig == c.Id) {
			config = c
			found = true
			break
		}
	}

	return config, found
}

func Commit(Config types.Config, pushHistory bool) {
	oldConfig, _ := getSelectedConfig()
	diffed, configs  := parser.DiffFiles(oldConfig.Files, Config.Files)

	if len(diffed) == 0 {
		fmt.Println("No file changes found.")
		os.Exit(1)
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


	if pushHistory {
		history := GetHistory()
		Config.Id = uuid.New().String()
		history = append([]types.Config{Config}, history... )
	
		filesystem.WriteJSON(constants.HISTORY_PATH, history)
	}

	manifest := GetManifest()

	manifest.SelectedConfig = Config.Id
	filesystem.WriteJSON(constants.MANIFEST_PATH, manifest)
}

func Apply() {
	fileString := filesystem.ReadFile(constants.CONFIG_PATH)
	Config := parser.Parse(fileString)
	Commit(Config, true)
}


func PrintRestorePoints() {
	history := GetHistory()

	if len(history) == 0 {
		fmt.Println("No return points found!")
		return
	}

	fmt.Println("Restore points\n")

	tbl := table.New("Index", "Id", "Message")

	headerFmt := color.New(color.FgHiMagenta, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	manifest := GetManifest()

	for i := 0; i < len(history); i++ {
		file := history[i]

		var index string

		if (manifest.SelectedConfig == file.Id) {
			index = strconv.Itoa(i) + " (current)"
		} else {
			index = strconv.Itoa(i)
		}

		tbl.AddRow(index, file.Id, file.Message)
	}

	// for i := len(history) - 1; i >= 0; i-- {
	// 	file := history[i]
	// 	tbl.AddRow(strconv.Itoa(i), file.Id, file.Message)
	// }

	tbl.Print()
}




func Restore(index int) {
	history := GetHistory()

	if (index < 0 || index > len(history) - 1) {
		fmt.Println("Invalid Index " + strconv.Itoa(index) + "!")
		return 
	}

	Config := history[index]	

	fmt.Println("Restoring to index:", index)
	Commit(Config, false)

	configString := parser.BuildConfigString(Config)

	filesystem.CreateFile(constants.CONFIG_PATH,configString)
}

func Revert() {
	fmt.Println("Not implemented")
}