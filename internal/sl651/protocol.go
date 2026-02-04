package sl651

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Protocol struct{}

func NewProtocol() *Protocol {
	return &Protocol{}
}

type Message struct {
	StationID        string
	FunctionCode     string
	FunctionCodeByte byte
	Direction        string
	RawData          string // Stores full hex string of the frame
	Payload          []byte
	Timestamp        time.Time
}

type DataType string

const (
	DataTypeRealtime DataType = "realtime"
	DataTypeAlarm    DataType = "alarm"
	DataTypeStatus   DataType = "status"
)

func (p *Protocol) Parse(hexData string) (*Message, error) {
	data, err := hex.DecodeString(hexData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex: %w", err)
	}
	if len(data) < 15 {
		return nil, errors.New("frame too short")
	}
	if data[0] != 0x7e || data[1] != 0x7e {
		return nil, errors.New("invalid start marker")
	}
	if !p.ValidateCRC(data) {
		calc0 := p.calculateCRC(data[:len(data)-2])
		calc2 := p.calculateCRC(data[2 : len(data)-2])
		actual := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])
		return nil, fmt.Errorf("CRC error: expected %04X, candidates [S0:%04X, S2:%04X]", actual, calc0, calc2)
	}
	msg := &Message{}
	msg.RawData = hexData

	// Determine Station ID range
	// Standard is data[3:8], but if it's mangled ASCII, it might be longer or start differently
	// For robustness, find index of Control Byte (0x01) and Function Code
	stationID := hex.EncodeToString(data[3:8])
	if data[2] == 0x01 && data[10] >= 0x2F {
		// Reasonable check
		stationID = hex.EncodeToString(data[3:8])
	}

	msg.StationID = stationID
	msg.FunctionCode = fmt.Sprintf("%02X", data[10])
	msg.FunctionCodeByte = data[10]
	flagLen := int(data[11])<<8 | int(data[12])
	bodyLen := flagLen & 0x0FFF
	if data[13] != 0x02 {
		return nil, errors.New("STX missing")
	}
	endIdx := 14 + bodyLen
	if endIdx > len(data)-2 {
		endIdx = len(data) - 2
	}
	msg.Payload = data[14:endIdx]
	msg.Timestamp = time.Now()
	return msg, nil
}

func (p *Protocol) ParseRealtimeData(payload []byte) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	cursor := 0
	if len(payload) >= 2 {
		result["serial"] = int(payload[0])<<8 | int(payload[1])
		cursor += 2
	}
	// Send Time: Usually 6 bytes (YYMMDDHHmmSS)
	if len(payload) >= cursor+6 {
		result["send_time"] = fmt.Sprintf("20%02x-%02x-%02x %02x:%02x:%02x",
			payload[cursor], payload[cursor+1], payload[cursor+2], payload[cursor+3], payload[cursor+4], payload[cursor+5])
		cursor += 6
	}
	for cursor < len(payload) {
		tag := payload[cursor]
		if tag == 0xFF || tag == 0x03 || tag == 0x00 {
			cursor++
			continue
		}
		if tag == 0xF1 && cursor+1 < len(payload) && payload[cursor+1] == 0xF1 {
			cursor += 2
			if cursor+6 <= len(payload) {
				result["body_station_id"] = hex.EncodeToString(payload[cursor : cursor+5])
				result["station_class"] = fmt.Sprintf("%02X", payload[cursor+5])
				cursor += 6
			}
			continue
		}
		// Observation Time Identifier (F0 F0)
		if tag == 0xF0 && cursor+1 < len(payload) && payload[cursor+1] == 0xF0 {
			cursor += 2
			// Standard says Obs Time is 5 bytes (YYMMDDHHmm)
			if cursor+5 <= len(payload) {
				result["observation_time"] = fmt.Sprintf("20%02x-%02x-%02x %02x:%02x",
					payload[cursor], payload[cursor+1], payload[cursor+2], payload[cursor+3], payload[cursor+4])
				cursor += 5
			}
			continue
		}
		if cursor+1 >= len(payload) {
			break
		}
		meta := payload[cursor+1]
		valLen := int(meta >> 3)
		decimals := int(meta & 0x07)
		if cursor+2+valLen > len(payload) {
			break
		}
		valBytes := payload[cursor+2 : cursor+2+valLen]
		val := decodeBCDToInt(valBytes)
		result[p.GetTagName(tag)] = float64(val) / math.Pow(10, float64(decimals))
		cursor += 2 + valLen
	}
	return result, nil
}

func (p *Protocol) GetTagName(tag byte) string {
	mapping := map[byte]string{
		0x39: "water_level",
		0x38: "voltage",
		0x30: "cumulative_flow",
		0x20: "total_rainfall",
		0x22: "rainfall",
		0x37: "instant_flow",
		0x26: "hourly_rainfall", // 1-hour
		0x27: "day_rainfall",    // period rainfall in user case
		0x1F: "period_rainfall", // user result shows 1F
		0x1A: "extra_rainfall",
	}
	if name, ok := mapping[tag]; ok {
		return name
	}
	return fmt.Sprintf("tag_%02X", tag)
}

func decodeBCDToInt(data []byte) int64 {
	res := int64(0)
	for _, b := range data {
		res = res*100 + int64(b>>4)*10 + int64(b&0x0F)
	}
	return res
}

func encodeIntToBCD(val int64, length int) []byte {
	res := make([]byte, length)
	temp := val
	for i := length - 1; i >= 0; i-- {
		low := temp % 10
		temp /= 10
		high := temp % 10
		temp /= 10
		res[i] = byte(high<<4 | low)
	}
	return res
}

func (p *Protocol) BuildMessage(stationID string, functionCode byte, bodyElements []byte) ([]byte, error) {
	frame := make([]byte, 0, 256)
	frame = append(frame, 0x7e, 0x7e, 0x01)
	addr, _ := hex.DecodeString(stationID)
	if len(addr) < 5 {
		padded := make([]byte, 5)
		copy(padded[5-len(addr):], addr)
		addr = padded
	} else {
		addr = addr[len(addr)-5:]
	}
	frame = append(frame, addr...)
	frame = append(frame, 0xa0, 0x00) // Password
	frame = append(frame, functionCode)

	serial := uint16(rand.Intn(65535))
	now := time.Now()

	sendTimeBCD := []byte{
		byte(p.intToBCD(now.Year() % 100)),
		byte(p.intToBCD(int(now.Month()))),
		byte(p.intToBCD(now.Day())),
		byte(p.intToBCD(now.Hour())),
		byte(p.intToBCD(now.Minute())),
		byte(p.intToBCD(now.Second())),
	}
	obsTimeBCD := []byte{
		byte(p.intToBCD(now.Year() % 100)),
		byte(p.intToBCD(int(now.Month()))),
		byte(p.intToBCD(now.Day())),
		byte(p.intToBCD(now.Hour())),
		byte(p.intToBCD(now.Minute())),
	}

	bodyPart := make([]byte, 0)
	bodyPart = append(bodyPart, byte(serial>>8), byte(serial))
	bodyPart = append(bodyPart, sendTimeBCD...)

	if functionCode != 0x2F {
		bodyPart = append(bodyPart, 0xF1, 0xF1)
		bodyPart = append(bodyPart, addr...)
		bodyPart = append(bodyPart, 0x50)
		bodyPart = append(bodyPart, 0xF0, 0xF0)
		bodyPart = append(bodyPart, obsTimeBCD...)
		bodyPart = append(bodyPart, bodyElements...)
	}

	l := uint16(len(bodyPart))
	frame = append(frame, byte(l>>8), byte(l))
	frame = append(frame, 0x02)
	frame = append(frame, bodyPart...)
	frame = append(frame, 0x03)
	// CRC is calculated on the full frame (including 7E 7E)
	crc := p.calculateCRC(frame)
	frame = append(frame, byte(crc>>8), byte(crc))
	return frame, nil
}

func (p *Protocol) intToBCD(n int) int {
	return ((n / 10) << 4) | (n % 10)
}

func (p *Protocol) calculateCRC(data []byte) uint16 {
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

func (p *Protocol) ValidateCRC(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	actual := uint16(data[len(data)-2])<<8 | uint16(data[len(data)-1])

	// Convention 1: Include 7E 7E (Start 0)
	if p.calculateCRC(data[:len(data)-2]) == actual {
		return true
	}

	// Convention 2: Skip 7E 7E (Start 2)
	if p.calculateCRC(data[2:len(data)-2]) == actual {
		return true
	}

	return false
}

func (p *Protocol) ConvertToStandardData(msg *Message) (map[string]interface{}, error) {
	res, err := p.ParseRealtimeData(msg.Payload)
	if err != nil {
		return nil, err
	}
	res["station_id"] = msg.StationID
	res["function_code"] = msg.FunctionCode
	res["timestamp"] = msg.Timestamp.Unix()
	return res, nil
}

func (p *Protocol) GetFunctionCodeName(code string) string {
	names := map[string]string{"2F": "链路维持", "31": "定时报", "32": "加报", "34": "小时报"}
	if name, ok := names[strings.ToUpper(code)]; ok {
		return name
	}
	return "未知"
}

func (p *Protocol) BuildTLV(tag byte, val interface{}, decimals int) []byte {
	floatVal := 0.0
	if f, ok := val.(float64); ok {
		floatVal = f
	} else if i, ok := val.(int); ok {
		floatVal = float64(i)
	}

	valInt := int64(math.Round(floatVal * math.Pow10(decimals)))
	// Length mapping based on user reference case
	l := 3
	switch tag {
	case 0x22, 0x20, 0x1F, 0x26, 0x37, 0x1A:
		l = 3
	case 0x39:
		l = 4
	case 0x30, 0x27:
		l = 5
	case 0x38:
		l = 2
	}
	valBCD := encodeIntToBCD(valInt, l)
	meta := byte(l<<3 | (decimals & 0x07))
	return append([]byte{tag, meta}, valBCD...)
}

func (p *Protocol) BuildTimeTLV(obsTime time.Time) []byte {
	// 0xF0 - Observation Time (YY MM DD HH mm) -> 5 bytes
	tag := byte(0xF0)
	// Manual BCD encoding for time
	y := obsTime.Year() % 100
	m := int(obsTime.Month())
	d := obsTime.Day()
	h := obsTime.Hour()
	min := obsTime.Minute()

	bcd := []byte{
		byte((y/10)<<4 | (y % 10)),
		byte((m/10)<<4 | (m % 10)),
		byte((d/10)<<4 | (d % 10)),
		byte((h/10)<<4 | (h % 10)),
		byte((min/10)<<4 | (min % 10)),
	}
	return append([]byte{tag, 0x28}, bcd...) // 0x28 = len 5, dec 0 (01010 000)
}

func (p *Protocol) GenerateID() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func (p *Protocol) GetDataType(code string) DataType {
	return DataTypeRealtime
}

func (p *Protocol) ParseCSVRecord(record []string) (*Message, error) {
	return nil, errors.New("CSV parsing not implemented")
}
