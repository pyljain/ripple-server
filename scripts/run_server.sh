#!/bin/bash

# Build and run the server
cd cmd/server
go build -o server
./server