package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"stash/src/types"
	"strings"
)

func RunCommand(command string) {
	args := strings.Split(command, " ")
	cmd := exec.Command(args[0], args[1:]...)

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Panic("Error getting stdout:", err)
    }

    stderr, err := cmd.StderrPipe()
    if err != nil {
        log.Panic("Error getting stderr:", err)
    }

    if err := cmd.Start(); err != nil {
        log.Panic("Error starting command:", err)
    }

    go func(reader io.ReadCloser) {
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }(stdout)

    go func(reader io.ReadCloser) {
        scanner := bufio.NewScanner(reader)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }(stderr)

    if err := cmd.Wait(); err != nil {
		log.Panic("Error while executing command '" + command + "':\n", err)
    }
}

func resolvePackage(packageType types.PackageType, pack types.Package) string {
    switch packageType {
    case types.DEFAULT_PACKAGE:
        if pack.Version == "any" { 
            return pack.Name
        }

        return pack.Name + "-" + pack.Version
    case types.FLATPAK_PACKAGE:
        name := ""

        if (pack.Remote != "") {
            name += "flathub"
        }

        name += " " + pack.Name

        if pack.Version != "any" {
            name += " " + pack.Version
        }
        
        return name
    default:
        log.Panic("package type ", packageType, " Isn't implemented!")
        return ""
    }
}


func RunInstallPackage(packageType types.PackageType, packages []types.Package) {
    var packageNames []string

    for _, pack := range packages {
        packageNames = append(packageNames, resolvePackage(packageType, pack))
    }

    switch packageType {
    case types.DEFAULT_PACKAGE:
        RunCommand("sudo dnf install -y " +  strings.Join(packageNames, " "))
    case types.FLATPAK_PACKAGE:
        RunCommand("sudo flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo")
        RunCommand("sudo flatpak install -y " + strings.Join(packageNames, " "))
    default:
        log.Panic("package type ", packageType, " Isn't implemented!")
    }
}


func RunChangePackage(packageType types.PackageType, packages []types.Package) {
	var packageNames []string

	for _, pack := range packages {
		packageNames = append(packageNames, resolvePackage(packageType, pack))
	}

    switch packageType {
    case types.DEFAULT_PACKAGE:
        RunCommand("sudo dnf install -y " +  strings.Join(packageNames, " "))
    case types.FLATPAK_PACKAGE:
        RunCommand("sudo flatpak remote-add --if-not-exists flathub https://flathub.org/repo/flathub.flatpakrepo")
        RunCommand("sudo flatpak install -y " + strings.Join(packageNames, " "))
    default:
        log.Panic("package type ", packageType, " Isn't implemented!")
    }
}

func RunRemovePackage(packageType types.PackageType, packages []types.Package){
	var packageNames []string

	for _, pack := range packages {
		packageNames = append(packageNames, pack.Name)
	}

    switch packageType {
    case types.DEFAULT_PACKAGE:
        RunCommand("sudo dnf remove -y " +  strings.Join(packageNames, " "))
    case types.FLATPAK_PACKAGE:
        RunCommand("sudo flatpak uninstall -y " + strings.Join(packageNames, " "))
    default:
        log.Panic("package type ", packageType, " Isn't implemented!")
    }
}