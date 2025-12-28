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

type Node struct {
	Name      string
	Outgoing  []string
	Incoming  []string
	Pos       rl.Vector2
	Vel       rl.Vector2
	LinkCount int
	Color     color.RGBA
}

func (n *Node) Links(graph map[string]*Node) []*Node {
	out := []*Node{}
	for _, link := range n.Incoming {
		if target, ok := graph[link]; ok {
			out = append(out, target)
		}
	}
	for _, link := range n.Outgoing {
		if target, ok := graph[link]; ok {
			out = append(out, target)
		}
	}
	return out
}

func (n *Node) Mass() float32 {
	return float32(5 + math.Log(float64(n.LinkCount+1)))
}
func (n *Node) Radius() float32 {
	return float32(5 + math.Log(float64(n.LinkCount+1)))
}

func (n *Node) IsHovered(camera rl.Camera2D) bool {
	mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	return rl.Vector2Length(rl.Vector2Subtract(mouseWorldPos, n.Pos)) < n.Mass()
}
func (n *Node) Overlap(b *Node) (bool, float32, rl.Vector2) {
	delta := rl.Vector2Subtract(b.Pos, n.Pos)
	dist := rl.Vector2Length(delta)
	minDist := n.Radius() + b.Radius()

	if dist == 0 {
		return true, 0, rl.NewVector2(1, 0)
	}

	return dist < minDist, dist, rl.Vector2Scale(delta, 1.0/dist)
}

var gravityStrength = float32(0.00000001)
var repulsionStrength = float32(500)
var connectionStrength = float32(0.02)
var connectionLength = float32(150)
var damping = 0.9

func graphStep(graph map[string]*Node) error {
	for _, node := range graph {
		force := rl.NewVector2(0, 0)

		gravDist := rl.Vector2Subtract(rl.NewVector2(0, 0), node.Pos)
		if rl.Vector2Length(gravDist) > 0.01 {
			toAdd := rl.Vector2Scale(
				gravDist,
				gravityStrength,
			)
			if !math.IsNaN(float64(toAdd.X)) {
				force = rl.Vector2Add(force, toAdd)
			} else {
				panic("nan grav")
			}
		}

		// hookes law
		for _, target := range node.Links(graph) {
			if target.Name == node.Name {
				continue
			}

			dist := rl.Vector2Subtract(node.Pos, target.Pos)
			if rl.Vector2Length(dist) == 0 {
				continue
			}

			toAdd := rl.Vector2Scale(
				rl.Vector2Normalize(dist),
				connectionStrength*(connectionLength-rl.Vector2Length(dist)),
			)
			if math.IsNaN(float64(toAdd.X)) {
				panic("nan hookes")
			}

			force = rl.Vector2Add(force, toAdd)
		}

		// repulsion
		for _, target := range graph {
			if target.Name == node.Name {
				continue
			}

			dir := rl.Vector2Subtract(node.Pos, target.Pos)
			distance := rl.Vector2Length(dir)

			if distance <= 1 {
				distance = 1
			}

			repulseCharge := repulsionStrength // float32(node.Mass()) * float32(target.Mass())

			toAdd := rl.Vector2Scale(rl.Vector2Normalize(dir), repulseCharge/(distance*distance))
			if math.IsNaN(float64(toAdd.X)) {
				fmt.Printf("%s -> %s NaN force\n", node.Name, target.Name)
				panic("nan repulse")
			}

			force = rl.Vector2Add(
				force,
				toAdd,
			)
		}

		// fmt.Println(force)
		node.Vel = rl.Vector2Add(
			node.Vel,
			rl.Vector2Scale(force, 1.0/float32(node.Mass())),
		)
		node.Vel = rl.Vector2Scale(node.Vel, float32(damping))
		// node.Pos = rl.Vector2Add(node.Pos, rl.NewVector2(rand.Float32(), rand.Float32()))
	}

	for _, node := range graph {
		node.Pos = rl.Vector2Add(node.Pos, node.Vel)
	}

	for _, node := range graph {
		for _, target := range graph {
			if node.Name == target.Name {
				continue
			}
			overlap, dist, normal := node.Overlap(target)
			if !overlap {
				continue
			}
			penetration := (node.Radius() + target.Radius()) - dist
			if penetration > 0 {
				correction := rl.Vector2Scale(normal, penetration*0.5)

				node.Pos = rl.Vector2Subtract(node.Pos, correction)
				target.Pos = rl.Vector2Add(target.Pos, correction)
			}
			relativeVel := rl.Vector2Subtract(target.Vel, node.Vel)
			velAlongNormal := rl.Vector2DotProduct(relativeVel, normal)

			if velAlongNormal > 0 {
				restitution := float32(0.3) // 0 = sticky, 1 = bouncy

				impulseMag := -(1 + restitution) * velAlongNormal
				impulseMag /= (1 / node.Mass()) + (1 / target.Mass())

				impulse := rl.Vector2Scale(normal, impulseMag)

				node.Vel = rl.Vector2Subtract(node.Vel, rl.Vector2Scale(impulse, 1/node.Mass()))
				target.Vel = rl.Vector2Add(target.Vel, rl.Vector2Scale(impulse, 1/target.Mass()))
			}

		}
	}

	return nil
}

func randVec() rl.Vector2 {
	return rl.NewVector2(rand.Float32()*500-250, rand.Float32()*500-250)
}

func main() {
	linkmap := extractVaultGraph("/home/dormierian/Documents/ObsidianVault/")
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
	ticks := 0

	for !rl.WindowShouldClose() {
		ticks++
		if ticks < 500 {
			for _ = range 10 {
				graphStep(graph)
			}
		}

		graphStep(graph)
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
				// draggedNode.Pos = rl.Vector2Subtract(draggedNode.Pos, delta)
				// draggedNode.Vel = rl.NewVector2(0, 0)
				draggedNode.Vel = rl.Vector2Scale(delta, -1)
			} else {
				camera.Target = rl.Vector2Add(camera.Target, delta)
			}
		}

		if wheel := rl.GetMouseWheelMove(); wheel != 0 {
			mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
			camera.Offset = rl.GetMousePosition()
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
