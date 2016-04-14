package network

import (
	. "../driver"
	. "../message"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

const (
	UDPPort = ":20011"
)

func CheckNetworkConnection(checkNetworkConChan chan bool) {
	network := true
	for {
		ip := GetNetworkIP()
		if ip == "::1" && network == true {
			network = false
			NetworkConnected(1)
			checkNetworkConChan <- false

		}
		if (ip != "::1") && !network {
			network = true
			NetworkConnected(0)
			checkNetworkConChan <- true
		}

	}
}

func GetNetworkIP() string {
	ipAdd, _ := net.InterfaceAddrs()
	ip := strings.Split(ipAdd[1].String(), "/")[0]
	return ip
}

func ClientConnectUDP(port string) *net.UDPConn {
	adress, err := net.ResolveUDPAddr("udp", "129.241.187.255"+port)
	if err != nil {
		fmt.Println("Could not resolve adress. Shutting down")
		os.Exit(0)
	}

	connection, err := net.DialUDP("udp", nil, adress)
	if err != nil {
		fmt.Println("Could not resolve socket. Shutting down")
		os.Exit(0)
	}
	return connection
}

func ServerConnectUDP() *net.UDPConn {
	adress, err := net.ResolveUDPAddr("udp", UDPPort)
	if err != nil {
		fmt.Println("Could not resolve adress. Shutting down")
		os.Exit(0)
	}

	connection, err := net.ListenUDP("udp", adress)
	if err != nil {
		fmt.Println("Could not resolve socket. Shutting down")
		os.Exit(0)
	}
	return connection

}

func StartUDPSend(UDPSendMsgChan chan UDPMessage, restartUDPSendChan chan bool, myIP string) {
	UDPSendConn := ClientConnectUDP(UDPPort)
	go UDPsend(UDPSendConn, UDPSendMsgChan, myIP, restartUDPSendChan)
}

func UDPsend(conn *net.UDPConn, UDPSendMsgChan chan UDPMessage, myIP string, restartUDPSendChan chan bool) {
	defer conn.Close()
	var ping UDPMessage
	ping.FromIP = myIP
	ping.MessageId = Ping
	encodedPing, _ := UDPMessageEncode(ping)
	ticker := time.NewTicker(time.Millisecond * 250).C
	for {
		select {
		case <-ticker:
			conn.Write(encodedPing)

		case msg := <-UDPSendMsgChan:
			msg.Checksum = msg.CalculateChecksum()
			encodedMsg, _ := UDPMessageEncode(msg)
			conn.Write(encodedMsg)

		case <-restartUDPSendChan:
			return
		}
	}

}

func UDPlisten(UDPPingReceivedChan chan UDPMessage, UDPMsgReceivedChan chan UDPMessage) {
	var msg UDPMessage
	conn := ServerConnectUDP()
	defer conn.Close()
	buf := make([]byte, 1024)

	for {

		numOfBytes, _, _ := conn.ReadFromUDP(buf)
		msgBuf := buf[0:numOfBytes]
		UDPMessageDecode(&msg, msgBuf)

		switch msg.MessageId {
		case Ping:
			UDPPingReceivedChan <- msg
			break
		case NewOrderFromMaster, NewOrder, ElevatorStateUpdate:
			if msg.CalculateChecksum() == msg.Checksum {
				UDPMsgReceivedChan <- msg
			}
			break
		}
	}
}
