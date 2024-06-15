package tor_reader

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

func zlibDecompress(buff []byte) ([]byte, error) {
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

type ExtractionResults struct {
	FilesAttempted  int
	FilesSuccessful int
	FilesNoHash     int
	TimeTaken       time.Duration
}

func StartExtractor(torFiles []string, outputDir string, hashPath string, workers int) *ExtractionResults {
	hashes := hash.Read(hashPath)

	if len(torFiles) == 1 {
		torName := torFiles[0]
		torFiles = []string{}

		f, _ := os.Open(torName)
		fi, _ := f.Stat()

		switch mode := fi.Mode(); {
		case mode.IsDir():
			files, _ := os.ReadDir(torName)

			for _, f := range files {
				file := filepath.Join(torName, f.Name())

				fileMode := f.Type()

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

	pool := workerpool.New(workers)

	filesNoHash := 0

	filesAttempted := 0
	filesSuccessful := 0
	start := time.Now()

	//log.Printf("using %d workerpools to instantiate server instances", runtime.NumCPU())
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
				if data.CompressionMethod == 1 {
					filesSuccessful++
					if data.Version >= 5 {
						newFileData, err := zstdDecompress(fileData)
						logger.Check(err)
						fileData = newFileData
					} else {
						newFileData, err := zlibDecompress(fileData)
						logger.Check(err)
						fileData = newFileData
					}
				}
				writeFile(fileData, hashData.Filename, outputDir)
				filesSuccessful++
			})
		} else {
			filesNoHash++
		}
	}
	pool.StopWait()

	diff := time.Since(start)

	return &ExtractionResults{
		FilesAttempted:  filesAttempted,
		FilesSuccessful: filesSuccessful,
		FilesNoHash:     filesNoHash,
		TimeTaken:       diff,
	}
}

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

	results := StartExtractor(torFiles, outputDir, hashPath, runtime.NumCPU())

	diff := results.TimeTaken
	log.Println("duration", diff.String())
	log.Println(results.FilesAttempted, results.FilesNoHash)

}
