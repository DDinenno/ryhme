package types

type Status int

const (
	BUILDING_PACKAGE_LIST Status = iota
	BUILDING_CONFIG
	SCANNING
)

type PackageType string

const (
	DEFAULT_PACKAGE  PackageType = "Packages"
	FLATPAK_PACKAGE  PackageType = "Flatpaks"
	SNAP_PACKAGE     PackageType = "Snaps"
	APPIMAGE_PACKAGE PackageType = "Appimages"
)

var PackageTypes = []PackageType{DEFAULT_PACKAGE, FLATPAK_PACKAGE, SNAP_PACKAGE, APPIMAGE_PACKAGE}

type Package struct {
	Name    string
	Version string
	Remote  string
}

type PackageDiff struct {
	Created    []Package
	Updated    []Package
	Removed    []Package
	HasChanges bool
}

type SourceFile struct {
	FilePath string
	Body     string
}

type ConfigFile struct {
	FilePath  string
	Source    string
	Start     int
	End       int
	MergeType string
	Body      string
}

type Repository struct {
	Name string
	Url  string
}

type FlatpakConfig struct {
	Repos    []Repository
	Packages []Package
}

type Config struct {
	Id          string
	Message     string
	SourceFiles []SourceFile
	Files       []ConfigFile
	Packages    []Package
	Flatpaks    []Package
}

type DiffAction string

const (
	DIFF_CREATE DiffAction = "create"
	DIFF_MODIFY DiffAction = "modify"
	DIFF_REMOVE DiffAction = "remove"
)

type DiffType struct {
	Name   string
	Action DiffAction
}

type DiffedConfig struct {
	Files    []DiffType
	Packages []DiffType
}

type RestorePoint struct {
	Id      int
	Message string
}

type History = []Config

type Manifest struct {
	SelectedConfig string
	History        History
}
