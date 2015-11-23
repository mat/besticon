package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mat/besticon/lettericon"
)

var (
	letter    = flag.String("letter", "X", "letter to draw")
	width     = flag.Int("width", 144, "width/height of the icon")
	iconColor = flag.String("color", "#909090", "icon color, as hex string")
)

func main() {
	flag.Parse()

	f, err := os.Create("out.png")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	col, err := lettericon.ColorFromHex(*iconColor)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	err = lettericon.Render(*letter, col, *width, f)
	if err != nil {
		os.Exit(1)
	}
	fmt.Println("Wrote out.png OK.")
}
