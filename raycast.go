package main

import (
	"math"
)

// Raycast performs DDA raycasting for one column
type RaycastResult struct {
	Dist     float64
	WallType int
	Side     int // 0=NS, 1=EW
	HitX     float64
}

func raycast(col int) RaycastResult {
	cameraX := 2*float64(col)/float64(ScreenWidth) - 1
	rayDirX := math.Cos(game.PlayerAngle) + math.Sin(game.PlayerAngle+math.Pi/2)*cameraX*0.66
	rayDirY := math.Sin(game.PlayerAngle) - math.Cos(game.PlayerAngle+math.Pi/2)*cameraX*0.66

	mapX := int(game.PlayerX)
	mapY := int(game.PlayerY)

	deltaDistX := math.Abs(1.0 / rayDirX)
	deltaDistY := math.Abs(1.0 / rayDirY)

	var stepX, stepY int
	var sideDistX, sideDistY float64

	if rayDirX < 0 {
		stepX = -1
		sideDistX = (game.PlayerX - float64(mapX)) * deltaDistX
	} else {
		stepX = 1
		sideDistX = (float64(mapX+1) - game.PlayerX) * deltaDistX
	}

	if rayDirY < 0 {
		stepY = -1
		sideDistY = (game.PlayerY - float64(mapY)) * deltaDistY
	} else {
		stepY = 1
		sideDistY = (float64(mapY+1) - game.PlayerY) * deltaDistY
	}

	// DDA
	side := 0
	hit := false
	var wallType int

	for !hit {
		if sideDistX < sideDistY {
			sideDistX += deltaDistX
			mapX += stepX
			side = 0
		} else {
			sideDistY += deltaDistY
			mapY += stepY
			side = 1
		}

		if mapX < 0 || mapX >= MapWidth || mapY < 0 || mapY >= MapHeight {
			break
		}

		wt := game.Map[mapY][mapX]
		if wt != WallEmpty {
			// Check if it's a door
			if wt == WallDoor {
				key := [2]int{mapX, mapY}
				if game.DoorOpen[key] {
					continue
				}
			}
			hit = true
			wallType = wt
		}
	}

	var perpWallDist float64
	if side == 0 {
		perpWallDist = (float64(mapX) - game.PlayerX + (1-float64(stepX))/2) / rayDirX
	} else {
		perpWallDist = (float64(mapY) - game.PlayerY + (1-float64(stepY))/2) / rayDirY
	}

	if perpWallDist < 0.01 {
		perpWallDist = 0.01
	}

	var hitX float64
	if side == 0 {
		hitX = game.PlayerY + perpWallDist*rayDirY
	} else {
		hitX = game.PlayerX + perpWallDist*rayDirX
	}
	hitX -= math.Floor(hitX)

	return RaycastResult{
		Dist:     perpWallDist,
		WallType: wallType,
		Side:     side,
		HitX:     hitX,
	}
}
