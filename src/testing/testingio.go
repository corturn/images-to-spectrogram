package main

import (
	"os"
	spectro "proj3-redesigned/spectro2"
	"strconv"
)

func main() {
	args := os.Args
	switch len(args) {
	case 3:
		// Assumes you're running sequential if 3 args
		spectro.ImgToWavSeq(args[1], true)
	case 4:
		numThreads, err := strconv.Atoi(args[3])

		if err != nil {
			break
		}

		if args[2] == "map" {
			spectro.MapReduce(args[1], numThreads, true)
		} else {
			spectro.WorkSteal(args[1], numThreads, true)
		}
	default:
		println("Usage: go run testingio.go file_path.json {seq}/{map/steal int(num threads)}")
	}
}
