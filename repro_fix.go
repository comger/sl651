package main

import (
	"encoding/hex"
	"fmt"
	"sl651-platform/internal/sl651"
)

func main() {
	p := sl651.NewProtocol()
	id := "1090330854"

	body := make([]byte, 0)
	// Try the exact sequence from the simulator
	body = append(body, p.BuildTLV(0x22, 12.2, 1)...)
	body = append(body, p.BuildTLV(0x39, 25.2, 1)...)
	body = append(body, p.BuildTLV(0x38, 12.5, 2)...)

	fullFrame, _ := p.BuildMessage(id, 0x31, body)
	fmt.Printf("Generated Frame: %x\n", fullFrame)

	msg, err := p.Parse(hex.EncodeToString(fullFrame))
	if err != nil {
		fmt.Printf("Parse Error: %v\n", err)
		return
	}

	fmt.Printf("Payload: %x\n", msg.Payload)
	res, _ := p.ParseRealtimeData(msg.Payload)

	fmt.Println("\n--- Final Parsed Results ---")
	for k, v := range res {
		fmt.Printf("%s: %v\n", k, v)
	}
}
