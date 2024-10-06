package actions

import (
	"fmt"
	"log"
	"os"
	"stash/src/constants"
	"stash/src/parser"
	"stash/src/types"

	array "stash/src/utils"
	commands "stash/src/utils"
	filesystem "stash/src/utils"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/manifoldco/promptui"
	"github.com/rodaine/table"
)

var version = "0.0.0"

func PrintVersion() {
	fmt.Println("Version: " + version)
}

func GetManifest() types.Manifest {
	files := filesystem.ReadJSON[types.Manifest](constants.MANIFEST_PATH, types.Manifest{})
	return files
}

func GetDefaultConfig() types.Config {
	return types.Config{
		Id:       "",
		Message:  "No message",
		Files:    []types.ConfigFile{},
		Packages: []types.Package{},
	}

}

func getLastConfig() (types.Config, bool) {
	history := GetManifest().History

	if len(history) == 0 {
		return types.Config{}, false
	}

	return history[0], true
}

func SyncPackages(packageType types.PackageType, diff types.PackageDiff) {
	commandsRan := 0

	if len(diff.Removed) > 0 {
		commands.RunRemovePackage(packageType, diff.Removed)
		commandsRan += 1
	}

	if len(diff.Created) > 0 {
		commands.RunInstallPackage(packageType, diff.Created)
		commandsRan += 1

	}

	if len(diff.Updated) > 0 {
		commands.RunChangePackage(packageType, diff.Updated)
		commandsRan += 1
	}

}

func getSelectedConfig() (types.Config, bool) {
	manifest := GetManifest()
	history := manifest.History

	if len(history) == 0 {
		return types.Config{}, false
	}

	var config types.Config
	var found bool = false

	for _, c := range history {
		if manifest.SelectedConfig == c.Id {
			config = c
			found = true
			break
		}
	}

	return config, found
}

func CheckForFileEdits() {
	config, foundConfig := getSelectedConfig()
	if !foundConfig {
		return
	}

	var changes []types.ConfigFile

	for i, file := range config.Files {
		targetPath := filesystem.ResolvePath(constants.RESOLVED_TREE_PATH, file.FilePath)

		content := filesystem.ReadFile(targetPath)

		if content != file.Body {
			file.Body = content
			config.Files[i] = file
			changes = append(changes, file)
		}
	}

	if len(changes) > 0 {
		message := fmt.Sprintf("%d file(s) have been changed since last commit. Do you want to sync these changes?", len(changes))

		for _, file := range changes {
			// message += "\n  " + file.FilePath
			fmt.Println(file.FilePath)

		}

		prompt := promptui.Select{
			Label: message,
			Items: []string{"Yes", "No", "View changes"},
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result {
		case "Yes":
			manifest := GetManifest()
			history := manifest.History

			for i, c := range history {
				if c.Id == config.Id {
					history[i] = config
				}
			}

			filesystem.WriteJSON(constants.MANIFEST_PATH, manifest)

			configString := parser.BuildConfigString(config)
			filesystem.CreateFile(constants.CONFIG_PATH, configString)

		case "No":
			// fmt.Println("Value is 2")
		case "View changes":
			fmt.Println("Not implemented")
		}

		prompt2 := promptui.Select{
			Label: "Continue with original command?",
			Items: []string{"Yes", "No"},
		}

		_, result2, err := prompt2.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		switch result2 {
		case "Yes":
			fmt.Println("")
		case "No":
			os.Exit(0)
		}
	}
}

var revertingCommit bool = false
var revertedConfig types.Config
var revertError interface{}
var committed []string

func runCommitStage(stage string, function func()) {
	if revertingCommit {
		fmt.Println("Reverting stage: ", stage)

		if !array.Includes(committed, stage) {
			log.Panic("ending-early")
		}
	}

	function()

	committed = append(committed, stage)
}

func GetRealPath(path string) string {
	homeDir := filesystem.GetHomeDir()
	return strings.ReplaceAll(constants.ROOT_PATH+path, "~", homeDir)
}

func Commit(_config types.Config, pushHistory bool) {
	Config := _config

	_oldConfig, _ := getSelectedConfig()

	oldConfig := _oldConfig
	if revertingCommit {
		oldConfig = revertedConfig
	}

	defer func() {
		if r := recover(); r != nil {
			if !revertingCommit {
				revertingCommit = true
				revertedConfig = Config
				revertError = r
				fmt.Println("Reverting...")
				Commit(oldConfig, false)
			} else {
				fmt.Println("Failed to Commit changes:", revertError)

				if r != "ending-early" {
					log.Panic("Failed to revert changes:", r)
					// fmt.Println("Failed to revert changes:", r)
				} else {
					log.Panic("Failed to revert changes:", revertError)
				}

			}
		}
	}()

	configs, diffedConfigs := parser.DiffFiles(oldConfig.Files, Config.Files)

	var diffPackages []types.PackageDiff
	var diffPackageTypes []types.PackageType

	for _, packageType := range types.PackageTypes {
		switch packageType {
		case types.DEFAULT_PACKAGE:
			diffs := parser.DiffPackages(oldConfig.Packages, Config.Packages)
			diffPackages = append(diffPackages, diffs)
			diffPackageTypes = append(diffPackageTypes, types.DEFAULT_PACKAGE)
		case types.FLATPAK_PACKAGE:
			diffs := parser.DiffPackages(oldConfig.Flatpaks, Config.Flatpaks)
			diffPackages = append(diffPackages, diffs)
			diffPackageTypes = append(diffPackageTypes, types.FLATPAK_PACKAGE)
		default:
		}
	}

	_, hasPackageChanges := array.Find(diffPackages, func(p types.PackageDiff) bool {
		return p.HasChanges
	})

	if len(diffedConfigs) == 0 && !hasPackageChanges {
		fmt.Println("No file changes found.")
		os.Exit(1)
	}

	runCommitStage("configs", func() {
		for i, conf := range configs {
			diffAction := diffedConfigs[i]

			originalPath := GetRealPath(conf.FilePath)
			originalTreePath := filesystem.ResolvePath(constants.ORIGINAL_TREE_PATH, conf.FilePath)
			resolvedPath := filesystem.ResolvePath(constants.RESOLVED_TREE_PATH, conf.FilePath)

			if diffAction == types.DIFF_REMOVE {
				if filesystem.FileExists(originalPath) {
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

			if diffAction == types.DIFF_MODIFY {
				fmt.Println("  Modifying file: ", originalPath)
			}

			if diffAction == types.DIFF_CREATE {
				fmt.Println("  Creating file: ", originalPath)
			}

			var content string

			if !filesystem.FileExists(originalTreePath) {
				if filesystem.FileExists(originalPath) {
					content = filesystem.ReadFile(originalPath)
					// copy file to originalTree
					filesystem.CreateFile(originalTreePath, content)
				}
			} else {
				content = filesystem.ReadFile(originalTreePath)
			}

			if content != "" {
				// file at target path doesn't exist
				resolvedContent := conf.Body
				if conf.MergeType == "append" {
					resolvedContent = string(content) + "\n" + conf.Body
				} else {
					resolvedContent = conf.Body
				}

				filesystem.CreateFile(resolvedPath, resolvedContent)
			} else {
				// file at target path doesn't exist
				filesystem.CreateFile(resolvedPath, conf.Body)
			}

			targetPath := GetRealPath(conf.FilePath)
			if filesystem.DoesSymlinkExist(targetPath) {
				log.Println("File exists, removing", targetPath)
				// TODO: try to see if the link already exists and prevent removing
				filesystem.RemoveFile(targetPath)
			}

			// resolvedConfigs = append(resolvedConfigs, resolved)
			filesystem.CreateSymlink(resolvedPath, targetPath)

		}
	})

	for i, packageDiff := range diffPackages {
		packageType := diffPackageTypes[i]

		commandsRan := 0

		runCommitStage(string(packageType)+"-create/update", func() {
			if len(packageDiff.Created) > 0 {
				commands.RunInstallPackage(packageType, packageDiff.Created)
				commandsRan += 1
			}
		})

		runCommitStage(string(packageType)+"-create/update", func() {
			if len(packageDiff.Updated) > 0 {
				commands.RunChangePackage(packageType, packageDiff.Updated)
				commandsRan += 1
			}
		})

		runCommitStage(string(packageType)+"-remove", func() {
			if len(packageDiff.Removed) > 0 {
				commands.RunRemovePackage(packageType, packageDiff.Removed)
				commandsRan += 1
			}
		})
	}

	manifest := GetManifest()

	if pushHistory {
		history := manifest.History
		Config.Id = uuid.New().String()
		history = append([]types.Config{Config}, history...)
		manifest.History = history
	}

	manifest.SelectedConfig = Config.Id
	filesystem.WriteJSON(constants.MANIFEST_PATH, manifest)
}

func Apply() {
	fileString := filesystem.ReadFile(constants.CONFIG_PATH)

	entryPoint := types.SourceFile{
		FilePath: constants.CONFIG_PATH,
		Body:     fileString,
	}

	Config := parser.Parse(entryPoint)
	Commit(Config, true)
}

func PrintRestorePoints() {
	manifest := GetManifest()
	history := manifest.History

	if len(history) == 0 {
		fmt.Println("No return points found!")
		return
	}

	fmt.Println("Restore points:")

	tbl := table.New("Index", "Id", "Message")

	headerFmt := color.New(color.FgHiMagenta, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	for i := 0; i < len(history); i++ {
		file := history[i]

		var index string

		if manifest.SelectedConfig == file.Id {
			index = strconv.Itoa(i) + " (current)"
		} else {
			index = strconv.Itoa(i)
		}

		tbl.AddRow(index, file.Id, file.Message)
	}

	tbl.Print()
}

func Restore(index int) {
	manifest := GetManifest()
	history := manifest.History

	if index < 0 || index > len(history)-1 {
		fmt.Printf("Invalid Index %d!", index)
		return
	}

	Config := history[index]

	fmt.Println("Restoring to index:", index)
	Commit(Config, false)

	// configString := parser.BuildConfigString(Config)
	for _, source := range Config.SourceFiles {
		filesystem.CreateFile(constants.CONFIG_PATH, source.Body)
	}
}

func Revert() {
	fmt.Println("Not implemented")
}
