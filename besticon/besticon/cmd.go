package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/mat/besticon/v3/besticon"
)

func main() {
	all := flag.Bool("all", false, "Display all Icons, not just the best.")
	flag.Parse()

	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stderr, "please provide a URL.\n")
		os.Exit(100)
	}

	url := os.Args[len(os.Args)-1]

	b := besticon.New(besticon.WithLogger(besticon.NewDefaultLogger(io.Discard))) // Disable verbose logging

	finder := b.NewIconFinder()
	icons, err := finder.FetchIcons(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:  failed to fetch icons: %s\n", url, err)
		os.Exit(1)
	}

	if *all {
		for _, img := range icons {
			if img.Width > 0 {
				fmt.Printf("%s:  %s\n", url, img.URL)
			}
		}
	} else {
		if len(icons) > 0 {
			best := icons[0]
			fmt.Printf("%s:  %s\n", url, best.URL)
		} else {
			fmt.Fprintf(os.Stderr, "%s:  no icons found\n", url)
			os.Exit(2)
		}
	}
}
