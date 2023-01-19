package main

import (
	"time"

	"github.com/sa6mwa/hfprop"
)

func main() {

	hfprop.GetGiroData("foF2", "JR055", time.Now().Add(-1*time.Hour), time.Now())

}
