package network

import (
	"net"
	"os"
	"strings"
	"fmt"
	
)







func ClientConnectUDP(port string)*net.UDPConn{
	adress,_ :=net.ResolveUDPAddr("udp","10.20.90.60"+port)
	conn,err := net.DialUDP("udp",nil,adress)
	if err == nil{
		fmt.Println("Connection achieved at : ",adress)
	}
	return conn
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

func ClientSend(conn *net.UDPConn,msg []byte ){
	_,_= conn.Write(msg)
	}




func ServerListenUDP(conn net.UDPConn,recvChan chan []byte){
	buf := make([]byte,1024)
	fmt.Println("Listening for messages on port")
	for{
		n,_,_ := conn.ReadFromUDP(buf)
		recvChan <- buf[0:n]
	}

}

func CheckNetworkConnection(networkAccessChannel chan bool){	
	for{
		ip := GetNetworkIP()
		if(ip == "::1"){
			networkAccessChannel<-false			
		}
	}
}

func GetNetworkIP()string{
	ipAdd,_ := net.InterfaceAddrs()		
	ip:=strings.Split(ipAdd[1].String(),"/")[0]
	return ip
}


//Main function for testing /// 


// func main(){
// 	ListenPort := ":54321"
// 	SendPort := ":12345"
// 	connectionChanListen := make(chan *net.UDPConn,10)
// 	connectionChanSend := make(chan *net.UDPConn)
// 	waitChan := make(chan int)
// 	recvChan := make(chan UDPmsg,5)



// 	fmt.Println("Starting server...")
// 	time.Sleep(time.Second *1)
// 	go ServerConnectUDP(ListenPort,connectionChanListen)
// 	go ServerListenUDP(connectionChanListen, recvChan)
// 	go serverPrint(recvChan)
// 	go ClientConnectUDP(SendPort,connectionChanSend)
// 	go ClientSend(connectionChanSend)

// 	//fmt.Println("Goroutines initialized")

// 	<-waitChan


// }

// func serverPrint(recvChan chan UDPmsg){
// 	for {
// 		fmt.Println("waiting ..")
// 		printmsg := <-recvChan
// 		fmt.Println("MSG: ",printmsg.msg)

		
// 	}
// }
