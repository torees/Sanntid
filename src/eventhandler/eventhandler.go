package main

import (
	. "../driver"
	"../elevManager"
	"../message"
	. "../network"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
	"os/signal"
)

const MAX_ORDER_COST = 100

type elevator struct {
	queue        elevManager.OrderQueue
	direction    int
	currentFloor int
	IP           string
}

func main() { 
	var myIP string
	for {
		myIP = GetNetworkIP()
		if !(myIP == "::1") {
			break
		}
		fmt.Println("No connection")
		time.Sleep(time.Second * 1)
	}
	NetworkConnected(0)
	offline := false

	UDPSendMsgChan := make(chan message.UDPMessage, 100)
	UDPPingReceivedChan := make(chan message.UDPMessage, 100)
	UDPMsgReceivedChan := make(chan message.UDPMessage, 100)
	restartUDPSendChan := make(chan bool)
	checkNetworkConChan := make(chan bool)
	

	newNetworkOrderToElevManagerChan := make(chan elevManager.OrderQueue, 100)
	newNetworkOrderFromElevManagerChan := make(chan elevManager.OrderQueue, 100)
	stateUpdateChan := make(chan message.UDPMessage, 100)
	requestStateUpdateChan := make(chan bool, 100)

	newMsgToMasterChan := make(chan message.UDPMessage, 100)
	newOrderFromMasterChan := make(chan message.UDPMessage, 100)
	setLightChan := make(chan elevManager.LightCommand, 100)
	elevatorAddedChan := make(chan string, 100)
	elevatorRemovedChan := make(chan string, 100)
	networkStatus := make(chan bool, 100)
	keyboardInterruptChan := make(chan os.Signal)


	StartUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)

	go UDPlisten(UDPPingReceivedChan, UDPMsgReceivedChan)
	go CheckNetworkConnection(checkNetworkConChan)
	go master(&offline,newNetworkOrderToElevManagerChan,networkStatus, setLightChan, elevatorAddedChan, elevatorRemovedChan, newMsgToMasterChan, newOrderFromMasterChan, myIP)
	go elevManager.ElevManager(&offline,stateUpdateChan,requestStateUpdateChan, setLightChan, newNetworkOrderFromElevManagerChan, newNetworkOrderToElevManagerChan )

	connectedElevTimers := make(map[string]*time.Timer)
	signal.Notify(keyboardInterruptChan, os.Interrupt) //Catch keyboard interrupts (Ctrl+C)	
	
	for {
		select {
		case msg := <-UDPPingReceivedChan:
			_, exists := connectedElevTimers[msg.FromIP]
			if exists {
				connectedElevTimers[msg.FromIP].Reset(time.Second)
			} else {
				elevatorAddedChan <- msg.FromIP
				connectedElevTimers[msg.FromIP] = time.AfterFunc(time.Millisecond*600, func() { deleteElevator(&connectedElevTimers, msg, elevatorRemovedChan) })
				requestStateUpdateChan <- true
			}

		case msg := <-newOrderFromMasterChan:
			UDPSendMsgChan <- msg

		case msg := <-UDPMsgReceivedChan:

			switch msg.MessageId {
			case message.ElevatorStateUpdate, message.NewOrder:
				newMsgToMasterChan <- msg

			case message.NewOrderFromMaster:
				//Light external orderbuttons for all connected elevators
				var light elevManager.LightCommand
				for i := 0; i < N_FLOORS; i++ {
					if msg.OrderQueue[i+4] == 1 {
						light = [3]int{1, i, 1}
						setLightChan <- light
						break

					}
					if msg.OrderQueue[i+8] == 1 {
						light = [3]int{0, i, 1}
						setLightChan <- light
						break
					}
				}

				// Make all connected elevators update copies of order queues
				newMsgToMasterChan <- msg

				if msg.ToIP == myIP {
					var order elevManager.OrderQueue
					for i := 0; i < 4; i++ {
						order.Internal[i] = msg.OrderQueue[i]
						order.Up[i] = msg.OrderQueue[(i + 8)]
						order.Down[i] = msg.OrderQueue[(i + 4)]
					}
					newNetworkOrderToElevManagerChan <- order
				}
			}

		case order := <-newNetworkOrderFromElevManagerChan:
			if !offline {
				var msg message.UDPMessage
				msg.MessageId = message.NewOrder
				msg.FromIP = myIP
				msg.OrderQueue = elevManager.AssembleQueueUpdate(order)			
				UDPSendMsgChan <- msg
			}


		case msg := <-stateUpdateChan:
			if !offline {
				msg.MessageId = message.ElevatorStateUpdate
				msg.FromIP = myIP			
				UDPSendMsgChan <- msg
			}

		case haveNetwork := <-checkNetworkConChan:
			if haveNetwork {
				offline = false
				StartUDPSend(UDPSendMsgChan, restartUDPSendChan, myIP)
				networkStatus <- true
			} else {
				offline = true
				restartUDPSendChan <- true
				networkStatus <-false
			}

		case <-keyboardInterruptChan:
			ElevStart(0)
			fmt.Println("Software killed by user")
			os.Exit(0)

		}

	}

}

func deleteElevator(connectedElevTimers *map[string]*time.Timer, msg message.UDPMessage, elevatorRemovedChan chan string) {
	elevatorRemovedChan <- msg.FromIP
	delete(*connectedElevTimers, msg.FromIP)
}

func master(offline *bool, newNetworkOrderToElevManagerChan chan elevManager.OrderQueue, networkStatus chan bool, setLightChan chan elevManager.LightCommand, elevatorAddedChan chan string, elevatorRemovedChan chan string,newMsgToMasterChan chan message.UDPMessage, newOrderFromMasterChan chan message.UDPMessage, myIP string) {
	numberOfElev := 0
	connectedElevMap := make(map[string]elevator)
	isMaster := true
	var IPlist []string
	var elev elevator

	for {
		select {
		
		case elevatorIP := <-elevatorRemovedChan:
			IPlist = IPlist[:0]
			numberOfElev -= 1
			elev = connectedElevMap[elevatorIP]
			tempqueue := elev.queue
			delete(connectedElevMap, elevatorIP)
			if numberOfElev != 0 {
				for key, _ := range connectedElevMap {
					IPlist = append(IPlist, key)
				}
				sort.Strings(IPlist)
				if IPlist[0] == myIP {
					isMaster = true
				
				} else {
					isMaster = false
				}
			}
			fmt.Println("Currently connected Elevators: ", IPlist)
			

			/*if numberOfElev == 0 {
				isMaster = true
			}*/
			var msg message.UDPMessage
			msg.MessageId = message.NewOrder
			msg.FromIP = myIP
			for floor := 0; floor < N_FLOORS; floor++ {
				if tempqueue.Up[floor] == 1 {
					msg.OrderQueue[floor+8] = 1
					newMsgToMasterChan <- msg
				}
				if tempqueue.Down[floor] == 1 {
					msg.OrderQueue[floor+4] = 1
					newMsgToMasterChan <- msg
				}
			}

		case id := <-elevatorAddedChan:
			numberOfElev += 1
			*offline = false
			if numberOfElev > N_ELEVATORS {
				//fault tolerance
				fmt.Println("To many elevators")
				os.Exit(0)
			}

			elev.IP = id
			connectedElevMap[id] = elev
			IPlist = IPlist[:0]
			for key, _ := range connectedElevMap{
				IPlist = append(IPlist, key)
			}
			sort.Strings(IPlist)
			if IPlist[0] == myIP {
				isMaster = true
			} else {
				isMaster = false
			}
			fmt.Println("Currently connected elevators: ", IPlist)

		case msg := <-newMsgToMasterChan:
			var IP string
			var newOrder elevManager.OrderQueue
			Msg:
			switch msg.MessageId {

			case message.NewOrder:
				for i := 0; i < N_FLOORS; i++ {
					newOrder.Internal[i] = msg.OrderQueue[i]
					newOrder.Down[i] = msg.OrderQueue[(i + 4)]
					newOrder.Up[i] = msg.OrderQueue[(i + 8)]

				}
				//uniqueOrder := true
				orderCost := MAX_ORDER_COST
				if isMaster && !*offline {
					for _, elev := range connectedElevMap{
						//fmt.Println(elev)
						if !elev.newOrder(newOrder) {

							//uniqueOrder = false
							fmt.Println("not unique")
							break Msg
						}
						tempOrderCost, tempIP := elev.cost(newOrder)
						if tempOrderCost < orderCost {
							orderCost = tempOrderCost
							//fmt.Println("temp ip", tempIP)
							IP = tempIP
						}
					}
					msg.ToIP = IP
					if IP == "" {
						fmt.Println("error in assigning elevator, order deleted")
						break
					}

					//update masters copy of the queue
						//fmt.Println("unique order, elev IP", IP)
					for i := 0; i < N_FLOORS; i++ {
						if newOrder.Up[i] == 1 {
							elev = connectedElevMap[IP]
							elev.queue.Up[i] = 1
							connectedElevMap[IP] = elev
						}
						if newOrder.Down[i] == 1 {
							elev = connectedElevMap[IP]
							elev.queue.Down[i] = 1
							connectedElevMap[IP] = elev
						}
					}
					msg.MessageId = message.NewOrderFromMaster
					newOrderFromMasterChan <- msg // send ON UDP

				}
				for i := 0; i < N_FLOORS; i++ {
					if newOrder.Internal[i] == 1 {
						elev = connectedElevMap[msg.FromIP]
						elev.queue.Internal[i] = 1
						connectedElevMap[msg.FromIP] = elev
					}
				}
				break

			case message.NewOrderFromMaster:
				for i := 0; i < N_FLOORS; i++ {
					newOrder.Internal[i] = msg.OrderQueue[i]
					newOrder.Down[i] = msg.OrderQueue[(i + 4)]
					newOrder.Up[i] = msg.OrderQueue[(i + 8)]

				}
				if !isMaster {
					IP := msg.ToIP
					elev = connectedElevMap[IP]
					for i := 0; i < N_FLOORS; i++ {
						if newOrder.Up[i] == 1 {
							elev.queue.Up[i] = 1
						}
						if newOrder.Down[i] == 1 {
							elev.queue.Down[i] = 1
						}
					}
					connectedElevMap[IP] = elev
				}

			case message.ElevatorStateUpdate:
				elev = connectedElevMap[msg.FromIP]
				elev.direction = msg.ElevatorStateUpdate[0]
				elev.currentFloor = msg.ElevatorStateUpdate[1]
				//fmt.Println("last queue : ", elev.queue)
				for i := 0; i < N_FLOORS; i++ {
					elev.queue.Internal[i] = msg.OrderQueue[i]
					elev.queue.Down[i] = msg.OrderQueue[i+4]
					elev.queue.Up[i] = msg.OrderQueue[i+8]
				}
				connectedElevMap[msg.FromIP] = elev
				//fmt.Println("newQueue : ", elev.queue)

				var lights elevManager.OrderQueue
				for _, elevator := range connectedElevMap {
					//fmt.Println("heis :",elevator.queue
					for floor := 0; floor < N_FLOORS; floor++ {
						if elevator.queue.Up[floor] == 1 {
							lights.Up[floor] = 1
						}
						if elevator.queue.Down[floor] == 1 {
							lights.Down[floor] = 1
						}
					}

				}
				//fmt.Println("lights UP: ", lights.Up, "lights down :", lights.Down)
				for floor := 0; floor < N_FLOORS; floor++ {
					if lights.Down[floor] == 0 && floor != BOTTOM_FLOOR {
						setLightChan <- elevManager.LightCommand{1, floor, 0}
					}
					if lights.Up[floor] == 0 && floor != TOP_FLOOR {
						setLightChan <- elevManager.LightCommand{0, floor, 0}
					}
				}
				break
			}
		case noNetwork := <- networkStatus:
			if(noNetwork){
				var order elevManager.OrderQueue
				for _, elevator := range connectedElevMap{
					for floor := 0; floor < N_FLOORS; floor++ {
						if elevator.queue.Up[floor] == 1 {
							order.Up[floor] = 1
						}
						if elevator.queue.Down[floor] == 1 {
							order.Down[floor] = 1
						}
					}
				}
				newNetworkOrderToElevManagerChan <- order	
			}



		}
	}
}



func (elev elevator) cost(order elevManager.OrderQueue) (int, string) {
	const dirCost = 2
	const distCost = 4
	const numOrderCost = 6
	cost := 5

	distanceCost := (elev.currentFloor - elev.findOrderFloor(order)) * distCost
	directionCost := 0

	if distanceCost < 0 {
		directionCost = dirCost
		distanceCost = int(math.Abs(float64(distanceCost)))
	}

	cost = elev.numOrdersInQueue()*numOrderCost + distanceCost + directionCost
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
