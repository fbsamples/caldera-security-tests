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
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/bitfield/script"
	"github.com/chromedp/chromedp"
	goutils "github.com/l50/goutils"
	log "github.com/sirupsen/logrus"
)

// Caldera contains parameters associated
// with MITRE Caldera.
type Caldera struct {
	Creds    Credentials
	Driver   ChromeDP
	RepoPath string
	URL      string
}

// Credentials contains the credentials
// to access Caldera.
type Credentials struct {
	User string
	Pass string
}

// Login logs into Caldera using Google Chrome with the input
// credentials and returns an authenticated session.
func Login(caldera Caldera) error {
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
	)

	if err != nil {
		return err
	}

	return nil

}

// GetRedCreds navigates to the input calderaPath to
// retrieve the red user credentials for MITRE Caldera.
func GetRedCreds(calderaPath string) (Credentials, error) {
	creds := Credentials{}
	cwd := goutils.Gwd()

	if err := goutils.Cd(calderaPath); err != nil {
		log.WithFields(log.Fields{
			"Target Path": calderaPath,
		}).WithError(err).Error("failed to change directory")
		return creds, err
	}

	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	re := regexp.MustCompile("red: [a-z][A-Z]*")
	_, err := script.Exec("docker compose exec caldera cat conf/local.yml").MatchRegexp(re).Stdout()
	if err != nil {
		log.WithError(err).Error("failed to get credentials")
		return creds, err
	}

	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		log.WithFields(log.Fields{
			"Target Path": calderaPath,
		}).WithError(err).Error("failed to retrieve credentials")
		return creds, err
	}
	os.Stdout = rescueStdout

	outSlice := strings.Split(string(out), ":")

	if err := goutils.Cd(cwd); err != nil {
		log.WithFields(log.Fields{
			"Target Path": cwd,
		}).WithError(err).Error("failed to change directory")
		return creds, err
	}

	creds.User = strings.TrimSpace(outSlice[0])
	creds.Pass = strings.TrimSpace(outSlice[1])
	return creds, nil
}
