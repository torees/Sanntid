package network

import (
	"../driver"
	"fmt"
	"net"
	"os"
	"strings"
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

func ServerConnectUDP(port string) *net.UDPConn {

	adress, err := net.ResolveUDPAddr("udp", port)
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

func ClientSend(conn *net.UDPConn, msg []byte) {
	_, _ = conn.Write(msg)
}

func ServerListenUDP(conn *net.UDPConn, buf []byte) int {
	n, _, _ := conn.ReadFromUDP(buf)
	return n

}

func CheckNetworkConnection(networkAccessChannel chan bool) {
	network := true
	for {
		ip := GetNetworkIP()
		if ip == "::1" && network == true {
			network = false
			driver.NetworkConnect(0)
			networkAccessChannel <- false

		}
		if (ip != "::1") && !network {
			network = true
			driver.NetworkConnect(1)
			networkAccessChannel <- true
		}

	}
}

func GetNetworkIP() string {
	ipAdd, _ := net.InterfaceAddrs()
	ip := strings.Split(ipAdd[1].String(), "/")[0]
	return ip
}
