package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Reverse map to get the Type names back from the uint8 IDs
var typeNames = map[uint8]string{
	0: "None", 1: "Normal", 2: "Fighting", 3: "Flying", 4: "Poison", 5: "Ground",
	6: "Rock", 7: "Bug", 8: "Ghost", 9: "Steel", 10: "Fire",
	11: "Water", 12: "Grass", 13: "Electric", 14: "Psychic", 15: "Ice",
	16: "Dragon", 17: "Dark", 18: "Fairy",
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Pokemon ID to debug: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	targetID, err := strconv.Atoi(input)
	if err != nil || targetID <= 0 {
		fmt.Println("Invalid ID. Please enter a positive number.")
		return
	}

	fmt.Printf("\n--- SEARCHING FOR ID %d ---\n", targetID)
	
	// 1. Fetch and print stats
	found := printPokemonData(uint16(targetID))
	if !found {
		fmt.Println("Pokemon not found in pokemon-data.bin")
		return
	}

	// 2. Fetch and render sprite
	fmt.Println("\n--- SPRITE RENDER ---")
	printSpriteData(uint16(targetID))
}

func printPokemonData(targetID uint16) bool {
	file, err := os.Open("../data-sources/pokemon-data.bin")
	if err != nil {
		fmt.Println("Error opening pokemon-data.bin:", err)
		return false
	}
	defer file.Close()

	var totalRecords uint16
	binary.Read(file, binary.LittleEndian, &totalRecords)

	for i := 0; i < int(totalRecords); i++ {
		// Read the 17-byte fixed block
		// Structure: ID(2), Species(2), Height(2), Weight(2), IsDefault(1), Type1(1), Type2(1), Stats(6)
		var id uint16
		binary.Read(file, binary.LittleEndian, &id)

		var species, height, weight uint16
		binary.Read(file, binary.LittleEndian, &species)
		binary.Read(file, binary.LittleEndian, &height)
		binary.Read(file, binary.LittleEndian, &weight)

		var isDefault, type1, type2 uint8
		binary.Read(file, binary.LittleEndian, &isDefault)
		binary.Read(file, binary.LittleEndian, &type1)
		binary.Read(file, binary.LittleEndian, &type2)

		var stats [6]uint8
		binary.Read(file, binary.LittleEndian, &stats)

		// Read variable-length name
		var nameLen uint8
		binary.Read(file, binary.LittleEndian, &nameLen)
		
		nameBytes := make([]byte, nameLen)
		file.Read(nameBytes)

		// If this is the one we want, print it and return
		if id == targetID {
			name := strings.Title(string(nameBytes))
			fmt.Printf("Name:   %s (#%d)\n", name, id)
			
			// Format Types beautifully
			typeStr := typeNames[type1]
			if type2 != 0 {
				typeStr += " / " + typeNames[type2]
			}
			fmt.Printf("Type:   %s\n", typeStr)
			fmt.Printf("Height: %.1f m | Weight: %.1f kg\n", float32(height)/10.0, float32(weight)/10.0)
			
			fmt.Printf("Stats:  HP:%d | Atk:%d | Def:%d | SpA:%d | SpD:%d | Spe:%d\n",
				stats[0], stats[1], stats[2], stats[3], stats[4], stats[5])
			return true
		}
	}
	return false
}

func printSpriteData(targetID uint16) {
	file, err := os.Open("../data-sources/pokemon-sprites.bin")
	if err != nil {
		fmt.Println("Error opening pokemon-sprites.bin:", err)
		return
	}
	defer file.Close()

	var totalSprites uint16
	binary.Read(file, binary.LittleEndian, &totalSprites)

	for i := 0; i < int(totalSprites); i++ {
		var id uint16
		binary.Read(file, binary.LittleEndian, &id)

		var pixels [1024]uint8
		binary.Read(file, binary.LittleEndian, &pixels)

		if id == targetID {
			renderASCII(pixels)
			return
		}
	}
	fmt.Println("Sprite not found.")
}

// renderASCII prints the 32x32 array using 16 vertical terminal rows
func renderASCII(pixels [1024]uint8) {
	const size = 32

	// We iterate Y by 2 because one terminal character represents 2 pixels
	for y := 0; y < size; y += 2 {
		for x := 0; x < size; x++ {
			topColor := pixels[y*size+x]
			bottomColor := pixels[(y+1)*size+x]

			// The ANSI rendering logic
			if topColor == 0 && bottomColor == 0 {
				// Both transparent: Print empty space
				fmt.Print(" ")
			} else if topColor != 0 && bottomColor == 0 {
				// Top colored, bottom transparent
				fmt.Printf("\033[38;5;%dm▀\033[0m", topColor)
			} else if topColor == 0 && bottomColor != 0 {
				// Top transparent, bottom colored (Use lower half block)
				fmt.Printf("\033[38;5;%dm▄\033[0m", bottomColor)
			} else {
				// Both colored (Foreground = top, Background = bottom, Character = Upper Half Block)
				fmt.Printf("\033[38;5;%dm\033[48;5;%dm▀\033[0m", topColor, bottomColor)
			}
		}
		// Newline after each row of 32 characters
		fmt.Println()
	}
}
