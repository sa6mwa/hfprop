package hfprop

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FastChar: https://giro.uml.edu/didbase/scaled.php
var (
	DMUF              string = DefaultDMUF
	DefaultDMUF       string = "3000"
	LgdcBaseUrl       string = "https://lgdc.uml.edu/common/DIDBGetValues"
	LgdcKeyUrsiCode   string = "ursiCode"
	LgdcKeyCharName   string = "charName" // char = characteristics
	LgdcKeyDMUF       string = "DMUF"
	LgdcKeyFromDate   string = "fromDate"
	LgdcKeyToDate     string = "toDate"
	GiroTimeFormatIn  string = "2006-01-02 15:04:05"
	GiroTimeFormatOut string = "2006-01-02T15:04:05.000Z"
	DefaultUrsiCode   string = "JR055"
	SkipVerifyTLS     bool   = false
)

// Configure distance for MUF (Maximum Usable Frequency (D)) when requesting the MUFD parameter
func SetDistanceForMUF(km float64) {
	DMUF = fmt.Sprintf("%.0f", km)
}

type GiroData struct {
	Time      time.Time
	Parameter string
	Value     float64
}

// GetGiroData retrieves data for a single characteristic (parameter) from the
// DIDB at lgdc.uml.edu between from time and to time. Returns a slice of
// GiroData objects or error if there was an error.
//
// Retrieve foF2 from Juliusruh (JR055) for the last hour. The gd slice is
// reversed meaning the latest value is the first entry in the slice.
//
//	gd, err := hfprop.GetGiroData("foF2", "JR055", time.Now().Add(-1*time.Hour), time.Now())
func GetGiroData(parameter string, ursiCode string, from time.Time, to time.Time) ([]GiroData, error) {
	gds := make([]GiroData, 0, 20)
	u, err := url.Parse(LgdcBaseUrl)
	if err != nil {
		return gds, err
	}
	values := url.Values{
		LgdcKeyUrsiCode: {ursiCode},
		LgdcKeyCharName: {parameter},
		LgdcKeyDMUF:     {DefaultDMUF},
		LgdcKeyFromDate: {from.UTC().Format(GiroTimeFormatIn)},
		LgdcKeyToDate:   {to.UTC().Format(GiroTimeFormatIn)},
	}
	u.RawQuery = values.Encode()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: SkipVerifyTLS},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 15,
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return gds, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return gds, err
	}

	// resp, err := http.Get(u.String())
	// if err != nil {
	// 	return gds, err
	// }
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return gds, err
	// }
	s := bufio.NewScanner(resp.Body)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "#") || strings.TrimSpace(s.Text()) == "" {
			continue
		}
		if strings.HasPrefix(s.Text(), "ERROR: ") {
			_, str, _ := strings.Cut(s.Text(), "ERROR: ")
			return gds, errors.New(strings.TrimSpace(str))
		}
		fields := strings.Fields(s.Text())
		if len(fields) < 3 {
			continue
		}
		gd := GiroData{}
		gd.Time, err = time.Parse(GiroTimeFormatOut, fields[0])
		if err != nil {
			return gds, err
		}
		gd.Parameter = parameter
		_, err = fmt.Sscanf(fields[2], "%f", &gd.Value)
		if err != nil {
			return gds, err
		}
		gds = append(gds, gd)
	}
	ReverseGiroData(gds)
	return gds, nil
}

// ReverseGiroData reverses a GiroData slice.
func ReverseGiroData(s []GiroData) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// ReverseStrings reverses a slice of strings.
func ReverseStrings(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
