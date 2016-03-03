package main

import(
	"fmt"
	"../network"
	"time"
	"net"
	"../message"
)

const(
	sendPingPort = "69696"
	ListenPingPort = "69697"
	listenElevCom = "69698"
	sendElevCom = "69699"
	)



func main(){ //function should be renamed afterwards, this is just for testing
	var myIp string
	waitChan := make(chan int)
	for{
		 myIp = network.GetNetworkIP()
		if(!(myIp == "::1")){
			break
		}
		fmt.Println("No network connection")
	}
	fmt.Println("My IP", myIp)

	//sendPingChan := make(chan *net.UDPConn,10)
	//listenPingChan := make(chan *net.UDPConn,10)
	// sendElevComChan :=make(chan *net.UDPConn, 10)
	// listenElevComChan := make(chan *net.UDPConn, 10)

	//init sockets for sending ping and messages 
	sendPingConn:= network.ClientConnectUDP(sendPingPort)
	//sendElevComConn := network.ClientConnectUDP(sendElevCom)

	//listenPingConn := network.ServerConnectUDP(ListenPingPort)
	//listenElevComConn := network.ServerConnectUDP(listenElevCom)

	go sendPing(sendPingConn)

	<-waitChan
	

}

func sendPing(conn *net.UDPConn){
	pingMsg := message.UDPMessage{message.Ping,"",0,0,0,0}
	encodedMsg,_ :=message.UDPMessageEncode(pingMsg)

	for{
		network.ClientSend(conn, encodedMsg)
		time.Sleep(time.Millisecond*250)

	}
}










