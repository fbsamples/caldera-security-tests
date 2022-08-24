/*
Copyright © 2022 Jayson Grace <jayson.e.grace@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/bitfield/script"
	"github.com/fatih/color"
	goutils "github.com/l50/goutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// TestEnvCmd represents the TestEnv command
	TestEnvCmd = &cobra.Command{
		Use:   "TestEnv",
		Short: "Create/Destroy test environment",
		Long: `Facilitate the creation or destruction
	of a test environment using docker compose.`,
		Run: func(cmd *cobra.Command, args []string) {
			create, _ := cmd.Flags().GetBool("create")
			destroy, _ := cmd.Flags().GetBool("destroy")
			cwd := goutils.Gwd()

			if err := goutils.Cd(caldera.RepoPath); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Repo Path": caldera.RepoPath,
				}).Error("failed to navigate to the caldera repo")
				os.Exit(1)
			}

			if create {
				if err = CreateTestEnv(); err != nil {
					log.WithError(err).Error("failed to create test environment")
					os.Exit(1)
				}
			} else if destroy {
				if err = DestroyTestEnv(); err != nil {
					log.WithError(err).Error("failed to create test environment")
					os.Exit(1)
				}

			}

			if err := goutils.Cd(cwd); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Current Working Directory": cwd,
				}).Error("failed to navigate back from the caldera repo")
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(TestEnvCmd)
	TestEnvCmd.Flags().BoolP(
		"create", "c", false, "Create the test environment.")
	TestEnvCmd.Flags().BoolP(
		"destroy", "d", false, "Destroy the test environment.")
}

// CreateTestEnv deploys an insecure version of Caldera using docker compose.
func CreateTestEnv() error {
	fmt.Println(color.YellowString(
		"Deploying Caldera container via docker compose, please wait..."))

	_, err = script.Exec("git checkout 9473dceefa4aee2ce43a88413c41247bda531ff7").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to checkout older branch")
		return err
	}

	_, err = script.Exec("docker compose up -d --force-recreate --build").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to deploy Caldera with docker compose")
		return err
	}

	return nil
}

// DestroyTestEnv destroys a Caldera deployment created using docker compose
func DestroyTestEnv() error {
	fmt.Println(color.YellowString(
		"Destroying Caldera container via docker compose, please wait..."))
	_, err := script.Exec("docker compose down -v").Stdout()
	if err != nil {
		return err
	}

	return nil
}
