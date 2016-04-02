package main

import(
	"fmt"
	"../network"
	"time"
	"net"
	"../message"
	"../statemachine"
	"sort"
)

const(
	UDPPort = ":20011"
	)

const N_ELEVATORS = 3
const N_FLOORS = 4


type elevator struct{
	queue statemachine.OrderQueue
	direction int
	currentFloor int 
	IP string 
}
func (elev elevator)cost(order OrderQueue)(float64, string){
	// do cost calculation on order
	//return cost value and IP 
	cost := 0.000
	return cost, elev.IP
}


type OrderQueue struct {
	internal []int
	down     []int
	up       []int
}



func main(){ //function should be renamed afterwards, this is just for testing
	var myIP string
	for{
		 myIP = network.GetNetworkIP()
		if(!(myIP == "::1")){
			break
		}
		fmt.Println("No network connection")
	}
	fmt.Println("My IP", myIP)


	UDPSendMsgChan := make(chan message.UDPMessage,100)
	UDPPingReceivedChan := make(chan message.UDPMessage,100)
	UDPOrderReceivedChan := make(chan message.UDPMessage,100)
	UDPElevatorStateUpdateChan := make(chan message.UDPMessage,100)
	checkNetworkConChan := make(chan bool)
	orderFromMasterChan := make(chan message.UDPMessage, 10)
	orderToMasterChan := make(chan message.UDPMessage, 10)
	stateUpdateToMasterChan := make(chan message.UDPMessage, 10)
	elevatorAddedChan := make(chan string, 10)
	elevatorRemovedChan := make(chan string, 10)


	//listenPingChan := make(chan bool, 1)
	// sendElevComChan :=make(chan *net.UDPConn, 10)
	// listenElevComChan := make(chan *net.UDPConn, 10)

	//init sockets for sending ping and messages 
	UDPSendConn:= network.ClientConnectUDP(UDPPort)
	//sendElevComConn := network.ClientConnectUDP(sendElevCom)

	UDPlistenConn := network.ServerConnectUDP(UDPPort)
	//listenElevComConn := network.ServerConnectUDP(listenElevCom)

	go UDPsend(UDPSendConn, UDPSendMsgChan, myIP)
	go UDPlisten(UDPlistenConn, UDPPingReceivedChan, UDPOrderReceivedChan,UDPElevatorStateUpdateChan)
	go network.CheckNetworkConnection(checkNetworkConChan)
	go masterThread(elevatorAddedChan, elevatorRemovedChan, stateUpdateToMasterChan,orderToMasterChan,orderFromMasterChan, myIP)
	go statemachine.StateMachine()
	//connectedElevIP := [N_ELEVATORS]string

	connectedElevTimers := make(map[string]*time.Timer)


	for{
		select{
			case msg := <-UDPPingReceivedChan:
				_,exists := connectedElevTimers[msg.FromIP]
				if exists{
					connectedElevTimers[msg.FromIP].Reset(time.Second)

				}else{
					elevatorAddedChan <- msg.FromIP
					connectedElevTimers[msg.FromIP] = time.AfterFunc(time.Second, func(){ deleteElevator(&connectedElevTimers,msg, elevatorRemovedChan)} )
					fmt.Println("adding new elevator")
				}

			case order := <- orderFromMasterChan:
				UDPSendMsgChan <- order	

			case msg := <- UDPOrderReceivedChan:
				orderToMasterChan <- msg
				fmt.Println("order received: ", msg.OrderQueue)
			
			case msg := <- UDPElevatorStateUpdateChan:
				stateUpdateToMasterChan <- msg
				fmt.Println("State update : ", msg.ElevatorStateUpdate)
			
			//case <- checkNetworkConChan:
				//network down, handle 

		}

	}
	

}


func deleteElevator(connectedElevTimers *map[string]*time.Timer, msg message.UDPMessage, elevatorRemovedChan chan string){
	elevatorRemovedChan <- msg.FromIP
	delete(*connectedElevTimers,msg.FromIP)
}



func UDPsend(conn *net.UDPConn, UDPMsgChan chan message.UDPMessage, IP string){
	defer conn.Close()
	var ping message.UDPMessage
	

	ping.FromIP = IP
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
			
		}
	}

}

func UDPlisten(conn *net.UDPConn, UDPPingReceivedChan chan message.UDPMessage, UDPOrderReceivedChan chan message.UDPMessage, UDPElevatorStateUpdateChan chan message.UDPMessage){
	defer conn.Close()
	var msg message.UDPMessage
	buf := make([]byte,1024)
	for{
		
		numOfBytes := network.ServerListenUDP(conn, buf)
		msgBuffer := buf[0:numOfBytes]
		message.UDPMessageDecode(&msg,msgBuffer)

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





func masterThread(elevatorAddedChan chan string, elevatorRemovedChan chan string, stateUpdateToMasterChan chan message.UDPMessage, orderToMasterChan chan message.UDPMessage,orderFromMasterChan chan message.UDPMessage, myIP string){
	numberOfelevators := 0
	connectedElev := make(map[string]elevator)
	master:= true
	var IPlist []string
	var elev elevator
	for{
		select{
			case id:=<-elevatorRemovedChan:
				
				IPlist = IPlist[:0]
				numberOfelevators -= 1
				delete(connectedElev,id)
				for key,_ := range connectedElev{
					IPlist = append(IPlist,key)
				}
				sort.Strings(IPlist)
				if(IPlist[0] == myIP){
					master = true
				}else{
					master = false
				}


				// remove elevator object from list via IP-address to lost elevator? 
			case id:=<- elevatorAddedChan:
				numberOfelevators += 1
				if(numberOfelevators > N_ELEVATORS){
					//fault tolerance
					fmt.Println("To many elevators") 
				}
				
				elev.IP = id
				connectedElev[id] = elev
				IPlist = IPlist[:0]
				for key,_ := range connectedElev{
					IPlist = append(IPlist,key)
				}
				sort.Strings(IPlist)
				if(IPlist[0] == myIP){
					master = true
				}else{
					master = false
				}

				
				
				//create new elevator object 


			case msg:= <- orderToMasterChan:
				var IP string
				var orderCost float64
				if(master){
					var newOrder OrderQueue
					newOrder.up = msg.OrderQueue[4:7]
					newOrder.down = msg.OrderQueue[8:11]
					for _,elev := range connectedElev{
						tempOrderCost,tempIP := elev.cost(newOrder)
						if(tempOrderCost < orderCost){
							orderCost = tempOrderCost
							IP = tempIP
						}
					}
					msg.ToIP = IP
					orderFromMasterChan <- msg
					//do something with msg, find out which elevator should take it.
				}
			case msg:= <- stateUpdateToMasterChan:
				elev.direction = msg.ElevatorStateUpdate[0]
				elev.currentFloor = msg.ElevatorStateUpdate[1]
				connectedElev[msg.FromIP] = elev

			}
	}

}

