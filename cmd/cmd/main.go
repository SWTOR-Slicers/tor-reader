package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	tor_reader "github.com/SWTOR-Slicers/tor-reader"
)

func main() {
	var torFiles []string

	hashPath := ""
	outputDir := ""
	if len(os.Args) >= 4 {
		fmt.Println(os.Args[1])
		// err := json.Unmarshal([]byte(os.Args[1]), &torFiles)
		// if err != nil {
		// 	fmt.Println(err)
		// }

		torFiles = strings.Split(os.Args[1], ",")
		fmt.Println(torFiles)
		outputDir = os.Args[2]
		hashPath = os.Args[3]
	} else {
		fmt.Println("Usage: tor-reader <torFiles> <outputDir> <hashPath>")
	}

	if len(torFiles) == 0 || outputDir == "" || hashPath == "" {
		return
	}

	log.Printf("using %d workerpools to instantiate server instances", runtime.NumCPU())

	results := tor_reader.StartExtractor(torFiles, outputDir, hashPath, runtime.NumCPU())

	diff := results.TimeTaken
	log.Println("duration", diff.String())
	log.Println(results.FilesAttempted, results.FilesNoHash)

}
