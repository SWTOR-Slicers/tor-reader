package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
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

func writeFile(data []byte, dir string) {
	if dir == "" {
		return
	}
	path := "./output/" + dir

	strings.Split(path, "/")

	os.MkdirAll(filepath.Dir(path), os.ModePerm)
	destination, err := os.Create(path)
	check(err)

	destination.Write(data)
	destination.Close()
}

func readHashes() map[uint64]HashData {
	hash := map[uint64]HashData{}

	file, err := os.Open("hashes_filename.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		obj := strings.Split(scanner.Text(), "#")

		pH, sH, filePath, crc := obj[0], obj[1], obj[2], obj[3]

		fileID, err := strconv.ParseUint(pH+sH, 16, 64)
		check(err)
		hash[uint64(fileID)] = HashData{pH, sH, filePath, crc}
	}
	return hash

}

func main() {
	hashes := readHashes()

	/*
		buff := []byte{120, 156, 202, 72, 205, 201, 201, 215, 81, 40, 207,
			47, 202, 73, 225, 2, 4, 0, 0, 255, 255, 33, 231, 4, 147}

		a, err := zlipDecompress(buff)

		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(a))
	*/
	torName := ""

	fmt.Println(os.Args)

	if len(os.Args) >= 2 {
		torName = os.Args[1]
	}
	if torName == "" {
		return
	}
	f, err := os.Open(torName)

	defer f.Close()
	check(err)
	reader := SWTORReader{*f}
	magicNumber := reader.ReadUInt32()

	if magicNumber != 0x50594D {
		fmt.Println("Not MYP File")
	}

	f.Seek(12, 0)

	fileTableOffset := reader.ReadUInt64()

	data := make([]SWTORFile, 0, 0)

	namedFiles := 0

	for fileTableOffset != 0 {
		f.Seek(int64(fileTableOffset), 0)
		numFiles := int32(reader.ReadUInt32())
		fileTableOffset = reader.ReadUInt64()
		namedFiles += int(numFiles)
		for i := int32(0); i < numFiles; i++ {
			offset := reader.ReadUInt64()
			if offset == 0 {
				f.Seek(26, 1)
				continue
			}
			info := SWTORFile{}
			info.HeaderSize = reader.ReadUInt32()
			info.Offset = offset
			info.CompressedSize = reader.ReadUInt32()
			info.UnCompressedSize = reader.ReadUInt32()
			current_position, _ := f.Seek(0, 1)
			info.SecondaryHash = reader.ReadUInt32()
			info.PrimaryHash = reader.ReadUInt32()
			f.Seek(current_position, 0)
			info.FileID = reader.ReadUInt64()
			info.Checksum = reader.ReadUInt32()
			info.CompressionMethod = reader.ReadUInt16()
			info.CRC = info.Checksum
			data = append(data, info)
		}
	}
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
					writeFile(fileData, hashData.filename)
					filesSuccessful++
				} else {
					fileData, err := zlipDecompress(fileData)
					check(err)
					writeFile(fileData, hashData.filename)
					filesSuccessful++
				}
				fmt.Println(filesSuccessful, filesAttempted)
				//fmt.Println(1)
			})
		} else {
			filesNoHash++
		}
	}
	pool.StopWait()

	diff := time.Now().Sub(start)
	log.Println("duration", fmt.Sprintf("%s", diff))

	fmt.Println(namedFiles, filesAttempted, filesNoHash)

}
