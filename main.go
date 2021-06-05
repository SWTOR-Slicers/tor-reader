package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
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
)

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

func writeFile(data []byte, dir string) {
	if dir == "" {
		return
	}
	path := "./output/" + dir

	strings.Split(path, "/")

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	destination, err := os.Create(path)
	logger.Check(err)

	destination.Write(data)
	destination.Close()
}

func main() {
	hashes := hash.Read("hashes_filename.txt")

	torName := ""
	if len(os.Args) >= 2 {
		torName = os.Args[1]
	}
	if torName == "" {
		return
	}

	torFiles := []string{torName}
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
				f, _ := os.Open(torName)
				defer f.Close()
				f.Seek(int64(data.Offset+uint64(data.HeaderSize)), 0)
				fileData := make([]byte, data.CompressedSize)
				f.Read(fileData)
				if data.CompressionMethod == 0 {
					writeFile(fileData, hashData.Filename)
					filesSuccessful++
				} else {
					fileData, err := zlipDecompress(fileData)
					logger.Check(err)
					writeFile(fileData, hashData.Filename)
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
