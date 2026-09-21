# OG Doom Terminal Edition

A terminal-based Doom-like FPS game written in Go, designed to run on feature phones via terminal/SSH with ANSI support.

![OG Doom](https://img.shields.io/badge/Platform-Terminal-green)
![Language](https://img.shields.io/badge/Language-Go-00ADD8)
![License](https://img.shields.io/badge/License-MIT-blue)

## Features

- **Raycasting Engine**: Classic Wolfenstein/Doom-style 3D rendering using ASCII art
- **Multiple Levels**: 3 unique levels with increasing difficulty
- **Enemies**: Demons, Imps, and Zombies with AI behavior
- **Weapons**: Pistol, Shotgun, and Plasma gun
- **Items**: Health packs, ammo, and keys
- **Minimap**: Toggle-able minimap for navigation
- **HUD**: Health, ammo, score, and level display
- **Feature Phone Compatible**: Works on any device with terminal access

## Controls

| Key | Action |
|-----|--------|
| W/↑ | Move Forward |
| S/↓ | Move Backward |
| A | Strafe Left |
| D | Strafe Right |
| J/← | Turn Left |
| L/→ | Turn Right |
| SPACE/F | Shoot |
| 1/2/3 | Switch Weapon |
| M | Toggle Minimap |
| E | Open Doors |
| P/ESC | Pause |
| Q | Quit |

## Installation

### From Source

```bash
git clone https://github.com/yourusername/og-doom.git
cd og-doom
go build -o doom
./doom
```

### Pre-built Binary

Download the latest release for your platform from the [Releases](https://github.com/yourusername/og-doom/releases) page.

## Cross-Compilation for Feature Phones

To build for ARM-based feature phones:

```bash
# For ARM devices (e.g., Raspberry Pi, feature phones)
GOOS=linux GOARCH=arm go build -o doom-arm

# For ARM64 devices
GOOS=linux GOARCH=arm64 go build -o doom-arm64

# For MIPS devices (common in routers/embedded)
GOOS=linux GOARCH=mips go build -o doom-mips
```

## Architecture

- `main.go` - Game initialization and main entry point
- `terminal.go` - Raw terminal input handling
- `raycast.go` - DDA raycasting engine
- `render.go` - 3D rendering and sprite drawing
- `gamelogic.go` - Game loop, physics, and enemy AI

## How It Works

The game uses a **DDA (Digital Differential Analyzer)** raycasting algorithm to cast rays from the player's viewpoint through each column of the screen. When a ray hits a wall, it calculates the distance and draws a vertical stripe with height inversely proportional to the distance.

Sprites (enemies and items) are rendered using **billboard rendering** - they always face the player and are sorted by distance for correct occlusion.

All rendering uses ANSI escape codes for colors and Unicode characters for wall textures, making it compatible with virtually any modern terminal.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- id Software for the original DOOM (1993)
- The raycasting tutorial at [lodev.org](https://lodev.org/cgtutor/raycasting.html)
- The Go programming language community
