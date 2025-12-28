package main

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var gravityStrength = float32(0.0005)
var repulsionStrength = float32(500)
var connectionStrength = float32(0.02)
var connectionLength = float32(150)
var damping = 0.97

func graphStep(graph []*Node) error {
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
		for _, target := range node.Links() {
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

			if distance <= 10 {
				distance = 10
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

		node.Vel = rl.Vector2Add(
			node.Vel,
			rl.Vector2Scale(force, 1.0/float32(node.Mass())),
		)
		node.Vel = rl.Vector2Scale(node.Vel, float32(damping))
	}

	for _, node := range graph {
		node.Pos = rl.Vector2Add(node.Pos, node.Vel)
	}

	return nil
}
