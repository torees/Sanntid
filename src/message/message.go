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
	OrderQueue          [12]int //[0-3] Internal, [4-7] Down, [8-11] Up
	ElevatorStateUpdate [2]int  // [0] Direction, [1] Position
	Checksum            int
}

func UDPMessageEncode(Msg UDPMessage) ([]byte, error) {
	return json.Marshal(Msg)
}

func UDPMessageDecode(Msg *UDPMessage, UDParray []byte) {
	json.Unmarshal(UDParray, Msg)
}

func (msg UDPMessage) CalculateChecksum() int {
	c := msg.MessageId%7 + msg.OrderQueue[0]%7 + msg.ElevatorStateUpdate[0]%7
	return c
}
