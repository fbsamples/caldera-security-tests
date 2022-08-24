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
	"fmt"
	"os"

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
			"Introducing stored XSS vulnerability, please wait..."))

		caldera.URL = viper.GetString("login_url")

		caldera.Creds, err = GetRedCreds(caldera.RepoPath)
		if err != nil {
			log.WithError(err).Errorf(
				"failed to get Caldera credentials: %v", err)
			os.Exit(1)
		}

		if err = Login(caldera); err != nil {
			log.WithError(err).Error("failed to login to caldera")
			os.Exit(1)
		}

		// options := append(chromedp.DefaultExecAllocatorOptions[:],
		// 	// Don't run chrome in headless mode
		// 	chromedp.Flag("headless", false),
		// )

		// // Create allocator context
		// allocatorCtx, cancel := chromedp.NewExecAllocator(
		// 	context.Background(), options...)
		// defer cancel()

		// // Create context
		// ctx, cancel := chromedp.NewContext(allocatorCtx)
		// defer cancel()

		fmt.Println("StoredXSSDos called")
	},
}

func init() {
	rootCmd.AddCommand(StoredXSSDosCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// StoredXSSDosCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// StoredXSSDosCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
