package main

import (
	"fmt"
	"log"
	"stash/src/constants"
	"stash/src/parser"
	filesystem "stash/src/utils"
	"strings"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func compareConfigStrings(configA string, configB string) {
	green := color.New(color.FgGreen).SprintFunc()
    blue := color.New(color.FgBlue).SprintFunc()

	splitA :=  strings.Split(configA, "\n")
	splitB := strings.Split(configB, "\n")

	tbl := table.New("Index", "Id", "Message")
	headerFmt := color.New(color.FgHiMagenta, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	arr := splitA
	if len(splitB) > len(splitA) {
		arr = splitB
	}

	longestLine := 0
	for _, line := range arr {
		if len(line) > longestLine {
			longestLine = len(line)
		}
	}

	errLinesPrint := 5
	var previousLines []string

	for i, _ := range arr {
		lineA := strings.TrimRight(splitA[i], " ")

		if i >= len(splitB) {
			previousLines = append(previousLines, fmt.Sprintf("[%d] " + green("A: " + lineA), i))
			continue;
		}

		lineB := strings.TrimRight(splitB[i], " ")
		lineSegment := fmt.Sprintf("[%d] " + lineA, i)
		paddedLineA := fmt.Sprintf("%-*s", longestLine + 10, lineSegment )
		previousLines = append(previousLines, paddedLineA + blue(lineB))

		if lineA != lineB {
			lastFive := previousLines[len(previousLines)-errLinesPrint:]
			for _, v := range lastFive {
				fmt.Println(v)
			}

			log.Panicln("Lines didn't match")
		}
	}

	if len(splitA) != len(splitB) {
		lastFive := previousLines[len(previousLines)-errLinesPrint:]
		for _, v := range lastFive {
			fmt.Println(v)
		}

		log.Panicln("Config lengths differed")
	} 
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Test Failed:", r)
		}
	}()

	fileString := filesystem.ReadFile(constants.TEST_CONFIG_PATH)
	Config := parser.Parse(fileString)
	recreatedBuild := parser.BuildConfigString(Config)
	compareConfigStrings(fileString, recreatedBuild)

	fmt.Println("Tests succeeded!")
}