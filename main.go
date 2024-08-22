/*
@author: sk
@date: 2024/8/21
*/
package main

import (
	"image/png"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

// https://www.bilibili.com/read/cv17617178/
// 暂时先只支持 RGB

func main() {
	//PngToQoi("test.png")
	//QoiToPng("test.qoi")
	ShowQoi("test.qoi")
}

func ShowQoi(name string) {
	file, err := os.Open(name)
	HandleErr(err)
	defer file.Close()

	img, err := Decode(file)
	HandleErr(err)
	ebiten.SetWindowTitle(name)
	bound := img.Bounds()
	ebiten.SetWindowSize(bound.Dx(), bound.Dy())
	HandleErr(ebiten.RunGame(NewQoiShow(img)))
}

func QoiToPng(name string) {
	in, err := os.Open(name)
	HandleErr(err)
	defer in.Close()
	img, err := Decode(in)
	HandleErr(err)

	idx := strings.LastIndex(name, ".")
	out, err := os.Create(name[:idx] + ".png")
	HandleErr(err)
	defer out.Close()

	HandleErr(png.Encode(out, img))
}

func PngToQoi(name string) {
	in, err := os.Open(name)
	HandleErr(err)
	defer in.Close()
	img, err := png.Decode(in)
	HandleErr(err)

	idx := strings.LastIndex(name, ".")
	out, err := os.Create(name[:idx] + ".qoi")
	HandleErr(err)
	defer out.Close()

	HandleErr(Encode(out, img))
}
