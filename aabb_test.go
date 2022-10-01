package main

import (
	"testing"

	"github.com/g3n/engine/util/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	log = logger.New("Game", nil)
	log.AddWriter(logger.NewConsole(true))
	log.SetFormat(logger.FTIME | logger.FMICROS)
	log.SetLevel(logger.DEBUG)
}

func TestSweptAABBNoVelNoCol(t *testing.T) {
	b1 := AABBox{
		x: 1, y: 1, z: 1,
		w: 1, h: 1, d: 1,
		vx: 0, vy: 0, vz: 0,
	}
	b2 := AABBox{
		x: 3, y: 3, z: 3,
		w: 1, h: 1, d: 1,
		vx: 0, vy: 0, vz: 0,
	}
	var normalx, normaly, normalz float32
	collisiontime := SweptAABB(b1, b2, &normalx, &normaly, &normalz)
	assert.Equal(t, float32(1.0), collisiontime, "collisiontime")
	assert.Equal(t, float32(0.0), normalx, "normalx")
	assert.Equal(t, float32(0.0), normaly, "normaly")
	assert.Equal(t, float32(0.0), normalz, "normalz")
}

// no col: movement in x, x differs
func TestSweptAABBNoCol(t *testing.T) {
	b1 := AABBox{
		x: 1, y: 1, z: 1,
		w: 1, h: 1, d: 1,
		vx: 3, vy: 0, vz: 0,
	}
	b2 := AABBox{
		x: 5, y: 3, z: 3,
		w: 1, h: 1, d: 1,
		vx: 0, vy: 0, vz: 0,
	}
	var normalx, normaly, normalz float32
	collisiontime := SweptAABB(b1, b2, &normalx, &normaly, &normalz)
	assert.Equal(t, float32(1.0), collisiontime, "collisiontime")
	assert.Equal(t, float32(0.0), normalx, "normalx")
	assert.Equal(t, float32(0.0), normaly, "normaly")
	assert.Equal(t, float32(0.0), normalz, "normalz")
}

func TestSweptAABBXCol(t *testing.T) {
	b1 := AABBox{
		x: 0, y: 0, z: 0,
		w: 1, h: 1, d: 1, // 0.5
		vx: 4, vy: 0, vz: 0,
	}
	b2 := AABBox{
		x: 2, y: 0, z: 0,
		w: 1, h: 1, d: 1,
		vx: 0, vy: 0, vz: 0,
	}
	var normalx, normaly, normalz float32
	collisiontime := SweptAABB(b1, b2, &normalx, &normaly, &normalz)
	assert.Equal(t, float32(0.25), collisiontime, "collisiontime")
	assert.Equal(t, float32(-1), normalx, "normalx")
	assert.Equal(t, float32(0.0), normaly, "normaly")
	assert.Equal(t, float32(0.0), normalz, "normalz")
}

func TestSweptAABBYCol(t *testing.T) {

}

func TestSweptAABBZCol(t *testing.T) {

}

func TestSweptAABB(t *testing.T) {
	b1 := AABBox{
		x: 1, y: 1, z: 1,
		w: 1, h: 1, d: 1,
		vx: 3, vy: 4, vz: 7,
	}
	b2 := AABBox{
		x: 3, y: 3, z: 3,
		w: 1, h: 1, d: 1,
		vx: 0, vy: 0, vz: 0,
	}
	var normalx, normaly, normalz float32
	collisiontime := SweptAABB(b1, b2, &normalx, &normaly, &normalz)
	t.Logf("collisionstime: %f", collisiontime)
}
