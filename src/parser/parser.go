package parser

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"stash/src/types"
	array "stash/src/utils"
	"strings"
)


func Parse(fileString string) types.Manifest {
	currentLine := 0
	files := []types.Config{}
	var message = "Did some stuff"
	var currentFile types.Config
	var packages = []types.Package{}


	filepathRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-/\\.\~]+(\.[a-zA-Z0-9]+)?`)
	doesntStartWithStringRegex := regexp.MustCompile(`^[^\s].*`)
	packageStartingRegexStr := `^[a-zA-Z]+ = \[`
	packageStartingRegex := regexp.MustCompile(packageStartingRegexStr)
	commentedLineRegex := regexp.MustCompile(`^#`)
	
	text := strings.ReplaceAll(fileString, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	status := types.SCANNING

	defer func() {
        if r := recover(); r != nil {
            // Catch the panic
            fmt.Println(fmt.Sprintf("%d %s\n^\n", currentLine, lines[currentLine]))
			log.Panic("Parsing Error: ", r)
			
        }
    }()
	
	for i := 0; i < len(lines); i++ {
		currentLine = i
		line := lines[i]

		if (strings.TrimSpace(line) == "" || commentedLineRegex.MatchString(line)) {
			continue
		}

		if status == types.SCANNING {

			if packageStartingRegex.MatchString(line) {
				status = types.BUILDING_PACKAGE_LIST
			} else if currentFile.FilePath == "" {
				
				if  filepathRegex.MatchString(line) {
					status = types.BUILDING_CONFIG

					currentFile = types.Config{
						FilePath: "",
						MergeType: "append",
						Start: -1,
						End: -1,
						Body: "",
					}
	
					splitParams := strings.Split(line, " ")	
					currentFile.FilePath = splitParams[0]

					_, duplicateFile := array.Find(files, func (c types.Config) bool {
						return c.FilePath == currentFile.FilePath
					})

					if duplicateFile {
						fmt.Println("Duplicate filename '" + currentFile.FilePath + "' at line:", i)
						os.Exit(1) // Use a non-zero exit code to indicate an error
					}
	
					for i := 1; i < len(splitParams); i++ {
						param := splitParams[i]
						
						if param == "-r" || param == "-replace" {
							currentFile.MergeType = "replace"
						}
	
						if param == "-a" || param == "-append" {
							currentFile.MergeType = "append"
						}
					}
				} else {
					log.Panicf("Expected filepath!")
				}
			}
		} else if (status == types.BUILDING_PACKAGE_LIST) {
		

			if(strings.TrimRight(line, " ") == "]") {
				status = types.SCANNING
			} else {
				packageStr := strings.ReplaceAll(line, " ", "")

				if (commentedLineRegex.MatchString(packageStr)) {
					continue
				}

				sections := strings.Split(packageStr,":")

				version := "any"

				if len(sections) > 1 && sections[1] != ""{
					version = sections[1]
				}
				

				currentPackage := types.Package{
					Name: sections[0],
					Version: version,
				}

				packages = append(packages, currentPackage)
			}
		} else if status == types.BUILDING_CONFIG {
			if (currentFile.Start == -1) {
				if(strings.TrimRight(line, " ") == "{") {
					currentFile.Start = i
					
				} else {
					log.Panicf("Expected '{' (opening bracket)")
				}
	
	
			} else if (currentFile.End == -1) {
				if(strings.TrimRight(line, " ") == "}") {
					currentFile.End = i
					files = append(files, currentFile)
	
					// create new config
					currentFile = types.Config{
						FilePath: "",
						MergeType: "append",
						Start: -1,
						End: -1,
						Body: "",
					}

					status = types.SCANNING
				} else {
					if (doesntStartWithStringRegex.MatchString(line)  &&  strings.TrimRight(line, " ") != "}"  ) {
						log.Panicf("Expected '}' (closing bracket)")
					}		
					line = strings.TrimPrefix(line, "    ")
					line = (line + "\n") 
					currentFile.Body += line
				}
			}
		}
	}

	return types.Manifest{
		Message: message,
		Files: files,
		Packages: packages,
	}
}


func DiffFiles(old []types.Config, new []types.Config) ([]types.DiffType, []types.Config) {
	var diffed []types.DiffType
	var configs []types.Config

	for _, config := range old {
		matched, found := array.Find(new, func (c types.Config) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_REMOVE }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		} else if (matched.Body != config.Body)  {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_MODIFY }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		}
	}

	// Find elements in new not in old
	for _, config := range new {
		_, found := array.Find(old, func (c types.Config) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_CREATE }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		} 
	}

	return diffed, configs
}

func DiffPackages(old []types.Config, new []types.Config) ([]types.DiffType, []types.Config) {
	var diffed []types.DiffType
	var configs []types.Config

	for _, config := range old {
		matched, found := array.Find(new, func (c types.Config) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_REMOVE }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		} else if (matched.Body != config.Body)  {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_MODIFY }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		}
	}

	// Find elements in new not in old
	for _, config := range new {
		_, found := array.Find(old, func (c types.Config) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_CREATE }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		} 
	}

	return diffed, configs
}