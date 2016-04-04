package message

import (
	"encoding/json"
)

//Message ID
const (
	Ping                = 1
	ElevatorStateUpdate = 2
	NewOrder            = 3
	NewOrderFromMaster  = 4
)

type UDPMessage struct {
	MessageId           int
	FromIP              string
	ToIP                string
	OrderQueue          [12]int //[0] internal [4] Down [8] Up
	ElevatorStateUpdate [2]int  // [0] = direction, [1] = position
	Checksum            int
}

func UDPMessageEncode(Msg UDPMessage) ([]byte, error) {
	return json.Marshal(Msg)
}

func UDPMessageDecode(Msg *UDPMessage, UDParray []byte) {
	json.Unmarshal(UDParray, Msg)
}

func CalculateChecksum(Msg *UDPMessage) int { // not a very good crc, just for testing
	c := Msg.MessageId%7 + Msg.OrderQueue[0]%7 + Msg.ElevatorStateUpdate[0]%7
	return c
}
