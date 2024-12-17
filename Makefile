.PHONY: clean build install

.EXPORT_ALL_VARIABLES:

VERSION = v0.0.1
UPXLVL = -9

all: bin/hfprop

clean:
	rm -rf bin

build: bin/hfprop

install:
	sudo install bin/hfprop /usr/local/bin/hfprop

bin/hfprop: bin
	go build -o bin/hfprop -trimpath -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/hfprop
	strip -s bin/hfprop
	#if which upx > /dev/null ; then upx $(UPXLVL) bin/bootstrap ; fi

bin:
	mkdir bin
