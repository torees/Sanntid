package network

import (
	"../driver"
	"../message"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	UDPPort = ":20011"
)

func ClientConnectUDP(port string) *net.UDPConn {
	adress, err := net.ResolveUDPAddr("udp", "129.241.187.255"+port)
	if err != nil {
		fmt.Println(adress, err)
	}

	connection, err := net.DialUDP("udp", nil, adress)
	if err == nil {
		fmt.Println("Connection achieved at : ", adress)
	}
	return connection
}

func ServerConnectUDP() *net.UDPConn {

	adress, err := net.ResolveUDPAddr("udp", UDPPort)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}

	connection, err := net.ListenUDP("udp", adress)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
	return connection

}

func clientSend(conn *net.UDPConn, msg []byte) {
	_, _ = conn.Write(msg)
}

func serverListenUDP(conn *net.UDPConn, buf []byte) int {
	n, _, _ := conn.ReadFromUDP(buf)
	return n

}

func CheckNetworkConnection(checkNetworkConChan chan bool) {
	network := true
	for {
		ip := GetNetworkIP()
		if ip == "::1" && network == true {
			network = false
			driver.NetworkConnect(1)
			checkNetworkConChan <- false

		}
		if (ip != "::1") && !network {
			network = true
			driver.NetworkConnect(0)
			checkNetworkConChan <- true
		}

	}
}

func GetNetworkIP() string {
	ipAdd, _ := net.InterfaceAddrs()
	ip := strings.Split(ipAdd[1].String(), "/")[0]
	return ip
}

func StartUDPSend(UDPSendMsgChan chan message.UDPMessage, restartUDPSendChan chan bool, myIP string) {
	UDPSendConn := ClientConnectUDP(UDPPort)
	go UDPsend(UDPSendConn, UDPSendMsgChan, myIP, restartUDPSendChan)
}

func UDPsend(conn *net.UDPConn, UDPSendMsgChan chan message.UDPMessage, IP string, restartUDPSendChan chan bool) {
	defer conn.Close()
	var ping message.UDPMessage
	ping.FromIP = IP
	ping.MessageId = message.Ping
	encodedPing, _ := message.UDPMessageEncode(ping)
	ticker := time.NewTicker(time.Millisecond * 250).C
	for {
		select {
		case <-ticker:
			clientSend(conn, encodedPing)

		case msg := <-UDPSendMsgChan:
			/*if msg.MessageId == 4 {
				fmt.Println("new UDP order:", msg.OrderQueue)
			}*/
			encodedMsg, _ := message.UDPMessageEncode(msg)
			clientSend(conn, encodedMsg)
			time.Sleep(time.Millisecond)
		case <-restartUDPSendChan:
			return
		}
	}

}

func UDPlisten(conn *net.UDPConn, UDPPingReceivedChan chan message.UDPMessage, UDPMsgReceivedChan chan message.UDPMessage) {
	defer conn.Close()
	var msg message.UDPMessage
	buf := make([]byte, 1024)
	for {

		numOfBytes := serverListenUDP(conn, buf)
		msgBuffer := buf[0:numOfBytes]
		message.UDPMessageDecode(&msg, msgBuffer)

		switch msg.MessageId {
		case message.Ping:
			UDPPingReceivedChan <- msg
			break
		case message.NewOrderFromMaster, message.NewOrder, message.ElevatorStateUpdate:
			UDPMsgReceivedChan <- msg
			break
			//Fault tolerance, shut down?

		}

	}
}
