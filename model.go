/*
@author: sk
@date: 2024/8/21
*/
package main

import (
	"image"
	"image/color"
)

type QoiModel struct {
}

func (q *QoiModel) Convert(c color.Color) color.Color {
	return c
}

type QoiImage struct {
	Model   color.Model
	Bounds0 image.Rectangle
	Colors  [][]color.Color
}

func NewQoiImage(bounds image.Rectangle) *QoiImage {
	colors := make([][]color.Color, bounds.Dx())
	for i := 0; i < bounds.Dx(); i++ {
		colors[i] = make([]color.Color, bounds.Dy())
	}
	return &QoiImage{Model: &QoiModel{}, Bounds0: bounds, Colors: colors}
}

func (q *QoiImage) ColorModel() color.Model {
	return q.Model
}

func (q *QoiImage) Bounds() image.Rectangle {
	return q.Bounds0
}

func (q *QoiImage) At(x, y int) color.Color {
	return q.Colors[x][y]
}
