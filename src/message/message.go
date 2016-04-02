package message

import (
	"encoding/json"
)
//Message ID
const(
	Ping = 1
	ElevatorStateUpdate = 2
	NewOrder = 3
	NewOrderFromMaster = 4
)

type UDPMessage struct{
	MessageId int
	FromIP string
	ToIP string 
	OrderQueue [12]int// 
	ElevatorStateUpdate [2]int // [0] = direction, [1] = position
	Checksum int 
}

func UDPMessageEncode(Msg UDPMessage)([]byte, error){
	return json.Marshal(Msg)
}

func UDPMessageDecode(Msg *UDPMessage, UDParray []byte){
	json.Unmarshal(UDParray, Msg)
}

func CalculateChecksum(Msg *UDPMessage)int{ // not a very good crc, just for testing 
	c := Msg.MessageId%7+Msg.OrderQueue[0]%7+Msg.ElevatorStateUpdate[0]%7
	return c
}

////Main function for testing/////////////////////////
// func main(){
// 	var msg = UDPMessage{10,"hallo",71,32,43, 23}
// 	b,_ :=UDPMessageEncode(msg)
// 	var msg2 UDPMessage
// 	UDPMessageDecode(&msg2,b)
// 	fmt.Println(msg)
// 	fmt.Println(b)
// 	fmt.Println(msg2)
// 	fmt.Println(CalculateChecksum(&msg2)
// 	fmt.Println(CalculateChecksum(&msg))
// }
