/*
@author: sk
@date: 2024/8/22
*/
package main

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type QoiShow struct {
	Image  *ebiten.Image
	Option *ebiten.DrawImageOptions
}

func NewQoiShow(img image.Image) *QoiShow {
	image0 := ebiten.NewImageFromImage(img)
	return &QoiShow{Image: image0, Option: &ebiten.DrawImageOptions{}}
}

func (q *QoiShow) Update() error {
	return nil
}

func (q *QoiShow) Draw(screen *ebiten.Image) {
	screen.DrawImage(q.Image, q.Option)
}

func (q *QoiShow) Layout(w, h int) (int, int) {
	return w, h
}
