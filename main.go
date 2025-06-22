package main
// Gravity Well Tetris - Full Instruction Set for Codex (Enhanced Version)

// ... existing imports ...
import (
    // existing imports
    "github.com/hajimehoshi/ebiten/v2/vector"
    "fmt"
)

// Add fields to Game struct
// ...
type Game struct {
    Grid      [49][49]Cell
    Active    *Block
    Rotation  int // 0, 90, 180, 270
    Tick      int
    GameOver  bool

    // New features
    AnimatingRotation bool
    RotationTarget    int
    RotationProgress  float64
    Score             int
    Level             int
    Particles         []Particle
}

// Add Particle struct

type Particle struct {
    X, Y     float64
    VX, VY   float64
    Life     int
    Color    color.Color
}

// Update rotateBoard to support animation
func (g *Game) StartRotation(degrees int) {
    if g.AnimatingRotation {
        return
    }
    g.AnimatingRotation = true
    g.RotationTarget = (g.Rotation + degrees) % 360
    g.RotationProgress = 0
}

// Modify Update to handle animated rotation
func (g *Game) Update() error {
    if g.GameOver { return nil }

    if ebiten.IsKeyPressed(ebiten.KeyLeft) {
        g.StartRotation(270)
    }
    if ebiten.IsKeyPressed(ebiten.KeyRight) {
        g.StartRotation(90)
    }
    if ebiten.IsKeyPressed(ebiten.KeyEscape) {
        return ebiten.Termination
    }

    if g.AnimatingRotation {
        g.RotationProgress += 6 // degrees per frame
        if g.RotationProgress >= 90 {
            g.rotateBoard((g.RotationTarget - g.Rotation + 360) % 360)
            g.AnimatingRotation = false
            g.RotationProgress = 0
        }
        return nil
    }

    g.Tick++
    if g.Tick%(30 - g.Level*2) == 0 && g.Active != nil {
        if len(g.Active.Path) > 0 {
            g.Active.Pos = g.Active.Path[0]
            g.Active.Path = g.Active.Path[1:]
        } else {
            g.lockBlock()
            g.clearLines()
            g.spawnBlock()
        }
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
    var out []Particle
    for _, pt := range p {
        if pt.Life > 0 {
            out = append(out, pt)
        }
    }
    return out
}

func (g *Game) clearLines() {
    linesCleared := 0
    // Clear filled rows
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
    // Clear filled columns
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
            X: float64(x*16 + 8),
            Y: float64(y*16 + 8),
            VX: math.Cos(angle) * speed,
            VY: math.Sin(angle) * speed,
            Life: 30 + rand.Intn(30),
            Color: randomColor(),
        })
    }
}

func (g *Game) Draw(screen *ebiten.Image) {
    size := 16
    angle := float64(g.Rotation + int(g.RotationProgress)) * math.Pi / 180
    center := ebiten.GeoM{}
    center.Translate(-392, -392)
    center.Rotate(angle)
    center.Translate(392, 392)

    // Draw grid
    for y := 0; y < 49; y++ {
        for x := 0; x < 49; x++ {
            if g.Grid[y][x].Filled {
                op := &ebiten.DrawImageOptions{}
                op.GeoM = center
                op.GeoM.Translate(float64(x*size), float64(y*size))
                rect := ebiten.NewImage(size, size)
                rect.Fill(g.Grid[y][x].Color)
                screen.DrawImage(rect, op)
            }
        }
    }

    // Draw active piece
    if g.Active != nil {
        for y := 0; y < len(g.Active.Shape); y++ {
            for x := 0; x < len(g.Active.Shape[0]); x++ {
                if g.Active.Shape[y][x] {
                    gx := g.Active.Pos.X + x
                    gy := g.Active.Pos.Y + y
                    op := &ebiten.DrawImageOptions{}
                    op.GeoM = center
                    op.GeoM.Translate(float64(gx*size), float64(gy*size))
                    rect := ebiten.NewImage(size, size)
                    rect.Fill(g.Active.Color)
                    screen.DrawImage(rect, op)
                }
            }
        }
    }

    // Draw particles
    for _, p := range g.Particles {
        ebitenutil.DrawRect(screen, p.X, p.Y, 2, 2, p.Color)
    }

    // Draw score and level
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.Score), 10, 10)
    ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level: %d", g.Level), 10, 30)
}

