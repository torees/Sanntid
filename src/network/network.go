package network

import (
	"net"
	"os"
	"strings"
	"fmt"
	
)







func ClientConnectUDP(port string)*net.UDPConn{
	adress,err :=net.ResolveUDPAddr("udp","129.241.187.255"+port)
	if (err != nil){
		fmt.Println(adress,err)
	}

	connection,err := net.DialUDP("udp",nil,adress)
	if err == nil{
		fmt.Println("Connection achieved at : ",adress)
	}
	return connection
}

func ServerConnectUDP(port string)*net.UDPConn{
	
	adress,err := net.ResolveUDPAddr("udp",port)
	if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }

    connection, err := net.ListenUDP("udp",adress)
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
    return connection

}

func ClientSend(conn *net.UDPConn,msg []byte ){
	_,_= conn.Write(msg)
	}




func ServerListenUDP(conn *net.UDPConn,buf []byte)int{
	n,_,_ := conn.ReadFromUDP(buf)
	return n

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
