package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"
	"strings"

	ctp "github.com/catppuccin/go"
	rl "github.com/gen2brain/raylib-go/raylib"
)

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

func main() {
	// TODO:
	// - move all the physics to own file
	// - make nodes point with pointers instead of ids
	// - optimise it all
	linkmap := extractVaultGraph("/home/dormierian/Documents/ObsidianVault/")
	// linkmap := extractVaultGraph("/home/dormierian/Downloads/obsidian-developer-docs/")
	flavour := ctp.Frappe

	graph := make(map[string]*Node)

	for name, links := range linkmap {
		graph[name] = &Node{
			Name:     name,
			Outgoing: links,
			Pos:      randVec(),
			Color:    colorize(flavour.Teal()),
		}
	}

	for _, node := range graph {
		for _, link := range node.Outgoing {
			if _, ok := graph[link]; !ok {
				if strings.HasPrefix(link, "#") {
					graph[link] = &Node{
						Name:     link,
						Outgoing: []string{},
						Pos:      randVec(),
						Color:    colorize(flavour.Green()),
					}
				} else {
					graph[link] = &Node{
						Name:     link,
						Outgoing: []string{},
						Pos:      randVec(),
						Color:    colorize(flavour.Subtext0()),
					}
				}
			}
			target := graph[link]
			target.LinkCount++

			target.Incoming = append(target.Incoming, node.Name)
		}
		node.LinkCount++
	}

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(800, 450, "raylib [core] example - basic window")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	camera := rl.Camera2D{
		Target:   rl.NewVector2(0, 0),
		Offset:   rl.NewVector2(400, 225),
		Rotation: 0,
		Zoom:     1.0,
	}

	var draggedNode *Node
	// ticks := 0
	paused := false

	for !rl.WindowShouldClose() {
		if !paused {
			// ticks++
			// if ticks < 500 {
			// 	for _ = range 10 {
			// 		graphStep(graph)
			// 	}
			// }

			graphStep(graph)
		}
		currentWidth := rl.GetScreenWidth()
		currentHeight := rl.GetScreenHeight()

		camera.Offset.X = float32(currentWidth) / 2
		camera.Offset.Y = float32(currentHeight) / 2

		cameraMoveSpeed := float32(9.0)
		if rl.IsKeyDown(rl.KeyH) || rl.IsKeyDown(rl.KeyA) {
			camera.Target.X -= cameraMoveSpeed / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyJ) || rl.IsKeyDown(rl.KeyS) {
			camera.Target.Y += cameraMoveSpeed / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyK) || rl.IsKeyDown(rl.KeyW) {
			camera.Target.Y -= cameraMoveSpeed / camera.Zoom
		}
		if rl.IsKeyDown(rl.KeyL) || rl.IsKeyDown(rl.KeyD) {
			camera.Target.X += cameraMoveSpeed / camera.Zoom
		}
		if rl.IsKeyPressed(rl.KeySpace) {
			paused = !paused
		}
		if rl.IsKeyPressed(rl.KeyZero) {
			camera.Target = rl.NewVector2(0, 0)
		}

		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			for _, node := range graph {
				if node.IsHovered(camera) {
					draggedNode = node
					break
				}
			}
		}
		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
			draggedNode = nil
		}

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			delta := rl.Vector2Scale(rl.GetMouseDelta(), -1.0/camera.Zoom)

			if draggedNode != nil {
				// draggedNode.Vel = rl.NewVector2(0, 0)
				draggedNode.Pos = rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
				draggedNode.Vel = rl.Vector2Scale(delta, -1)
			} else {
				camera.Target = rl.Vector2Add(camera.Target, delta)
			}
		}

		if wheel := rl.GetMouseWheelMove(); wheel != 0 {
			mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
			if wheel < 0 {
				camera.Offset = rl.GetMousePosition()
			}
			camera.Target = mouseWorldPos

			scale := 0.2 * wheel
			camera.Zoom = rl.Clamp(float32(math.Exp(math.Log(float64(camera.Zoom))+float64(scale))), 0.125, 64.0)
		}

		rl.BeginDrawing()
		rl.ClearBackground(colorize(flavour.Base()))

		rl.BeginMode2D(camera)

		for _, node := range graph {
			for _, link := range node.Outgoing {
				target := graph[link]
				var color color.RGBA
				if node.IsHovered(camera) || target.IsHovered(camera) {
					color = colorize(flavour.Teal())
				} else {
					color = colorize(flavour.Surface0())
				}
				rl.DrawLineV(node.Pos, target.Pos, color)
			}
		}

		for _, node := range graph {
			rl.DrawCircleV(node.Pos, float32(node.Radius()), node.Color)
			rl.DrawText(node.Name, int32(node.Pos.X), int32(node.Pos.Y), 10, colorize(flavour.Text()))
		}

		rl.EndMode2D()
		rl.DrawText(fmt.Sprint(camera.Zoom), 0, 0, 20, colorize(flavour.Red()))
		rl.EndDrawing()
	}
}
