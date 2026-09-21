package main

import (
	"fmt"
	"math"
	"strings"
)

type ScreenBuffer struct {
	Chars  [ScreenHeight][ScreenWidth]string
	Colors [ScreenHeight][ScreenWidth]string
	Depth  [ScreenWidth]float64
}

func render() {
	var buf ScreenBuffer

	// Cast rays
	for x := 0; x < ScreenWidth; x++ {
		result := raycast(x)
		buf.Depth[x] = result.Dist

		lineHeight := int(float64(ScreenHeight) / result.Dist)
		if lineHeight > ScreenHeight*2 {
			lineHeight = ScreenHeight * 2
		}

		drawStart := -lineHeight/2 + ScreenHeight/2
		drawEnd := lineHeight/2 + ScreenHeight/2

		if drawStart < 0 {
			drawStart = 0
		}
		if drawEnd >= ScreenHeight {
			drawEnd = ScreenHeight - 1
		}

		wallChar := getWallChar(result.WallType, result.HitX, result.Side)
		distFactor := 1.0 - math.Min(result.Dist/16.0, 0.8)

		for y := drawStart; y <= drawEnd; y++ {
			buf.Chars[y][x] = wallChar
			buf.Colors[y][x] = getWallColor(result.WallType, result.Side, distFactor)
		}
	}

	renderSprites(&buf)
	renderWeapon(&buf)
	renderHUD(&buf)
	if game.MinimapOn {
		renderMinimap(&buf)
	}

	// Output
	var sb strings.Builder
	sb.WriteString("\033[H")

	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			if buf.Colors[y][x] != "" {
				sb.WriteString(buf.Colors[y][x])
			}
			sb.WriteString(buf.Chars[y][x])
		}
		if buf.Colors[y][ScreenWidth-1] != "" {
			sb.WriteString(ColorReset)
		}
		sb.WriteString("\n")
	}

	fmt.Print(sb.String())
}

func getWallChar(wallType int, hitX float64, side int) string {
	chars := wallChars[wallType]
	if chars == [2]string{} {
		return "█"
	}
	if side == 0 {
		if hitX < 0.5 {
			return chars[0]
		}
		return chars[1]
	}
	return chars[1]
}

func getWallColor(wallType int, side int, distFactor float64) string {
	var base string
	switch wallType {
	case WallStone:
		base = "\033[37m"
	case WallBrick:
		base = "\033[33m"
	case WallMetal:
		base = "\033[36m"
	case WallWood:
		base = "\033[32m"
	case WallDoor:
		base = "\033[35m"
	case WallExit:
		base = "\033[1;33m"
	default:
		base = "\033[37m"
	}

	if side == 1 {
		if strings.Contains(base, "37") {
			return "\033[37;2m"
		}
		return base + ";2m"
	}
	return base
}

func renderSprites(buf *ScreenBuffer) {
	type sortedEntity struct {
		entity *Entity
		dist   float64
		angle  float64
	}
	var sprites []sortedEntity

	for i := range game.Entities {
		e := &game.Entities[i]
		if !e.Alive && e.Dying == 0 {
			continue
		}

		dx := e.X - game.PlayerX
		dy := e.Y - game.PlayerY
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 0.5 {
			continue
		}

		angle := math.Atan2(dy, dx) - game.PlayerAngle
		for angle < -math.Pi {
			angle += 2 * math.Pi
		}
		for angle > math.Pi {
			angle -= 2 * math.Pi
		}

		sprites = append(sprites, sortedEntity{e, dist, angle})
	}

	// Sort far to near
	for i := 0; i < len(sprites); i++ {
		for j := i + 1; j < len(sprites); j++ {
			if sprites[i].dist < sprites[j].dist {
				sprites[i], sprites[j] = sprites[j], sprites[i]
			}
		}
	}

	for _, s := range sprites {
		e := s.entity
		if s.angle < -FOV/2 || s.angle > FOV/2 {
			continue
		}

		spriteScreenX := int((0.5 + s.angle/FOV) * ScreenWidth)
		spriteHeight := int(float64(ScreenHeight) / s.dist)
		spriteWidth := spriteHeight / 2
		if spriteWidth < 1 {
			spriteWidth = 1
		}
		if spriteHeight < 1 {
			spriteHeight = 1
		}

		startY := -spriteHeight/2 + ScreenHeight/2
		startX := spriteScreenX - spriteWidth/2

		var char string
		var color string

		if e.Dying > 0 {
			char = "†"
			color = ColorRed
		} else if e.Alive {
			switch e.Type {
			case EntityDemon, EntityImp, EntityZombie:
				char = string(e.Sprite)
				if e.DamageFlash > 0 {
					color = ColorBrightRed
				} else {
					color = ColorRed
				}
			case EntityBarrel:
				char = "█"
				color = ColorYellow
			case EntityHealth:
				char = "+"
				color = ColorGreen
			case EntityAmmo:
				char = "▓"
				color = ColorCyan
			case EntityKey:
				char = "♦"
				color = ColorMagenta
			default:
				char = "?"
				color = ColorWhite
			}
		}

		for sy := 0; sy < spriteHeight; sy++ {
			for sx := 0; sx < spriteWidth; sx++ {
				drawX := startX + sx
				drawY := startY + sy

				if drawX >= 0 && drawX < ScreenWidth && drawY >= 0 && drawY < ScreenHeight {
					if s.dist < buf.Depth[drawX] {
						buf.Chars[drawY][drawX] = char
						buf.Colors[drawY][drawX] = color
					}
				}
			}
		}
	}
}

func renderWeapon(buf *ScreenBuffer) {
	weaponX := ScreenWidth/2 - 5
	weaponY := ScreenHeight - 4

	sprites := weaponSprites[game.Weapon]

	for i, line := range sprites {
		for j, ch := range line {
			x := weaponX + j
			y := weaponY + i

			if x >= 0 && x < ScreenWidth && y >= 0 && y < ScreenHeight {
				if ch != ' ' {
					buf.Chars[y][x] = string(ch)
					buf.Colors[y][x] = ColorWhite
				}
			}
		}
	}

	if game.IsShooting {
		flashX := ScreenWidth / 2
		flashY := ScreenHeight - 5
		if flashX >= 0 && flashX < ScreenWidth && flashY >= 0 && flashY < ScreenHeight {
			buf.Chars[flashY][flashX] = "*"
			buf.Colors[flashY][flashX] = ColorYellow
		}
	}
}

func renderHUD(buf *ScreenBuffer) {
	hudY := ScreenHeight - 1

	healthStr := fmt.Sprintf("HP:%d/%d", game.Health, game.MaxHealth)
	for i, ch := range healthStr {
		buf.Chars[hudY][i] = string(ch)
		if game.Health > 50 {
			buf.Colors[hudY][i] = ColorGreen
		} else if game.Health > 25 {
			buf.Colors[hudY][i] = ColorYellow
		} else {
			buf.Colors[hudY][i] = ColorRed
		}
	}

	ammoStr := fmt.Sprintf("AMMO:%d", game.Ammo)
	ammoX := 15
	for i, ch := range ammoStr {
		buf.Chars[hudY][ammoX+i] = string(ch)
		if game.Ammo > 20 {
			buf.Colors[hudY][ammoX+i] = ColorCyan
		} else {
			buf.Colors[hudY][ammoX+i] = ColorRed
		}
	}

	scoreStr := fmt.Sprintf("SCORE:%d", game.Score)
	scoreX := 30
	for i, ch := range scoreStr {
		buf.Chars[hudY][scoreX+i] = string(ch)
		buf.Colors[hudY][scoreX+i] = ColorYellow
	}

	levelStr := fmt.Sprintf("LVL:%d", game.Level+1)
	levelX := 50
	for i, ch := range levelStr {
		buf.Chars[hudY][levelX+i] = string(ch)
		buf.Colors[hudY][levelX+i] = ColorMagenta
	}

	weaponNames := []string{"PISTOL", "SHOTGUN", "PLASMA"}
	weaponStr := weaponNames[game.Weapon]
	weaponX := 65
	for i, ch := range weaponStr {
		buf.Chars[hudY][weaponX+i] = string(ch)
		buf.Colors[hudY][weaponX+i] = ColorWhite
	}
}

func renderMinimap(buf *ScreenBuffer) {
	mapSize := 10
	startX := 1
	startY := 1

	for x := 0; x <= mapSize+1; x++ {
		buf.Chars[startY-1][startX+x-1] = "-"
		buf.Chars[startY+mapSize][startX+x-1] = "-"
	}
	for y := 0; y <= mapSize+1; y++ {
		buf.Chars[startY+y-1][startX-1] = "|"
		buf.Chars[startY+y-1][startX+mapSize] = "|"
	}

	centerX := int(game.PlayerX)
	centerY := int(game.PlayerY)

	for dy := 0; dy < mapSize; dy++ {
		for dx := 0; dx < mapSize; dx++ {
			mapX := centerX - mapSize/2 + dx
			mapY := centerY - mapSize/2 + dy

			if mapX < 0 || mapX >= MapWidth || mapY < 0 || mapY >= MapHeight {
				continue
			}

			x := startX + dx
			y := startY + dy

			if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
				continue
			}

			if mapX == int(game.PlayerX) && mapY == int(game.PlayerY) {
				buf.Chars[y][x] = "@"
				buf.Colors[y][x] = ColorGreen
				continue
			}

			found := false
			for _, e := range game.Entities {
				if e.Alive && int(e.X) == mapX && int(e.Y) == mapY {
					buf.Chars[y][x] = "!"
					buf.Colors[y][x] = ColorRed
					found = true
					break
				}
			}
			if found {
				continue
			}

			wt := game.Map[mapY][mapX]
			if wt != WallEmpty {
				buf.Chars[y][x] = "#"
				buf.Colors[y][x] = ColorWhite
			} else {
				buf.Chars[y][x] = " "
			}
		}
	}
}
