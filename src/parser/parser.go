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

func HandleParsingPackages(lines []string, startingLine int) () {

}

func Parse(fileString string) types.Config {
	currentLine := 0
	files := []types.ConfigFile{}
	var message = "Did some stuff"
	var currentFile types.ConfigFile
	var packages = []types.Package{}
	var flatpaks = []types.Package{}

	filepathRegex := regexp.MustCompile(`^[a-zA-Z0-9_\-/\\.\~]+(\.[a-zA-Z0-9]+)?`)
	openingBracketRegex := regexp.MustCompile("^" + openingBracket + "(\\s+|)$")
	closingBracketRegex := regexp.MustCompile("^" + closingBracket + "(\\s+|)$")

	packageStartingRegexStr := `^[a-zA-Z0-9]+ = \[`
	packageStartingRegex := regexp.MustCompile(packageStartingRegexStr)
	commentedLineRegex := regexp.MustCompile(`^#`)
	
	text := strings.ReplaceAll(fileString, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	status := types.SCANNING
	var packageType types.PackageType

	defer func() {
        if r := recover(); r != nil {
            // Catch the panic
            fmt.Println(fmt.Sprintf("%d %s\n^\n", currentLine, lines[currentLine]))
			log.Panic("Parsing Error: ", r)
			
        }
    }()

	for i := 0; i < len(lines); i++ {
		currentLine = i
		line := lines[currentLine]

		if (strings.TrimSpace(line) == "" || commentedLineRegex.MatchString(line)) {
			continue
		}

		if status == types.SCANNING {

			if packageStartingRegex.MatchString(line) {
				status = types.BUILDING_PACKAGE_LIST
				fields := strings.Fields(line)
				packageType = "unset"

				for _, pType := range types.PackageTypes {
					if (fields[0] == strings.ToLower(string(pType))) {
						packageType = pType
						break
					}
				}

				if packageType == "unset" {
					log.Panic("Package type ", fields[0], " doesn't exist!" )
				}

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
						fmt.Println("Duplicate filename '" + currentFile.FilePath + "' at line:", currentLine)
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
				fields := strings.Fields(line)
				version := "any"

				if (fields[0] == "#") {
					continue
				}

				if (packageType == types.DEFAULT_PACKAGE) {
					if len(fields) > 1 && fields[1] != ""{
						version = fields[1]
					}
					
					currentPackage := types.Package{
						Name: fields[0],
						Version: version,
					}
	
					packages = append(packages, currentPackage)
				} else if(packageType == types.FLATPAK_PACKAGE) {
					var remote string = "flathub"
					version = "any"

					if len(fields) == 2{
						remote = fields[1]
					} else if len(fields) == 3 {
						remote = fields[1]
						version = fields[2]
					}
					
					currentPackage := types.Package{
						Name: fields[0],
						Version: version,
						Remote: remote,
					}



					flatpaks = append(flatpaks, currentPackage)
				} else {	
					log.Panic("package type ", packageType, " not implemented!")
				}


			
			}
		} else if status == types.BUILDING_CONFIG {
			if (currentFile.Start == -1) {
				if(openingBracketRegex.MatchString(openingBracket)) {
					currentFile.Start = currentLine
					
				} else {
					log.Panicf("Expected '" + openingBracket + "'")
				}
	
	
			} else if (currentFile.End == -1) {
				if(closingBracketRegex.MatchString(line)) {
					currentFile.End = currentLine
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
		Flatpaks: flatpaks,
	}
}

func BuildConfigString(config types.Config) string {
	configString := ""

	for _, file := range config.Files {
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
				configString += " " + value.Version + "\n"
			}
			configString += "\n"
		}

		configString += "]"
	}

	return configString
}

func DiffFiles(old []types.ConfigFile, new []types.ConfigFile) ([]types.ConfigFile, []types.DiffAction) {
	var diffed []types.DiffAction
	var configs []types.ConfigFile

	for _, config := range old {
		matched, found := array.Find(new, func (c types.ConfigFile) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			diffed = append(diffed, types.DIFF_REMOVE)
			configs = append(configs, config)
		} else if (matched.Body != config.Body)  {
			diffed = append(diffed, types.DIFF_MODIFY)
			configs = append(configs, matched)
		}
	}

	// Find elements in new not in old
	for _, config := range new {
		_, found := array.Find(old, func (c types.ConfigFile) bool {
			return c.FilePath == config.FilePath
		})

		if !found {
			diffed = append(diffed, types.DIFF_CREATE)
			configs = append(configs, config)
		} 
	}

	return configs, diffed
}

func DiffPackages(old []types.Package, new []types.Package) types.PackageDiff {
	var created []types.Package
	var updated []types.Package
	var removed []types.Package
	hasChanges := false


	for _, pack := range old {
		matched, found := array.Find(new, func (p types.Package) bool {
			return p.Name == pack.Name
		})

		if !found {
			removed = append(removed, pack)
			hasChanges = true
		} else if (matched.Version != pack.Version ||  matched.Remote != pack.Remote)  {
			updated = append(updated, matched)
			hasChanges = true
		}
	}

	// Find elements in new not in old
	for _, pack := range new {
		_, found := array.Find(old, func (p types.Package) bool {
			return p.Name == pack.Name
		})

		if !found {
			created = append(created, pack)
			hasChanges = true
		} 
	}


	return types.PackageDiff{
		Created: created, 
		Updated: updated,
		Removed: removed,
		HasChanges: hasChanges,
	}
}