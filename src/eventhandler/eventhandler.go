package main

import(
	"fmt"
	"../network"
	"time"
	"net"
	"../message"
	"../statemachine"
)

const(
	UDPPort = ":20011"
	)

const N_ELEVATORS = 3


type elevator struct{
	var queue statemachine.OrderQueue
	var direction int
	var currentFloor int 
	


}


func main(){ //function should be renamed afterwards, this is just for testing
	var myIp string
	for{
		 myIp = network.GetNetworkIP()
		if(!(myIp == "::1")){
			break
		}
		fmt.Println("No network connection")
	}
	fmt.Println("My IP", myIp)


	UDPSendMsgChan := make(chan message.UDPMessage,100)
	UDPPingReceivedChan := make(chan message.UDPMessage,100)
	UDPOrderReceivedChan := make(chan message.UDPMessage,100)
	UDPElevatorStateUpdateChan := make(chan message.UDPMessage,100)
	checkNetworkConChan := make(chan bool)

	//listenPingChan := make(chan bool, 1)
	// sendElevComChan :=make(chan *net.UDPConn, 10)
	// listenElevComChan := make(chan *net.UDPConn, 10)

	//init sockets for sending ping and messages 
	UDPSendConn:= network.ClientConnectUDP(UDPPort)
	//sendElevComConn := network.ClientConnectUDP(sendElevCom)

	UDPlistenConn := network.ServerConnectUDP(UDPPort)
	//listenElevComConn := network.ServerConnectUDP(listenElevCom)

	go UDPsend(UDPSendConn, UDPSendMsgChan, myIp)
	go UDPlisten(UDPlistenConn, UDPPingReceivedChan, UDPOrderReceivedChan,UDPElevatorStateUpdateChan)
	go network.CheckNetworkConnection(checkNetworkConChan)
	//connectedElevIP := [N_ELEVATORS]string

	connectedElevTimers := make(map[string]*time.Timer)


	for{
		select{
			case msg := <-UDPPingReceivedChan:
				fmt.Println("ping received from: ", msg.IP)
				_,exists := connectedElevTimers[msg.IP]
				if exists{
					connectedElevTimers[msg.IP].Reset(time.Second)
				}else{
					connectedElevTimers[msg.IP] = time.AfterFunc(time.Second, func(){ delete(connectedElevTimers,msg.IP) } )
				}



				//add msg.IP to ip list IFNOT there already OR number of elevators = N

			//case msg := <- UDPOrderReceivedChan:
				//fmt.Println("order received: ", msg.OrderQueue)
			//case msg := <- UDPElevatorStateUpdateChan:
				//fmt.Println("State update : ", msg.ElevatorStateUpdate)
			case <- checkNetworkConChan:
				//network down, handle 

		}
		for key,_:= range connectedElevTimers{
			fmt.Println(key)
		}

	}
	

}






func UDPsend(conn *net.UDPConn, UDPMsgChan chan message.UDPMessage, IP string){
	defer conn.Close()
	var ping message.UDPMessage
	
	
//msg created for testing purposes --------------------
	var msg message.UDPMessage
	msg.IP = IP
	msg.MessageId = message.NewOrder
	msg.OrderQueue = [12]int{1,0,0,0,0,0,0,0,0,0,0,0}
	ticker2 := time.NewTicker(time.Millisecond*2500).C
	var msg2 message.UDPMessage
	msg2.IP = IP
	msg2.MessageId = message.ElevatorStateUpdate
	msg.ElevatorStateUpdate = [2]int{1,0}
	ticker3 := time.NewTicker(time.Millisecond*3500).C
//----------------------------

	ping.IP = IP
	ping.MessageId = message.Ping
	encodedPing,_ :=message.UDPMessageEncode(ping)
	ticker := time.NewTicker(time.Millisecond*250).C



	defer conn.Close()
	for{
		select{
			case <- ticker:
				network.ClientSend(conn, encodedPing)

			case msg := <-UDPMsgChan:
				encodedMsg,_:= message.UDPMessageEncode(msg)
				network.ClientSend(conn, encodedMsg)
			
			// testing --------------------------	
			case <- ticker2:
				UDPMsgChan <- msg
			case <- ticker3:
				UDPMsgChan <- msg2		
			//-------------------------------------	
		}
	}

}

func UDPlisten(conn *net.UDPConn, UDPPingReceivedChan chan message.UDPMessage, UDPOrderReceivedChan chan message.UDPMessage, UDPElevatorStateUpdateChan chan message.UDPMessage){
	defer conn.Close()
	var msg message.UDPMessage
	buf := make([]byte,1024)
	for{
		
		n := network.ServerListenUDP(conn, buf)
		b := buf[0:n]
		message.UDPMessageDecode(&msg,b)

		switch msg.MessageId{
			case message.Ping:
				UDPPingReceivedChan <- msg
				break
			case message.ElevatorStateUpdate:
				UDPElevatorStateUpdateChan <- msg
				break
			case message.NewOrder:
				UDPOrderReceivedChan <- msg
				break
			default:
				//Fault tolerance, shut down?  

		}


	}
}





func masterThread(flagChan){
	numberOfelevators := 1

	select{
	case id:=<-elevatorRemoved:
		numberOfelevators -= 1
		// remove elevator object from list via IP-address to lost elevator? 
	case id:=<- elevatorAdded:
		numberOfelevators += 1
		if(numberOfelevators > N_ELEVATORS){
			//fault tolerance
			fmt.Println("To many elevators") 
		}
		//create new elevator object 


	case msg:= <- newOrders:
		//do something with msg, find out which elevator should take it.
	case msg:= <- stateUpdates:
	default

	}
}

