/*
Copyright © 2022-present, Meta Platforms, Inc. and affiliates

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
	"path/filepath"

	"github.com/bitfield/script"
	"github.com/fatih/color"
	goutils "github.com/l50/goutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// TestEnvCmd represents the TestEnv command
	TestEnvCmd = &cobra.Command{
		Use:   "TestEnv",
		Short: "Create/Destroy test environment",
		Long: `Facilitate the creation or destruction
	of a test environment using docker compose.`,
		Run: func(cmd *cobra.Command, args []string) {
			vuln, _ := cmd.Flags().GetBool("vuln")
			recent, _ := cmd.Flags().GetBool("recent")
			destroy, _ := cmd.Flags().GetBool("destroy")
			cwd := goutils.Gwd()

			caldera.RepoPath = viper.GetString("repo_path")
			if err := goutils.Cd(caldera.RepoPath); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Repo Path": caldera.RepoPath,
				}).Error("failed to navigate to the caldera repo")
				os.Exit(1)
			}

			if vuln {
				if err = CreateTestEnvVuln(); err != nil {
					log.WithError(err).Error("failed to create vulnerable test environment")
					os.Exit(1)
				}
			} else if destroy {
				if err = DestroyTestEnv(); err != nil {
					log.WithError(err).Error("failed to destroy test environment")
					os.Exit(1)
				}
			} else if recent {
				if err = CreateTestEnvRecent(); err != nil {
					log.WithError(err).Error("failed to create recent test environment")
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
		"vuln", "v", false, "Create vulnerable test environment.")
	TestEnvCmd.Flags().BoolP(
		"recent", "r", false, "Create test environment with the most "+
			"recent commit to the CALDERA's default branch.")
	TestEnvCmd.Flags().BoolP(
		"destroy", "d", false, "Destroy the test environment.")
}

// CreateTestEnvVuln deploys an insecure version of Caldera using docker compose.
func CreateTestEnvVuln() error {
	fmt.Println(color.YellowString(
		"Deploying Caldera container via docker compose, please wait..."))

	_, err = script.Exec("git checkout 9473dceefa4aee2ce43a88413c41247bda531ff7").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to checkout older branch")
		return err
	}

	if err := goutils.Cd(filepath.Join("plugins", "debrief")); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Repo Path": caldera.RepoPath,
		}).Error("failed to navigate to the caldera repo")
	}

	_, err = script.Exec("git checkout 7ea5d726538a27bdc33613b1c23d822f73935c6f").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to checkout older debrief plugin branch")
		return err
	}

	if err := goutils.Cd("../../"); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Repo Path": caldera.RepoPath,
		}).Error("failed to navigate to the caldera repo")
	}

	_, err = script.Exec("docker compose up -d --force-recreate --build").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to deploy Caldera with docker compose")
		return err
	}

	return nil
}

// CreateTestEnvRecent deploys the most recent version of Caldera using docker compose.
func CreateTestEnvRecent() error {
	fmt.Println(color.YellowString(
		"Deploying CALDERA container via docker compose, please wait..."))

	_, err = script.Exec("git checkout master").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to checkout master branch")
		return err
	}

	_, err = script.Exec("docker compose up -d --force-recreate --build").Stdout()
	if err != nil {
		log.WithError(err).Error("failed to deploy CALDERA with docker compose")
		return err
	}

	return nil
}

// DestroyTestEnv destroys a CALDERA deployment created using docker compose
func DestroyTestEnv() error {
	fmt.Println(color.YellowString(
		"Destroying CALDERA container via docker compose, please wait..."))
	_, err := script.Exec("docker compose down -v").Stdout()
	if err != nil {
		return err
	}

	return nil
}
