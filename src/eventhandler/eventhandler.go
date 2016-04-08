package main

import (
	. "../driver"
	"../elevManager"
	"../message"
	"../network"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

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
	//fmt.Println(cost, elev.IP, elev.queue)

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

func (elev elevator) newOrder(order elevManager.OrderQueue) bool {

	for floor := 0; floor < N_FLOORS; floor++ {
		if elev.queue.Up[floor] == 1 && order.Up[floor] == 1 {
			return false
		}
		if elev.queue.Down[floor] == 1 && order.Down[floor] == 1 {
			return false
		}

	}
	return true
}

func main() { //function should be renamed afterwards, this is just for testing
	var myIP string
	for {
		myIP = network.GetNetworkIP()
		if !(myIP == "::1") {
			break
		}
		fmt.Println("No network connection")
		time.Sleep(time.Second * 1)
	}
	NetworkConnect(0)
	fmt.Println("My IP", myIP)

	//UDP channels
	UDPSendMsgChan := make(chan message.UDPMessage, 100)
	UDPPingReceivedChan := make(chan message.UDPMessage, 100)
	UDPMsgReceivedChan := make(chan message.UDPMessage, 100)

	checkNetworkConChan := make(chan bool)
	restartUDPSendChan := make(chan bool)

	//Channels to elevManager
	NewNetworkOrderToSM := make(chan elevManager.OrderQueue, 100)
	NewNetworkOrderFromSM := make(chan elevManager.OrderQueue, 100)
	stateUpdateFromSM := make(chan message.UDPMessage, 100)
	requestStateUpdateChan := make(chan bool, 100)

	// Channels to master thread
	NewMsgToMasterChan := make(chan message.UDPMessage, 100)
	NewOrderFromMasterChan := make(chan message.UDPMessage, 100)
	lightCommandChan := make(chan elevManager.LightCommand, 100)
	elevatorAddedChan := make(chan string, 100)
	elevatorRemovedChan := make(chan string, 100)

	//Init sockets for sending ping and messages
	UDPlistenConn := network.ServerConnectUDP()
	network.StartUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)

	// Goroutines
	go network.UDPlisten(UDPlistenConn, UDPPingReceivedChan, UDPMsgReceivedChan)
	go network.CheckNetworkConnection(checkNetworkConChan)
	go masterThread(lightCommandChan, elevatorAddedChan, elevatorRemovedChan, NewMsgToMasterChan, NewOrderFromMasterChan, myIP)
	go elevManager.ElevManager(requestStateUpdateChan, lightCommandChan, NewNetworkOrderFromSM, NewNetworkOrderToSM, stateUpdateFromSM)

	connectedElevTimers := make(map[string]*time.Timer)
	offline := false
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

			// send udpmessage to correct routine
			switch msg.MessageId {
			case message.ElevatorStateUpdate, message.NewOrder:
				fmt.Println("hello debuggin new order received on UDP")
				NewMsgToMasterChan <- msg

			case message.NewOrderFromMaster:
				//tenn lys på n heiser
				var light elevManager.LightCommand
				for i := 0; i < N_FLOORS; i++ {
					if msg.OrderQueue[i+4] == 1 {
						light = [3]int{0, i, 1}
						lightCommandChan <- light
						break

					}
					if msg.OrderQueue[i+8] == 1 {
						light = [3]int{1, i, 1}
						lightCommandChan <- light
						break
					}
				}

				///////////////////////////
				// Make all copies update elevators queues
				NewMsgToMasterChan <- msg
				/////////////////////////////////////////////////////////
				if msg.ToIP == myIP {
					var order elevManager.OrderQueue
					for i := 0; i < 4; i++ {
						order.Internal[i] = msg.OrderQueue[i]
						order.Up[i] = msg.OrderQueue[(i + 8)]
						order.Down[i] = msg.OrderQueue[(i + 4)]
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
				msg.OrderQueue[(i + 4)] = order.Down[i]
				msg.OrderQueue[(i + 8)] = order.Up[i]

			}
			//calculate checksum?
			if !offline {
				UDPSendMsgChan <- msg
			}
		case msg := <-stateUpdateFromSM:
			msg.MessageId = message.ElevatorStateUpdate
			msg.FromIP = myIP
			//msg.Checksum = CalculateCheckSum(msg)
			if !offline {
				UDPSendMsgChan <- msg
			}
		case haveNetwork := <-checkNetworkConChan:
			if haveNetwork {
				offline = false
				network.StartUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)
			} else {
				offline = true
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

func masterThread(lightCommandChan chan elevManager.LightCommand, elevatorAddedChan chan string, elevatorRemovedChan chan string, NewMsgToMasterChan chan message.UDPMessage, NewOrderFromMasterChan chan message.UDPMessage, myIP string) {
	numberOfelevators := 0
	connectedElev := make(map[string]elevator)
	master := true
	offline := false
	var IPlist []string
	var elev elevator
	for {
		select {
		case elevatorIP := <-elevatorRemovedChan:
			IPlist = IPlist[:0]
			numberOfelevators -= 1

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
			//get all orders external from elevatorIP and send internally
			elev = connectedElev[elevatorIP]
			tempqueue := elev.queue
			delete(connectedElev, elevatorIP)
			//ask elevManager do delete queue externals
			if elevatorIP == myIP {

				offline = true
			}

			if numberOfelevators == 0 {
				master = true
			}
			var msg message.UDPMessage
			msg.MessageId = message.NewOrder
			msg.FromIP = myIP
			for floor := 0; floor < N_FLOORS; floor++ {
				if tempqueue.Up[floor] == 1 {
					msg.OrderQueue[floor+8] = 1
					NewMsgToMasterChan <- msg
				}
				if tempqueue.Down[floor] == 1 {
					msg.OrderQueue[floor+4] = 1
					NewMsgToMasterChan <- msg
				}
			}

		case id := <-elevatorAddedChan:
			numberOfelevators += 1
			offline = false
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

			switch msg.MessageId {

			case message.NewOrder:
				for i := 0; i < N_FLOORS; i++ {
					newOrder.Internal[i] = msg.OrderQueue[i]
					newOrder.Down[i] = msg.OrderQueue[(i + 4)]
					newOrder.Up[i] = msg.OrderQueue[(i + 8)]

				}

				fmt.Println("order recieved ", newOrder)
				uniqueOrder := true
				orderCost := MAX_ORDER_COST
				if master && !offline {
					for _, elev := range connectedElev {
						if !elev.newOrder(newOrder) {

							uniqueOrder = false
							fmt.Println("not unique")
						}
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
					if uniqueOrder {
						//fmt.Println("unique order, elev IP", IP)
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
				if !master {
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

				elev = connectedElev[msg.FromIP]
				elev.direction = msg.ElevatorStateUpdate[0]
				elev.currentFloor = msg.ElevatorStateUpdate[1]

				for i := 0; i < N_FLOORS; i++ {
					elev.queue.Internal[i] = msg.OrderQueue[i]
					elev.queue.Down[i] = msg.OrderQueue[i+4]
					elev.queue.Up[i] = msg.OrderQueue[i+8]
				}

				if elev.direction == 1 {
					lightCommandChan <- elevManager.LightCommand{0, elev.currentFloor, 0}
				}
				if elev.direction == -1 {
					lightCommandChan <- elevManager.LightCommand{1, elev.currentFloor, 0}
				}

				/*if elev.queue.Up[elev.currentFloor] == 1 {
					light = [3]int{0, elev.currentFloor, 0}
					fmt.Println("turning up light of in floor", elev.currentFloor)
					lightCommandChan <- light
				}
				if elev.queue.Down[elev.currentFloor] == 1 {
					light = [3]int{1, elev.currentFloor, 0}
					fmt.Println("turning down light of in floor", elev.currentFloor)
					lightCommandChan <- light
				}*/

				//elev.queue.Up[elev.currentFloor] = 0
				//elev.queue.Down[elev.currentFloor] = 0
				//elev.queue.Internal[elev.currentFloor] = 0
				connectedElev[msg.FromIP] = elev
				break
			}
		}
	}

}
