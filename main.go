package main

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/SWTOR-Slicers/tor-reader/logger"
	"github.com/SWTOR-Slicers/tor-reader/reader/hash"
	"github.com/SWTOR-Slicers/tor-reader/reader/tor"
	"github.com/gammazero/workerpool"
	"github.com/klauspost/compress/zstd"
)

func zstdDecompress(buff []byte) ([]byte, error) {
	r, err := zstd.NewReader(bytes.NewReader(buff))
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	defer r.Close()

	var out bytes.Buffer
	io.Copy(&out, r)

	return out.Bytes(), nil
}

func zlipDecompress(buff []byte) ([]byte, error) {
	b := bytes.NewReader(buff)
	r, err := zlib.NewReader(b)

	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	var out bytes.Buffer
	io.Copy(&out, r)

	return out.Bytes(), nil
}

func writeFile(data []byte, dir string, outputDir string) {
	if dir == "" {
		return
	}
	path := outputDir + dir

	strings.Split(path, "/")

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	destination, err := os.Create(path)
	logger.Check(err)

	destination.Write(data)
	destination.Close()
}

func main() {
	var torFiles []string

	hashPath := ""
	outputDir := ""
	if len(os.Args) >= 4 {
		fmt.Println(os.Args[1])
		err := json.Unmarshal([]byte(os.Args[1]), &torFiles)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(torFiles)
		outputDir = os.Args[2]
		hashPath = os.Args[3]
	}
	if len(torFiles) == 0 || outputDir == "" || hashPath == "" {
		return
	}

	hashes := hash.Read(hashPath)

	if len(torFiles) == 1 {
		torName := torFiles[0]
		torFiles = []string{}

		f, _ := os.Open(torName)
		fi, _ := f.Stat()

		switch mode := fi.Mode(); {
		case mode.IsDir():
			files, _ := ioutil.ReadDir(torName)

			for _, f := range files {
				file := filepath.Join(torName, f.Name())

				fileMode := f.Mode()

				if fileMode.IsRegular() {
					if filepath.Ext(file) == ".tor" {
						torFiles = append(torFiles, file)
					}
				}
			}
		case mode.IsRegular():
			torFiles = append(torFiles, torName)
		}
	}

	data := tor.ReadAll(torFiles)

	pool := workerpool.New(runtime.NumCPU())

	filesNoHash := 0

	filesAttempted := 0
	filesSuccessful := 0
	start := time.Now()

	log.Printf("using %d workerpools to instantiate server instances", runtime.NumCPU())
	for _, data := range data {
		if hashData, ok := hashes[data.FileID]; ok {
			filesAttempted++
			hashData := hashData
			data := data
			pool.Submit(func() {
				f, _ := os.Open(data.TorFile)
				defer f.Close()
				f.Seek(int64(data.Offset+uint64(data.HeaderSize)), 0)
				fileData := make([]byte, data.CompressedSize)
				f.Read(fileData)
				if data.CompressionMethod == 0 {
					writeFile(fileData, hashData.Filename, outputDir)
					filesSuccessful++
				} else {
					fileData, err := zstdDecompress(fileData)
					logger.Check(err)
					writeFile(fileData, hashData.Filename, outputDir)
					filesSuccessful++
				}
				fmt.Println(filesSuccessful, filesAttempted)
			})
		} else {
			filesNoHash++
		}
	}
	pool.StopWait()

	diff := time.Now().Sub(start)
	log.Println("duration", fmt.Sprintf("%s", diff))

	fmt.Println(filesAttempted, filesNoHash)
}
