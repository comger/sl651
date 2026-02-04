package main

import (
	"encoding/hex"
	"fmt"
)

func calculateCRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for i := 0; i < 8; i++ {
			if crc&0x0001 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

func main() {
	// Full message
	hexStr := "7e7e011090330853a00032004d0200da260202134021f1f1109033085350f0f026020213402219000000272b0000000000392300000138371b000000302b000004300020190000001f1900000026190000001a190000003812118903fb3e"
	data, _ := hex.DecodeString(hexStr)

	// CRC is calculated on data[2 : len(data)-2]
	// data[0:2] is 7E 7E
	// data[len-2:] is CRC
	calcCRC := calculateCRC16(data[2 : len(data)-2])
	actualCRC := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])

	fmt.Printf("Calculated CRC: %04X\n", calcCRC)
	fmt.Printf("Actual CRC:     %02X%02X\n", data[len(data)-2], data[len(data)-1])

	if calcCRC == actualCRC {
		fmt.Println("CRC Match!")
	} else {
		// Try little endian or different ranges
		calcCRC_LE := uint16(data[len(data)-2]) | uint16(data[len(data)-1])<<8
		fmt.Printf("Calculated CRC (LE): %04X\n", calcCRC_LE)
	}
}
