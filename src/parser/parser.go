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

var (
	openingBracket = "{{"
	closingBracket = "}}"
	indent = "    "
	lineEndingsRegexp = regexp.MustCompile("\r\n|\n")
)

func Parse(fileString string) types.Config {
	currentLine := 0
	files := []types.ConfigFile{}
	var message = "Did some stuff"
	var currentFile types.ConfigFile
	var packages = []types.Package{}


	filepathRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-/\\.\~]+(\.[a-zA-Z0-9]+)?`)
	// doesntStartWithStringRegex := regexp.MustCompile(`^[^\s].*`)
	openingBracketRegex := regexp.MustCompile("^" + openingBracket + "(\\s+|)$")
	closingBracketRegex := regexp.MustCompile("^" + closingBracket + "(\\s+|)$")

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

					currentFile = types.ConfigFile{
						FilePath: "",
						MergeType: "append",
						Start: -1,
						End: -1,
						Body: "",
					}
	
					splitParams := strings.Split(line, " ")	
					currentFile.FilePath = splitParams[0]

					_, duplicateFile := array.Find(files, func (c types.ConfigFile) bool {
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
				if(openingBracketRegex.MatchString(openingBracket)) {
					currentFile.Start = i
					
				} else {
					log.Panicf("Expected '" + openingBracket + "'")
				}
	
	
			} else if (currentFile.End == -1) {
				if(closingBracketRegex.MatchString(line)) {
					currentFile.End = i
					files = append(files, currentFile)
	
					// create new config
					currentFile = types.ConfigFile{
						FilePath: "",
						MergeType: "append",
						Start: -1,
						End: -1,
						Body: "",
					}

					status = types.SCANNING
				} else {
					line = strings.TrimPrefix(line, "    ")
					line = (line + "\n") 
					currentFile.Body += line
				}
			}
		}
	}

	return types.Config{
		Message: message,
		Files: files,
		Packages: packages,
	}
}

func BuildConfigString(config types.Config) string {
	configString := ""

	for _, file := range config.Files {
		// fmt.Println("Index:", index, "Value:", value)

		// configString +=	value.Body + "\n"



		lines := lineEndingsRegexp.Split(file.Body, -1)

		configString += file.FilePath 
		if (file.MergeType == "append") {
			configString += " " + "-a\n"
		} else if (file.MergeType == "replace") {
			configString += " " + "-r\n"
		} else {
			configString += "\n"
		}

		configString += openingBracket + "\n"
	
		for _, line := range lines {
			configString += indent + line + "\n"
		}

		configString += closingBracket + "\n\n"

	}

	if (len(config.Packages) > 0) {
		configString += "packages = [\n"

		for _, value := range config.Packages {
			configString += indent + value.Name 

			if(value.Version != "any") {
				configString +=  value.Version + "\n"
			}
			configString += "\n"
		}

		configString += "]"
	}

	return configString
}

func DiffFiles(old []types.ConfigFile, new []types.ConfigFile) ([]types.DiffType, []types.ConfigFile) {
	var diffed []types.DiffType
	var configs []types.ConfigFile

	for _, config := range old {
		matched, found := array.Find(new, func (c types.ConfigFile) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_REMOVE }
			diffed = append(diffed, entry)
			configs = append(configs, config)
		} else if (matched.Body != config.Body)  {
			entry := types.DiffType{ Name: config.FilePath, Action: types.DIFF_MODIFY }
			diffed = append(diffed, entry)
			configs = append(configs, matched)
		}
	}

	// Find elements in new not in old
	for _, config := range new {
		_, found := array.Find(old, func (c types.ConfigFile) bool {
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

func DiffPackages(old []types.ConfigFile, new []types.ConfigFile) ([]types.DiffType, []types.ConfigFile) {
	var diffed []types.DiffType
	var configs []types.ConfigFile

	for _, config := range old {
		matched, found := array.Find(new, func (c types.ConfigFile) bool {
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
		_, found := array.Find(old, func (c types.ConfigFile) bool {
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