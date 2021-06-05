package hash

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/SWTOR-Slicers/tor-reader/logger"
)

type HashData struct {
	PH       string
	SH       string
	Filename string
	CRC      string
}

func Read(filePath string) map[uint64]HashData {
	hash := map[uint64]HashData{}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		obj := strings.Split(scanner.Text(), "#")

		pH, sH, filePath, crc := obj[0], obj[1], obj[2], obj[3]

		fileID, err := strconv.ParseUint(pH+sH, 16, 64)
		logger.Check(err)
		hash[uint64(fileID)] = HashData{pH, sH, filePath, crc}
	}
	return hash
}
