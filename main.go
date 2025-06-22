package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Coord represents a coordinate on the board.
type Coord struct {
	X int
	Y int
}

// Cell represents a single grid cell.
type Cell struct {
	Filled bool
	Color  color.Color
}

// Block is a falling tetromino travelling toward the center.
type Block struct {
	Shape [][]bool
	Pos   Coord
	Path  []Coord
	Color color.Color
}

// Particle represents a small visual effect when a line clears.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   int
	Color  color.Color
}

// Game holds all state for Gravity Well Tetris.
type Game struct {
	Grid     [49][49]Cell
	Active   *Block
	Rotation int
	Tick     int
	GameOver bool

	// Animated rotation fields
	AnimatingRotation bool
	RotationTarget    int
	RotationProgress  float64

	// Score state
	Score int
	Level int

	// Particles
	Particles []Particle
}

// Predefined tetromino shapes.
var tetrominoes = [][][]bool{
	// I
	{{true, true, true, true}},
	// O
	{{true, true}, {true, true}},
	// T
	{{true, true, true}, {false, true, false}},
	// S
	{{false, true, true}, {true, true, false}},
	// Z
	{{true, true, false}, {false, true, true}},
	// J
	{{true, false, false}, {true, true, true}},
	// L
	{{false, false, true}, {true, true, true}},
}

var colors = []color.Color{
	color.RGBA{0x00, 0xff, 0xff, 0xff}, // cyan
	color.RGBA{0xff, 0xff, 0x00, 0xff}, // yellow
	color.RGBA{0xaa, 0x00, 0xaa, 0xff}, // purple
	color.RGBA{0x00, 0xff, 0x00, 0xff}, // green
	color.RGBA{0xff, 0x00, 0x00, 0xff}, // red
	color.RGBA{0x00, 0x00, 0xff, 0xff}, // blue
	color.RGBA{0xff, 0x7f, 0x00, 0xff}, // orange
}

func randomColor() color.Color {
	return colors[rand.Intn(len(colors))]
}

// rotateCoord90CW rotates a coordinate 90 degrees clockwise around the center.
func rotateCoord90CW(c Coord) Coord {
	return Coord{X: 48 - c.Y, Y: c.X}
}

// rotateBoard rotates the grid, active block and paths.
func (g *Game) rotateBoard(deg int) {
	steps := ((deg % 360) + 360) % 360 / 90
	for i := 0; i < steps; i++ {
		var newGrid [49][49]Cell
		for y := 0; y < 49; y++ {
			for x := 0; x < 49; x++ {
				nc := rotateCoord90CW(Coord{X: x, Y: y})
				newGrid[nc.Y][nc.X] = g.Grid[y][x]
			}
		}
		g.Grid = newGrid
		if g.Active != nil {
			g.Active.Pos = rotateCoord90CW(g.Active.Pos)
			for j := range g.Active.Path {
				g.Active.Path[j] = rotateCoord90CW(g.Active.Path[j])
			}
		}
		g.Rotation = (g.Rotation + 90) % 360
	}
}

// collides checks if a block placed at pos would collide with the board.
func (g *Game) collides(pos Coord, shape [][]bool) bool {
	for y := 0; y < len(shape); y++ {
		for x := 0; x < len(shape[0]); x++ {
			if !shape[y][x] {
				continue
			}
			gx := pos.X + x
			gy := pos.Y + y
			if gx < 0 || gx >= 49 || gy < 0 || gy >= 49 {
				return true
			}
			if g.Grid[gy][gx].Filled {
				return true
			}
		}
	}
	return false
}

// bresenham returns coordinates along a line from start to end (exclusive of start).
func bresenham(start, end Coord) []Coord {
	var pts []Coord
	x0, y0 := start.X, start.Y
	x1, y1 := end.X, end.Y
	dx := abs(x1 - x0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	dy := -abs(y1 - y0)
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for x0 != x1 || y0 != y1 {
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
		pts = append(pts, Coord{X: x0, Y: y0})
	}
	return pts
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// lockBlock merges the active block into the grid.
func (g *Game) lockBlock() {
	if g.Active == nil {
		return
	}
	for y := 0; y < len(g.Active.Shape); y++ {
		for x := 0; x < len(g.Active.Shape[0]); x++ {
			if !g.Active.Shape[y][x] {
				continue
			}
			gx := g.Active.Pos.X + x
			gy := g.Active.Pos.Y + y
			if gx >= 0 && gx < 49 && gy >= 0 && gy < 49 {
				g.Grid[gy][gx] = Cell{Filled: true, Color: g.Active.Color}
			}
		}
	}
	g.Active = nil
}

// spawnBlock creates a new block at a random edge and computes its path.
func (g *Game) spawnBlock() {
	idx := rand.Intn(len(tetrominoes))
	shape := tetrominoes[idx]
	clr := colors[idx%len(colors)]
	w := len(shape[0])
	h := len(shape)

	var pos Coord
	side := rand.Intn(4)
	switch side {
	case 0: // top
		pos = Coord{X: rand.Intn(49 - w), Y: 0}
	case 1: // bottom
		pos = Coord{X: rand.Intn(49 - w), Y: 49 - h}
	case 2: // left
		pos = Coord{X: 0, Y: rand.Intn(49 - h)}
	case 3: // right
		pos = Coord{X: 49 - w, Y: rand.Intn(49 - h)}
	}

	center := Coord{X: 24 - w/2, Y: 24 - h/2}
	path := bresenham(pos, center)

	if g.collides(pos, shape) {
		g.GameOver = true
		return
	}
	g.Active = &Block{Shape: shape, Pos: pos, Path: path, Color: clr}
}

// step advances the active block one step along its path.
func (g *Game) step() {
	if g.Active == nil {
		return
	}
	if len(g.Active.Path) == 0 {
		g.lockBlock()
		g.clearLines()
		g.spawnBlock()
		return
	}
	next := g.Active.Path[0]
	if g.collides(next, g.Active.Shape) {
		g.lockBlock()
		g.clearLines()
		g.spawnBlock()
		return
	}
	g.Active.Pos = next
	g.Active.Path = g.Active.Path[1:]
}

// StartRotation initiates a smooth 90 degree rotation.
func (g *Game) StartRotation(degrees int) {
	if g.AnimatingRotation {
		return
	}
	g.AnimatingRotation = true
	g.RotationTarget = (g.Rotation + degrees) % 360
	g.RotationProgress = 0
}

// Update handles input and game logic each frame.
func (g *Game) Update() error {
	if g.GameOver {
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.StartRotation(270)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.StartRotation(90)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.step()
	}

	if g.AnimatingRotation {
		g.RotationProgress += 6
		if g.RotationProgress >= 90 {
			g.rotateBoard((g.RotationTarget - g.Rotation + 360) % 360)
			g.AnimatingRotation = false
			g.RotationProgress = 0
		}
		return nil
	}

	g.Tick++
	delay := 30 - g.Level*2
	if delay < 5 {
		delay = 5
	}
	if g.Tick%delay == 0 {
		g.step()
	}

	for i := range g.Particles {
		g.Particles[i].X += g.Particles[i].VX
		g.Particles[i].Y += g.Particles[i].VY
		g.Particles[i].Life--
	}
	g.Particles = filterParticles(g.Particles)

	return nil
}

func filterParticles(p []Particle) []Particle {
	out := p[:0]
	for _, pt := range p {
		if pt.Life > 0 {
			out = append(out, pt)
		}
	}
	return out
}

// clearLines removes filled rows or columns and spawns particles.
func (g *Game) clearLines() {
	linesCleared := 0
	for y := 0; y < 49; y++ {
		full := true
		for x := 0; x < 49; x++ {
			if !g.Grid[y][x].Filled {
				full = false
				break
			}
		}
		if full {
			for x := 0; x < 49; x++ {
				g.Grid[y][x] = Cell{}
				g.spawnParticles(x, y)
			}
			linesCleared++
		}
	}
	for x := 0; x < 49; x++ {
		full := true
		for y := 0; y < 49; y++ {
			if !g.Grid[y][x].Filled {
				full = false
				break
			}
		}
		if full {
			for y := 0; y < 49; y++ {
				g.Grid[y][x] = Cell{}
				g.spawnParticles(x, y)
			}
			linesCleared++
		}
	}
	g.Score += linesCleared * 100
	if linesCleared > 0 && g.Score/500 > g.Level {
		g.Level++
	}
}

func (g *Game) spawnParticles(x, y int) {
	for i := 0; i < 5; i++ {
		angle := rand.Float64() * 2 * math.Pi
		speed := rand.Float64() * 2
		g.Particles = append(g.Particles, Particle{
			X:     float64(x*16 + 8),
			Y:     float64(y*16 + 8),
			VX:    math.Cos(angle) * speed,
			VY:    math.Sin(angle) * speed,
			Life:  30 + rand.Intn(30),
			Color: randomColor(),
		})
	}
}

// Draw renders the entire game scene.
func (g *Game) Draw(screen *ebiten.Image) {
	size := 16
	angle := float64(g.Rotation+int(g.RotationProgress)) * math.Pi / 180
	geom := ebiten.GeoM{}
	geom.Translate(-392, -392)
	geom.Rotate(angle)
	geom.Translate(392, 392)

	for y := 0; y < 49; y++ {
		for x := 0; x < 49; x++ {
			if g.Grid[y][x].Filled {
				op := &ebiten.DrawImageOptions{}
				op.GeoM = geom
				op.GeoM.Translate(float64(x*size), float64(y*size))
				rect := ebiten.NewImage(size, size)
				rect.Fill(g.Grid[y][x].Color)
				screen.DrawImage(rect, op)
			}
		}
	}

	if g.Active != nil {
		for y := 0; y < len(g.Active.Shape); y++ {
			for x := 0; x < len(g.Active.Shape[0]); x++ {
				if g.Active.Shape[y][x] {
					gx := g.Active.Pos.X + x
					gy := g.Active.Pos.Y + y
					op := &ebiten.DrawImageOptions{}
					op.GeoM = geom
					op.GeoM.Translate(float64(gx*size), float64(gy*size))
					rect := ebiten.NewImage(size, size)
					rect.Fill(g.Active.Color)
					screen.DrawImage(rect, op)
				}
			}
		}
	}

	for _, p := range g.Particles {
		ebitenutil.DrawRect(screen, p.X, p.Y, 2, 2, p.Color)
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.Score), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d", g.Level), 10, 30)
}

func (g *Game) Layout(outW, outH int) (int, int) { return 784, 784 }

// newGame creates and initializes a new Game instance.
func newGame() *Game {
	g := &Game{Level: 0}
	g.spawnBlock()
	return g
}

func main() {
	rand.Seed(time.Now().UnixNano())
	ebiten.SetWindowSize(784, 784)
	ebiten.SetWindowTitle("Gravity Well Tetris")
	if err := ebiten.RunGame(newGame()); err != nil {
		panic(err)
	}
}
