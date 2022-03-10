#! /bin/bash

# rm -rf ./start
go build .
GIN_MODE=release ./start -TEST=false -ENV=local