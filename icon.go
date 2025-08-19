package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"minefetch/internal/image/print"
	"minefetch/internal/image/scale"
	"minefetch/internal/image/sixel"
	"minefetch/internal/term"
	"os"
)

const iconAspectRatio = 0.5

//go:embed embed/default.png
var defaultIcon []byte

func iconHeight() uint {
	return uint(float64(cfg.icon.size) * iconAspectRatio)
}

func printIcon(b []byte) {
	var img image.Image
	if b != nil {
		img, _ = png.Decode(bytes.NewReader(b))
	}
	if img == nil {
		var err error
		img, err = png.Decode(bytes.NewReader(defaultIcon))
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
	}

	switch cfg.icon.format {
	case "sixel":
		cellWidth, cellHeight, _ := term.QueryCellSize()
		if cellWidth == 0 || cellHeight == 0 {
			log.Fatalln("Failed to get cell size")
		}
		img = scale.NearestNeighbor(img, cfg.icon.size*cellWidth, iconHeight()*cellHeight)
		sixel.Encode(os.Stdout, img, nil)
		// Some terminals print a newline after sixel images, some don't
		// fmt.Println()
		fmt.Print(term.Up(iconHeight()) + term.Back(cfg.icon.size))
	case "half":
		f := float64(cfg.icon.size) / float64(img.Bounds().Dx())
		if f != 1 {
			img = scale.Lanczos(img, f)
		}
		print.HalfPrint(img, 255/2)
		fmt.Print(term.Up(iconHeight()-1) + term.Back(cfg.icon.size))
	case "shade":
		f := float64(cfg.icon.size) / float64(img.Bounds().Dx()) / 2
		if f != 1 {
			img = scale.Lanczos(img, f)
		}
		print.ShadePrint(img, true)
		fmt.Println()
	}
}
