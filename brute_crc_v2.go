package main

import (
	"encoding/hex"
	"fmt"
)

func crcModbus(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func main() {
	hexStr := "7e7e011090330853a00032004d0200da260202134021f1f1109033085350f0f026020213402219000000272b0000000000392300000138371b000000302b000004300020190000001f1900000026190000001a190000003812118903fb3e"
	data, _ := hex.DecodeString(hexStr)

	// Transmitted CRC is FB 3E (at indices 92, 93)
	// Let's try many targets
	targets := []uint16{
		0xFB3E, // Big Endian
		0x3EFB, // Little Endian
	}

	for _, target := range targets {
		for start := 0; start < 20; start++ {
			for end := 20; end <= 92; end++ {
				if end <= start {
					continue
				}
				if crcModbus(data[start:end]) == target {
					fmt.Printf("MATCH MODBUS! Start: %d, End: %d, Target: %04X\n", start, end, target)
				}
			}
		}
	}
}
