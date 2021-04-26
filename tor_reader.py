import os
from struct import unpack
import json
import zlib


class Reader():

	def __init__(self, file):
		self.file = file

	# DEPRECATED FUNCTIONS
	def ruint8(self):  # Function to read unsigned byte
	    return unpack(b'B', self.file.read(1))[0]

	def rfloat8(self):  # Function to read a float that's been encoded as a single byte
	    return float((unpack(b'B', self.file.read(1))[0] - 127.5) / 127.5)

	def ruint16(self):  # Function to read unsigned int16
	    return unpack(b'<H', self.file.read(2))[0]

	def rfloat16(self):  # Function to read float16
	    return unpack(b'<e', self.file.read(2))[0]

	def rint32(self):  # Function to read signed int32
	    return unpack(b'<i', self.file.read(4))[0]

	def ruint32(self):  # Function to read unsigned int32
	    return unpack(b'<I', self.file.read(4))[0]


	def rfloat32(self):  # Function to read float32
	    return unpack(b'<f', self.file.read(4))[0]

	def rstring(self):
		file = self.file
		offset = file.tell()

		file.seek(unpack(b'<I', file.read(4))[0])

		string = ""
		byte = file.read(1)

		while byte != b'\x00':
			string += byte.decode('utf-8')
			byte = file.read(1)

		file.seek(offset + 4)

		return string

	def ReadInt32(self):
		return self.rint32()
	def ReadUInt32(self):
		return self.ruint32()
	def ReadUInt64(self):
		return unpack(b'<Q', self.file.read(8))[0]
	def ReadUInt16(self):
		return self.ruint16()

class FileInfo():
	HeaderSize = 0
	Offset = 0
	CompressedSize = 0
	UnCompressedSize = 0
	SecondaryHash = 0
	PrimaryHash = 0
	FileID = 0
	Checksum = 0
	CompressionMethod = 0
	CRC = 0

	def __str__(self):
		return json.dumps({
			"HeaderSize": self.HeaderSize,
			"Offset": self.Offset,
			"CompressedSize": self.CompressedSize,
			"UnCompressedSize": self.UnCompressedSize,
			"SecondaryHash": self.SecondaryHash,
			"PrimaryHash": self.PrimaryHash,
			"FileID": self.FileID,
			"Checksum": self.Checksum,
			"CompressionMethod": self.CompressionMethod,
			"CRC": self.CRC
		})


def writeFile(data, foundName, filename):
	os.makedirs(os.path.dirname("./output/" + filename), exist_ok=True)
	
	with open("./output/" + filename, "wb") as f:
		f.write(data)


hashes = {}

with open("hashes_filename.txt", "r") as f:
	
	for line in f:
		obj = line.strip().split("#")
		pH, sH, filePath, crc = obj[0], obj[1], obj[2], obj[3]

		hashes[int(pH + sH, 16)] =	{
			"filename": filePath,
			"crc": crc,
		}

with open("swtor_main_art_dynamic_chest_1.tor", "rb") as f:

	reader = Reader(f)
	magicNumber = reader.ReadInt32()
	if (magicNumber != 0x50594D):
		raise "Not an MYP File"
	
	f.seek(12)

	fileTableOffset = reader.ReadUInt64()

	data = []
	while fileTableOffset != 0:
		f.seek(fileTableOffset)

		numFiles = reader.ReadUInt32()
		fileTableOffset = reader.ReadUInt64()
		for i in range(0, numFiles):
			info = FileInfo()
			info.Offset = reader.ReadUInt64()
			if (info.Offset == 0):
				
				f.seek(26, 1)
				continue

			info.HeaderSize = reader.ReadUInt32()
			info.CompressedSize = reader.ReadUInt32()
			info.UnCompressedSize = reader.ReadUInt32()
			current_position = f.tell()
			info.SecondaryHash = reader.ReadUInt32() # Legacy ids for writing formats
			info.PrimaryHash = reader.ReadUInt32() # Legacy support
			f.seek(current_position)
			info.FileID = reader.ReadUInt64()
			info.Checksum = reader.ReadUInt32()
			info.CompressionMethod = reader.ReadUInt16()
			info.CRC = info.Checksum

			data.append(info)

	for a in data:
		f.seek(a.Offset + a.HeaderSize, 0)
		data = f.read(a.CompressedSize)
		if a.CompressionMethod == 0:

			if a.FileID in hashes:
				hashData = hashes[a.FileID]
				writeFile(data, True, hashData["filename"])
		else:
			if a.FileID in hashes:
				hashData = hashes[a.FileID]
				try:
					b = zlib.decompress(data)
					writeFile(b, True, hashData["filename"])
				except zlib.error:
					print("Fuck")