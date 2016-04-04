package main

import (
	"../message"
	"../network"
	"../statemachine"
	"fmt"
	"math"
	"net"
	"sort"
	"time"
)

const (
	UDPPort = ":20011"
)

const N_ELEVATORS = 3
const N_FLOORS = 4

type elevator struct {
	queue        statemachine.OrderQueue
	direction    int
	currentFloor int
	IP           string
}

func (elev elevator) cost(order statemachine.OrderQueue) (int, string) {
	// do cost calculation on order
	//return cost value and IP
	const dirCost = 10
	const distCost = 5
	const numOrderCost = 4
	cost := 0

	distanceCost := (elev.currentFloor - elev.findOrderFloor(order)) * distCost
	directionCost := 0

	if distanceCost < 0 {
		directionCost = dirCost
		distanceCost = int(math.Abs(float64(distanceCost)))
	}

	cost = elev.numOrdersInQueue()*numOrderCost + distanceCost + directionCost
	fmt.Println(cost, elev.IP)

	return cost, elev.IP
}

func (elev elevator) findOrderFloor(order statemachine.OrderQueue) int {
	for i := 0; i < N_FLOORS; i++ {
		if order.Up[i] == 1 || order.Down[i] == 1 {
			return i
		}
	}
	return -1
}

func (elev elevator) numOrdersInQueue() int {
	numOrders := 0
	for i := 0; i < N_FLOORS; i++ {
		if elev.queue.Up[i] == 1 {
			numOrders += 1
		}
		if elev.queue.Down[i] == 1 {
			numOrders += 1
		}
	}
	return numOrders
}

func main() { //function should be renamed afterwards, this is just for testing
	var myIP string
	for {
		myIP = network.GetNetworkIP()
		if !(myIP == "::1") {
			break
		}
		fmt.Println("No network connection")
	}
	fmt.Println("My IP", myIP)

	//UDP channels
	UDPSendMsgChan := make(chan message.UDPMessage, 100)
	UDPPingReceivedChan := make(chan message.UDPMessage, 100)
	UDPMsgReceivedChan := make(chan message.UDPMessage, 100)

	checkNetworkConChan := make(chan bool)
	restartUDPSendChan := make(chan bool)

	//Channels to statemachine
	NewNetworkOrderToSM := make(chan statemachine.OrderQueue, 10)
	NewNetworkOrderFromSM := make(chan statemachine.OrderQueue, 10)
	stateUpdateFromSM := make(chan [2]int, 10)

	// Channels to master thread
	NewMsgToMasterChan := make(chan message.UDPMessage, 10)
	NewOrderFromMasterChan := make(chan message.UDPMessage, 10)

	elevatorAddedChan := make(chan string, 10)
	elevatorRemovedChan := make(chan string, 10)

	//Init sockets for sending ping and messages
	UDPlistenConn := network.ServerConnectUDP(UDPPort)
	startUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)

	// Goroutines
	go UDPlisten(UDPlistenConn, UDPPingReceivedChan, UDPMsgReceivedChan)
	go network.CheckNetworkConnection(checkNetworkConChan)
	go masterThread(elevatorAddedChan, elevatorRemovedChan, NewMsgToMasterChan, NewOrderFromMasterChan, myIP)
	go statemachine.StateMachine(NewNetworkOrderFromSM, NewNetworkOrderToSM, stateUpdateFromSM)

	connectedElevTimers := make(map[string]*time.Timer)

	for {
		select {
		case msg := <-UDPPingReceivedChan:
			_, exists := connectedElevTimers[msg.FromIP]
			if exists {
				connectedElevTimers[msg.FromIP].Reset(time.Second)

			} else {
				elevatorAddedChan <- msg.FromIP
				connectedElevTimers[msg.FromIP] = time.AfterFunc(time.Second, func() { deleteElevator(&connectedElevTimers, msg, elevatorRemovedChan) })
				fmt.Println("adding new elevator")

			}

		case msg := <-NewOrderFromMasterChan:
			UDPSendMsgChan <- msg

		case msg := <-UDPMsgReceivedChan:
			// send udpmessage to correct routine
			switch msg.MessageId {
			case message.ElevatorStateUpdate, message.NewOrder:
				fmt.Println("something to master")
				NewMsgToMasterChan <- msg
				//fmt.Println("New msg sent to master: ")

			case message.NewOrderFromMaster:
				if msg.ToIP == myIP {
					var order statemachine.OrderQueue
					for i := 0; i < 4; i++ {
						order.Up[i] = msg.OrderQueue[(i + 4)]
						order.Down[i] = msg.OrderQueue[(i + 8)]
					}
					NewNetworkOrderToSM <- order
				}
			}

		case order := <-NewNetworkOrderFromSM:
			//create UDP message and send via UDP
			fmt.Println("received new network order from SM")
			var msg message.UDPMessage
			msg.MessageId = message.NewOrder
			msg.FromIP = myIP
			for i := 0; i < 4; i++ {
				msg.OrderQueue[(i + 4)] = order.Up[i]
				msg.OrderQueue[(i + 8)] = order.Down[i]
			}
			//calculate checksum?
			fmt.Println("new order sent on network")
			UDPSendMsgChan <- msg

		case stateUpdate := <-stateUpdateFromSM:
			var msg message.UDPMessage
			msg.MessageId = message.ElevatorStateUpdate
			msg.ElevatorStateUpdate = stateUpdate
			msg.FromIP = myIP
			//msg.Checksum = CalculateCheckSum(msg)
			UDPSendMsgChan <- msg

		case haveNetwork := <-checkNetworkConChan:
			if haveNetwork {
				startUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)
			} else {
				restartUDPSendChan <- true
			}

		}

	}

}

func deleteElevator(connectedElevTimers *map[string]*time.Timer, msg message.UDPMessage, elevatorRemovedChan chan string) {
	elevatorRemovedChan <- msg.FromIP
	delete(*connectedElevTimers, msg.FromIP)
	fmt.Println("deleting elevator :", msg.FromIP)
}

func startUDPSend(UDPSendMsgChan chan message.UDPMessage, restartUDPSendChan chan bool, myIP string) {
	UDPSendConn := network.ClientConnectUDP(UDPPort)
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
			network.ClientSend(conn, encodedPing)

		case msg := <-UDPSendMsgChan:
			encodedMsg, _ := message.UDPMessageEncode(msg)
			network.ClientSend(conn, encodedMsg)
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

		numOfBytes := network.ServerListenUDP(conn, buf)
		msgBuffer := buf[0:numOfBytes]
		message.UDPMessageDecode(&msg, msgBuffer)

		switch msg.MessageId {
		case message.Ping:
			UDPPingReceivedChan <- msg
			break
		case message.NewOrderFromMaster, message.NewOrder, message.ElevatorStateUpdate:
			//fmt.Println("order received" ,msg)
			UDPMsgReceivedChan <- msg
			//fmt.Println("new network order received")
			break
			//Fault tolerance, shut down?

		}

	}
}

func masterThread(elevatorAddedChan chan string, elevatorRemovedChan chan string, NewMsgToMasterChan chan message.UDPMessage, NewOrderFromMasterChan chan message.UDPMessage, myIP string) {
	numberOfelevators := 0
	connectedElev := make(map[string]elevator)
	master := true
	var IPlist []string
	var elev elevator
	for {

		select {
		case elevatorIP := <-elevatorRemovedChan:
			IPlist = IPlist[:0]
			numberOfelevators -= 1
			delete(connectedElev, elevatorIP)
			if numberOfelevators != 0 {
				for key, _ := range connectedElev {
					IPlist = append(IPlist, key)
				}
				sort.Strings(IPlist)
				if IPlist[0] == myIP {
					master = true
				} else {
					master = false
				}
			}
			fmt.Println(IPlist)
			if numberOfelevators == 0 {
				master = true
			}

			// remove elevator object from list via IP-address to lost elevator?
		case id := <-elevatorAddedChan:
			numberOfelevators += 1
			if numberOfelevators > N_ELEVATORS {
				//fault tolerance
				fmt.Println("To many elevators")
			}

			elev.IP = id
			connectedElev[id] = elev
			IPlist = IPlist[:0]
			for key, _ := range connectedElev {
				IPlist = append(IPlist, key)
			}
			sort.Strings(IPlist)
			if IPlist[0] == myIP {
				master = true
			} else {
				master = false
			}
			fmt.Println(IPlist)

			//create new elevator object

		case msg := <-NewMsgToMasterChan:
			switch msg.MessageId {
			case message.NewOrder:
				var IP string
				var orderCost int
				var newOrder statemachine.OrderQueue
				if master {
					for i := 0; i < N_FLOORS; i++ {
						newOrder.Up[i] = msg.OrderQueue[(i + 4)]
						newOrder.Down[i] = msg.OrderQueue[(i + 8)]
					}

					for _, elev := range connectedElev {
						fmt.Println(elev)
						tempOrderCost, tempIP := elev.cost(newOrder)

						if tempOrderCost < orderCost {
							orderCost = tempOrderCost
							IP = tempIP
						}
					}
					// this handles single elevator on network
					if IP == "" {
						msg.ToIP = myIP
						IP = myIP
					} else {
						msg.ToIP = IP
					}

					//end of comment

					for i := 0; i < N_FLOORS; i++ {
						if newOrder.Up[i] == 1 {
							elev = connectedElev[IP]
							elev.queue.Up[i] = 1
							connectedElev[IP] = elev
						}
						if newOrder.Down[i] == 1 {
							elev = connectedElev[IP]
							elev.queue.Down[i] = 1
							connectedElev[IP] = elev
						}
					}
					msg.MessageId = message.NewOrderFromMaster
					//fmt.Print("happening all the time")
					NewOrderFromMasterChan <- msg
					//do something with msg, find out which elevator should take it.
				}
				break
			case message.ElevatorStateUpdate:
				elev = connectedElev[msg.FromIP]
				fmt.Println("Previous info on elevator: ", elev)
				elev.direction = msg.ElevatorStateUpdate[0]

				elev.currentFloor = msg.ElevatorStateUpdate[1]
				elev.queue.Up[elev.currentFloor] = 0
				elev.queue.Down[elev.currentFloor] = 0
				connectedElev[msg.FromIP] = elev
				fmt.Println("master knows this of elev: ", elev)

				//fmt.Println(connectedElev)
				break
			}
		}
	}

}
