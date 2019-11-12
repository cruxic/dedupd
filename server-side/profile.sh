#!/bin/bash

#note: -run=$^ avoids running the test functions

#go test '-run=$^' -cpuprofile cpu.prof -bench=. digestmap
#go test '-run=$^' -memprofile mem.prof -bench=randFill$ digestmap
go test '-run=$^' -memprofile mem.prof -bench=randFillGoMap$ digestmap


