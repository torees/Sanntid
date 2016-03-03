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
	MessageSource string
	MessageSource string

	MessageId int

	OrderButton int
	ElevatorFloor int
	ElevatorFloorTarget int 



	Checksum int 
}

func UDPMessageEncode(Msg UDPMessage){

}

func UDPMessageDecode(Msg UDPMessage, UDParray []byte){
	
}