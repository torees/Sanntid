package main

import(
	"fmt"
	"../network"
	"time"
	"net"
	"../message"
)

const(
	sendPingPort = ":20011"
	ListenPingPort = ":33333"
	listenElevCom = ":40002"
	sendElevCom = ":40003"
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
	//listenPingChan := make(chan bool, 1)
	// sendElevComChan :=make(chan *net.UDPConn, 10)
	// listenElevComChan := make(chan *net.UDPConn, 10)

	//init sockets for sending ping and messages 
	sendPingConn:= network.ClientConnectUDP(sendPingPort)
	//sendElevComConn := network.ClientConnectUDP(sendElevCom)

	listenPingConn := network.ServerConnectUDP(ListenPingPort)
	//listenElevComConn := network.ServerConnectUDP(listenElevCom)

	go sendPing(sendPingConn)
	go listenPing(listenPingConn)

	<-waitChan
	

}

func sendPing(conn *net.UDPConn){
	pingMsg := message.UDPMessage{message.Ping,"",0,0,0,0}
	encodedMsg,_ :=message.UDPMessageEncode(pingMsg)
	defer conn.Close()
	for{
		network.ClientSend(conn, encodedMsg)
		time.Sleep(time.Millisecond*250)
		fmt.Println("ping sent")

	}
}

func listenPing(conn *net.UDPConn){
	var ping message.UDPMessage
	buf := make([]byte,1024)
	defer conn.Close()
	for{
		n := network.ServerListenUDP(conn, buf)
		//fmt.Println("buf", buf[0:n])
		b := buf[0:n]
		message.UDPMessageDecode(&ping,b)
		fmt.Println(ping)
	}
}










