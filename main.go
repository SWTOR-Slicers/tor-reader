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
	"strconv"
	"strings"
	"time"

	"github.com/SWTOR-Slicers/tor-reader/logger"
	"github.com/SWTOR-Slicers/tor-reader/reader/hash"
	"github.com/SWTOR-Slicers/tor-reader/reader/tor"
	"github.com/gammazero/workerpool"
	"github.com/klauspost/compress/zstd"
)

var file_types map[string]string
var xml_types map[string]string

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
	fmt.Println(path)
	destination, err := os.Create(path)
	logger.Check(err)

	destination.Write(data)
	destination.Close()
}
func GetASCII(bytes []byte, offset int, length int) string {
	return string(bytes[offset : offset+length])
}

func GuessFileName(file tor.TorFile, data []byte) string {
	bytes := make([]byte, 200)
	if int(file.CompressedSize) < len(bytes) {
		bytes = make([]byte, file.CompressedSize)
	}

	bytes = data[0:len(bytes)]

	if ((bytes[0] == 0x01) && (bytes[1] == 0x00)) && (bytes[2] == 0x00) {
		return "stb"
	}
	if ((bytes[0] == 0x02) && (bytes[1] == 0x00)) && (bytes[2] == 0x00) {
		return "mph"
	}
	if ((bytes[0] == 0x21) && (bytes[1] == 0x0d)) && ((bytes[2] == 0x0a) && (bytes[3] == 0x21)) {
		str5 := GetASCII(bytes, 0, 64)
		if strings.Index(str5, "Particle Specification") >= 0 {
			return "prt"
		} else {
			return "dat"
		}
	}
	if ((bytes[0] == 0) && (bytes[1] == 1)) && (bytes[2] == 0) {
		return "ttf"
	}
	if ((bytes[0] == 10) && (bytes[1] == 5)) && ((bytes[2] == 1) && (bytes[3] == 8)) {
		return "pcx"
	}
	if ((bytes[0] == 0x38) && (bytes[1] == 0x03)) && ((bytes[2] == 0x00) && (bytes[3] == 0x00)) {
		return "spt"
	}
	if ((bytes[0] == 0x18) && (bytes[1] == 0x00)) && ((bytes[2] == 0x00) && (bytes[3] == 0x00)) {
		strCheckDAT := GetASCII(bytes, 4, 22)
		if strCheckDAT == "AREA_DAT_BINARY_FORMAT" || strCheckDAT == "ROOM_DAT_BINARY_FORMAT" {
			return "dat"
		}
	}

	str := string(bytes)
	str2 := GetASCII(bytes, 0, 4)
	for key, val := range file_types {
		if strings.Index(str2, key) >= 0 {
			if key == "RIFF" {
				if strings.Index(GetASCII(bytes, 8, 4), "WAVE") >= 0 {
					return "wav"
				}
			} else if key == "lua" {
				if strings.Index(str, "lua") > 50 {
					continue
				}
			} else if key == "\\0\\0\\0\\0" {
				if bytes[0x0b] == 0x41 {
					return "jba"
				}
			}
			return val
		}
	}

	if strings.Index(str2, "<") >= 0 {
		str4 := GetASCII(bytes, 0, 64)
		for key, val := range xml_types {
			if strings.Index(str4, key) >= 0 {
				return val
			}
		}
		return "xml"
	}

	var str6 string

	if len(bytes) < 128 {
		str6 = string(bytes)
	} else {
		str6 = GetASCII(bytes, 0, 128)
	}

	if strings.Index(str6, "[SETTINGS]") >= 0 && strings.Index(str6, "gr2") >= 0 {
		return "dyc"
	}

	if strings.Index(str, "cnv_") >= 1 && strings.Index(str, ".wem") >= 1 {
		return "acb"
	}

	length := len(strings.Split(str, ","))
	if length >= 10 {
		return "csv"
	} else {
		return "txt"
	}
}

func main() {
	file_types = make(map[string]string)
	xml_types = make(map[string]string)
	file_types["CWS"] = "swf"
	file_types["CFX"] = "gfx"
	file_types["PROT"] = "node"
	file_types["GAWB"] = "gr2"
	file_types["SCPT"] = "scpt"
	file_types["FACE"] = "fxe"
	file_types["PK"] = "zip"
	file_types["lua"] = "lua"
	file_types["DDS"] = "dds"
	file_types["XSM"] = "xsm"
	file_types["XAC"] = "xac"
	file_types["8BPS"] = "8bps"
	file_types["bdLF"] = "db"
	file_types["gsLF"] = "geom"
	file_types["idLF"] = "diffuse"
	file_types["psLF"] = "specular"
	file_types["amLF"] = "mask"
	file_types["ntLF"] = "tint"
	file_types["lgLF"] = "glow"
	file_types["Gamebry"] = "nif"
	file_types["WMPHOTO"] = "lmp"
	file_types["BKHD"] = "bnk"
	file_types["AMX"] = "amx"
	file_types["OLCB"] = "clo"
	file_types["PNG"] = "png"
	file_types["; Zo"] = "zone.txt"
	file_types["RIFF"] = "riff"
	file_types["WAVE"] = "wav"
	file_types["\\0\\0\\0\\0"] = "zero.txt"

	xml_types["<Material>"] = "mat"
	xml_types["<TextureObject"] = "tex"
	xml_types["<manifest>"] = "manifest"
	xml_types["<\\0n\\0o\\0d\\0e\\0W\\0C\\0l\\0a\\0s\\0s\\0e\\0s\\0"] = "fxspec"
	xml_types["<\\0A\\0p\\0p\\0e\\0a\\0r\\0a\\0n\\0c\\0e"] = "epp"
	xml_types["<ClothData>"] = "clo"
	xml_types["<v>"] = "not"
	xml_types["<Rules>"] = "rul"
	xml_types["<SurveyInstance>"] = "svy"
	xml_types["<DataTable>"] = "tbl"
	xml_types["<TextureObject xmlns"] = "tex"
	xml_types["<EnvironmentMaterial"] = "emt"

	var torFiles []string

	hashPath := ""
	outputDir := ""
	unnamed := false
	if len(os.Args) >= 4 {
		torFilesLoc := os.Args[1]
		torsDat, err05 := os.ReadFile(torFilesLoc)
		if err05 != nil {
			fmt.Println(err05)
		}
		err := json.Unmarshal(torsDat, &torFiles)
		if err != nil {
			fmt.Println(err)
		}
		os.Remove(torFilesLoc)

		outputDir = os.Args[2]
		hashPath = os.Args[3]
		unnamed = os.Args[4] == "true"
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
	filesWithHash := 0
	filesSuccessful := 0
	start := time.Now()

	log.Printf("using %d workerpools to instantiate server instances", runtime.NumCPU())
	for _, data := range data {
		if hashData, ok := hashes[data.FileID]; ok {
			if !unnamed {
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
				filesWithHash++
			}
		} else {
			filesNoHash++
			if unnamed {
				filesAttempted++
				data := data
				pool.Submit(func() {
					f, _ := os.Open(data.TorFile)
					defer f.Close()
					f.Seek(int64(data.Offset+uint64(data.HeaderSize)), 0)
					fileData := make([]byte, data.CompressedSize)
					f.Read(fileData)
					if data.CompressionMethod == 0 {
						writeFile(fileData, "\\"+filepath.Join(filepath.Base(data.TorFile)[0:strings.LastIndex(filepath.Base(data.TorFile), ".")], strconv.Itoa(int(data.Checksum))+"_"+strconv.Itoa(int(data.PrimaryHash))+strconv.Itoa(int(data.SecondaryHash))+"."+GuessFileName(data, fileData)), outputDir)
						filesSuccessful++
					} else {
						fileData, err := zstdDecompress(fileData)
						logger.Check(err)
						writeFile(fileData, "\\"+filepath.Join(filepath.Base(data.TorFile)[0:strings.LastIndex(filepath.Base(data.TorFile), ".")], strconv.Itoa(int(data.Checksum))+"_"+strconv.Itoa(int(data.PrimaryHash))+strconv.Itoa(int(data.SecondaryHash))+"."+GuessFileName(data, fileData)), outputDir)
						filesSuccessful++
					}
					fmt.Println(filesSuccessful, filesAttempted)
				})
			}
		}
	}
	pool.StopWait()

	diff := time.Now().Sub(start)
	log.Println("duration", fmt.Sprintf("%s", diff))

	fmt.Println(filesAttempted, filesNoHash, filesWithHash)
	log.Println(filesAttempted, filesNoHash, filesWithHash, len(data))
}
