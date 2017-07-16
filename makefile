# go-between makefile
# Copyright (C) 2017 Daniel Wilson
# MIT License - See LICENSE.md
# https://github.com/Danw33/go-between

# Default: Make the "build" version without debug symbols
all: build

# Build without debug symbols (Smaller output executable)
build:
	go build -ldflags "-s -w" ./src/go-between.go
	if [ -a ./go-between ]; then chmod +X ./go-between; fi;

# Debug build with debug symbols (Larger output executable)
debug:
	go build ./src/go-between.go
	if [ -a ./go-between ]; then chmod +X ./go-between; fi;

# Install the build (with systemd service if the host OS uses systemd)
install:
	cp ./go-between /usr/local/bin/go-between
	if [ -a /lib/systemd/system/ ]; then cp ./go-between.service /lib/systemd/system/go-between.service; fi;

# Clean the working directory of any existing builds
clean:
	if [ -a ./go-between ]; then rm ./go-between; fi;
