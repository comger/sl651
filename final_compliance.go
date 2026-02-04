package main

import (
	"encoding/json"
	"fmt"
	"sl651-platform/internal/sl651"
)

func main() {
	hexStr := "7e7e011090330853a00032004d0200da260202134021f1f1109033085350f0f026020213402219000000272b0000000000392300000138371b000000302b000004300020190000001f1900000026190000001a190000003812118903fb3e"

	p := sl651.NewProtocol()
	msg, err := p.Parse(hexStr)
	if err != nil {
		fmt.Printf("Parse Error: %v\n", err)
		return
	}

	fmt.Printf("StationID: %s\n", msg.StationID)
	fmt.Printf("FunctionCode: %s\n", msg.FunctionCode)

	data, err := p.ConvertToStandardData(msg)
	if err != nil {
		fmt.Printf("Conversion Error: %v\n", err)
		return
	}

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}
