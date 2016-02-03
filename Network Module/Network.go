package main

import (
	"fmt"
	"net"
	"os"
	"time"
	"bufio"
	
)

type UDPmsg struct{
	size int
	msg string

}




func main(){
	ListenPort := ":12345"
	SendPort := ":12345"
	connectionChanListen := make(chan *net.UDPConn,10)
	connectionChanSend := make(chan *net.UDPConn)
	waitChan := make(chan int)
	recvChan := make(chan UDPmsg,5)



	fmt.Println("Starting server...")
	time.Sleep(time.Second *1)
	go ServerConnectUDP(ListenPort,connectionChanListen)
	go ServerListenUDP(connectionChanListen, recvChan)
	go serverPrint(recvChan)
	go ClientConnectUDP(SendPort,connectionChanSend)
	go ClientSend(connectionChanSend)

	//fmt.Println("Goroutines initialized")

	<-waitChan


}

func ClientConnectUDP(port string, connectionChanSend chan *net.UDPConn){
	adress,_ :=net.ResolveUDPAddr("udp"," "+port)
	conn,err := net.DialUDP("udp",nil,adress)
	if err == nil{
		fmt.Println("Connection achieved at : ",adress)
	}
	connectionChanSend <- conn
}

func ClientSend(connectionChanSend chan *net.UDPConn){
	conn := <-connectionChanSend
	keyread := bufio.NewReader(os.Stdin)
	for{
		fmt.Println("Type: ")
		text,_ := keyread.ReadString('\n')
		msg := []byte(text)
		conn.Write(msg)
		time.Sleep(time.Second*1)
	}

}






func ServerConnectUDP(port string, connectionChanListen chan *net.UDPConn){
	
	ServAddr,err := net.ResolveUDPAddr("udp",port)
	if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }

    ServConn, err := net.ListenUDP("udp",ServAddr)
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
    fmt.Println("UDP connection established...")
    connectionChanListen <- ServConn

}



func serverPrint(recvChan chan UDPmsg){
	for {
		fmt.Println("waiting ..")
		printmsg := <-recvChan
		fmt.Println("MSG: ",printmsg.msg)

		
	}
}


func ServerListenUDP(connectionChanListen chan *net.UDPConn,recvChan chan UDPmsg){
	buf := make([]byte,1024)
	fmt.Println("Listening for messages on port")
	for{
		fmt.Println("Listening...")
		conn := <-connectionChanListen
		
		n,_,_ := conn.ReadFromUDP(buf)
		
		dummymsg := UDPmsg{ n, string(buf[0:n])}
		recvChan <- dummymsg
		connectionChanListen <- conn

		time.Sleep(time.Second*1)

	}

}



