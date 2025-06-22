# Gravity Well Tetris - Design Brief

## Overview
Gravity Well Tetris is a novel twist on the classic Tetris formula. Instead of pieces falling downwards, they are pulled toward the center of a square board by a gravity well. The player does not rotate individual blocks — instead, they rotate the entire world. The objective is to form and clear full horizontal or vertical lines, score points, and advance through levels.

## Target
- Should build on macos, windows, linux.

## Core Mechanics

### Board
- Size: 49x49 grid.
- Center: Cell (24,24) is the gravity well.
- Grid is square and symmetrical.

### Block Behavior
- Blocks are standard Tetris tetrominoes.
- Blocks spawn on a random edge and follow a path toward the center.
- Movement follows a straight line (Bresenham path) to the gravity well.
- Blocks stop when they collide or reach the center.

### Rotation
- Player rotates the entire board (not the block).
- Rotation is animated in 90° steps.
- Rotation is clockwise or counterclockwise via arrow keys.

### Clearing
- Full **horizontal or vertical** lines are cleared.
- Cleared cells generate particle effects.
- Score is awarded for each cleared line.

## Input
- **Left / Right Arrow**: Rotate the board (counter/clockwise).
- **Down / Space**: Advance the active block one step.
- **Escape**: Quit the game.

## Scoring and Levels
- 100 points per line cleared.
- Every 500 points increases level.
- Higher levels reduce delay between block movements.

## Visuals & Effects
- Grid blocks rendered with `ebitenutil.DrawRect`.
- Particle effects emitted on cell clearing.
- Smooth rotation using animated transformation.
- HUD displays score and level.

## Technologies
- **Language**: Go
- **Framework**: [Ebiten](https://ebiten.org) for 2D graphics.
- **Platform**: Desktop (cross-platform via Ebiten)

## Files and Structure
- `main.go`: Contains full game logic.
- `Game` struct handles core state.
- `Block`, `Coord`, `Cell`, and `Particle` types define gameplay elements.

## Future Improvements
- Background animations (e.g. gravitational lensing).
- Sound effects and music.
- High score system.
- Piece preview and hold mechanics.
- Game over animation.

---
This brief summarizes the full technical and design scope of the Gravity Well Tetris project.

