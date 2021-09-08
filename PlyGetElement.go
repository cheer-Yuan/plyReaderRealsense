package plyReaderRealsense

import (
	"bytes"
	"encoding/binary"
)

/* PlyGetElementHuge retrieves all the elements of a type from the PLY file to reduce the time cost on calling system functions. */
func PlyGetElementHuge(plyfile *PlyFile, element interface{}, size int) {
	switch plyfile.file_type {
	case PLY_BINARY_LE:
		buffbytes := make([]byte, size)
		r := bytes.NewReader(buffbytes)
		_, _ = plyfile.Fp.Read(buffbytes)
		_ = binary.Read(r, binary.LittleEndian, element)

	case PLY_BINARY_BE:
		buffbytes := make([]byte, size)
		r := bytes.NewReader(buffbytes)
		_, _ = plyfile.Fp.Read(buffbytes)
		_ = binary.Read(r, binary.BigEndian, element)
	}
}