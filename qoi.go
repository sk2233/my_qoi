/*
@author: sk
@date: 2024/8/21
*/
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

const (
	ChannelRGB  = 3
	ChannelRGBA = 4
)

const (
	ColorRGB = 0
	ColorAll = 1
)

const (
	QoiMagic = 0x716f6966 // qoif
)

const (
	OpRGB  = 0b1111_1110 // r g b 3byte
	OpRGBA = 0b1111_1111 // r g b a 4byte
	// (3*r + 5*g + 7*b)%64 对应的索引
	OpIndex = 0b0000_0000 // 0~5 index 6bit
	OpDiff  = 0b0100_0000 // 4~5 dr 2bit # 2~3 dg 2bit # 0~1 db 2bit
	OpLuma  = 0b1000_0000 // 0~5 dg 6bit # 4~7 dr-dg 4bit # 0~3 db-dg 4bit
	OpRun   = 0b1100_0000 // 0~5 run 6bit
)

type QoiHeader struct {
	Magic      uint32
	Width      uint32
	Height     uint32
	Channel    uint8
	ColorSpace uint8
}

func Encode(w io.Writer, m image.Image) error {
	buff := &bytes.Buffer{}
	bound := m.Bounds()
	header := &QoiHeader{
		Magic:      QoiMagic,
		Width:      uint32(bound.Dx()),
		Height:     uint32(bound.Dy()),
		Channel:    uint8(ChannelRGB),
		ColorSpace: uint8(ColorRGB),
	}
	if err := binary.Write(buff, binary.BigEndian, header); err != nil {
		return err
	}

	clrIdxes := make([]uint32, 64)
	lastR, lastG, lastB := uint8(0), uint8(0), uint8(0) // 初始上一个值都是 0
	l := uint8(0)
	for y := 0; y < bound.Dy(); y++ {
		for x := 0; x < bound.Dx(); x++ {
			tr, tg, tb, _ := m.At(x, y).RGBA()
			r := uint8(tr >> 8)
			g := uint8(tg >> 8)
			b := uint8(tb >> 8)
			u32 := RGBToU32(r, g, b)
			idx := RGBIndex(r, g, b)
			// 先从小编码尝试  长度要限制防止与 OpRGB 碰撞
			if r == lastR && g == lastG && b == lastB && l < 0b11_1110-1 { // 长度编码
				l++
			} else {
				if l > 0 { // 写入长度编码
					buff.WriteByte(OpRun | l)
					l = 0
				}
				if clrIdxes[idx] == u32 { // 索引编码
					buff.WriteByte(OpIndex | idx)
				} else { // 偏移编码
					dr := int(r) - int(lastR)
					dg := int(g) - int(lastG)
					db := int(b) - int(lastB)
					if (dr+2) >= 0 && (dr+2) < 4 && // 小偏移
						(dg+2) >= 0 && (dg+2) < 4 &&
						(db+2) >= 0 && (db+2) < 4 {
						buff.WriteByte(OpDiff | (uint8(dr+2) << 4) | (uint8(dg+2) << 2) | uint8(db+2))
					} else if (dg+32) >= 0 && (dg+32) < 64 && // 大偏移
						(dr-dg+8) >= 0 && (dr-dg+8) < 16 &&
						(db-dg+8) >= 0 && (db-dg+8) < 16 {
						buff.WriteByte(OpLuma | uint8(dg+32))
						buff.WriteByte((uint8(dr-dg+8) << 4) | uint8(db-dg+8))
					} else { // 原始编码
						buff.WriteByte(OpRGB)
						buff.WriteByte(r)
						buff.WriteByte(g)
						buff.WriteByte(b)
					}
				}
			}
			lastR, lastG, lastB = r, g, b
			clrIdxes[idx] = u32 // 更新缓存索引
		}
	}
	if l > 0 { // 写入长度编码
		buff.WriteByte(OpRun | l)
		l = 0
	}
	_, err := w.Write(buff.Bytes())
	return err
}

func Decode(r0 io.Reader) (image.Image, error) {
	buff := &bytes.Buffer{}
	_, err := io.Copy(buff, r0)
	if err != nil {
		return nil, err
	}

	header := &QoiHeader{}
	if err := binary.Read(buff, binary.BigEndian, header); err != nil {
		return nil, err
	}
	if header.Magic != QoiMagic {
		return nil, fmt.Errorf("magic number mismatch")
	}
	img := NewQoiImage(image.Rect(0, 0, int(header.Width), int(header.Height)))

	clrIdxes := make([]uint32, 64)
	lastR, lastG, lastB := uint8(0), uint8(0), uint8(0) // 初始上一个值都是 0
	l := uint8(0)
	for y := 0; y < int(header.Height); y++ {
		for x := 0; x < int(header.Width); x++ {
			r, g, b := uint8(0), uint8(0), uint8(0)
			if l > 0 { // 还有可以消耗的长度，消耗长度
				l--
				r, g, b = lastR, lastG, lastB
			} else {
				by, err := buff.ReadByte()
				if err != nil {
					return nil, err
				}
				if by == OpRGB { // 原始编码
					if r, err = buff.ReadByte(); err != nil {
						return nil, err
					}
					if g, err = buff.ReadByte(); err != nil {
						return nil, err
					}
					if b, err = buff.ReadByte(); err != nil {
						return nil, err
					}
				} else if (by & 0b1100_0000) == OpIndex { // 索引编码
					r, g, b = U32ToRGB(clrIdxes[by&0b0011_1111])
				} else if (by & 0b1100_0000) == OpRun { // 长度编码
					r, g, b = lastR, lastG, lastB
					l = (by & 0b0011_1111) - 1
				} else if (by & 0b1100_0000) == OpDiff { // 小偏移
					r = uint8(int((by>>4)&0b11) - 2 + int(lastR))
					g = uint8(int((by>>2)&0b11) - 2 + int(lastG))
					b = uint8(int(by&0b11) - 2 + int(lastB))
					val := RGBToU32(r, g, b)
					val++
				} else if (by & 0b1100_0000) == OpLuma { // 大偏移
					dg := int((by & 0b11_1111) - 32)
					g = uint8(dg + int(lastG))
					nb, err := buff.ReadByte()
					if err != nil {
						return nil, err
					}
					r = uint8(int((nb>>4)&0b1111) - 8 + dg + int(lastR))
					b = uint8(int(nb&0b1111) - 8 + dg + int(lastB))
					val := RGBToU32(r, g, b)
					val++
				}
			}
			img.Colors[x][y] = color.RGBA{R: r, G: g, B: b, A: 0xFF}
			lastR, lastG, lastB = r, g, b
			clrIdxes[RGBIndex(r, g, b)] = RGBToU32(r, g, b) // 更新缓存索引
		}
	}
	return img, nil
}
