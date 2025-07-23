package main

import (
	"bytes"
	_ "embed"
	"image/png"
	"minefetch/internal/image/print"
	"minefetch/internal/image/scale"
	"minefetch/internal/mc"
)

const iconAspectRatio = 0.5

//go:embed unknown_server.png
var defaultIcon []byte

func iconHeight() uint {
	return uint(float64(flagIconSize) * iconAspectRatio)
}

func printIcon(icon *mc.Icon) error {
	img := icon.Image
	var err error
	if img == nil {
		img, err = png.Decode(bytes.NewReader(defaultIcon))
	}
	if err != nil {
		return err
	}

	f := float64(flagIconSize) / float64(img.Bounds().Dy())
	if f != 1 {
		img = scale.Lanczos(img, f)
	}
	print.HalfPrint(img, 255/2)

	return nil
}
