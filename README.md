# Pokm: Optimized Binary Pokémon CLI

A high-performance Pokémon data viewer and team analyzer built from scratch in Go. This project focuses on extreme data density and native terminal rendering without external dependencies.

## Highlights
- **Custom Binary Format:** Transformed 1.7GB of raw CSV/PNG data into a **1.5MB** unified binary database.
- **Data Pipeline:** Implemented an ETL (Extract, Transform, Load) process to denormalize relational data into fixed-width records for $O(1)$ seek performance.
- **ANSI Sprite Renderer:** Developed a native terminal engine using Unicode Half-Blocks (`▀`) and ANSI-256 color mapping (1 byte per pixel).
- **Zero-Dependency:** Compiled as a single static binary that runs in any standard Linux terminal.

## Project Structure
- `/data-sources`: The binary "databases".
- `/pokm`: The main application and rendering logic.
- `/tools`: The "Generator" scripts used to compile raw assets.

## Note

Only pokemon and moves up to generation 9 were included.

## Quick Start
```bash
# Clone and run
go run pokm/main.go
