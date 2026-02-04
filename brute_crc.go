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

func crcCCITT(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}

func main() {
	hexStr := "7e7e011090330853a00032004d0200da260202134021f1f1109033085350f0f026020213402219000000272b0000000000392300000138371b000000302b000004300020190000001f1900000026190000001a190000003812118903fb3e"
	data, _ := hex.DecodeString(hexStr)
	target := uint16(0xFB3E)

	fmt.Printf("Full Data Len: %d\n", len(data))

	for start := 0; start < 10; start++ {
		for end := len(data) - 5; end < len(data); end++ {
			if end <= start {
				continue
			}
			slice := data[start:end]
			if crcModbus(slice) == target {
				fmt.Printf("MATCH MODBUS! Start: %d, End: %d\n", start, end)
			}
			// Try swapped bytes target
			if crcModbus(slice) == 0x3EFB {
				fmt.Printf("MATCH MODBUS (Swapped)! Start: %d, End: %d\n", start, end)
			}
			if crcCCITT(slice) == target {
				fmt.Printf("MATCH CCITT! Start: %d, End: %d\n", start, end)
			}
		}
	}
}
