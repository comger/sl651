package main

import (
	"fmt"
)

func parseTestMessage(data []byte) {
	fmt.Printf("报文长度: %d\n", len(data))
	fmt.Printf("完整报文: %x\n", data)
	fmt.Println()

	if len(data) >= 2 && data[0] == 0x7e && data[1] == 0x7e {
		fmt.Println("起始符: 7E 7E")
	}

	if len(data) >= 3 {
		fmt.Printf("地址域: %02X\n", data[2])
	}

	if len(data) >= 12 {
		stationID := string(data[3:12])
		fmt.Printf("站点ID: %s\n", stationID)
	}

	if len(data) >= 13 {
		fmt.Printf("密码: %02X\n", data[12])
	}

	if len(data) >= 14 {
		fmt.Printf("功能码: %02X\n", data[13])
	}

	if len(data) >= 16 {
		length := int(data[14])<<8 | int(data[15])
		fmt.Printf("数据长度: %d\n", length)
	}

	if len(data) >= 20 {
		voltage := float64(int(data[16])<<8|int(data[17])) / 100.0
		fmt.Printf("电源电压: %.2fV\n", voltage)
	}

	if len(data) >= 24 {
		waterLevel := float64(int(data[18])<<24|int(data[19])<<16|int(data[20])<<8|int(data[21])) / 1000.0
		fmt.Printf("水位: %.3f米\n", waterLevel)
	}

	if len(data) >= 30 {
		year := 2000 + int(data[22])
		month := int(data[23])
		day := int(data[24])
		hour := int(data[25])
		minute := int(data[26])
		second := int(data[27])
		fmt.Printf("观测时间: %04d-%02d-%02d %02d:%02d:%02d\n", year, month, day, hour, minute, second)
	}

	if len(data) >= 32 {
		crc := int(data[30])<<8 | int(data[31])
		fmt.Printf("CRC: %04X\n", crc)
	}

	fmt.Println()
	fmt.Println("详细字节分析:")
	for i := 0; i < len(data) && i < 40; i++ {
		fmt.Printf("[%2d] %02X ", i, data[i])
		if (i+1)%8 == 0 {
			fmt.Println()
		}
	}
}

func main() {
	hexStr := "7e7e011090330853a00032004d0200da260202134021f1f1109033085350f0f026020213402219000000272b0000000000392300000138371b000000302b000004300020190000001f1900000026190000001a190000003812118903fb3e"
	
	data := make([]byte, len(hexStr)/2)
	for i := 0; i < len(hexStr); i += 2 {
		b := byte(0)
		if hexStr[i] >= '0' && hexStr[i] <= '9' {
			b = (hexStr[i]-'0') << 4
		} else if hexStr[i] >= 'a' && hexStr[i] <= 'f' {
			b = (hexStr[i]-'a'+10) << 4
		} else if hexStr[i] >= 'A' && hexStr[i] <= 'F' {
			b = (hexStr[i]-'A'+10) << 4
		}
		if hexStr[i+1] >= '0' && hexStr[i+1] <= '9' {
			b |= hexStr[i+1] - '0'
		} else if hexStr[i+1] >= 'a' && hexStr[i+1] <= 'f' {
			b |= hexStr[i+1] - 'a' + 10
		} else if hexStr[i+1] >= 'A' && hexStr[i+1] <= 'F' {
			b |= hexStr[i+1] - 'A' + 10
		}
		data[i/2] = b
	}

	parseTestMessage(data)
}
