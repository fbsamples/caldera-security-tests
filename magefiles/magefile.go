//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	goutils "github.com/l50/goutils"

	// mage utility functions
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var (
	err error
)

func init() {
	os.Setenv("GO111MODULE", "on")
}

// Compile Compiles caldera-security-tests for the input operating system.
//
// # Example:
// ```
// ./magefile compile darwin
// ```
//
// If an operating system is not input, binaries will be created for
// windows, linux, and darwin.
func Compile(ctx context.Context, osCli string) error {
	var operatingSystems []string
	binName := "cst"
	binDir := "bin"
	supportedOS := []string{"windows", "linux", "darwin"}

	if osCli == "all" {
		operatingSystems = supportedOS
	} else if goutils.StringInSlice(osCli, supportedOS) {
		operatingSystems = []string{osCli}
	}

	// Create bin/ if it doesn't already exist
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		if err := os.Mkdir(binDir, os.ModePerm); err != nil {
			return fmt.Errorf(color.RedString(
				"failed to create bin dir: %v", err))
		}
	}

	for _, os := range operatingSystems {
		fmt.Printf(color.YellowString(
			"Compiling caldera-security-tests bin "+
				"for %s OS, please wait.\n", os))
		env := map[string]string{
			"GOOS":   os,
			"GOARCH": "amd64",
		}

		binPath := filepath.Join(binDir, fmt.Sprintf("%s-%s", binName, os))

		if err := sh.RunWith(env, "go", "build", "-o", binPath); err != nil {
			return fmt.Errorf(color.RedString(
				"failed to create %s bin: %v", binPath, err))
		}
	}

	return nil
}

// InstallDeps Installs go dependencies
func InstallDeps() error {
	fmt.Println(color.YellowString("Installing dependencies."))

	if err := goutils.Tidy(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install dependencies: %v", err))
	}

	if err := goutils.InstallGoPCDeps(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install pre-commit dependencies: %v", err))
	}

	if err := goutils.InstallVSCodeModules(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install vscode-go modules: %v", err))
	}

	return nil
}

// RunPreCommit runs all pre-commit hooks locally
func RunPreCommit() error {
	mg.Deps(InstallDeps)
	mg.Deps(goutils.InstallPCHooks)
	mg.Deps(goutils.UpdatePCHooks)
	mg.Deps(goutils.ClearPCCache)

	fmt.Println(color.YellowString("Running all pre-commit hooks locally."))
	mg.Deps(goutils.RunPCHooks)

	return nil
}
