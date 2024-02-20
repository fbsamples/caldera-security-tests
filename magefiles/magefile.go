//go:build mage

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/l50/goutils/v2/dev/lint"
	mageutils "github.com/l50/goutils/v2/dev/mage"
	"github.com/l50/goutils/v2/docs"
	"github.com/l50/goutils/v2/git"
	"github.com/l50/goutils/v2/str"
	"github.com/l50/goutils/v2/sys"
	"github.com/spf13/afero"

	// mage utility functions

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
	} else if str.InSlice(osCli, supportedOS) {
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

// InstallDeps installs the Go dependencies necessary for developing
// on the project.
//
// Example usage:
//
// ```go
// mage installdeps
// ```
//
// **Returns:**
//
// error: An error if any issue occurs while trying to
// install the dependencies.
func InstallDeps() error {
	fmt.Println(color.YellowString("Running go mod tidy on magefiles and repo root."))
	cwd := sys.Gwd()
	if err := sys.Cd("magefiles"); err != nil {
		return fmt.Errorf("failed to cd into magefiles directory: %v", err)
	}

	if err := mageutils.Tidy(); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	if err := sys.Cd(cwd); err != nil {
		return fmt.Errorf("failed to cd back into repo root: %v", err)
	}

	if err := mageutils.Tidy(); err != nil {
		return fmt.Errorf("failed to install dependencies: %v", err)
	}

	fmt.Println(color.YellowString("Installing dependencies."))
	if err := lint.InstallGoPCDeps(); err != nil {
		return fmt.Errorf("failed to install pre-commit dependencies: %v", err)
	}

	if err := mageutils.InstallVSCodeModules(); err != nil {
		return fmt.Errorf(color.RedString(
			"failed to install vscode-go modules: %v", err))
	}

	return nil
}

// RunPreCommit updates, clears, and executes all pre-commit hooks
// locally. The function follows a three-step process:
//
// First, it updates the pre-commit hooks.
// Next, it clears the pre-commit cache to ensure a clean environment.
// Lastly, it executes all pre-commit hooks locally.
//
// Example usage:
//
// ```go
// mage runprecommit
// ```
//
// **Returns:**
//
// error: An error if any issue occurs at any of the three stages
// of the process.
func RunPreCommit() error {
	if !sys.CmdExists("pre-commit") {
		return fmt.Errorf("pre-commit is not installed")
	}

	fmt.Println(color.YellowString("Updating pre-commit hooks."))
	if err := lint.UpdatePCHooks(); err != nil {
		return err
	}

	fmt.Println(color.YellowString("Clearing the pre-commit cache to ensure we have a fresh start."))
	if err := lint.ClearPCCache(); err != nil {
		return err
	}

	fmt.Println(color.YellowString("Running all pre-commit hooks locally."))
	if err := lint.RunPCHooks(); err != nil {
		return err
	}

	return nil
}

// GeneratePackageDocs creates documentation for the various packages
// in the project.
//
// Example usage:
//
// ```go
// mage generatepackagedocs
// ```
//
// **Returns:**
//
// error: An error if any issue occurs during documentation generation.
func GeneratePackageDocs() error {
	fs := afero.NewOsFs()

	repoRoot, err := git.RepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repo root: %v", err)
	}
	sys.Cd(repoRoot)

	repo := docs.Repo{
		Owner: "l50",
		Name:  "goutils/v2",
	}

	templatePath := filepath.Join("magefiles", "tmpl", "README.md.tmpl")
	if err := docs.CreatePackageDocs(fs, repo, templatePath); err != nil {
		return fmt.Errorf("failed to create package docs: %v", err)
	}

	return nil
}
