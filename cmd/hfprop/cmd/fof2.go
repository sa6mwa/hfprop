/*
Copyright © 2024 SA6MWA Michel

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
	"time"

	"github.com/sa6mwa/hfprop"
	"github.com/spf13/cobra"
)

// fof2Cmd represents the fof2 command
var fof2Cmd = &cobra.Command{
	Use:   "fof2",
	Short: "Retrieves the foF2 from DIDBase",
	Long: `Retrieves the foF2 in kHz reported by selected digisonde from DIDBase.

Examples:
  hfprop fof2
  3225

Present foF2 in MHz instead of kHz:
  hfprop fof2 -m
  3.225

Use TR169 (Tromsö) instead of default digisonde Juliusruh (JR055):
  hfprop fof2 -u tr169
  3000

Retrieve data from yesterday up until 9 hours ago:
  hfprop fof2 -m -f yesterday -t "nine hours ago""

`,
	Run: func(cmd *cobra.Command, args []string) {
		mhz, err := cmd.Flags().GetBool("mhz")
		exitOnError(err)
		from, err := cmd.Flags().GetString("from")
		exitOnError(err)
		to, err := cmd.Flags().GetString("to")
		exitOnError(err)
		timeFrom, err := parseTime(from)
		exitOnError(err)
		timeTo, err := parseTime(to)
		exitOnError(err)

		// Adjust time from -f to be 30 minutes in the past in order to
		// have a change to get some data when specifying, e.g -f
		// yesterday -t yesterday (or even -f now -t now).
		timeFrom = timeFrom.Add(time.Duration(-30) * time.Minute)

		h := hfprop.New(ursiCode).SetLgdcBaseURL(didbURL).SetFromTime(timeFrom).SetToTime(timeTo)

		err = h.GetFoF2()
		exitOnError(err)

		_, latestFoF2, err := h.Latest("foF2")
		exitOnError(err)

		//origF2 := latestFoF2

		// Convert to kHz
		if !mhz {
			latestFoF2 = latestFoF2 * 1000.0
		}

		//fmt.Printf("foF2=%v mhz=%v from=%v to=%v ursi=%v\n", origF2, mhz, timeFrom.UTC().Format(time.RFC3339), timeTo.UTC().Format(time.RFC3339), ursiCode)
		//fmt.Println(h.GiroData)

		fmt.Println(hfprop.FormatFloat(latestFoF2))
	},
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(fof2Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fof2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fof2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	fof2Cmd.Flags().BoolP("mhz", "m", false, "Report in MHz instead of kHz")
	fof2Cmd.Flags().StringP("from", "f", "two hours ago", "Query DIDB for data from this time up until -t")
	fof2Cmd.Flags().StringP("to", "t", "now", "Query DIDB for data from time specified by -f up until this time")
}
