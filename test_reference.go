package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sl651-platform/internal/sl651"
)

func main() {
	p := sl651.NewProtocol()
	hexStr := "7e7e011090330853a00032004d020121260202182021f1f1109033085350f0f026020218202219000000272b0000000000392300000134371b000000302b000004300020190000001f1900000026190000001a19000000381211530388ad"

	msg, err := p.Parse(hexStr)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("Station ID: %s, Function: %s\n", msg.StationID, msg.FunctionCode)

	data, err := p.ParseRealtimeData(msg.Payload)
	if err != nil {
		log.Fatalf("ParseRealtimeData error: %v", err)
	}

	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println("Parsed Results:")
	fmt.Println(string(b))
}
