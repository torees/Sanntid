package message

import{
	
	"encoding/json"
}

//Message types 
const{
	Ping = 0
	OrderButtonPushed = 1
	ElevatorStateUpdate = 2
	QueueNewOrder = 3
}

type UDPMessage struct{
	MessageId int
	MessageTargetIP string
	OrderButton int
	ElevatorFloor int
	ElevatorFloorTarget int 



	Checksum int 
}

func UDPMessageEncode(Msg UDPMessage){

}

func UDPMessageDecode(Msg UDPMessage, UDParray []byte){
	
}

func (msg *UDPMessage)CalculateChecksum() int{
	msg.Checksum = msg.MessageId%7+msg.OrderButton%7+msg.OrderButton%7
	msg.Checksum += msg.ElevatorFloor%7
}