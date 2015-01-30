package main

import (
	"fmt"
	"github.com/mat/besticon/ico"
	"os"
)

func main() {
	for _, filename := range os.Args[1:] {
		fmt.Printf("%s: ", filename)
		f, err := os.Open(filename)
		if err != nil {
			fmt.Printf("failed to open %s\n", filename)
			continue
		}
		defer f.Close()

		dir, err := ico.ParseIco(f)
		if err != nil {
			fmt.Printf("Failed to parse %s as icon file\n", filename)
		} else if dir.Count == 1 {
			fmt.Printf("MS Windows icon resource - 1 icon\n")
		} else {
			best := dir.FindBestIcon()
			fmt.Printf("MS Windows icon resource - %d icons, %dx%d, %d-colors\n", dir.Count,
				best.Width,
				best.Height,
				best.ColorCount())
			//		fmt.Printf("%+v\n", dir)
		}
	}
}
