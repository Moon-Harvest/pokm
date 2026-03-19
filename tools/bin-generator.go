package main

import (
    "encoding/binary"
    "encoding/csv"
    "fmt"
    "os"
    "strconv"
    "image"
    _ "image/png"
    "path/filepath"
    "sort"
    "strings"
)

const (
	targetSize      = 32
	transparentFlag = 0 // We'll use 0 to represent "transparent"
)

func main() {
    generatePokemonBin()
    generateMovesBin()

    fmt.Printf("[+] Created ../data-sources/pokemon.bin\n")
    fmt.Printf("[+] Created ../data-sources/moves.bin\n")
}

func generatePokemonBin() {
    pokeBin, err := os.Create("../data-sources/pokemon.bin")
    if err != nil {
        fmt.Println("Error creating pokemon.bin:", err)
        return
    }
    defer pokeBin.Close()

    // 1. PRE-LOAD STATS INTO A MAP
    statsFile, _ := os.Open("../original-data-sources/pokemon_stats.csv")
    statsReader := csv.NewReader(statsFile)
    statsRecords, _ := statsReader.ReadAll()
    statsFile.Close()

    // map[pokemon_id][6]uint8 (Index: 0=HP, 1=Atk, 2=Def, 3=SpA, 4=SpD, 5=Spe)
    statsMap := make(map[int][6]uint8)
    for i, row := range statsRecords {
        if i == 0 { continue }
        pID, _ := strconv.Atoi(row[0])
        sID, _ := strconv.Atoi(row[1])
        val, _ := strconv.Atoi(row[2])
        
        if sID >= 1 && sID <= 6 {
            stats := statsMap[pID]
            stats[sID-1] = uint8(val)
            statsMap[pID] = stats
        }
    }

    // 2. PRE-LOAD TYPES INTO A MAP
    typesFile, _ := os.Open("../original-data-sources/pokemon_types.csv")
    typesReader := csv.NewReader(typesFile)
    typesRecords, _ := typesReader.ReadAll()
    typesFile.Close()

    // map[pokemon_id][2]uint8 (Index: 0=Primary, 1=Secondary)
    typesMap := make(map[int][2]uint8)
    for i, row := range typesRecords {
        if i == 0 { continue }
        pID, _ := strconv.Atoi(row[0])
        tID, _ := strconv.Atoi(row[1])
        slot, _ := strconv.Atoi(row[2])
        
        types := typesMap[pID]
        if slot == 1 {
            types[0] = uint8(tID)
        } else if slot == 2 {
            types[1] = uint8(tID)
        }
        typesMap[pID] = types
    }

    // 3. READ MAIN POKEMON DATA AND JOIN
    pokeFile, _ := os.Open("../original-data-sources/pokemon.csv")
    pokeReader := csv.NewReader(pokeFile)
    pokeRecords, _ := pokeReader.ReadAll()
    pokeFile.Close()

    pokemonCount := uint16(len(pokeRecords) - 1)
    binary.Write(pokeBin, binary.LittleEndian, pokemonCount)

    for i, row := range pokeRecords {
        if i == 0 { continue }

        id, _ := strconv.Atoi(row[0])
        species_id, _ := strconv.Atoi(row[2])
        height, _ := strconv.Atoi(row[3])
        weight, _ := strconv.Atoi(row[4])
        is_default, _ := strconv.Atoi(row[5])

        // The "Join" execution
        stats := statsMap[id]
        types := typesMap[id]

        // --- WRITE THE DENORMALIZED RECORD ---
        
        // Physics & Identity
        binary.Write(pokeBin, binary.LittleEndian, uint16(id))
        binary.Write(pokeBin, binary.LittleEndian, uint16(species_id))
        binary.Write(pokeBin, binary.LittleEndian, uint16(height))
        binary.Write(pokeBin, binary.LittleEndian, uint16(weight))
        binary.Write(pokeBin, binary.LittleEndian, uint8(is_default))

        // Types (2 bytes)
        binary.Write(pokeBin, binary.LittleEndian, types[0])
        binary.Write(pokeBin, binary.LittleEndian, types[1])

        // Stats (6 bytes)
        for _, stat := range stats {
            binary.Write(pokeBin, binary.LittleEndian, stat)
        }

        // Variable Length String (Name)
        nameBytes := []byte(row[1])
        binary.Write(pokeBin, binary.LittleEndian, uint8(len(nameBytes)))
        pokeBin.Write(nameBytes)
    }
}

func generateMovesBin() {
    movesBin, err := os.Create("../data-sources/moves.bin")
    if err != nil {
        fmt.Println("Error creating moves.bin:", err)
        return
    }
    defer movesBin.Close()

    movesFile, _ := os.Open("../original-data-sources/moves.csv")
    movesReader := csv.NewReader(movesFile)
    movesRecords, _ := movesReader.ReadAll()
    movesFile.Close()

    movesCount := uint16(len(movesRecords) - 1)
    binary.Write(movesBin, binary.LittleEndian, movesCount)

    for i, row := range movesRecords {
        if i == 0 { continue }

        // id, identifier, type_id, power, pp, accuracy, priority, damage_class_id
        id, _ := strconv.Atoi(row[0])
        type_id, _ := strconv.Atoi(row[2])
        power, _ := strconv.Atoi(row[3])
        pp, _ := strconv.Atoi(row[4])
        
        raw_accuracy := row[5]
        var accuracy uint8
        if raw_accuracy == "0" || raw_accuracy == "" {
            accuracy = 255
        } else {
            val, _ := strconv.Atoi(raw_accuracy)
            accuracy = uint8(val)
        }
        
        priority, _ := strconv.Atoi(row[6])
        damage_class_id, _ := strconv.Atoi(row[7])

        // Write Fixed Data
        binary.Write(movesBin, binary.LittleEndian, uint16(id))
        binary.Write(movesBin, binary.LittleEndian, uint8(type_id))
        binary.Write(movesBin, binary.LittleEndian, uint8(power))
        binary.Write(movesBin, binary.LittleEndian, uint8(pp))
        binary.Write(movesBin, binary.LittleEndian, accuracy)
        binary.Write(movesBin, binary.LittleEndian, int8(priority))
        binary.Write(movesBin, binary.LittleEndian, uint8(damage_class_id))

        // Write Variable String
        nameBytes := []byte(row[1])
        binary.Write(movesBin, binary.LittleEndian, uint8(len(nameBytes)))
        movesBin.Write(nameBytes)
    }
}

// Note: The PNG sprites were removed from the folders, but you can download them from this specific github repo folderi
// https://github.com/PokeAPI/sprites/tree/master/sprites/pokemon
func generateSpritesBin() {
	inputDir := "../original-data-sources/pokemon-sprites"
	outputDir := "../data-sources/ascii-art-sprites"
	outputFile := filepath.Join(outputDir, "sprites.bin")

	// Ensure output directory exists
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	binFile, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating %s: %v\n", outputFile, err)
		return
	}
	defer binFile.Close()

	// Read all files in the directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		fmt.Printf("Error reading input directory: %v\n", err)
		return
	}

	// Filter for PNGs and sort them numerically by ID
	var pngFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".png") {
			pngFiles = append(pngFiles, f.Name())
		}
	}

	sort.Slice(pngFiles, func(i, j int) bool {
		id1, _ := strconv.Atoi(strings.TrimSuffix(pngFiles[i], ".png"))
		id2, _ := strconv.Atoi(strings.TrimSuffix(pngFiles[j], ".png"))
		return id1 < id2
	})

	// Write total count of sprites at the top of the file
	spriteCount := uint16(len(pngFiles))
	binary.Write(binFile, binary.LittleEndian, spriteCount)

	fmt.Printf("[*] Processing %d sprites...\n", spriteCount)

	processed := 0
	for _, filename := range pngFiles {
		id, _ := strconv.Atoi(strings.TrimSuffix(filename, ".png"))

		filePath := filepath.Join(inputDir, filename)
		imgFile, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Failed to open %s: %v\n", filename, err)
			continue
		}

		img, _, err := image.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			fmt.Printf("Failed to decode %s: %v\n", filename, err)
			continue
		}

		// Write the Pokemon ID first (2 bytes)
		binary.Write(binFile, binary.LittleEndian, uint16(id))

		// Downsample to 32x32 and write pixel array (1024 bytes)
		bounds := img.Bounds()
		origWidth := bounds.Dx()
		origHeight := bounds.Dy()

		for y := 0; y < targetSize; y++ {
			for x := 0; x < targetSize; x++ {
				// Nearest neighbor coordinate mapping
				srcX := bounds.Min.X + (x * origWidth / targetSize)
				srcY := bounds.Min.Y + (y * origHeight / targetSize)

				r, g, b, a := img.At(srcX, srcY).RGBA()

				// Convert 16-bit color components to ANSI 8-bit index
				ansiColor := rgbaToANSI(r, g, b, a)
				binary.Write(binFile, binary.LittleEndian, ansiColor)
			}
		}

		processed++
		if processed%100 == 0 {
			fmt.Printf("[-] Processed %d/%d sprites\n", processed, spriteCount)
		}
	}

	fmt.Printf("[+] Successfully generated %s\n", outputFile)
}

// rgbaToANSI converts Go's 16-bit RGBA color into an 8-bit ANSI 256-color index
func rgbaToANSI(r, g, b, a uint32) uint8 {
	// If it's more than 50% transparent, mark as transparent flag
	// Go's color.RGBA returns values in the range 0-65535
	if a < 32768 {
		return transparentFlag
	}

	// Scale down from 0-65535 to the 0-5 range used by the ANSI color cube
	r5 := (r * 5) / 65535
	g5 := (g * 5) / 65535
	b5 := (b * 5) / 65535

	// ANSI 256 color formula for the 6x6x6 color cube: 16 + 36*r + 6*g + b
	return uint8(16 + (36 * r5) + (6 * g5) + b5)
}
