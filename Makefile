build:
	cd cmd/gsvnc && go build -o ../../dist/gsvnc .

ARGS ?= 
run: build
	dist/gsvnc $(ARGS)