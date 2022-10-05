package main

import (
	"testing"

	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/util/logger"
	"github.com/stretchr/testify/assert"
)

// TODO size 0.5, etc. tests

func init() {
	log = logger.New("Game", nil)
	log.AddWriter(logger.NewConsole(true))
	log.SetFormat(logger.FTIME | logger.FMICROS)
	log.SetLevel(logger.DEBUG)
}

func TestSweptAABBNoVelNoCol(t *testing.T) {
	b1 := NewAABBoxFromBox3(math32.Vector3{1, 1, 1}, math32.Vector3{})
	b2 := NewAABBoxFromBox3(math32.Vector3{3, 3, 3}, math32.Vector3{})
	collision := detect(b1, []AABBox{b2})
	assert.Equal(t, AABBCollision{dt: 1}, collision, "collision")
}

// no col: movement in x, x differs
func TestSweptAABBNoCol(t *testing.T) {
	b1 := NewAABBoxFromBox3(math32.Vector3{1, 1, 1}, math32.Vector3{3, 0, 0})
	b2 := NewAABBoxFromBox3(math32.Vector3{5, 3, 3}, math32.Vector3{})
	collision := detect(b1, []AABBox{b2})
	assert.Equal(t, AABBCollision{dt: 1}, collision, "collision")
}

func TestSweptAABBXCol(t *testing.T) {
	b1 := NewAABBoxFromBox3(math32.Vector3{0, 0, 0}, math32.Vector3{4, 0, 0})
	b2 := NewAABBoxFromBox3(math32.Vector3{2, 0, 0}, math32.Vector3{})
	collision := detect(b1, []AABBox{b2})
	assert.Equal(t, AABBCollision{dt: 0.25, nx: -1}, collision, "collision")
}

func TestSweptAABBYCol(t *testing.T) {

}

func TestSweptAABBZCol(t *testing.T) {

}

func TestSweptAABB(t *testing.T) {
	// source: https://gist.github.com/iam4722202468/590c032cb5eeb60437e47e8e2cbb5091
	b1 := NewAABBoxFromBox3(math32.Vector3{1, 1, 1}, math32.Vector3{3, 4, 7})
	b2 := NewAABBoxFromBox3(math32.Vector3{3, 3, 3}, math32.Vector3{})
	collision := detect(b1, []AABBox{b2})
	assert.Equal(t, AABBCollision{dt: float32(1) / 3, nx: -1}, collision, "collision")
}
