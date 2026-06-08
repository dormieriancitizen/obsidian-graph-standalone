package main

import (
	"image/color"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Node struct {
	Name        string
	Outgoing    []*Node
	Incoming    []*Node
	Pos         rl.Vector2
	Vel         rl.Vector2
	LinkCount   int
	Color       color.RGBA
	StringLinks []string
}

func (n *Node) Links() []*Node {
	out := []*Node{}
	for _, target := range n.Incoming {
		if target != nil {
			out = append(out, target)
		}
	}
	for _, target := range n.Outgoing {
		if target != nil {
			out = append(out, target)
		}
	}
	return out
}

func (n *Node) Mass() float32 {
	return float32(5 + math.Log(float64(n.LinkCount+1)))
}
func (n *Node) Radius() float32 {
	return float32(5 + math.Sqrt(float64(n.LinkCount+1)))
}

func (n *Node) IsHovered(camera rl.Camera2D) bool {
	mouseWorldPos := rl.GetScreenToWorld2D(rl.GetMousePosition(), camera)
	return rl.Vector2Length(rl.Vector2Subtract(mouseWorldPos, n.Pos)) < n.Radius()
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
