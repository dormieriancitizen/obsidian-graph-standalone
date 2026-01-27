package main

import (
	"image/color"
	"math"
	"math/rand/v2"
	"os"
	"strings"
	"unicode"

	ctp "github.com/catppuccin/go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type State struct {
	graph       []*Node
	camera      rl.Camera2D
	paused      bool
	draggedNode *Node
	ticks       int
	search      string
	searching   bool
}

func colorize(ctpColor ctp.Color) color.RGBA {
	return color.RGBA{
		R: ctpColor.RGB[0],
		G: ctpColor.RGB[1],
		B: ctpColor.RGB[2],
		A: 0xff,
	}
}

func randVec() rl.Vector2 {
	return rl.NewVector2(rand.Float32()*5000-2500, rand.Float32()*5000-2500)
}

func doMove(state *State) {
	cameraMoveSpeed := float32(9.0)
	if rl.IsKeyDown(rl.KeyH) || rl.IsKeyDown(rl.KeyA) {
		state.camera.Target.X -= cameraMoveSpeed / state.camera.Zoom
	}
	if rl.IsKeyDown(rl.KeyJ) || rl.IsKeyDown(rl.KeyS) {
		state.camera.Target.Y += cameraMoveSpeed / state.camera.Zoom
	}
	if rl.IsKeyDown(rl.KeyK) || rl.IsKeyDown(rl.KeyW) {
		state.camera.Target.Y -= cameraMoveSpeed / state.camera.Zoom
	}
	if rl.IsKeyDown(rl.KeyL) || rl.IsKeyDown(rl.KeyD) {
		state.camera.Target.X += cameraMoveSpeed / state.camera.Zoom
	}
	if rl.IsKeyPressed(rl.KeySpace) {
		state.paused = !state.paused
	}
	if rl.IsKeyPressed(rl.KeyZero) {
		state.camera.Target = rl.NewVector2(0, 0)
	}
	if rl.IsKeyPressed(rl.KeySlash) {
		state.searching = true
		state.search = ""
	}
}
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
func deleteBack(s string) string {
	r := []rune(s)
	i := len(r)

	for i > 0 && unicode.IsSpace(r[i-1]) {
		i--
	}

	for i > 0 && isWordChar(r[i-1]) {
		i--
	}

	return string(r[:i])
}

func doSearch(state *State) {
	if rl.IsKeyPressed(rl.KeyEnter) {
		state.searching = false
		return
	}
	if rl.IsKeyPressed(rl.KeyBackspace) {
		if len(state.search) > 0 {
			r := []rune(state.search)
			state.search = string(r[:len(r)-1])
			return
		} else {
			state.searching = false
			return
		}
	}
	if (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) && (rl.IsKeyPressed(rl.KeyW) || rl.IsKeyPressed(rl.KeyBackspace)) {
		state.search = deleteBack(state.search)
		return
	}

	for {
		key := rl.GetCharPressed()
		if key == 0 {
			break
		}
		if key >= 32 && key <= 126 {
			state.search += string(key)
		}
	}
}

func doInput(state *State) {
	if !state.searching {
		doMove(state)
	} else {
		doSearch(state)
	}

	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		for _, node := range state.graph {
			if node.IsHovered(state.camera) {
				state.draggedNode = node
				break
			}
		}
	}
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		state.draggedNode = nil
	}

	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		delta := rl.Vector2Scale(rl.GetMouseDelta(), -1.0/state.camera.Zoom)

		if state.draggedNode != nil {
			// draggedNode.Vel = rl.NewVector2(0, 0)
			state.draggedNode.Pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), state.camera)
			state.draggedNode.Vel = rl.Vector2Scale(delta, -1)
		} else {
			state.camera.Target = rl.Vector2Add(state.camera.Target, delta)
		}
	}

	if wheel := rl.GetMouseWheelMove(); wheel != 0 {
		mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), state.camera)
		if wheel < 0 {
			state.camera.Offset = rl.GetMousePosition()
		}
		state.camera.Target = mouseWorldPos

		scale := 0.2 * wheel
		state.camera.Zoom = rl.Clamp(float32(math.Exp(math.Log(float64(state.camera.Zoom))+float64(scale))), 0.125, 64.0)
	}
}

func main() {
	// TODO:
	// - optimise it all
	flavour := ctp.Frappe

	if len(os.Args) < 2 {
		panic("No path to vault passed")
	}

	graph := extractVaultGraph(os.Args[1], flavour)

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(800, 450, "Vault Graph")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	state := State{
		graph: graph,
		camera: rl.Camera2D{
			Target:   rl.NewVector2(0, 0),
			Offset:   rl.NewVector2(400, 225),
			Rotation: 0,
			Zoom:     1.0,
		},
	}

	for !rl.WindowShouldClose() {
		if !state.paused {
			state.ticks++
			if state.ticks < 500 {
				for _ = range 10 {
					graphStep(graph)
				}
			}

			graphStep(state.graph)
		}

		doInput(&state)

		rl.BeginDrawing()
		rl.ClearBackground(colorize(flavour.Base()))
		rl.BeginMode2D(state.camera)

		hovered := false
		for _, node := range graph {
			for _, target := range node.Outgoing {
				var color color.RGBA
				if node.IsHovered(state.camera) || target.IsHovered(state.camera) {
					color = colorize(flavour.Teal())
					hovered = true
				} else {
					color = colorize(flavour.Surface0())
				}
				rl.DrawLineEx(node.Pos, target.Pos, 2, color)
			}
		}

		for _, node := range state.graph {
			dimmed := false
			highlight := false

			if hovered {
				highlight = true
				dimmed = !node.IsHovered(state.camera)
				for _, target := range node.Links() {
					if target.IsHovered(state.camera) {
						dimmed = false
					}
				}
			}

			if state.searching {
				highlight = true
				dimmed = !strings.Contains(node.Name, state.search) && len(state.search) > 0
			}

			if dimmed {
				rl.DrawCircleV(node.Pos, float32(node.Radius()), colorize(flavour.Crust()))
				alphad := color.RGBA{
					R: node.Color.R,
					G: node.Color.G,
					B: node.Color.B,
					A: 255 / 3,
				}
				rl.DrawCircleV(node.Pos, float32(node.Radius()-1), alphad)
			} else if highlight {
				rl.DrawCircleV(node.Pos, float32(node.Radius()+4), colorize(flavour.Mauve()))
				rl.DrawCircleV(node.Pos, float32(node.Radius()+2), node.Color)
			} else {
				rl.DrawCircleV(node.Pos, float32(node.Radius()), colorize(flavour.Crust()))
				rl.DrawCircleV(node.Pos, float32(node.Radius()-1), node.Color)

			}

			rl.DrawText(node.Name, int32(node.Pos.X), int32(node.Pos.Y), 10, colorize(flavour.Text()))
		}

		rl.EndMode2D()
		if state.searching {
			rl.DrawText("/"+state.search+"|", 0, 0, 20, colorize(flavour.Text()))
		}
		rl.EndDrawing()
	}
}
