package hfprop

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	LgdcBaseUrl       string = "https://lgdc.uml.edu/common/DIDBGetValues"
	LgdcKeyUrsiCode   string = "ursiCode"
	LgdcKeyCharName   string = "charName"
	LgdcKeyDMUF       string = "DMUF"
	LgdcKeyFromDate   string = "fromDate"
	LgdcKeyToDate     string = "toDate"
	DefaultUrsiCode   string = "JR055"
	DefaultDMUF       string = "3000"
	GiroTimeFormatIn  string = "2006-01-02 15:04:05"
	GiroTimeFormatOut string = "2006-01-02T15:04:05.000Z"
)

func GetGiroData(parameter string, ursiCode string, from time.Time, to time.Time) error {

	// 2023-01-18T19:08:16.000Z

	u, err := url.Parse(LgdcBaseUrl)
	if err != nil {
		return err
	}

	values := url.Values{
		LgdcKeyUrsiCode: {ursiCode},
		LgdcKeyCharName: {parameter},
		LgdcKeyDMUF:     {DefaultDMUF},
		LgdcKeyFromDate: {from.UTC().Format(GiroTimeFormatIn)},
		LgdcKeyToDate:   {to.UTC().Format(GiroTimeFormatIn)},
	}

	u.RawQuery = values.Encode()

	fmt.Println(u.String())

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Print(string(body))
	return nil
}
