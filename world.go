package main

import (
	"fmt"
	"time"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
)

const CHUNK_SIZE = 2

type World struct {
	scene *core.Node

	gopherNode *core.Node
	velocity   float32 // linear velocity (m/s)
	rotvel     float32 // rotation velocity (rad/s)
	gravity    float32 // linear acceleration (m/s^2)

	// shared
	CubeGeo  *geometry.Geometry
	Mat      Materials
	MakeCube func(mat *material.Standard) func() *graphic.Mesh
}

type Materials struct {
	Grass *material.Standard
	Dirt  *material.Standard
	Stone *material.Standard
}

func (w *World) Update(renderer *renderer.Renderer, deltaTime time.Duration) {
	gbox := w.gopherNode.BoundingBox()
	children := w.scene.Children()
	colls := 0
	log.Debug("checking #" + fmt.Sprint(len(children)) + " (-1) collisions")
	for _, v := range children {
		bbox := v.BoundingBox()
		if gbox.IsIntersectionBox(&bbox) && v.Name() != "gopher" {
			log.Debug("coll: " + v.Name())
			colls++

			// step by step from last position to this position before intersection
			// TODO also reset camera & orbit
		}
	}
	log.Debug("colls: #" + fmt.Sprint(colls))
}

func (w *World) NewChunk(chunkX, chunkY, chunkZ float32) {
	for x := 0; x < CHUNK_SIZE; x++ {
		for y := 0; y < CHUNK_SIZE; y++ {
			for z := 0; z < CHUNK_SIZE; z++ {
				var cube *graphic.Mesh
				if y == 0 {
					cube = w.MakeCube(w.Mat.Grass)()
				} else if y < 4 {
					cube = w.MakeCube(w.Mat.Dirt)()
				} else {
					cube = w.MakeCube(w.Mat.Stone)()
				}
				cube.SetPosition(chunkX+float32(x), chunkY-float32(y), chunkZ-float32(z))
				log.Debug("cube pos: " + fmt.Sprint(cube.Position()) + " / " + fmt.Sprint(cube.BoundingBox()))
				w.scene.Add(cube)
			}
		}
	}
	cube := w.MakeCube(w.Mat.Stone)()
	cube.SetPosition(chunkX+1, chunkY+1, chunkZ+1)
	w.scene.Add(cube)
}

// y: green, up/down
// z: blue, horizontal
// x: red, horizontal
func NewWorld() *World {
	w := new(World)
	w.CubeGeo = geometry.NewCube(1)
	w.Mat.Grass = material.NewStandard(math32.NewColorHex(0x008A13))
	w.Mat.Dirt = material.NewStandard(math32.NewColorHex(0x9B7653))
	w.Mat.Stone = material.NewStandard(math32.NewColorHex(0xB7B09C))

	// WithMaterial
	w.MakeCube = func(mat *material.Standard) func() *graphic.Mesh {
		return func() *graphic.Mesh { return graphic.NewMesh(w.CubeGeo, mat) }
	}

	w.scene = core.NewNode()

	return w
}
