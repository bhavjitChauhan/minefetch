package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/image/print"
	"minefetch/internal/image/scale"
)

const iconAspectRatio = 0.5

//go:embed unknown_server.png
var defaultIcon []byte

func iconHeight() uint {
	return uint(float64(cfg.iconSize) * iconAspectRatio)
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

	f := float64(cfg.iconSize) / float64(img.Bounds().Dy())
	if f != 1 {
		img = scale.Lanczos(img, f)
	}
	print.HalfPrint(img, 255/2)
	fmt.Print(ansi.Up(iconHeight()-1) + ansi.Back(cfg.iconSize))
}
