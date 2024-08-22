/*
@author: sk
@date: 2024/8/21
*/
package main

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func RGBIndex(r, g, b uint8) uint8 {
	return uint8((int(r)*3 + int(g)*5 + int(b)*7) % 64)
}

func RGBToU32(r, g, b uint8) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func U32ToRGB(rgb uint32) (uint8, uint8, uint8) {
	return uint8(rgb), uint8(rgb >> 8), uint8(rgb >> 16)
}
