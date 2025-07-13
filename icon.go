package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"image"
	"image/png"
	"minefetch/internal/image/print"
	"minefetch/internal/image/scale"
	"strings"
)

const iconWidth = 32
const iconHeight = iconWidth / 2

//go:embed unknown_server.png
var defaultIcon []byte

func printIcon(s *string) error {
	var img image.Image
	var err error
	if *s != "" {
		d := base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(*s, "data:image/png;base64,")))
		img, err = png.Decode(d)
	} else {
		img, err = png.Decode(bytes.NewReader(defaultIcon))
	}
	if err != nil {
		return err
	}

	f := float64(iconWidth) / float64(img.Bounds().Dy())
	if f != 1 {
		img = scale.Lanczos(img, f)
	}
	print.HalfPrint(img, 255/2)

	return nil
}
