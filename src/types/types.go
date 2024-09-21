package types

type Status int

const (
    BUILDING_PACKAGE_LIST Status = iota
    BUILDING_CONFIG
    SCANNING
)

type PackageType string


const (
	DEFAULT PackageType = "packages"
    FLATPAK PackageType = "flatpacks"
    SNAP PackageType = "snaps"
    APP_IMAGE PackageType = "appimages"
)

var PackageTypes = []PackageType{DEFAULT, FLATPAK, SNAP, APP_IMAGE}

type Package struct {
	Name string
	Version string
}

type ConfigFile struct {
	FilePath string
	Start int
	End int
	MergeType string
	Body  string
}

type Config struct {
	Id string
	Message string
	Files []ConfigFile
	Packages []Package
	// flatpaks []Package
}

type History = []Config

type DiffAction string

const (
	DIFF_CREATE DiffAction = "create"
    DIFF_MODIFY DiffAction = "modify"
    DIFF_REMOVE DiffAction = "remove"
)

type DiffType struct {
	Name string
	Action DiffAction
}

type DiffedConfig struct {
	Files []DiffType
	Packages []DiffType
}

type RestorePoint struct {
    Id   int
    Message string
}

type Manifest struct {
	SelectedConfig string
}