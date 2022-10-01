package main

import (
	"fmt"
	"math"

	"github.com/g3n/engine/core"
	"github.com/g3n/engine/math32"
)

// swept aabb collision detection & response
// source:
// - 2d: https://www.gamedev.net/articles/programming/general-and-gameplay-programming/swept-aabb-collision-detection-and-response-r3084/
// - 3d: https://gist.github.com/iam4722202468/590c032cb5eeb60437e47e8e2cbb5091

type AABBox struct {
	x, y, z    float32
	w, h, d    float32
	vx, vy, vz float32
}

/*func NewAABBoxFromBox3(box3 math32.Box3, vx, vy, vz float32) AABBox {
	return AABBox{x: box3.Min.X, y: box3.Min.Y, z: box3.Min.Z, w: box3.Max.X - box3.Min.X, h: box3.Max.Y - box3.Min.Y, d: box3.Max.Z - box3.Min.Z,
		vx: vx, vy: vy, vz: vz}
}*/

// TODO try WorldPosition
// TODO try node.Position().X + node.BoundingBox().Max.X though size was always correct anyway
// TODO node.BoundingBox().Max.Y - node.BoundingBox().Min.Y -> -Inf?? get w,h,d as param sicne only v.BoundingBox() returns the correct Box...
func NewAABBoxFromBox3(node *core.Node, vx, vy, vz float32) AABBox {
	var worldPos math32.Vector3
	node.WorldPosition(&worldPos)
	return AABBox{
		//x: node.Position().X, y: node.Position().Y, z: node.Position().Z,
		x: worldPos.X, y: worldPos.Y, z: worldPos.Z,
		w: 1, h: 1, d: 1,
		vx: vx, vy: vy, vz: vz,
	}
}

// returns the maximum multiplier which can be applied to the direction without collision (0-1)
func SweptAABB(b1, b2 AABBox, normalx, normaly, normalz *float32) float32 {
	// xInvEntry and yInvEntry specify how far away the closest edges of the objects are from each other.
	// xInvExit and yInvExit are the distance to the far side of the object.
	// You can think of this is a being like shooting through an object; the entry point is where the bullet goes through,
	// and the exit point is where it exits from the other side. These values are the inverse time until it hits the other object on the axis.
	// We will now use these values to take the velocity into account.
	var xInvEntry, yInvEntry, zInvEntry float32
	var xInvExit, yInvExit, zInvExit float32

	// find the distance between the objects on the near and far sides for both x and z
	if b1.vx > 0.0 {
		xInvEntry = b2.x - (b1.x + b1.w)
		xInvExit = (b2.x + b2.w) - b1.x
	} else {
		xInvEntry = (b2.x + b2.w) - b1.x
		xInvExit = b2.x - (b1.x + b1.w)
	}

	if b1.vy > 0.0 {
		yInvEntry = b2.y - (b1.y + b1.h)
		yInvExit = (b2.y + b2.h) - b1.y
	} else {
		yInvEntry = (b2.y + b2.h) - b1.y
		yInvExit = b2.y - (b1.y + b1.h)
	}

	if b1.vz > 0.0 {
		zInvEntry = b2.z - (b1.z + b1.d)
		zInvExit = (b2.z + b2.d) - b1.z
	} else {
		zInvEntry = (b2.z + b2.d) - b1.z
		zInvExit = b2.z - (b1.z + b1.d)
	}

	// What we are doing here is dividing the xEntry, yEntry, xExit and yExit by the object's velocity.
	// Of course, if the velocity is zero on any axis, it will cause a divide-by-zero error.
	// These new variables will give us our value between 0 and 1 of when each collision occurred on each axis.
	// The next step is to find which axis collided first.
	// find time of collision and time of leaving for each axis (if statement is to prevent divide by zero)
	var xEntry, yEntry, zEntry float32
	var xExit, yExit, zExit float32
	if b1.vx == 0.0 {
		xEntry = float32(math.Inf(-1))
		xExit = float32(math.Inf(1))
	} else {
		xEntry = xInvEntry / b1.vx
		xExit = xInvExit / b1.vx
	}

	if b1.vy == 0.0 {
		yEntry = float32(math.Inf(-1))
		yExit = float32(math.Inf(1))
	} else {
		yEntry = yInvEntry / b1.vy
		yExit = yInvExit / b1.vy
	}

	if b1.vz == 0.0 {
		zEntry = float32(math.Inf(-1))
		zExit = float32(math.Inf(1))
	} else {
		zEntry = zInvEntry / b1.vz
		zExit = zInvExit / b1.vz
	}

	// entryTime will tell use when the collision first occurred and exitTime will tell us when it exited the object from the other side.
	// This can be useful for certain effects, but at the moment, we just need it to calculate if a collision occurred at all.
	// find the earliest/latest times of collisionfloat
	entryTime := float32(math.Max(float64(xEntry), math.Max(float64(yEntry), float64(zEntry))))
	exitTime := float32(math.Min(float64(xExit), math.Min(float64(yExit), float64(zExit))))

	// if there was no collision
	// (xEntry < 0.0 && yEntry < 0.0) || (xEntry < 0.0 && zEntry < 0.0) || (yEntry < 0.0 && zEntry < 0.0) -> (xEntry < 0.0 && yEntry < 0.0 && zEntry < 0.0)
	if entryTime > exitTime || (xEntry < 0.0 && yEntry < 0.0 && zEntry < 0.0) || xEntry > 1.0 || yEntry > 1.0 || zEntry > 1.0 {
		*normalx = 0.0
		*normaly = 0.0
		*normalz = 0.0
		return 1.0
	} else { // if there was a collision
		log.Info("collision: entryTime: " + fmt.Sprint(entryTime) + ", exitTime: " + fmt.Sprint(exitTime))
		// calculate normal of collided surface
		if xEntry > yEntry && xEntry > zEntry {
			if xInvEntry < 0.0 {
				*normalx = 1.0
				*normaly = 0.0
				*normalz = 0.0
			} else {
				*normalx = -1.0
				*normaly = 0.0
				*normalz = 0.0
			}
		} else if yEntry > zEntry {
			if yInvEntry < 0.0 {
				*normalx = 0.0
				*normaly = 1.0
				*normalz = 0.0
			} else {
				*normalx = 0.0
				*normaly = -1.0
				*normalz = 0.0
			}
		} else {
			if zInvEntry < 0.0 {
				*normalx = 0.0
				*normaly = 0.0
				*normalz = 1.0
			} else {
				*normalx = 0.0
				*normaly = 0.0
				*normalz = -1.0
			}
		}
		return entryTime
	}
}
