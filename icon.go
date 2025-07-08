package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strings"
)

const iconInvRatio = 2
const iconWidth = 32
const iconHeight = iconWidth / iconInvRatio

var levels = [...]string{" ", "░", "▒", "▓", "█"}

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

	dx := img.Bounds().Dx() / iconWidth
	dy := dx * iconInvRatio
	for y := range iconHeight {
		for x := range iconWidth {
			c := color.NRGBAModel.Convert(img.At(x*dx, y*dy)).(color.NRGBA)
			// https://pkg.go.dev/image/png#example-Decode
			level := c.A / 51 // 51 * 5 = 255
			if level == 5 {
				level--
			}
			fmt.Print(trueColor(c.R, c.G, c.B) + levels[level])
		}
		if y != iconHeight-1 {
			fmt.Print("\n")
		}
	}
	fmt.Print(reset)

	return nil
}
