package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

type UDPmsg struct{
	size int,
	msg string

}


func main(){
	port := ":12345"
	connectionChan := make(chan UDPConn)
	recvChan := make(chan UDPmsg,5)

	go ServerConnectUDP(port,connectionChan)
	go ServerListenUDP(connectionChan, recvChan)


}


func ServerPrint(recvChan chan UDPmsg){
	for {
		printmsg := <-recvChan
		fmt.Println("MSG: ",printmsg.msg)

}


func ServerListenUDP(connectionChan chan UDPConn,recvChan chan UDPmsg){
	buf := make([]byte,1024)
	for{
		
		conn :=<-connectionChan
		n,addr,err := conn.ReadFromUDP(buf)

		dummymsg = UDPmsg{n,buf(0:n)}
		recvChan <- dummymsg
		connectionChan <- conn
	}

}



func ServerConnectUDP(port string, connectionChan chan UDPConn){
	ServAddr,err := net.ResolveUDPAddr("udp",port)
	if err  != nil {
        fmt.Println("Error in resolve: " , err)
        os.Exit(0)
    }

    ServConn, err := net.ListenUDP("udp",ServAddr)
    if err  != nil {
        fmt.Println("Error in resolve: " , err)
        os.Exit(0)
    }
    connectionChan <- ServConn



}