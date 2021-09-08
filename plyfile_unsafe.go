package plyReaderRealsense

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
)


type Vertex struct {
	X, Y, Z float32
	R, G, B uint8
}

type VertexMono struct {
	X, Y, Z float32
}

type VertexMono64 struct {
	X, Y, Z float64
}

type FaceReading struct {
	Nverts    byte
	Vert1 [4]byte
	Vert2 [4]byte
	Vert3 [4]byte
}

type FaceReadingHuge struct {
	Nverts    uint8
	Vert1 int32
	Vert2 int32
	Vert3 int32
}

type Face64 struct {
	X, Y, Z int64
}

type Face32 struct {
	X, Y, Z int32
}

//all unsafe functions are here

/* PlyOpenForWriting opens a file and returns a pointer to the root struct PlyFile which will contain the header of the data to write */
func PlyOpenForWriting(filename string, nelems int, elem_names []string, file_type int, version *float32) *PlyFile {
	// announce variables
	var list_elems []PlyElement
	header_vol := 0
	var list_comments []string
	var obj_info []string

	// create a file
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
	}

	// initialize the slice of PlyElement
	for i := 0; i < len(elem_names); i ++ {
		list_props := make([]PlyProperty, 0)
		elem := New_element(elem_names[i], 0, list_props)
		list_elems = append(list_elems, *elem)
	}

	return New_file(filename, f, file_type, header_vol, *version, list_elems, list_comments, obj_info)
}

/* PlyElementCount specifies the total number of an element in the struct PlyFile */
func PlyElementCount(plyfile *PlyFile, elem_name string, num_this int) {
	elem_miss := true

	// iterate in the slice of  among all element to find the one we will modify
	for i := 0; i < len(plyfile.elems); i++ {
		if plyfile.elems[i].name == elem_name {
			plyfile.elems[i].num = num_this
			elem_miss = false
			break
		}
	}

	if elem_miss {
		fmt.Println("Element not found for", elem_name)
	}
}

/* PlyDescribeProperty describes a property of an element. */
func PlyDescribeProperty(plyfile *PlyFile, element_name string, prop PlyProperty) {
	prop_miss := true

	// iterate in the slice of  among all element to find the one we will modify
	for i := 0; i < len(plyfile.elems); i++ {
		if plyfile.elems[i].name == element_name {
			plyfile.elems[i].props = append(plyfile.elems[i].props, prop)
			prop_miss = false
			break
		}
	}

	if prop_miss {
		fmt.Println("Element not found for", element_name)
	}
}

/* */
func PlyPutComment(plyfile *PlyFile, comment string) {
	plyfile.comments = append(plyfile.comments, comment)
}

func PlyPutObjInfo(plyfile *PlyFile, obj_info string) {
	plyfile.obj_info = append(plyfile.obj_info, obj_info)
}

/* PlyHeaderComplete writes the header to the file*/
func PlyHeaderComplete(plyfile *PlyFile) {

	// write the start of the header
	_, err := plyfile.Fp.WriteString("ply\n")
	if err != nil {
		fmt.Println("Error when writing")
	}

	// write the file type
	switch plyfile.file_type {
	case PLY_ASCII:
		_, _ = plyfile.Fp.WriteString("format ascii 1.0\n")
	case PLY_BINARY_BE:
		_, _ = plyfile.Fp.WriteString("format binary_big_endian 1.0\n")
	case PLY_BINARY_LE:
		_, _ = plyfile.Fp.WriteString("format binary_little_endian 1.0\n")
	default:
		_, _ = plyfile.Fp.WriteString("bad file type entered\n")
	}

	// write the comments
	if len(plyfile.comments) > 0 {
		for i := 0; i < len(plyfile.comments); i++ {
			_, _ = plyfile.Fp.WriteString("comment " + plyfile.comments[i] + "\n")
		}
	}

	// write object information
	if len(plyfile.obj_info) > 0 {
		for i := 0; i < len(plyfile.obj_info); i++ {
			_, _ = plyfile.Fp.WriteString("obj_info " + plyfile.obj_info[i] + "\n")
		}
	}

	// write the information for each element
	for i := 0; i < len(plyfile.elems); i++ {
		_, _ = plyfile.Fp.WriteString("element " + plyfile.elems[i].name + " " + strconv.Itoa(plyfile.elems[i].num) + "\n")

		// write the corresponding properties
		for j := 0; j < len(plyfile.elems[i].props); j++ {
			if plyfile.elems[i].props[j].Is_list == 0 {
				_, _ = plyfile.Fp.WriteString("property " + TypeConverterInverse(plyfile.elems[i].props[j].Internal_type) + " " + plyfile.elems[i].props[j].Name + "\n")
			} else {
				_, _ = plyfile.Fp.WriteString("property list " + TypeConverterInverse(plyfile.elems[i].props[j].Count_internal) + " " + TypeConverterInverse(plyfile.elems[i].props[j].Internal_type) + " " + plyfile.elems[i].props[j].Name + "\n")
			}
		}
	}

	// write the end of the header
		_, _ = plyfile.Fp.WriteString("end_header" + "\n")

}

/* This function is meaningless if we specifies which element to input using different functions such as PlyPutElement and PlyPutElementFace.
	PlyPutElementSetup marks the element which is to be written next */
func PlyPutElementSetup(plyfile *PlyFile, b string) {
	ElementMiss := true

	for i := 0; i < len(plyfile.elems); i++ {
		if plyfile.elems[i].name == b {
			plyfile.elems[i].marker = 1
			ElementMiss = false
		}
	}

	if ElementMiss {
		fmt.Println("Element to be written not found")
	}
}


/* PlyPutElement writes the element Vertex. In binary mode it is compatible with elements having numbers of scalar properties. In ascii mode it is compatible with elements having numbers of 3 scalar properties named X, Y and Z. */
func PlyPutElement(plyfile *PlyFile, b Vertex) {
	switch plyfile.file_type {
	case PLY_BINARY_LE:
		// write one data
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, b)
		if err != nil {
			fmt.Println("Error when writing the vertex")
		}
		_, err = plyfile.Fp.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error when writing to the file")
		}

	case PLY_BINARY_BE:
		// write one data
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, b)
		if err != nil {
			fmt.Println("Error when writing the vertex")
		}
		_, err = plyfile.Fp.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error when writing to the file")
		}

	case PLY_ASCII:
		// write one data
		str := strconv.FormatFloat(float64(b.X), 'g', 8, 64) + " " + strconv.FormatFloat(float64(b.Y), 'g', 6, 64) + " " + strconv.FormatFloat(float64(b.Z), 'g', 8, 64) + " " + "\n"
		_, err := plyfile.Fp.WriteString(str)
		if err != nil {
			fmt.Println("Error when writing to the file")
		}
	}
}

/* PlyPutElementFace writes the element FaceReading. */
func PlyPutElementFace(plyfile *PlyFile, b FaceReading) {
	switch plyfile.file_type {
	case PLY_BINARY_LE:
		// write data
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.LittleEndian, b)
		if err != nil {
			fmt.Println("Error when writing the vertex")
		}
		_, err = plyfile.Fp.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error when writing to the file")
		}

	case PLY_BINARY_BE:
		// write one data
		buf := new(bytes.Buffer)
		err := binary.Write(buf, binary.BigEndian, b)
		if err != nil {
			fmt.Println("Error when writing the vertex")
		}
		_, err = plyfile.Fp.Write(buf.Bytes())
		if err != nil {
			fmt.Println("Error when writing to the file")
		}

	case PLY_ASCII:
		// write one data

		break

		//for i := 0; i < 4; i++ {
		//	buf := make([]byte, 0)
		//	buf = append(buf, b.Verts[i])
		//	buf = append(buf, b.Verts[i + 1])
		//	fmt.Println(&buf)
		//	bytebuff := bytes.NewBuffer(buf)
		//	var data int16
		//	binary.Read(bytebuff, binary.BigEndian, &data)
		//	fmt.Println(data)
		//
		//}
		//fmt.Println(b)

		//_, err := plyfile.Fp.WriteString(str)
		//if err != nil {
		//	fmt.Println("Error when writing to the file")
		//}
	}
}

/* PlyUseExistingForWriting creates a PlyFile object using an existing file pointer */
func PlyUseExistingForWriting(fp *os.File, nelems int, elem_names []string, file_type int, version *float32) *PlyFile {
	elems := make([]PlyElement, nelems)
	for i := 0;i < nelems; i++ {
		elems[i].name = elem_names[i]
	}

	comments := make([]string, 0)
	obj_info := make([]string, 0)

	plyfile := New_file("", fp, file_type, 0, *version, elems, comments, obj_info)

	return plyfile
}


