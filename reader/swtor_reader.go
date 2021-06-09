package reader

import (
	"encoding/binary"
	"os"

	"github.com/SWTOR-Slicers/tor-reader/logger"
)

type SWTORReader struct {
	File *os.File
}

func (self SWTORReader) ReadUInt64() uint64 {
	bs := make([]byte, 8)
	_, err := self.File.Read(bs)
	logger.Check(err)

	return binary.LittleEndian.Uint64(bs)
}

func (self SWTORReader) ReadUInt16() uint16 {
	bs := make([]byte, 2)
	_, err := self.File.Read(bs)
	logger.Check(err)

	return binary.LittleEndian.Uint16(bs)
}

func (self SWTORReader) ReadUInt32() uint32 {
	bs := make([]byte, 4)
	_, err := self.File.Read(bs)
	logger.Check(err)

	return binary.LittleEndian.Uint32(bs)
}
func (self SWTORReader) ReadInt32() int32 {
	bs := make([]byte, 4)
	_, err := self.File.Read(bs)
	logger.Check(err)

	return int32(binary.LittleEndian.Uint32(bs))
}
