package plyReaderRealsense

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
)

// PLY definitions, for consistency with original test file
const (
	PLY_ASCII     = 1 /* ascii PLY file */
	PLY_BINARY_BE = 2 /* binary PLY file, big endian */
	PLY_BINARY_LE = 3 /* binary PLY file, little endian */

	PLY_OKAY  = 0  /* ply routine worked okay */
	PLY_ERROR = -1 /* error in ply routine */

	/* scalar data types supported by PLY format */
	PLY_START_TYPE = 0
	PLY_CHAR       = 1
	PLY_SHORT      = 2 // int16
	PLY_INT        = 3
	PLY_UCHAR      = 4
	PLY_USHORT     = 5
	PLY_UINT       = 6
	PLY_FLOAT      = 7
	PLY_DOUBLE     = 8
	PLY_END_TYPE   = 9

	PLY_SCALAR = 0
	PLY_LIST   = 1
)

//
type PlyProperty struct {
	Name          string /* property name */
	External_type int    /* file's data type */
	Internal_type int    /* program's data type */
	Offset        int    /* offset bytes of prop in a struct */

	Is_list        int /* 1 = list, 0 = scalar */
	Count_external int /* file's count type */
	Count_internal int /* program's count type */
	Count_offset   int /* offset byte for list count */
}

//// description of a property and its constructor
//type PlyPropertyNew struct {
//	Name    string   // property nme
//	is_list bool     // 1 = list, 0 = scalar
//	types   []string // slice of the types in the properties
//}

func New_property(name string, et int, it int, offs int, il int, ce int, ci int, co int) *PlyProperty {
	return &PlyProperty{
		Name:           name,
		External_type:  et,
		Internal_type:  it,
		Offset:         offs,
		Is_list:        il,
		Count_external: ce,
		Count_internal: ci,
		Count_offset:   co,
	}
}

// description of an element and its constructor
type PlyElement struct {
	name   string        // element name
	num    int           // number of elements in this object
	marker int           // marker during the reading or writing
	props  []PlyProperty // list of properties in the file
}

func New_element(name string, num int, props []PlyProperty) *PlyElement {
	return &PlyElement{
		name:   name,
		num:    num,
		marker: 0,
		props:  props,
	}
}

// description of an .ply file and its constructor
type PlyFile struct {
	name       string
	Fp         *os.File     // file pointer
	file_type  int          // 1 : ascii; 3 : binary little endian; 2 : binary big endian
	header_vol int          // number of bytes occupied bt the header
	version    float32      // version number of file
	elems      []PlyElement // list of elements
	comments   []string     // list of comments
	obj_info   []string     // list of oject ifo
}

func New_file(name string, fp *os.File, file_type int, hv int, version float32, elems []PlyElement, comments []string, obj_info []string) *PlyFile {
	return &PlyFile{
		name:       name,
		Fp:         fp,
		file_type:  file_type,
		header_vol: hv,
		version:    version,
		elems:      elems,
		comments:   comments,
		obj_info:   obj_info,
	}
}

/* PlyOpenForReading opens a PLY file (specified by filename) and reads in the header information. The returned PlyFile object is used to access header information and data stored in the PLY file. */
func PlyOpenForReading(filename string) (*PlyFile, []string) {

	// Variables of PlyFile
	var file_type int
	var version float64
	var comments []string
	var obj_info []string
	var elems []PlyElement

	// Variables of PlyElement
	var props []PlyProperty
	var ele_name string
	var ele_num int

	prop_edited := false

	// Open the file
	file, err_F := os.Open(filename)
	if err_F != nil {
		log.Fatal(err_F)
	}

	// Read lines until the end of the header
	var lines []string
	linecount := 0
	bytecount := 0
	buffer := bufio.NewReader(file)
	for true {
		buf, vol_bytes := ReadlineAscii(buffer)
		lines = append(lines, buf)
		bytecount += vol_bytes

		//indentify the end of the header
		if len(lines[linecount]) > 3 {
			if lines[linecount][0:3] == "end" {
				break
			}
		}

		linecount++
	}

	// Parsing : deciding what does each line describe
	for _, i := range lines {
		split := strings.Fields(i)
		switch split[0] {
		// define the format
		case "format":
			version, _ = strconv.ParseFloat(split[2], 64)
			switch split[1] {
			case "ascii":
				file_type = PLY_ASCII
			case "binary_little_endian":
				file_type = PLY_BINARY_LE
			case "binary_big_endian":
				file_type = PLY_BINARY_BE
			}

		// add comment
		case "comment":
			comments = append(comments, i[8:])

		// add obj_info
		case "obj_info":
			if len(split) > 1 {
				for i := 1; i < len(split); i++ {
					obj_info = append(obj_info, split[i])
				}
			}

		//  add a new element
		case "element":
			// if we meet a new element and we have edited the properties of an element : we pack this element and add it in to the slice
			if prop_edited == true {
				elems = append(elems, *New_element(ele_name, ele_num, props))
				prop_edited = false
				props = nil
			}
			ele_name = split[1]
			ele_num, _ = strconv.Atoi(split[2])

		case "property":
			// if this line describes a property list : the second keyword will be "list", the third keyword will describe the type of first data (length of the list) and the fourth keyword for the rest data
			var isList, count, typ int
			var name string
			if split[1] == "list" {
				isList = 1
				count = TypeConverter(split[2])
				typ = TypeConverter(split[3])
				name = split[4]
			} else {
				isList = 0
				count = 0
				typ = TypeConverter(split[1])
				name = split[2]
			}
			prop := New_property(name, typ, typ, 0, isList, count, count, 0)
			props = append(props, *prop)

			// indicate whether the property is modified
			prop_edited = true

		case "end_header":
			// we pack the last element and add it in to the slice
			elems = append(elems, *New_element(ele_name, ele_num, props))
		}
	}

	plyfile := New_file(filename, file, file_type, bytecount, float32(version), elems, comments, obj_info)

	nelems := len(plyfile.elems)
	elem_names := make([]string, nelems)
	for i := 0; i < nelems; i++ {
		elem_names[i] = plyfile.elems[i].name
	}

	// positioning the file pointer for the next step
	plyfile.Fp.Close()
	plyfile.Fp, _ = os.Open(plyfile.name)
	buf_trash := make([]byte, plyfile.header_vol)
	_, _ = plyfile.Fp.Read(buf_trash)

	return plyfile, elem_names
}

/* PlyGetElementDescription reads information about a specified element from an open PLY file. Return : the list of properties for this element ; the total number of elements in this file ; the number of properties for this element*/
func PlyGetElementDescription(plyfile *PlyFile, element_name string) ([]PlyProperty, int, int) {
	for i := 0; i < len(plyfile.elems); i++ {
		if plyfile.elems[i].name == element_name {
			return plyfile.elems[i].props, plyfile.elems[i].num, len(plyfile.elems[i].props)
		}
	}
	return nil, 0, 0
}

func LocateElement(plyfile *PlyFile, size uintptr) {
	// determine the start of the element to read
	for elem_index := 0; elem_index < len(plyfile.elems); elem_index++ {
		if plyfile.elems[elem_index].marker != 1 {
			buff := make([]byte, size)
			for index := 0; index < plyfile.elems[elem_index].num; index++ {
				_, _ = plyfile.Fp.Read(buff)
			}
		}
	}
}

/* PlyGetComments returns the comments contained in the open PLY file header. */
func PlyGetComments(plyfile *PlyFile) []string {
	return plyfile.comments
}

/* PlyGetObjInfo returns the object info contained in the open PLY file header. */
func PlyGetObjInfo(plyfile *PlyFile) []string {
	return plyfile.obj_info
}

/* PlyClose closes the open plyfile */
func PlyClose(plyfile *PlyFile) {
	if plyfile != nil {
		err := plyfile.Fp.Close()
		if err != nil {
		}
	}
}

// read a line in format ascii, read the appended list of string and its volume in bytes
func ReadlineAscii(reader *bufio.Reader) (string, int) {
	a, _, _ := reader.ReadLine()
	buf := string(a)
	return buf, len(string(a)) + 1
}

func TypeConverter(typeStr string) int {
	switch typeStr {
	case "float":
		return PLY_FLOAT
	case "uchar":
		return PLY_UCHAR
	case "int":
		return PLY_INT
	case "short":
		return PLY_SHORT
	case "float32":
		return PLY_FLOAT
	}
	return 0
}

func TypeConverterInverse(typeInt int) string {
	switch typeInt {
	case PLY_FLOAT:
		return "float32"
	case PLY_INT:
		return "int"
	case PLY_UCHAR:
		return "uchar"
	case PLY_SHORT:
		return "short"
	}
	return ""
}
