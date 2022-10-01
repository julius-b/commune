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

	commands [CMD_LAST]bool // commands states
}

// Update is called every frame.
func (g *Game) Update(renderer *renderer.Renderer, deltaTime time.Duration) {

	// Get player world position
	var position math32.Vector3
	g.world.gopherNode.WorldPosition(&position)
	// smooth / wall-sliding collision
	// position.Add(&direction)
	// Get player world direction
	var quat math32.Quaternion
	g.world.gopherNode.WorldQuaternion(&quat)
	fullDir := math32.Vector3{}

	gg := g.world.gravity * float32(deltaTime.Seconds())
	gravity := math32.Vector3{0, -1, 0}
	gravity.ApplyQuaternion(&quat)
	gravity.Normalize()
	gravity.MultiplyScalar(gg)
	//fullDir.Add(&gravity)

	// apply commands to player relative to in-game direction
	if g.commands[CMD_LEFT] || g.commands[CMD_RIGHT] {
		// Calculates angle delta to rotate
		angle := g.world.rotvel * float32(deltaTime.Seconds())
		if g.commands[CMD_RIGHT] {
			angle = -angle
		}
		g.world.gopherNode.RotateY(angle)
	}
	if g.commands[CMD_FORWARD] || g.commands[CMD_BACKWARD] {
		// Calculates the distance to move
		dist := g.world.velocity * float32(deltaTime.Seconds())

		direction := math32.Vector3{1, 0, 0} // player always move towards own x
		direction.ApplyQuaternion(&quat)     // make player forward into world forward
		direction.Normalize()
		direction.MultiplyScalar(dist)
		if g.commands[CMD_BACKWARD] {
			direction.Negate()
		}
		fullDir.Add(&direction)

		log.Debug("direction: " + fmt.Sprint(direction))
	}

	log.Debug("fullDir: " + fmt.Sprint(fullDir))
	var ignored float32
	gbox := g.world.gopherNode
	gbbox := NewAABBoxFromBox3(gbox, fullDir.X, fullDir.Y, fullDir.Z)
	log.Debug("gbbox: " + fmt.Sprint(gbbox))
	children := g.world.scene.Children()
	log.Debug("checking #" + fmt.Sprint(len(children)) + " (-1) collisions")
	collisiontimes := make([]float32, 0, len(children))
	for _, v := range children {
		if v.Name() != "gopher" {
			bbox := v
			bbbox := NewAABBoxFromBox3(bbox.GetNode(), 0, 0, 0)
			log.Debug("bbbox: " + fmt.Sprintf("%#v", bbbox))
			collisiontimes = append(collisiontimes, SweptAABB(gbbox, bbbox, &ignored, &ignored, &ignored))
		}
	}
	log.Debug("collisiontimes: " + fmt.Sprint(collisiontimes))
	mincollisiontime := Min(collisiontimes)
	log.Debug("mincollisiontime: " + fmt.Sprint(mincollisiontime))
	fullDir.MultiplyScalar(mincollisiontime)

	position.Add(&fullDir)
	g.world.gopherNode.SetPositionVec(&position)

	g.camera.LookAt(&position, &math32.Vector3{0, 2, 0})
	g.orbit.SetTarget(position) // updated independently

	cpos := g.camera.Position()
	cpos.Add(&fullDir)
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

	g.world.velocity = 10.0
	g.world.gravity = 9.81
	g.world.rotvel = 2.0
	g.LoadGopher()
	g.world.gopherNode.SetPosition(0, 1.5, -1)
	g.world.scene.Add(g.world.gopherNode)

	// Create camera
	g.camera = camera.New(1)
	//g.camera.SetPosition(0, 1, -2)
	gpos := g.world.gopherNode.Position()
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
func (g *Game) LoadGopher() {
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

	g.world.gopherNode = core.NewNode()
	g.world.gopherNode.SetName("gopher")
	g.world.gopherNode.Add(gopherTop)

	log.Debug("Done decoding gopher model")
}
