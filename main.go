package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/loader/obj"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/util/logger"
	"github.com/g3n/engine/window"
)

const (
	CMD_FORWARD = iota
	CMD_BACKWARD
	CMD_LEFT
	CMD_RIGHT
	CMD_LAST
)

var log *logger.Logger

type Game struct {
	*app.Application

	scene  *core.Node
	camera *camera.Camera
	orbit  *camera.OrbitControl

	world *World

	commands [CMD_LAST]bool // command states
}

// Update is called every frame.
func (g *Game) Update(renderer *renderer.Renderer, deltaTime time.Duration) {
	dt := float32(deltaTime.Seconds())

	// can be used after collision detection by others
	g.world.gopher.Velocity = math32.Vector3{}

	// Get player world position
	var position math32.Vector3
	g.world.gopher.node.WorldPosition(&position)

	// apply commands to player relative to world direction
	if g.commands[CMD_LEFT] || g.commands[CMD_RIGHT] {
		// Calculates angle delta to rotate
		angle := g.world.rotvel * dt
		if g.commands[CMD_RIGHT] {
			angle = -angle
		}
		g.world.gopher.node.RotateY(angle)
	}
	if g.commands[CMD_FORWARD] || g.commands[CMD_BACKWARD] {
		// Calculates the distance to move
		dist := g.world.velocity

		// player always moves towards own x
		direction := math32.Vector3{1, 0, 0}
		if g.commands[CMD_BACKWARD] {
			direction.Negate()
		}
		direction.MultiplyScalar(dist)
		g.world.gopher.Velocity.Add(&direction)

		log.Debug("direction: " + fmt.Sprint(direction))
	}

	// direction needs to be applied every time because rotation might change

	// gravity is applied every time to velocity, works
	// direction is applied only on just pressed, only works when -> velocity looses direction
	gg := g.world.gravity
	gravity := math32.Vector3{0, -1, 0}
	gravity.MultiplyScalar(gg)
	g.world.gopher.Velocity.Add(&gravity)

	// Get player world direction (every update), after rotation
	var quat math32.Quaternion
	g.world.gopher.node.WorldQuaternion(&quat)

	// velocity (x/z) is only updated when a key is just pressed or unpressed
	// but when rotating
	// translate player movement into world movement
	g.world.gopher.Velocity.ApplyQuaternion(&quat)
	//g.world.gopher.Velocity.Normalize() // TODO certainly not here... anywhere?
	g.world.gopher.Velocity.MultiplyScalar(dt)

	// run multiple times: to implement sliding, the dt gets subtracted and the colliding axis (|normal|>0) set to zero, then movement in other directions can be allowed
	// one vector, pressing & unpressing keys updates it. wall-sliding resets one value temporarily

	children := g.world.scene.Children()
	log.Debug("checking #" + fmt.Sprint(len(children)) + " (-1) collisions")
	staticBoxes := make([]AABBox, 0, len(children)-1)
	for _, v := range children {
		if v.Name() != "gopher" {
			staticBoxes = append(staticBoxes, NewAABBoxFromBox3(v.GetNode().Position(), math32.Vector3{}))
		}
	}
	cnt := 0
colldet:
	cnt++
	log.Debug("gopher vel: " + fmt.Sprint(g.world.gopher.Velocity))
	gbbox := NewAABBoxFromBox3(position, g.world.gopher.Velocity)
	log.Debug("gbbox: " + fmt.Sprint(gbbox))
	collision := detect(gbbox, staticBoxes)
	log.Debug("collision: " + fmt.Sprintf("%#v", collision))

	if collision.dt < 1 {
		// coll.dt may be 0, when walking into a wall -- why not when gravity?
		//dt -= collision.dt
		if collision.nx != 0 {
			println("x coll o.O")
			g.world.gopher.Velocity.X = 0
		}
		if collision.ny != 0 {
			g.world.gopher.Velocity.Y = 0
		}
		if collision.nz != 0 {
			println("z coll o.O")
			g.world.gopher.Velocity.Z = 0
		}
		if cnt < 2 {
			goto colldet
		}
	}
	g.world.gopher.Velocity.MultiplyScalar(collision.dt)

	log.Debug("oldpos: " + fmt.Sprintf("%#v", position))
	position.Add(&g.world.gopher.Velocity)
	log.Debug("newpos: " + fmt.Sprintf("%#v", position))
	g.world.gopher.node.SetPositionVec(&position)

	g.camera.LookAt(&position, &math32.Vector3{0, 2, 0})
	g.orbit.SetTarget(position) // updated independently

	cpos := g.camera.Position()
	cpos.Add(&g.world.gopher.Velocity)
	g.camera.SetPositionVec(&cpos)
	//position.Add(&math32.Vector3{-3, 1, 0})
	//g.camera.SetPositionVec(&position)
	//g.camera.SetQuaternionQuat(&quat)
	//position.Sub()
	//g.camera.SetDirectionVec()
	g.camera.SetChanged(true)

	// collision detection after movement, may reset
	// TODO if g.world != nil (ie. has been loaded)
	// TODO move gopher, etc. into World
	//g.world.Update(renderer, deltaTime)

	// Clear the color, depth, and stencil buffers
	g.Gls().Clear(gls.COLOR_BUFFER_BIT | gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT)
	// Render scene
	err := renderer.Render(g.scene, g.camera)
	if err != nil {
		panic(err)
	}
}

func main() {
	oDebug := flag.Bool("debug", false, "display the debug log and check OpenGL errors")
	flag.Parse()

	log = logger.New("Game", nil)
	log.AddWriter(logger.NewConsole(true))
	log.SetFormat(logger.FTIME | logger.FMICROS)
	if *oDebug {
		log.SetLevel(logger.DEBUG)
	} else {
		log.SetLevel(logger.INFO)
	}
	log.Info("Initializing...")

	g := new(Game)
	g.Application = app.App(1280, 920, "Game")
	log.Debug("OpenGL version: %s", g.Gls().GetString(gls.VERSION))

	// Speed up a bit by not checking OpenGL errors (only if not debugging)
	if !*oDebug {
		g.Gls().SetCheckErrors(false)
	}

	g.scene = core.NewNode()

	g.world = NewWorld()
	g.scene.Add(g.world.scene)
	gui.Manager().Set(g.scene)
	g.world.NewChunk(0, 0, 0)

	// TODO config
	g.world.velocity = 3.0
	g.world.gravity = 9.81
	g.world.rotvel = 4.0

	g.world.gopher = Entity{
		node: g.LoadGopher(),
	}
	g.world.gopher.node.SetPosition(0, 1.5, -1)
	g.world.scene.Add(g.world.gopher.node)

	// Create camera
	g.camera = camera.New(1)
	//g.camera.SetPosition(0, 1, -2)
	gpos := g.world.gopher.node.Position()
	gpos2 := gpos.Clone()
	gpos2.Add(&math32.Vector3{-3, 1, 0})
	g.camera.SetPositionVec(gpos2)
	g.camera.LookAt(gpos2, &math32.Vector3{0, 0, 0})
	g.scene.Add(g.camera)

	// Create orbit control and set limits
	g.orbit = camera.NewOrbitControl(g.camera)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := g.GetSize()
		g.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		g.camera.SetAspect(float32(width) / float32(height))
	}
	g.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Create and add a button to the scene
	btn := gui.NewButton("Make Red")
	btn.SetPosition(100, 40)
	btn.SetSize(40, 40)
	//btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
	//	mat.SetColor(math32.NewColor("DarkRed"))
	//})
	g.scene.Add(btn)

	// Create and add lights to the scene
	g.scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	g.scene.Add(pointLight)

	// Create and add an axis helper to the scene
	g.scene.Add(helper.NewAxes(1.5))

	// Set background color to gray
	g.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	// Subscribe to key events
	g.Subscribe(window.OnKeyDown, g.onKey)
	g.Subscribe(window.OnKeyUp, g.onKey)

	// Run the application
	g.Run(g.Update)
}

// Process key events
func (g *Game) onKey(evname string, ev interface{}) {
	state := evname == window.OnKeyDown

	kev := ev.(*window.KeyEvent)
	switch kev.Key {
	case window.KeyW:
		g.commands[CMD_FORWARD] = state
	case window.KeyS:
		g.commands[CMD_BACKWARD] = state
	case window.KeyA:
		g.commands[CMD_LEFT] = state
	case window.KeyD:
		g.commands[CMD_RIGHT] = state
	}
}

// LoadGopher loads the gopher model and adds to it the sound players associated to it
func (g *Game) LoadGopher() *core.Node {
	log.Debug("Decoding gopher model...")

	// Decode model in OBJ format
	dec, err := obj.Decode("./gopher/gopher.obj", "./gopher/gopher.mtl")
	if err != nil {
		panic(err.Error())
	}

	// Create a new node with all the objects in the decoded file and adds it to the scene
	gopherTop, err := dec.NewGroup()
	if err != nil {
		panic(err.Error())
	}

	node := core.NewNode()
	node.SetName("gopher")
	node.Add(gopherTop)

	log.Debug("Done decoding gopher model")
	return node
}
