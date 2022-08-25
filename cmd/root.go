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
	"context"
	"embed"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/chromedp/chromedp"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfigName = "config"
	defaultConfigType = "yaml"
)

var (
	//go:embed config/*
	configContentsFs embed.FS

	cfgFile string
	debug   bool
	err     error
	caldera Caldera

	logToFile bool
	logFile   os.File

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "caldera-security-tests",
		Short: "Replicate vulnerabilities in MITRE Caldera found by Jayson Grace.",
	}
)

// ChromeDP contains parameters associated with
// running ChromeDP.
type ChromeDP struct {
	Context  context.Context
	Options  *[]chromedp.ExecAllocatorOption
	Headless bool
}

// Execute adds child commands to the root
// command and sets flags appropriately.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
	defer logFile.Close()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config", "",
		"Config file (default is config/config.yaml)")

	rootCmd.PersistentFlags().BoolVarP(
		&debug, "debug", "", false, "Show debug messages.")

	rootCmd.PersistentFlags().BoolVarP(
		&logToFile, "enableLog", "l",
		true, "Enable logging.")

	err = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	if err != nil {
		log.WithError(err).Error("failed to bind to debug in the yaml config")
	}
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// Wait is used to wait for a period
// of time.
func Wait(near float64) time.Duration {
	zoom := int(near / 10)
	x := rand.Intn(zoom) + int(0.95*near)
	return time.Duration(x) * time.Millisecond
}

func getConfigFile() ([]byte, error) {
	configFileData, err := configContentsFs.ReadFile(
		path.Join("config",
			fmt.Sprintf("%s.%s",
				defaultConfigName,
				defaultConfigType)))
	if err != nil {
		log.WithError(err).Error(
			"error reading config/ contents")
		return configFileData, err
	}

	return configFileData, nil
}

func createConfigFile(cfgPath string) error {
	err := os.MkdirAll(path.Dir(cfgPath), os.ModePerm)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Config Path": cfgPath,
		}).Error("failed to create config file")
		return err
	}

	configFileData, err := getConfigFile()
	if err != nil {
		log.WithError(err).Error("failed to read config file")
		return err
	}

	err = os.WriteFile(cfgPath, configFileData, os.ModePerm)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Config Path": cfgPath,
		}).Error("failed to write the config file")
		return err
	}

	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config yaml file in the config directory
		viper.AddConfigPath(defaultConfigName)
		viper.SetConfigType(defaultConfigType)
		viper.SetConfigName(defaultConfigName)
	}

	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		log.Info("No config file found - creating with default values")
		err = createConfigFile(
			path.Join(defaultConfigName,
				fmt.Sprintf("%s.%s", defaultConfigName, defaultConfigType)))

		if err != nil {
			log.WithError(err).Error(
				"failed to create the config file")
			os.Exit(1)
		}

		if err = viper.ReadInConfig(); err != nil {
			log.WithError(err).Error(
				"error reading config file")
			os.Exit(1)
		} else {
			log.Debug("Using config file: ",
				viper.ConfigFileUsed())
		}
	} else {
		log.Debug("Using config file: ", viper.ConfigFileUsed())
	}

	err := configLogging()
	if err != nil {
		log.WithError(err).Error("failed to set up logging")
	}
}
