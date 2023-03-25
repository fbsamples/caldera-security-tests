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
	"context"
	"errors"
	"fmt"
	"math"
	"os"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// StoredXSSUnoCmd runs the XSS vulnerability found before DEF CON 30.
	StoredXSSUnoCmd = &cobra.Command{
		Use:   "StoredXSSUno",
		Short: "Stored XSS found during DEF CON 30.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.YellowString(
				"Introducing stored XSS vulnerability #1, please wait..."))

			caldera.URL = viper.GetString("login_url")
			caldera.RepoPath = viper.GetString("repo_path")
			caldera.Creds, err = GetRedCreds(caldera.RepoPath)
			if err != nil {
				log.WithError(err).Fatalf(
					"failed to get Caldera credentials: %v", err)
			}

			caldera.Driver.Headless = viper.GetBool("headless")
			driver, cancels, err := setupChrome(caldera)
			if err != nil {
				log.WithError(err).Fatal("failed to setup Chrome")
			}

			defer cancelAll(cancels)

			caldera.Driver = driver

			caldera, err = Login(caldera)
			if err != nil {
				log.WithError(err).Fatal("failed to login to caldera")
			}

			caldera.Payload = viper.GetString("payload")

			if err = storedXSSUnoVuln(caldera.Payload); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Payload": caldera.Payload,
				}).Error(color.RedString(err.Error()))
			}
		},
	}
	storedXSSUnoSuccess bool
)

func init() {
	rootCmd.AddCommand(StoredXSSUnoCmd)
	storedXSSUnoSuccess = false
}

func storedXSSUnoVuln(payload string) error {
	var buf []byte

	// Selectors for chromeDP
	pageSelector := "#nav-menu > ul:nth-child(2) > li:nth-child(4) > a"
	createOPSelector := "#select-operation > div:nth-child(3) > button"
	opNameSelector := "#op-name"
	startSelector := "#operationsPage > div > div.modal.is-active > div.modal-card > footer > nav > div.level-right > div > button"

	imagePath := viper.GetString("image_path")

	// listen network event
	listenForNetworkEvent(caldera.Driver.Context)

	// handle payload that use alerts, prompts, etc.
	chromedp.ListenTarget(caldera.Driver.Context, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				err := chromedp.Run(caldera.Driver.Context,
					page.HandleJavaScriptDialog(true))

				// If we have gotten here, the exploit succeeded.
				storedXSSUnoSuccess = true

				if err != nil {
					panic(err)
				}
			}()
		}
	})
	if err := chromedp.Run(caldera.Driver.Context,
		network.Enable(),
		chromedp.Click(pageSelector),
		chromedp.Sleep(Wait(1000)),
		chromedp.Click(createOPSelector),
		chromedp.SendKeys(opNameSelector, payload),
		chromedp.Click(startSelector),
		chromedp.Sleep(Wait(2000)),
		chromedp.ActionFunc(func(ctx context.Context) error {

			_, _, contentSize, _, _, _, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				log.WithError(err).Error("failed to get layout metrics")
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)),
				int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).Do(ctx)
			if err != nil {
				log.WithError(err).Error("failed to override device metrics")
				return err
			}

			// capture screenshot
			buf, err = page.CaptureScreenshot().
				WithQuality(100).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  2,
				}).Do(ctx)
			if err != nil {
				log.WithError(err).Error("failed to take screenshot")
				return err
			}
			return nil
		})); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"Payload": payload,
		}).Error("unexpected error while introducing the exploit")
		return err
	}

	if err := os.WriteFile(imagePath+"1.png", buf, 0644); err != nil {
		log.WithError(err).Error("failed to write screenshot to disk")
	}

	if storedXSSUnoSuccess {
		errMsg := "failure: Stored XSS Uno ran successfully"
		return errors.New(errMsg)
	}

	log.WithFields(log.Fields{
		"Payload": payload,
	}).Info(color.GreenString("Success: Stored XSS Uno failed to run"))

	return nil
}
