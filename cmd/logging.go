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
	"io"
	"os"

	goutils "github.com/l50/goutils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func configLogging() error {
	logger, err := goutils.CreateLogFile()
	if err != nil {
		log.WithError(err).Error("error creating the log file")
	}

	// Set log level
	configLogLevel := viper.GetString("log.level")
	if logLevel, err := log.ParseLevel(configLogLevel); err != nil {
		log.WithFields(log.Fields{
			"level":    logLevel,
			"fallback": "info"}).Warn("Invalid log level")
	} else {
		if debug {
			log.Debug("Debug logs enabled")
			logLevel = log.DebugLevel
		}
		log.SetLevel(logLevel)
	}

	// Set log format
	switch viper.GetString("log.format") {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     true,
			PadLevelText:    true,
		})
	}

	// Output to both stdout and the log file
	mw := io.MultiWriter(os.Stdout, logger.FilePtr)
	log.SetOutput(mw)

	return nil
}
