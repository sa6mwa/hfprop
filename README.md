# hfprop
HF propagation API and CLI utilizing ionosonde data from LGDC - [Lowell's Global Ionosphere Radio Observatory "GIRO" Data Center](https://digisonde.com/digisonde.html#global-section).

## Installation

If you have [Go](https://go.dev) you can build and install the CLI with the
oneliner below, otherwise you find binaries in the tarball in each release...

```console
go install github.com/sa6mwa/hfprop/cmd/hfprop@latest
```

## Usage

```console
# Get latest foF2 for the default Digisonde (Juliusruh, URSI code JR055):
$ hfprop fof2
3225

# Return in MHz instead of kHz:
$ hfprop fof2 -m
3.225

# Get foF2 from URSI-code TR169 instead (Tromsö):
$ hfprop fof2 -u tr169
3000

# Get the height of peak density of the F2 layer in km (default ionosonde):
$ hfprop hmf2
267.4

# Calculate the skywave Take Off Angle in degrees between two stations
# 100 km apart using latest hmF2 from the default Digisonde (Juliusruh):
$ hfprop toa 100
82.45

# Approximate the distance to a station when the receive/transmit
# take off angle is known. Will, in this example, derive current
# hmF2 from the default Digisonde (JR055, Juliusruh):
$ hfprop distance 82.45
100.0

```


https://lgdc.uml.edu/common/DIDBGetValues?ursiCode=JR055&charName=foF2,hF2,hmF2,yF2&DMUF=3000&fromDate=2023%2F01%2F18+19%3A00%3A00&toDate=2023%2F01%2F18+20%3A00%3A00
