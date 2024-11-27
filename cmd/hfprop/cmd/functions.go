package cmd

import (
	"strings"
	"time"

	"github.com/jinzhu/now"
)

func parseTime(input string) (time.Time, error) {
	n := time.Now()
	words := strings.Split(input, " ")

	switch {
	case len(words) == 1:
		switch strings.ToLower(words[0]) {
		case "now":
			return n, nil
		case "yesterday":
			return n.AddDate(0, 0, -1), nil
		case "tomorrow":
			return n.AddDate(0, 0, 1), nil
		default:
			return now.Parse(input)
		}
	case len(words) == 2:
		switch strings.ToLower(input) {
		case "last week", "previous week":
			return n.AddDate(0, 0, -7), nil
		case "next week", "coming week", "upcoming week":
			return n.AddDate(0, 0, 7), nil
		case "second ago":
			return n.Add(time.Duration(-1) * time.Second), nil
		case "minute ago":
			return n.Add(time.Duration(-1) * time.Minute), nil
		case "hour ago":
			return n.Add(time.Duration(-1) * time.Hour), nil
		case "day ago":
			return n.AddDate(0, 0, -1), nil
		default:
			return now.Parse(input)
		}
	case len(words) == 3:
	default:
		return now.Parse(input)
	}

	switch strings.ToLower(words[2]) {
	case "ago":
	default:
		return now.Parse(input)
	}

	quantity := 0
	switch strings.ToLower(words[0]) {
	case "one", "1":
		quantity = 1
	case "two", "2":
		quantity = 2
	case "three", "3":
		quantity = 3
	case "four", "4":
		quantity = 4
	case "five", "5":
		quantity = 5
	case "six", "6":
		quantity = 6
	case "seven", "7":
		quantity = 7
	case "eight", "8":
		quantity = 8
	case "nine", "9":
		quantity = 9
	case "ten", "10":
		quantity = 10
	case "eleven", "11":
		quantity = 11
	case "twelve", "12":
		quantity = 12
	case "thirteen", "13":
		quantity = 13
	case "fourteen", "14":
		quantity = 14
	case "fifteen", "15":
		quantity = 15
	case "sixteen", "16":
		quantity = 16
	case "eighteen", "18":
		quantity = 18
	case "nineteen", "19":
		quantity = 19
	case "twenty", "20":
		quantity = 20
	case "twenty-one", "21":
		quantity = 21
	case "twenty-two", "22":
		quantity = 22
	case "twenty-three", "23":
		quantity = 23
	case "twenty-four", "24":
		quantity = 24
	default:
		return now.Parse(input)
	}

	unit := words[1]
	switch unit {
	case "second", "seconds":
		return n.Add(time.Duration(-quantity) * time.Second), nil
	case "minute", "minutes":
		return n.Add(time.Duration(-quantity) * time.Minute), nil
	case "hour", "hours":
		return n.Add(time.Duration(-quantity) * time.Hour), nil
	case "day", "days":
		return n.AddDate(0, 0, -quantity), nil
	default:
		return now.Parse(input)
	}
}
