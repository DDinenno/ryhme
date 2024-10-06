package stash

import (
	"fmt"
	"os"
	"stash/src/actions"
	"stash/src/constants"
	"stash/src/parser"
	"stash/src/types"
	filesystem "stash/src/utils"

	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	TEST_FILESYSTEM_DIR = "mock-filesystem"
	TEST_CONFIG_PATH    = "test.conf"
)

func TestMain(m *testing.M) {
	// Setup code (runs before all tests)
	fmt.Println("Setting up before all tests")

	constants.ROOT_PATH = TEST_FILESYSTEM_DIR + "/real-tree/"
	constants.MANIFEST_PATH = TEST_FILESYSTEM_DIR + "/.manifest.json"
	constants.RESOLVED_TREE_PATH = TEST_FILESYSTEM_DIR + "/home-tree"
	constants.ORIGINAL_TREE_PATH = TEST_FILESYSTEM_DIR + "/original-tree"
	constants.CONFIG_PATH = TEST_FILESYSTEM_DIR + TEST_CONFIG_PATH

	// Run all the tests
	code := m.Run()

	// Teardown code (runs after all tests)
	fmt.Println("Tearing down after all tests")

	// Exit with the correct code
	os.Exit(code)
}

func RecreateFileSystem() {
	if filesystem.FileExists(TEST_FILESYSTEM_DIR) {
		filesystem.RemoveFolderRecursively(TEST_FILESYSTEM_DIR)
	}

	filesystem.CreateFolder(TEST_FILESYSTEM_DIR)

}

func TestParsing(t *testing.T) {
	fileString := filesystem.ReadFile(TEST_CONFIG_PATH)

	source := types.SourceFile{
		FilePath: TEST_CONFIG_PATH,
		Body:     fileString,
	}

	Config := parser.Parse(source)
	fmt.Println(Config)

	// TODO

	// recreatedBuild := parser.BuildConfigString(Config)
	// test.CompareConfigStrings(fileString, recreatedBuild)
}

func TestCommittingChanges(t *testing.T) {
	RecreateFileSystem()

	configs := []types.ConfigFile{
		{
			FilePath: "~/test/.test",
			Body:     "Something=true\nSomethingElse=true",
		},
		{
			FilePath: "~/test/multiple/paths/nested.file",
			Body:     "This\nis\na\nnested\nfile!",
		},
		{
			FilePath: "~/.bashrc",
			Body:     "This is a bashrc file",
		},
	}

	mockedConfigFile := ""

	for _, conf := range configs {
		// Append config to a mocked config file
		mockedConfigFile += conf.FilePath + "\n{{\n" + conf.Body + "\n}}\n\n"
	}

	source := types.SourceFile{
		FilePath: TEST_CONFIG_PATH,
		Body:     mockedConfigFile,
	}

	Config := parser.Parse(source)
	actions.Commit(Config, true)

	// Ensure all files have been created on the filesystem and match what's in the stash config
	for _, conf := range configs {
		configContent := filesystem.ReadFile(actions.GetRealPath(conf.FilePath))
		assert.Equal(t, conf.Body, configContent, "Config doesn't match the one on the filesystem!")
	}
}
