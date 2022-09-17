/*
Copyright © 2022-present, Meta, Inc. All rights reserved.

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
	"os"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	goutils "github.com/l50/goutils"
	"github.com/magefile/mage/sh"
	log "github.com/sirupsen/logrus"
)

// Caldera contains parameters associated
// with MITRE CALDERA.
type Caldera struct {
	Creds    Credentials
	Driver   ChromeDP
	HomeURL  string
	RepoPath string
	URL      string
	Payloads []string
}

// Credentials contains the credentials
// to access CALDERA.
type Credentials struct {
	User string
	Pass string
}

func cancelAll(cancels []func()) {
	for _, cancel := range cancels {
		cancel()
	}
}

// setChromeOptions is used to set the chrome
// parameters required by ChromeDP.
func setChromeOptions(headless bool) ChromeDP {
	chromeOpts := ChromeDP{
		Options: &[]chromedp.ExecAllocatorOption{
			chromedp.DisableGPU,
			// chromedp.IgnoreCertErrors,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoFirstRun,
			chromedp.Flag("headless", headless),
		},
	}

	return chromeOpts
}

func setupChrome(caldera Caldera) (ChromeDP, []func(), error) {
	var cancels []func()

	// Configure Chrome
	chrome := setChromeOptions(caldera.Driver.Headless)
	allocatorCtx, cancel := chromedp.NewExecAllocator(
		context.Background(), *chrome.Options...)

	cancels = append([]func(){cancel}, cancels...)

	chrome.Context, cancel = chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	cancels = append([]func(){cancel}, cancels...)

	return chrome, cancels, nil
}

// Login logs into CALDERA using Google Chrome with the input
// credentials and returns an authenticated session.
func Login(caldera Caldera) (Caldera, error) {
	// Selectors for chromeDP
	rocketSelector := "#home > div.modal.is-active > div.modal-card > footer > button"
	userSelector := "body > div > div > form > div:nth-child(1) > div > input"
	passSelector := "body > div > div > form > div:nth-child(2) > div > input"
	loginSelector := "body > div > div > form > button"

	err = chromedp.Run(caldera.Driver.Context,
		chromedp.Navigate(caldera.URL),
		chromedp.Sleep(Wait(1000)),
		chromedp.SendKeys(userSelector, caldera.Creds.User),
		chromedp.SendKeys(passSelector, caldera.Creds.Pass),
		chromedp.Sleep(Wait(1000)),
		chromedp.Click(loginSelector),
		chromedp.Sleep(Wait(2000)),
		chromedp.Click(rocketSelector),
	)

	if err != nil {
		log.WithError(err).Error("failed to login to CALDERA")
		return caldera, err
	}

	return caldera, nil

}

// GetRedCreds navigates to the input calderaPath to
// retrieve the red user credentials for MITRE CALDERA.
func GetRedCreds(calderaPath string) (Credentials, error) {
	creds := Credentials{}
	cwd := goutils.Gwd()
	found := false
	outStr := ""

	if err := os.Chdir(calderaPath); err != nil {
		log.WithFields(log.Fields{
			"Target Path": calderaPath,
		}).WithError(err).Error("failed to change directory")
		return creds, err
	}

	output, err := sh.Output("docker",
		"compose",
		"exec",
		"-T",
		"caldera",
		"cat",
		"conf/local.yml")

	if err != nil {
		log.WithFields(log.Fields{
			"Target Path": calderaPath,
		}).WithError(err).Error("failed to retrieve credentials")
		return creds, err
	}

	outSlice := goutils.StringToSlice(output, " ")
	for _, out := range outSlice {
		if found {
			outStr += out
		}
		if out == "red:" {
			found = true
			outStr = out
		}
	}
	cSlice := strings.Split(outStr, ":")
	creds.User = strings.TrimSpace(cSlice[0])
	creds.Pass = strings.TrimSpace(cSlice[1])

	if err := os.Chdir(cwd); err != nil {
		log.WithFields(log.Fields{
			"Target Path": cwd,
		}).WithError(err).Error("failed to change directory")
		return creds, err
	}

	return creds, nil
}

func listenForNetworkEvent(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventResponseReceived:
			resp := ev.Response
			if len(resp.Headers) != 0 {
				log.WithFields(log.Fields{
					"Response Headers": resp.Headers,
					"Response Status":  resp.Status,
					"Response Body":    resp.StatusText,
				}).Debug("HTTP Response Information")
			}
		default:
			return
		}
	})
}
