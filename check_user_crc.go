package main

import (
	"encoding/hex"
	"fmt"
)

func calculateCRC(data []byte) uint16 {
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
	hexStr := "7e7e011090330854a00031001a02a443260202152538220000118939230000013830230004300003"
	data, _ := hex.DecodeString(hexStr)

	crc := calculateCRC(data)
	fmt.Printf("Calculated CRC (MODBUS): %04X\n", crc)
}
