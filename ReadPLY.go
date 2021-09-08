package plyReaderRealsense

import (
	"dataprocessing/mymath"
	"dataprocessing/plyfile"
	"encoding/binary"
	"math/rand"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))



// read a monochrome .ply file, for 32 bits data and 64 bits data
func ReadPLYMono64(filename string) ([]plyfile.VertexMono64, []plyfile.Face64) {
	var vertices []plyfile.VertexMono64
	var faces []plyfile.Face64

	// open the PLY file for reading
	cplyfile, elem_names := plyfile.PlyOpenForReading(filename)

	// read each element
	for _, name := range elem_names {

		// get element description
		_, num_elems, _ := plyfile.PlyGetElementDescription(cplyfile, name)

		if name == "vertex" {
			// read all the vertices
			vlisthuge := make([]float32, num_elems * 3)
			plyfile.PlyGetElementHuge(cplyfile, &vlisthuge, len(vlisthuge) * 4)
			for i := 0; i < num_elems; i++ {
				var buff plyfile.VertexMono64
				buff.X, buff.Y, buff.Z = float64(vlisthuge[i * 3]), float64(vlisthuge[i * 3 + 1]), float64(vlisthuge[i * 3 + 2])
				vertices = append(vertices, buff)
			}
		} else if name == "face" {
			// read all the faces
			flisthuge := make([]byte, num_elems * 25)
			_, _ = cplyfile.Fp.Read(flisthuge)

			// decode the integers from the memory
			for i := 0; i < num_elems; i++ {
				var buff plyfile.Face64
				buff.X, buff.Y, buff.Z = int64(binary.LittleEndian.Uint64(flisthuge[i * 25 + 1 : i * 25 + 9])), int64(binary.LittleEndian.Uint64(flisthuge[i * 25 + 9 : i * 25 + 17])), int64(binary.LittleEndian.Uint64(flisthuge[i * 25 + 17 : i * 25 + 25]))
				faces = append(faces, buff)
			}
		}


	}
	// close the PLY file
	plyfile.PlyClose(cplyfile)
	return vertices, faces
}
func ReadPLYMono32(filename string) ([]plyfile.VertexMono, []plyfile.Face32) {
	var vertices []plyfile.VertexMono
	var faces []plyfile.Face32

	// open the PLY file for reading
	cplyfile, elem_names := plyfile.PlyOpenForReading(filename)

	// read each element
	for _, name := range elem_names {

		// get element description
		_, num_elems, _ := plyfile.PlyGetElementDescription(cplyfile, name)

		if name == "vertex" {
			// read all the vertices
			vlisthuge := make([]float32, num_elems * 3)
			plyfile.PlyGetElementHuge(cplyfile, &vlisthuge, len(vlisthuge) * 4)
			for i := 0; i < num_elems; i++ {
				var buff plyfile.VertexMono
				buff.X, buff.Y, buff.Z = vlisthuge[i * 3], vlisthuge[i * 3 + 1], vlisthuge[i * 3 + 2]
				vertices = append(vertices, buff)
			}
		} else if name == "face" {
			// read all the faces
			flisthuge := make([]byte, num_elems * 13)
			_, _ = cplyfile.Fp.Read(flisthuge)

			// decode the integers from the memory
			for i := 0; i < num_elems; i++ {
				var buff plyfile.Face32
				buff.X, buff.Y, buff.Z = int32(binary.LittleEndian.Uint32(flisthuge[i * 13 + 1 : i * 13 + 5])), int32(binary.LittleEndian.Uint32(flisthuge[i * 13 + 5 : i * 13 + 9])), int32(binary.LittleEndian.Uint32(flisthuge[i * 13 + 9 : i * 13 + 13]))
				faces = append(faces, buff)
			}
		}
		//for i := 0; i < num_props; i++ {
		//	fmt.Println("property", plist[i].Name)
		//}
	}
	// close the PLY file
	plyfile.PlyClose(cplyfile)
	return vertices, faces
}


// AddNoise add noise to a given percentage of the total points, for 32 bits data and 64 bits data
func AddNoise32(vertices []plyfile.VertexMono, percent float64, minNoise float64, maxNoise float64) {
	// determine on which vertices to add the noise
	var ListIndex []int
	for i := 0; i < int(percent * float64(len(vertices))); i++ {
		index := r.Intn(len(vertices) - 1)
		if mymath.ExistIntList(ListIndex, index) {
			i--
			continue
		} else {
			ListIndex = append(ListIndex, index)
		}
	}

	// multiply a noise between ]minNoise, maxNoise[
	for _, i := range ListIndex {
		vertices[i].X *= float32(minNoise + r.Float64() * (maxNoise - minNoise))
		vertices[i].Y *= float32(minNoise + r.Float64() * (maxNoise - minNoise))
		vertices[i].Z *= float32(minNoise + r.Float64() * (maxNoise - minNoise))
	}
}
func AddNoise64(vertices []plyfile.VertexMono64, percent float64, minNoise float64, maxNoise float64) {
	// determine on which vertices to add the noise
	var ListIndex []int
	for i := 0; i < int(percent * float64(len(vertices))); i++ {
		index := r.Intn(len(vertices) - 1)
		if mymath.ExistIntList(ListIndex, index) {
			i--
			continue
		} else {
			ListIndex = append(ListIndex, index)
		}
	}

	// multiply a noise between ]minNoise, maxNoise[
	for _, i := range ListIndex {
		vertices[i].X *= minNoise + r.Float64() * (maxNoise - minNoise)
		vertices[i].Y *= minNoise + r.Float64() * (maxNoise - minNoise)
		vertices[i].Z *= minNoise + r.Float64() * (maxNoise - minNoise)
	}
}

