package main

import (
	"fmt"
	"math"
	"time"
)

func gameLoop() {
	hideCursor()

	// Game timing
	tickRate := time.Second / 30 // 30 FPS
	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	// Input channel
	keyChan := make(chan byte, 10)
	go func() {
		for {
			key, ok := readKey()
			if ok {
				keyChan <- key
			}
		}
	}()

	// Main loop
	for {
		select {
		case key := <-keyChan:
			if game.GameOver || game.Won {
				if key == 'r' {
					initGame()
					continue
				}
				if key == 'q' || key == 3 { // Ctrl+C
					restoreTerminal()
					fmt.Println("Thanks for playing!")
					return
				}
				continue
			}

			if key == 'q' || key == 3 {
				restoreTerminal()
				fmt.Println("Thanks for playing!")
				return
			}

			if game.Paused {
				if key == 'p' || key == 27 {
					game.Paused = false
				}
				continue
			}

			handleInput(key)

		case <-ticker.C:
			if game.GameOver || game.Won || game.Paused {
				continue
			}

			update()
			render()
		}
	}
}

func handleInput(key byte) {
	moveSpeed := 0.08
	turnSpeed := 0.05

	switch key {
	case 'w', 'W':
		// Move forward
		newX := game.PlayerX + math.Cos(game.PlayerAngle)*moveSpeed
		newY := game.PlayerY + math.Sin(game.PlayerAngle)*moveSpeed
		if canMove(newX, newY) {
			game.PlayerX = newX
			game.PlayerY = newY
		}

	case 's', 'S':
		// Move backward
		newX := game.PlayerX - math.Cos(game.PlayerAngle)*moveSpeed
		newY := game.PlayerY - math.Sin(game.PlayerAngle)*moveSpeed
		if canMove(newX, newY) {
			game.PlayerX = newX
			game.PlayerY = newY
		}

	case 'a', 'A':
		// Strafe left
		newX := game.PlayerX + math.Cos(game.PlayerAngle-math.Pi/2)*moveSpeed
		newY := game.PlayerY + math.Sin(game.PlayerAngle-math.Pi/2)*moveSpeed
		if canMove(newX, newY) {
			game.PlayerX = newX
			game.PlayerY = newY
		}

	case 'd', 'D':
		// Strafe right
		newX := game.PlayerX + math.Cos(game.PlayerAngle+math.Pi/2)*moveSpeed
		newY := game.PlayerY + math.Sin(game.PlayerAngle+math.Pi/2)*moveSpeed
		if canMove(newX, newY) {
			game.PlayerX = newX
			game.PlayerY = newY
		}

	case 'j', 'J':
		// Turn left
		game.PlayerAngle -= turnSpeed
		if game.PlayerAngle < 0 {
			game.PlayerAngle += 2 * math.Pi
		}

	case 'l', 'L':
		// Turn right
		game.PlayerAngle += turnSpeed
		if game.PlayerAngle >= 2*math.Pi {
			game.PlayerAngle -= 2 * math.Pi
		}

	case ' ', 'f', 'F':
		// Shoot
		shoot()

	case '1':
		game.Weapon = WeaponPistol
	case '2':
		game.Weapon = WeaponShotgun
	case '3':
		game.Weapon = WeaponPlasma

	case 'm', 'M':
		game.MinimapOn = !game.MinimapOn

	case 'p', 'P', 27: // ESC
		game.Paused = true

	case 'e', 'E':
		// Interact (open doors)
		tryOpenDoor()
	}
}

func canMove(x, y float64) bool {
	// Check corners
	margin := 0.2
	corners := [][2]float64{
		{x - margin, y - margin},
		{x + margin, y - margin},
		{x - margin, y + margin},
		{x + margin, y + margin},
	}

	for _, c := range corners {
		mx, my := int(c[0]), int(c[1])
		if mx < 0 || mx >= MapWidth || my < 0 || my >= MapHeight {
			return false
		}
		wt := game.Map[my][mx]
		if wt != WallEmpty {
			if wt == WallDoor {
				key := [2]int{mx, my}
				if !game.DoorOpen[key] {
					return false
				}
			} else {
				return false
			}
		}
	}
	return true
}

func shoot() {
	if game.Ammo <= 0 {
		beep()
		return
	}

	game.IsShooting = true
	game.ShootFrame = 5
	game.Ammo--

	// Hitscan
	var hitEntity *Entity
	hitDist := math.MaxFloat64

	// Cast a ray forward
	rayX := math.Cos(game.PlayerAngle)
	rayY := math.Sin(game.PlayerAngle)

	// Check entities
	for i := range game.Entities {
		e := &game.Entities[i]
		if !e.Alive {
			continue
		}

		dx := e.X - game.PlayerX
		dy := e.Y - game.PlayerY

		// Project onto ray
		dot := dx*rayX + dy*rayY
		if dot < 0 {
			continue
		}

		// Perpendicular distance
		perpDist := math.Abs(dx*rayY - dy*rayX)

		// Check if in FOV (generous for gameplay)
		if perpDist > 0.5 {
			continue
		}

		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < hitDist {
			hitDist = dist
			hitEntity = e
		}
	}

	if hitEntity != nil {
		// Calculate damage based on distance
		damage := 0
		switch game.Weapon {
		case WeaponPistol:
			damage = int(25.0 / hitDist)
		case WeaponShotgun:
			damage = int(50.0 / hitDist)
		case WeaponPlasma:
			damage = int(40.0 / hitDist)
		}

		if damage < 1 {
			damage = 1
		}

		hitEntity.Health -= damage
		hitEntity.DamageFlash = 3

		if hitEntity.Health <= 0 {
			hitEntity.Alive = false
			hitEntity.Dying = 10

			// Score
			switch hitEntity.Type {
			case EntityDemon:
				game.Score += 100
			case EntityImp:
				game.Score += 75
			case EntityZombie:
				game.Score += 50
			case EntityBarrel:
				game.Score += 25
			}
		}
	}
}

func tryOpenDoor() {
	// Check for doors in front of player
	checkX := int(game.PlayerX + math.Cos(game.PlayerAngle)*1.5)
	checkY := int(game.PlayerY + math.Sin(game.PlayerAngle)*1.5)

	if checkX >= 0 && checkX < MapWidth && checkY >= 0 && checkY < MapHeight {
		if game.Map[checkY][checkX] == WallDoor {
			game.DoorOpen[[2]int{checkX, checkY}] = true
			beep()
		}
	}
}

func update() {
	// Decrease shoot animation
	if game.ShootFrame > 0 {
		game.ShootFrame--
		if game.ShootFrame == 0 {
			game.IsShooting = false
		}
	}

	// Update entities
	for i := range game.Entities {
		e := &game.Entities[i]

		// Dying animation
		if e.Dying > 0 {
			e.Dying--
			continue
		}

		if !e.Alive {
			continue
		}

		// Decrease damage flash
		if e.DamageFlash > 0 {
			e.DamageFlash--
		}

		// Enemy AI
		if e.Type == EntityDemon || e.Type == EntityImp || e.Type == EntityZombie {
			enemyAI(e)
		}

		// Check pickup
		if e.Type == EntityHealth || e.Type == EntityAmmo || e.Type == EntityKey {
			dist := math.Sqrt((e.X-game.PlayerX)*(e.X-game.PlayerX) + (e.Y-game.PlayerY)*(e.Y-game.PlayerY))
			if dist < 0.8 {
				pickupItem(e)
			}
		}
	}

	// Check for exit
	if game.Map[int(game.PlayerY)][int(game.PlayerX)] == WallExit {
		// Next level
		loadLevel(game.Level + 1)
		if game.Won {
			// Victory screen handled in render
		}
	}
}

func enemyAI(e *Entity) {
	dx := game.PlayerX - e.X
	dy := game.PlayerY - e.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	// Only chase if within detection range
	if dist > 10 {
		return
	}

	// Move toward player
	if dist > 1.5 {
		speed := e.Speed
		if dist < 3 {
			speed *= 2
		}

		moveX := dx / dist * speed
		moveY := dy / dist * speed

		newX := e.X + moveX
		newY := e.Y + moveY

		// Check if can move
		mx, my := int(newX), int(newY)
		if mx >= 0 && mx < MapWidth && my >= 0 && my < MapHeight {
			if game.Map[my][mx] == WallEmpty {
				e.X = newX
				e.Y = newY
			}
		}
	}

	// Attack if close enough
	if dist < 2.0 {
		e.AttackCooldown--
		if e.AttackCooldown <= 0 {
			// Deal damage
			damage := 5
			if e.Type == EntityDemon {
				damage = 10
			}
			game.Health -= damage
			e.AttackCooldown = 30

			if game.Health <= 0 {
				game.Health = 0
				game.GameOver = true
			}
		}
	}
}

func pickupItem(e *Entity) {
	switch e.Type {
	case EntityHealth:
		if game.Health < game.MaxHealth {
			game.Health += 25
			if game.Health > game.MaxHealth {
				game.Health = game.MaxHealth
			}
			e.Alive = false
			game.Score += 10
		}
	case EntityAmmo:
		if game.Ammo < game.MaxAmmo {
			game.Ammo += 20
			if game.Ammo > game.MaxAmmo {
				game.Ammo = game.MaxAmmo
			}
			e.Alive = false
			game.Score += 10
		}
	case EntityKey:
		// Find which key color
		e.Alive = false
		game.Score += 50
	}
}
