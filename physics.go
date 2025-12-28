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
