package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sa6mwa/hfprop"
)

func main() {

	hfprop.SetDistanceForMUF(100.0)
	gd, err := hfprop.GetGiroData("foF2", "JR055", time.Now().Add(-1*time.Hour), time.Now())
	if err != nil {
		log.Fatal(err)
	}

	if len(gd) != 0 {
		fmt.Println(gd[0].Parameter, "=", gd[0].Value)
	}
}
