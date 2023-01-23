package hfprop

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/* Characteristics (parameters):
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

// LatestDistance predicts the distance to a transceiver based on take-off angle
// and the latest hmF2 from default (JR055) Digisonde or the optional ursiCode
// Digisonde. Returns distance in kilometers as float64 or error.
func DistanceByTOA(toa float64, ursiCode ...string) (distance float64, err error) {
	digisonde := DefaultUrsiCode
	if len(ursiCode) > 0 {
		digisonde = ursiCode[0]
	}
	gd, err := GetGiroData("hmF2", digisonde, time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return 0.0, err
	}
	if len(gd) == 0 {
		return 0.0, fmt.Errorf("unable to get latest hmF2 from %s", digisonde)
	} else if gd[0].Value < 10.0 {
		return 0.0, fmt.Errorf("unable to get a valid hmF2 value from %s", digisonde)
	}
	return Distance(toa, gd[0].Value), nil
}

// LatestTOA predicts the single-hop take-off angle in degrees to a transceiver
// distance kilometers away based on latest hmF2 value from default (JR055)
// Digisonde or the optional ursiCode Digisonde. Function returns the number of
// degrees above the horizon a transmission path enters or exits the ionosphere
// as a float64 or error if something failed.
func LatestTOA(distance float64, ursiCode ...string) (degrees float64, err error) {
	digisonde := DefaultUrsiCode
	if len(ursiCode) > 0 {
		digisonde = ursiCode[0]
	}
	gd, err := GetGiroData("hmF2", digisonde, time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		return 0.0, err
	}
	if len(gd) == 0 {
		return 0.0, fmt.Errorf("unable to get latest hmF2 from %s", digisonde)
	} else if gd[0].Value < 10.0 {
		return 0.0, fmt.Errorf("unable to get a valid hmF2 value from %s", digisonde)
	}
	return TOA(distance, gd[0].Value), nil
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
