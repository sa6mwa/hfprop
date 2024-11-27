/*
hfprop (properties/propagation) is package to access Lowell's DIDBase and query
for ionogram-derived "characteristics" (parameters/properties) from Lowell
Digisondes (ionosondes). The package also feature functions to calculate
take-off angle, distance to a transceiver based on take-off angle, etc using
values from DIDBase.

See FastChar for more information: https://giro.uml.edu/didbase/scaled.php

There is also a CLI in the hfprop module that you can install by running:

	go install github.com/sa6mwa/hfprop/cmd/hfprop@latest

Binaries usually end-up under ~/go/bin.

Author of hfprop is amateur radio operator SA6MWA Michel Blomgren
sa6mwa@gmail.com.

The full set of characteristics in DIDBase that can be requested:

	foF2 -- F2 layer critical frequency
	foF1 -- F1 layer critical frequency
	foE -- E layer critical frequency
	foEs -- Es layer critical frequency
	fbEs -- Blanketing frequency of Es-layer
	foEa -- Critical frequency of auroral E-layer
	foP -- Critical frequency of F region patch trace
	fxI -- Maximum frequency of F trace
	MUFD -- Maximum usable frequency, 3000 km
	MD -- MUF(3000)/foF2
	hF2 -- Minimum virtual height of F2 trace
	hF -- Minimum virtual height of F trace
	hE -- Minimum virtual height of E trace
	hEs -- Minimum virtual height of Es trace
	hEa -- Minimum virtual height of auroral E trace
	hP -- Minimum virtual height of F patch trace
	TypeEs -- Type of Es layer(s)
	hmF2 -- Peak height F2-layer
	hmF1 -- Peak height F1-layer
	hmE -- Peak height of E-layer
	zhalfNm -- True height at 1/2 NmF2
	yF2 -- Half thickness of F2-layer
	yF1 -- Half thickness of F1-layer
	yE -- Half thickness of E-layer
	scaleF2 -- Scale height at the F2-peak
	B0 -- IRI thickness parameter
	B1 -- IRI profile shape parameter
	D1 -- IRI profile shape parameter
	TEC -- Ionogram-derived total electron content
	FF -- Frequence spread between fxF2 and fxI
	FE -- Frequence spread beyond foE
	QF -- Range spread of F-layer
	QE -- Range spread of E-layer
	fmin -- Minimum frequency of echoes
	fminF -- Minimum frequency of F-layer echoes
	fminE -- Minimum frequency of E-layer echoes
	fminEs -- Minimum frequency of Es-layer
	foF2p -- foF2 prediction by IRI no-storm option
*/
package hfprop

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FastChar: https://giro.uml.edu/didbase/scaled.php
var (
	DefaultDMUF          string = "3000"
	LgdcBaseURL          string = "https://lgdc.uml.edu/common/DIDBGetValues"
	LgdcKeyUrsiCode      string = "ursiCode"
	LgdcKeyCharName      string = "charName" // char = characteristics
	LgdcKeyDMUF          string = "DMUF"
	LgdcKeyFromDate      string = "fromDate"
	LgdcKeyToDate        string = "toDate"
	GiroTimeFormatIn     string = "2006-01-02 15:04:05"
	GiroTimeFormatOut    string = "2006-01-02T15:04:05.000Z"
	DefaultUrsiCode      string = "JR055"
	DefaultSkipVerifyTLS bool   = false
)

type HFProp struct {
	DMUF          string
	LgdcBaseURL   string
	UrsiCode      string
	FromTime      time.Time
	ToTime        time.Time
	GiroData      []GiroData
	SkipVerifyTLS bool
}

type GiroData struct {
	Time      time.Time
	Parameter string
	Value     float64
	instance  *HFProp
}

// New creates, configures and returns a HFProp instance. Optionally, you can
// set which Digisonde by URSI code to use by specifying it in the variadic
// digisonde. hfprop.DefaultUrsiCode will be the default Digisonde if option is
// left out. FromTime defaults one hour from invocation and ToTime is the time
// of calling New (ready to use for latest data). Use the Set* functions on
// HFProp to change the configuration in subsequent calls to the other receiver
// functions.
//
//	h := New().SetDistanceForMUF(100.0).SetWithinLastHalfHour().SetDigisonde("JR055")
//	distance, err := h.DistanceByTOA(85.0)
//	if err != nil {
//		// handle error
//	}
//	fmt.Printf("Distance to transmitter is %.1f km\n", distance)
func New(digisonde ...string) *HFProp {
	ursiCode := DefaultUrsiCode
	if len(digisonde) > 0 {
		ursiCode = digisonde[0]
	}
	return &HFProp{
		DMUF:          DefaultDMUF,
		LgdcBaseURL:   LgdcBaseURL,
		UrsiCode:      ursiCode,
		FromTime:      time.Now().Add(-1 * time.Hour),
		ToTime:        time.Now(),
		GiroData:      make([]GiroData, 0), // Initialize slice even if nil slices work
		SkipVerifyTLS: DefaultSkipVerifyTLS,
	}
}

func (h *HFProp) SetLgdcBaseURL(url string) *HFProp {
	h.LgdcBaseURL = url
	return h
}

func (h *HFProp) SetDistanceForMUF(km float64) *HFProp {
	h.DMUF = fmt.Sprintf("%.0f", km)
	return h
}

func (h *HFProp) SetWithinLastHour() *HFProp {
	return h.SetSinceUntilNow(1 * time.Hour)
}

func (h *HFProp) SetWithinLastFifteenMinutes() *HFProp {
	return h.SetSinceUntilNow(15 * time.Minute)
}

func (h *HFProp) SetWithinLastHalfHour() *HFProp {
	return h.SetSinceUntilNow(30 * time.Minute)
}

func (h *HFProp) SetSinceUntilNow(d time.Duration) *HFProp {
	h.SetFromTime(time.Now().Add(-d))
	return h.SetToTime(time.Now())
}

func (h *HFProp) SetFromTime(from time.Time) *HFProp {
	h.FromTime = from
	return h
}

// SetFromTime using string formatted as in the Giro output or input
// data, e.g:
//
//	GiroTimeFormatIn     string = "2006-01-02 15:04:05"
//	GiroTimeFormatOut    string = "2006-01-02T15:04:05.000Z"
//
// Populates the instance's FromTime field or return error on failure.
func (h *HFProp) SetFromTimeByGiroTimeFormatString(from string) error {
	t, err := time.Parse(GiroTimeFormatOut, from)
	if err != nil {
		t2, err2 := time.Parse(GiroTimeFormatIn, from)
		if err2 != nil {
			return err
		}
		h.FromTime = t2
		return nil
	}
	h.FromTime = t
	return nil
}

func (h *HFProp) SetToTime(to time.Time) *HFProp {
	h.ToTime = to
	return h
}

// SetToTime using string formatted as in the Giro output or input
// data, e.g:
//
//	GiroTimeFormatIn     string = "2006-01-02 15:04:05"
//	GiroTimeFormatOut    string = "2006-01-02T15:04:05.000Z"
//
// Populates the instance's ToTime field or return error on failure.
func (h *HFProp) SetToTimeByGiroTimeFormatString(to string) error {
	t, err := time.Parse(GiroTimeFormatOut, to)
	if err != nil {
		t2, err2 := time.Parse(GiroTimeFormatIn, to)
		if err2 != nil {
			return err
		}
		h.ToTime = t2
		return nil
	}
	h.ToTime = t
	return nil
}

func (h *HFProp) SetDigisonde(ursiCode string) *HFProp {
	h.UrsiCode = ursiCode
	return h
}
func (h *HFProp) SetURSI(ursiCode string) *HFProp {
	return h.SetDigisonde(ursiCode)
}

func (h *HFProp) GetFoF2() error {
	return h.GetGiroData("foF2")
}

func (h *HFProp) GetFoF1() error {
	return h.GetGiroData("foF1")
}

func (h *HFProp) GetFoE() error {
	return h.GetGiroData("foE")
}

func (h *HFProp) GetFxI() error {
	return h.GetGiroData("fxI")
}

// Maximum usable frequency, 3000 km
func (h *HFProp) GetMUFD() error {
	return h.GetGiroData("MUFD")
}

// MD = MUF(3000)/foF2
func (h *HFProp) GetMD() error {
	return h.GetGiroData("MD")
}

// hF2 = Minimum virtual height of F2 trace
func (h *HFProp) GethF2() error {
	return h.GetGiroData("hF2")
}

// hF1 = Minimum virtual height of F1 trace
func (h *HFProp) GethF1() error {
	return h.GetGiroData("hF1")
}

// hF = Minimum virtual height of F trace
func (h *HFProp) GethF() error {
	return h.GetGiroData("hF")
}

// hE = Minimum virtual height of E trace
func (h *HFProp) GethE() error {
	return h.GetGiroData("hE")
}

func (h *HFProp) GetHmF2() error {
	return h.GetGiroData("hmF2")
}

func (h *HFProp) GetHmF1() error {
	return h.GetGiroData("hmF1")
}

func (h *HFProp) GetHmE() error {
	return h.GetGiroData("hmE")
}

func (h *HFProp) GetFmin() error {
	return h.GetGiroData("fmin")
}

func (h *HFProp) GetFminF() error {
	return h.GetGiroData("fminF")
}

func (h *HFProp) GetFminE() error {
	return h.GetGiroData("fminE")
}

func (h *HFProp) GetFoF2p() error {
	return h.GetGiroData("foF2p")
}

// GetGiroData retrieves data for a single characteristic (parameter)
// from the DIDB at lgdc.uml.edu between from time and to
// time. Prepends a slice of GiroData objects to h.GiroData or error
// if there was an error.
//
// Retrieve foF2 from Juliusruh (JR055) for the last hour. The gd slice is
// reversed meaning the latest value is the first entry in the slice.
//
//	hfprop.SetURSI("JR055")
//	hfprop.SetFromTime(time.Now().Add(-1*time.Hour))
//	hfprop.SetToTime(time.Now())
//	err := hfprop.GetGiroData("foF2")
//
//	# Possible Characteristics (parameters):
//	foF2 -- F2 layer critical frequency
//	foF1 -- F1 layer critical frequency
//	foE -- E layer critical frequency
//	foEs -- Es layer critical frequency
//	fbEs -- Blanketing frequency of Es-layer
//	foEa -- Critical frequency of auroral E-layer
//	foP -- Critical frequency of F region patch trace
//	fxI -- Maximum frequency of F trace
//	MUFD -- Maximum usable frequency, 3000 km
//	MD -- MUF(3000)/foF2
//	hF2 -- Minimum virtual height of F2 trace
//	hF -- Minimum virtual height of F trace
//	hE -- Minimum virtual height of E trace
//	hEs -- Minimum virtual height of Es trace
//	hEa -- Minimum virtual height of auroral E trace
//	hP -- Minimum virtual height of F patch trace
//	TypeEs -- Type of Es layer(s)
//	hmF2 -- Peak height F2-layer
//	hmF1 -- Peak height F1-layer
//	hmE -- Peak height of E-layer
//	zhalfNm -- True height at 1/2 NmF2
//	yF2 -- Half thickness of F2-layer
//	yF1 -- Half thickness of F1-layer
//	yE -- Half thickness of E-layer
//	scaleF2 -- Scale height at the F2-peak
//	B0 -- IRI thickness parameter
//	B1 -- IRI profile shape parameter
//	D1 -- IRI profile shape parameter
//	TEC -- Ionogram-derived total electron content
//	FF -- Frequence spread between fxF2 and fxI
//	FE -- Frequence spread beyond foE
//	QF -- Range spread of F-layer
//	QE -- Range spread of E-layer
//	fmin -- Minimum frequency of echoes
//	fminF -- Minimum frequency of F-layer echoes
//	fminE -- Minimum frequency of E-layer echoes
//	fminEs -- Minimum frequency of Es-layer
//	foF2p -- foF2 prediction by IRI no-storm option
func (h *HFProp) GetGiroData(parameter string) error {
	gds := make([]GiroData, 0, 20)
	u, err := url.Parse(LgdcBaseURL)
	if err != nil {
		return err
	}
	values := url.Values{
		LgdcKeyUrsiCode: {h.UrsiCode},
		LgdcKeyCharName: {parameter},
		LgdcKeyDMUF:     {h.DMUF},
		LgdcKeyFromDate: {h.FromTime.UTC().Format(GiroTimeFormatIn)},
		LgdcKeyToDate:   {h.ToTime.UTC().Format(GiroTimeFormatIn)},
	}
	u.RawQuery = values.Encode()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: h.SkipVerifyTLS},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 15,
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
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
			return errors.New(strings.TrimSpace(str))
		}
		fields := strings.Fields(s.Text())
		if len(fields) < 3 {
			continue
		}
		gd := GiroData{}
		gd.Time, err = time.Parse(GiroTimeFormatOut, fields[0])
		if err != nil {
			return err
		}
		gd.Parameter = parameter
		_, err = fmt.Sscanf(fields[2], "%f", &gd.Value)
		if err != nil {
			return err
		}
		gds = append(gds, gd)
	}
	ReverseGiroData(gds)
	h.GiroData = append(gds, h.GiroData...)
	return nil
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

// Get latest value from the instance's GiroData. Returns the
// parameter, value or error in case GiroData is empty.
func (h *HFProp) Latest(parameter ...string) (parameterInGiroData string, value float64, err error) {
	if len(h.GiroData) == 0 {
		return "", 0.0, fmt.Errorf("unable to get latest data from %s", h.UrsiCode)
	}
	if len(parameter) > 0 {
		p := parameter[0]
		for _, gd := range h.GiroData {
			if gd.Parameter == p {
				return gd.Parameter, gd.Value, nil
			}
		}
		return "", 0.0, fmt.Errorf("parameter not found: %x", p)
	}
	return h.GiroData[0].Parameter, h.GiroData[0].Value, nil
}

// Distance by take-off angle based on latest hmF2 value. GiroData in
// h must have an entry where parameter is hmF2 or this will fail.
func (h *HFProp) DistanceByTOA(toa float64) (distance float64, err error) {
	for _, gd := range h.GiroData {
		if gd.Parameter == "hmF2" {
			if gd.Value < 10.0 {
				return 0.0, fmt.Errorf("unable to get valid hmF2 value from %s", h.UrsiCode)
			}
			return Distance(toa, gd.Value), nil
		}
	}
	return 0.0, errors.New("no hmF2 parameter and value in instance")
}

// LatestTOA predicts the single-hop take-off angle in degrees to a
// transceiver distance kilometers away based on latest hmF2 value
// from h's Digisonde (UrsiCode). Function returns the number of
// degrees above the horizon a transmission path enters or exits the
// ionosphere as a float64 or error if something failed.
func (h *HFProp) LatestTOA(distance float64) (degrees float64, err error) {
	for _, gd := range h.GiroData {
		if gd.Parameter == "hmF2" && gd.Value > 10.0 {
			return TOA(distance, gd.Value), nil
		}
	}
	return 0.0, fmt.Errorf("unable to get latest hmF2 from %s", h.UrsiCode)
}

// Distance predicts the distance to a transceiver based on single-hop take-off
// angle in degrees above the horizon and peak height of the F2 layer (hmF2).
// Function returns the distance in kilometers as a float64.
func Distance(toa, hmf2 float64) (distance float64) {
	maxDistance := math.Round(40000 / 2 / math.Pi)
	for distance = 1.0; distance < maxDistance; distance++ {
		calculatedTOA := TOA(distance, hmf2)
		if calculatedTOA < toa {
			return distance - 1.0
		}
	}
	// close enough
	return
}

// TOA predicts the single-hop take-off angle in degrees to a transceiver
// distance kilometers away based on specified hmF2 value (peak height of the F2
// layer). Function returns the number of degrees above the horizon a
// transmission path enters or exits the ionosphere as a float64.
func TOA(distance float64, hmf2 float64) (degrees float64) {
	earthRadius := 40000 / 2 / math.Pi
	earthAngleA := distance / 40000 * (2 * math.Pi)
	horizontal := earthRadius * math.Sin(earthAngleA/2)
	tangentvalue := (math.Pi - earthAngleA/2) / 2
	vertical := horizontal / (math.Sin(tangentvalue) / math.Cos(tangentvalue))
	takeOffAngle := math.Atan((vertical+hmf2)/horizontal) - earthAngleA/2
	degrees = takeOffAngle / math.Pi * 180
	return
}

// FormatFloat return floating pointer num as a string with just the
// right amount of decimals, without trailing zeroes.
func FormatFloat(num float64) string {
	if num == float64(int(num)) {
		return fmt.Sprintf("%d", int(num))
	} else {
		return strconv.FormatFloat(num, 'f', -1, 64)
	}
}
