package main

import (
	"fmt"
	"math"

	"github.com/g3n/engine/math32"
)

// smooth / wall-sliding collision
// swept-aabb collision detection & response
// source:
// - 2d: https://www.gamedev.net/articles/programming/general-and-gameplay-programming/swept-aabb-collision-detection-and-response-r3084/
// - 3d: https://gist.github.com/iam4722202468/590c032cb5eeb60437e47e8e2cbb5091 (can't handle collisions when only moving one 1 axis)
// - 3d: https://github.com/rpereira-dev/CubeEngine/blob/master/VoxelEngine/src/com/grillecube/common/world/entity/collision/CollisionDetection.java
// - another 3d example using the Minkowski's difference: https://luisreis.net/blog/aabb_collision_handling/ (seems to work, attention: requires closest point to the origin and size along each axis)

// depends on player speed & smallest object size
const COLLISION_EPSILON float32 = 0.0000001

type AABBox struct {
	math32.Vector3
	sx, sy, sz float32 // size
	vx, vy, vz float32
}

func NewAABBoxFromBox3(pos math32.Vector3, vel math32.Vector3) AABBox {
	return AABBox{
		Vector3: pos,
		sx:      1, sy: 1, sz: 1,
		vx: vel.X, vy: vel.Y, vz: vel.Z,
	}
}

type AABBCollision struct {
	nx, ny, nz float32 // normals
	dt         float32 // maximum multiplier, 0-1
}

// detectAABB does a swept AABB detection on the (potentially) moving box b1 and the static box b2
func detectAABB(b1, b2 AABBox) AABBCollision {
	// no collision possible
	if b1.vx == 0 && (b1.X+b1.sx <= b2.X || b1.X >= b2.X+b2.sx) {
		return AABBCollision{dt: 1}
	}
	if b1.vy == 0 && (b1.Y+b1.sy <= b2.Y || b1.Y >= b2.Y+b2.sy) {
		return AABBCollision{dt: 1}
	}
	if b1.vz == 0 && (b1.Z+b1.sz <= b2.Z || b1.Z >= b2.Z+b2.sz) {
		return AABBCollision{dt: 1}
	}

	// find the distance between the objects on the near and far sides of each axis
	var xInvEntry, yInvEntry, zInvEntry float32
	var xInvExit, yInvExit, zInvExit float32
	if b1.vx > 0.0 {
		xInvEntry = b2.X - (b1.X + b1.sx)
		xInvExit = (b2.X + b2.sx) - b1.X
	} else {
		xInvEntry = (b2.X + b2.sx) - b1.X
		xInvExit = b2.X - (b1.X + b1.sx)
	}

	if b1.vy > 0.0 {
		yInvEntry = b2.Y - (b1.Y + b1.sy)
		yInvExit = (b2.Y + b2.sy) - b1.Y
	} else {
		yInvEntry = (b2.Y + b2.sy) - b1.Y
		yInvExit = b2.Y - (b1.Y + b1.sy)
	}

	if b1.vz > 0.0 {
		zInvEntry = b2.Z - (b1.Z + b1.sz)
		zInvExit = (b2.Z + b2.sz) - b1.Z
	} else {
		zInvEntry = (b2.Z + b2.sz) - b1.Z
		zInvExit = b2.Z - (b1.Z + b1.sz)
	}

	log.Debug(fmt.Sprintf("xInvEntry/exit = %f/%f, yInvEntry/Exit = %f/%f, zInvEntry/Exit = %f/%f", xInvEntry, xInvExit, yInvEntry, yInvExit, zInvEntry, zInvExit))

	// find time of collision and time of leaving for each axis (`if` statement is to prevent divide by zero)
	var xEntry, yEntry, zEntry float32
	var xExit, yExit, zExit float32
	if b1.vx == 0.0 {
		// TODO if xInvExit <= 0: entry without movement
		xEntry = float32(math.Inf(-1))
		xExit = float32(math.Inf(1))
	} else {
		xEntry = xInvEntry / b1.vx
		xExit = xInvExit / b1.vx
	}

	// TODO what if z & y are already 'colliding'?
	// but then yEntry must not be 0!! that'd be bad as well
	if b1.vy == 0.0 {
		// TODO if yInvExit <= 0: entry without movement
		yEntry = float32(math.Inf(-1))
		yExit = float32(math.Inf(1))
	} else {
		yEntry = yInvEntry / b1.vy
		yExit = yInvExit / b1.vy
	}

	if b1.vz == 0.0 {
		// TODO if zInvExit <= 0: entry without movement
		zEntry = float32(math.Inf(-1))
		zExit = float32(math.Inf(1))
	} else {
		zEntry = zInvEntry / b1.vz
		zExit = zInvExit / b1.vz
	}

	log.Debug(fmt.Sprintf("xEntry/exit = %f/%f, yEntry/Exit = %f/%f, zEntry/Exit = %f/%f", xEntry, xExit, yEntry, yExit, zEntry, zExit))

	entryTime := float32(math.Max(float64(xEntry), math.Max(float64(yEntry), float64(zEntry))))
	exitTime := float32(math.Min(float64(xExit), math.Min(float64(yExit), float64(zExit))))
	log.Debug(fmt.Sprintf("entry/exitTime = %f/%f", entryTime, exitTime))

	// the main source does not consider that collisions can occur when only one entry is < 0
	if entryTime >= exitTime || entryTime < 0.0 {
		return AABBCollision{dt: 1}
	}

	var nx, ny, nz float32
	log.Info("collision: entryTime: " + fmt.Sprint(entryTime) + ", exitTime: " + fmt.Sprint(exitTime))
	// calculate normal of collided surface (consider if velocity is negative)
	if xEntry > yEntry && xEntry > zEntry {
		ny = 0.0
		nz = 0.0
		if b1.vx < 0.0 { // xInvEntry
			nx = 1.0
		} else {
			nx = -1.0
		}
	} else if yEntry > xEntry && yEntry > zEntry {
		nx = 0.0
		nz = 0.0
		if b1.vy < 0.0 { // yInvEntry
			ny = 1.0
		} else {
			ny = -1.0
		}
	} else {
		nz = 0.0
		ny = 0.0
		if b1.vz < 0.0 { // zInvEntry
			nz = 1.0
		} else {
			nz = -1.0
		}
	}
	return AABBCollision{nx, ny, nz, entryTime}
}

// objects only contains world (can only handle static objects, v=0), not entities (otherwise add b1 != obj check)
func detect(b1 AABBox, objects []AABBox) AABBCollision {
	// earliest
	collision := AABBCollision{dt: 1}
	for i := 0; i < len(objects); i++ {
		c := detectAABB(b1, objects[i])
		if c.dt < collision.dt {
			collision = c
		}
	}
	return collision
}
