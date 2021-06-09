package tor

type TorFile struct {
	HeaderSize        uint32
	Offset            uint64
	CompressedSize    uint32
	UnCompressedSize  uint32
	SecondaryHash     uint32
	PrimaryHash       uint32
	FileID            uint64
	Checksum          uint32
	CompressionMethod uint16
	CRC               uint32 // Same as logger.Checksum
	TorFile           string
}
