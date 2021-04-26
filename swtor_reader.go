package main

import (
	"encoding/binary"
	"os"
)

type SWTORReader struct {
	file os.File
}

func (self SWTORReader) ReadUInt64() uint64 {
	bs := make([]byte, 8)
	_, err := self.file.Read(bs)
	check(err)

	return binary.LittleEndian.Uint64(bs)
}

func (self SWTORReader) ReadUInt16() uint16 {
	bs := make([]byte, 2)
	_, err := self.file.Read(bs)
	check(err)

	return binary.LittleEndian.Uint16(bs)
}

func (self SWTORReader) ReadUInt32() uint32 {
	bs := make([]byte, 4)
	_, err := self.file.Read(bs)
	check(err)

	return binary.LittleEndian.Uint32(bs)
}
func (self SWTORReader) ReadInt32() int32 {
	bs := make([]byte, 4)
	_, err := self.file.Read(bs)
	check(err)

	return int32(binary.LittleEndian.Uint32(bs))
}
