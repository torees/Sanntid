package main

import (
	"../driver"
	"../elevManager"
	"../message"
	"../network"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"time"
)

const (
	UDPPort = ":20011"
)

const N_ELEVATORS = 3
const N_FLOORS = 4
const MAX_ORDER_COST = 25

type elevator struct {
	queue        elevManager.OrderQueue
	direction    int
	currentFloor int
	IP           string
}

func (elev elevator) cost(order elevManager.OrderQueue) (int, string) {
	// do cost calculation on order
	//return cost value and IP
	const dirCost = 2
	const distCost = 4
	const numOrderCost = 6
	cost := 0

	distanceCost := (elev.currentFloor - elev.findOrderFloor(order)) * distCost
	directionCost := 0

	if distanceCost < 0 {
		directionCost = dirCost
		distanceCost = int(math.Abs(float64(distanceCost)))
	}

	cost = elev.numOrdersInQueue()*numOrderCost + distanceCost + directionCost
	fmt.Println(cost, elev.IP, elev.queue)

	return cost, elev.IP
}

func (elev elevator) findOrderFloor(order elevManager.OrderQueue) int {
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
		if elev.queue.Internal[i] == 1 {
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
	driver.NetworkConnect(0)
	fmt.Println("My IP", myIP)

	//UDP channels
	UDPSendMsgChan := make(chan message.UDPMessage, 100)
	UDPPingReceivedChan := make(chan message.UDPMessage, 100)
	UDPMsgReceivedChan := make(chan message.UDPMessage, 100)

	checkNetworkConChan := make(chan bool)
	restartUDPSendChan := make(chan bool)

	//Channels to elevManager
	NewNetworkOrderToSM := make(chan elevManager.OrderQueue, 10)
	NewNetworkOrderFromSM := make(chan elevManager.OrderQueue, 10)
	stateUpdateFromSM := make(chan [2]int, 10)
	requestStateUpdateChan := make(chan bool)

	// Channels to master thread
	NewMsgToMasterChan := make(chan message.UDPMessage, 10)
	NewOrderFromMasterChan := make(chan message.UDPMessage, 10)
	lightCommandChan := make(chan elevManager.LightCommand, 10)
	elevatorAddedChan := make(chan string, 10)
	elevatorRemovedChan := make(chan string, 10)

	//Init sockets for sending ping and messages
	UDPlistenConn := network.ServerConnectUDP(UDPPort)
	startUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)

	// Goroutines
	go UDPlisten(UDPlistenConn, UDPPingReceivedChan, UDPMsgReceivedChan)
	go network.CheckNetworkConnection(checkNetworkConChan)
	go masterThread(lightCommandChan, elevatorAddedChan, elevatorRemovedChan, NewMsgToMasterChan, NewOrderFromMasterChan, myIP)
	go elevManager.ElevManager(requestStateUpdateChan, lightCommandChan, NewNetworkOrderFromSM, NewNetworkOrderToSM, stateUpdateFromSM)

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
				requestStateUpdateChan <- true
				fmt.Println("adding new elevator")


			}

		case msg := <-NewOrderFromMasterChan:
			UDPSendMsgChan <- msg

		case msg := <-UDPMsgReceivedChan:
			fmt.Println("new message on UPD", msg)
			fmt.Println("")
			// send udpmessage to correct routine
			switch msg.MessageId {
			case message.ElevatorStateUpdate, message.NewOrder:

				NewMsgToMasterChan <- msg
				//fmt.Println("New msg sent to master: ")

			case message.NewOrderFromMaster:
				//tenn lys på n heiser
				var light elevManager.LightCommand
				for i := 0; i < N_FLOORS; i++ {
					if msg.OrderQueue[i+4] == 1 {
						light = [3]int{0, i, 1}
						break

					}
					if msg.OrderQueue[i+8] == 1 {
						light = [3]int{1, i, 1}
						break
					}
				}
				lightCommandChan <- light
				///////////////////////////
				// Make all copies update elevators queues 
				NewMsgToMasterChan <- msg
				/////////////////////////////////////////////////////////
				if msg.ToIP == myIP {
					var order elevManager.OrderQueue
					for i := 0; i < 4; i++ {
						order.Internal[i] = msg.OrderQueue[i]
						order.Up[i] = msg.OrderQueue[(i + 4)]
						order.Down[i] = msg.OrderQueue[(i + 8)]
					}
					NewNetworkOrderToSM <- order
				}
			}

		case order := <-NewNetworkOrderFromSM:
			//create UDP message and send via UDP
			var msg message.UDPMessage
			msg.MessageId = message.NewOrder
			msg.FromIP = myIP
			for i := 0; i < 4; i++ {
				msg.OrderQueue[i] = order.Internal[i]
				msg.OrderQueue[(i + 4)] = order.Up[i]
				msg.OrderQueue[(i + 8)] = order.Down[i]
			}
			//calculate checksum?
			fmt.Println("order sent on UDP")

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
			time.Sleep(time.Millisecond*10)
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
			UDPMsgReceivedChan <- msg
			break
			//Fault tolerance, shut down?

		}

	}
}

func masterThread(lightCommandChan chan elevManager.LightCommand, elevatorAddedChan chan string, elevatorRemovedChan chan string, NewMsgToMasterChan chan message.UDPMessage, NewOrderFromMasterChan chan message.UDPMessage, myIP string) {
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

		case id := <-elevatorAddedChan:
			numberOfelevators += 1
			if numberOfelevators > N_ELEVATORS {
				//fault tolerance
				fmt.Println("To many elevators")
				os.Exit(0)
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

		case msg := <-NewMsgToMasterChan:
			
			var IP string
			var newOrder elevManager.OrderQueue

			for i := 0; i < N_FLOORS; i++ {
				newOrder.Internal[i] = msg.OrderQueue[i]
				newOrder.Up[i] = msg.OrderQueue[(i + 4)]
				newOrder.Down[i] = msg.OrderQueue[(i + 8)]

			}

			switch msg.MessageId {
			case message.NewOrder:
				
				orderCost := MAX_ORDER_COST
				if master {
					for _, elev := range connectedElev {
						tempOrderCost, tempIP := elev.cost(newOrder)
						if tempOrderCost < orderCost {
							orderCost = tempOrderCost
							IP = tempIP
						}
					}
					// this handles single elevator on network

					if IP == "" {
						msg.ToIP = IP
						IP = myIP

					} else {
						msg.ToIP = IP
					}

					//end of comment
					//update masters copy of the queue
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
					NewOrderFromMasterChan <- msg // send ON UDP

				}
				for i := 0; i < N_FLOORS; i++ {
					if newOrder.Internal[i] == 1 {
						elev = connectedElev[msg.FromIP]
						elev.queue.Internal[i] = 1
						connectedElev[msg.FromIP] = elev
					}
				}
				break

			case message.NewOrderFromMaster:

				if(!master){
					IP := msg.ToIP

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
				}

		case message.ElevatorStateUpdate:
			var light elevManager.LightCommand
			elev = connectedElev[msg.FromIP]
			fmt.Println("heis: ", msg.FromIP)
			elev.direction = msg.ElevatorStateUpdate[0]
			elev.currentFloor = msg.ElevatorStateUpdate[1]
			fmt.Println("current floor ", elev.currentFloor)
			fmt.Println("up queue ", elev.queue.Up)
			fmt.Println("\n\n\n\n\n\n\n\n")

			if elev.queue.Up[elev.currentFloor] == 1 {
				light = [3]int{0, elev.currentFloor, 0}
				lightCommandChan <- light
			}
			if elev.queue.Down[elev.currentFloor] == 1 {
				light = [3]int{1, elev.currentFloor, 0}
				lightCommandChan <- light
			}

			elev.queue.Up[elev.currentFloor] = 0
			elev.queue.Down[elev.currentFloor] = 0
			elev.queue.Internal[elev.currentFloor] = 0
			connectedElev[msg.FromIP] = elev
			break
		}
	}
}

}
