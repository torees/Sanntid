package elevManager

import (
	"../driver"
	"fmt"
	"time"
	. "../internalOrders"
)

type Direction int
type Command int



const (
	stop Command = iota
	openDoor
	goUp
	goDown
)

const (
	up_dir   Direction = 1
	down_dir           = -1
)

type OrderQueue struct {
	Internal [driver.N_FLOORS]int
	Down     [driver.N_FLOORS]int
	Up       [driver.N_FLOORS]int
}

type LightCommand [3]int //index: [0]: button [1]: floor [2] lightvalue

func elevatorController(commandChan chan Command) {
	doorOpen := false
	doorTimeoutChan := make(chan bool)
	doorTimer := time.AfterFunc(time.Second*3, func() { doorTimeoutChan <- true })

	//legge inn fault tolerance ved manuell flytting av heis? Timer på ny command, restart ved timeout

	for {
		select {
		case command := <-commandChan:
			switch command {
			case stop:
				driver.ElevStart(0)
				break
			case openDoor:
				doorOpen = true
				doorTimer.Reset(time.Second * 3)
				driver.DoorOpen(1)
				break

			case goUp:
				if !doorOpen {
					driver.ElevStart(1)
				}
				break
			case goDown:
				if !doorOpen {
					driver.ElevStart(-1)
				}
				break
			default:
				//fault tolerance?
			}
		case <-doorTimeoutChan:
			driver.DoorOpen(0)
			doorOpen = false
		}
	}
}

func nextDirection(elevDir *Direction, queue *OrderQueue, currentFloor int) Command {
	if currentFloor == 3 {
		*elevDir = down_dir
	} else if currentFloor == 0 {
		*elevDir = up_dir
	}

	if *elevDir == up_dir {
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			if (queue.Up[i] != 0) || (queue.Internal[i] != 0) || (queue.Down[i] != 0) {
				return goUp
			}
		}
		for i := currentFloor - 1; i >= driver.BOTTOM_FLOOR; i-- {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				*elevDir = down_dir
				return goDown
			}

		}

	}
	if *elevDir == down_dir {
		for i := currentFloor - 1; i >= driver.BOTTOM_FLOOR; i-- {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				return goDown
			}

		}
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				*elevDir = up_dir
				return goUp
			}
		}
	}
	return stop

}

func initializeElevator(positionChan chan int, stateUpdateFromSM chan [2]int) {
	driver.HardwareInit()
	var stateUpdate [2]int
	fmt.Println("Starting Elevator 3000...")
	driver.ElevStart(1)
	initialFloor:=<-positionChan
	driver.ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan+1)
	stateUpdate[0], stateUpdate[1] = int(up_dir), initialFloor
	stateUpdateFromSM <- stateUpdate

}

func ElevManager(requestStateUpdateChan chan bool, lightCommandChan chan LightCommand, NewNetworkOrderFromSM chan OrderQueue, NewNetworkOrderToSM chan OrderQueue, stateUpdateFromSM chan [2]int) {

	var queue OrderQueue


	//channels
	positionChan := make(chan int)
	commandChan := make(chan Command, 100)
	orderButtonChan := make(chan OrderQueue)

	//goroutines
	go elevatorController(commandChan)
	go ElevPosition(positionChan)
	go CheckOrderButton(orderButtonChan)

	initializeElevator(positionChan, stateUpdateFromSM)
	var order OrderQueue
	//WriteInternals(order.Internal)
	queue.Internal = ReadInternals()
	elevDir := up_dir
	defer driver.ElevStart(0)
	var stateUpdate [2]int
	for {
		
		select {
		case orderButtonPushed := <-orderButtonChan:
			for i := 0; i < driver.N_FLOORS; i++ {
				if (orderButtonPushed.Internal[i] != queue.Internal[i]) && (orderButtonPushed.Internal[i] == 1) {
					queue.Internal[i] = 1
					order.Internal[i] = 1
					driver.ButtonLamp(2, i, 1)

					NewNetworkOrderFromSM <- order
					break
				}
				if (orderButtonPushed.Up[i] != queue.Up[i]) && (orderButtonPushed.Up[i] == 1) {
					order.Up[i] = 1
					NewNetworkOrderFromSM <- order
					break
				}
				if (orderButtonPushed.Down[i] != queue.Down[i]) && (orderButtonPushed.Down[i] == 1) {
					order.Down[i] = 1
					NewNetworkOrderFromSM <- order
					break
				}

			}
			WriteInternals(queue.Internal)

		case neworder := <-NewNetworkOrderToSM:
			for i := 0; i < driver.N_FLOORS; i++ {
				if neworder.Up[i] == 1 {
					queue.Up[i] = 1
					driver.ButtonLamp(0, i, 1)
				}
				if neworder.Down[i] == 1 {
					queue.Down[i] = 1
					driver.ButtonLamp(1, i, 1)
				}
			}
			break

		case currentFloor := <-positionChan:
			stateUpdate[0], stateUpdate[1] = int(elevDir), currentFloor
			if stopOnFloor(elevDir, currentFloor, &queue) == true {
				commandChan <- stop
				commandChan <- openDoor
				stateUpdateFromSM <- stateUpdate

			}
			nextDir := nextDirection(&elevDir, &queue, currentFloor)

			if nextDir != stop {
				commandChan <- nextDir
			}

		case light := <-lightCommandChan:
			driver.ButtonLamp(driver.Button_type(light[0]), light[1], light[2])
		
		case <- requestStateUpdateChan:
			stateUpdateFromSM <- stateUpdate			

		}

	}

}
func removeFloorFromQueue(currentFloor int, queue *OrderQueue) {
	queue.Internal[currentFloor] = 0
	queue.Up[currentFloor] = 0
	queue.Down[currentFloor] = 0
	driver.ButtonLamp(2, currentFloor, 0)
	//driver.ButtonLamp(1, currentFloor, 0)
	//driver.ButtonLamp(2, currentFloor, 0)
}

func stopOnFloor(elevDir Direction, currentFloor int, queue *OrderQueue) bool {
	if currentFloor == driver.TOP_FLOOR && queue.Down[currentFloor] == 1 || currentFloor == driver.BOTTOM_FLOOR && queue.Up[currentFloor] == 1 {
		removeFloorFromQueue(currentFloor, queue)
		return true
	}
	if queue.Internal[currentFloor] == 1 {
		removeFloorFromQueue(currentFloor, queue)
		return true
	}
	if elevDir == up_dir {
		if queue.Up[currentFloor] == 1 {
			removeFloorFromQueue(currentFloor, queue)
			return true
		}

	} else {
		if queue.Down[currentFloor] == 1 {
			removeFloorFromQueue(currentFloor, queue)
			return true
		}
	}

	if elevDir == up_dir {
		for i := currentFloor + 1; i < driver.N_FLOORS; i++ {
			if queue.Up[i] == 1 || queue.Internal[i] == 1 || queue.Down[i] == 1 {
				return false
			} else if queue.Down[currentFloor] == 1 {
				removeFloorFromQueue(currentFloor, queue)
				return true
			}
		}
	} else {
		for i := currentFloor - 1; i > driver.BOTTOM_FLOOR-1; i-- {
			if queue.Up[i] == 1 || queue.Internal[i] == 1 || queue.Down[i] == 1 {
				return false
			} else if queue.Up[currentFloor] == 1 {
				removeFloorFromQueue(currentFloor, queue)
				return true
			}
		}
	}
	return false

}

func CheckOrderButton(orderButtonChan chan OrderQueue) {

	var prevbuttonsPressed OrderQueue
	var buttonsPressed OrderQueue

	for {

		for floor := 0; floor < driver.N_FLOORS; floor++ {
			for button := 0; button < 3; button++ {
				if (floor == 0 && button == 1) || (floor == 3 && button == 0) {

				} else {
					switch button {
					case 0:
						buttonsPressed.Up[floor] = driver.ButtonPushed(driver.Button_type(button), floor)
						break
					case 1:
						buttonsPressed.Down[floor] = driver.ButtonPushed(driver.Button_type(button), floor)
						break
					case 2:
						buttonsPressed.Internal[floor] = driver.ButtonPushed(driver.Button_type(button), floor)
						break
					default:

					}
				}
			}
		}
		var num int
		for i:= 0; i< driver.N_FLOORS; i++{
			if(buttonsPressed.Up[i] ==1){
				num +=1}
			if (buttonsPressed.Down[i] ==1){
				num +=1
			} 
			if(buttonsPressed.Internal[i] ==1) {
				num += 1
			}
		}
		if ((prevbuttonsPressed != buttonsPressed) && num == 1) {
			orderButtonChan <- buttonsPressed
		}
		prevbuttonsPressed = buttonsPressed
	}
}

func ElevPosition(positionChan chan int) {
	for {
		floor := driver.FloorSensor()
		if floor != -1 {
			positionChan <- floor
			driver.FloorIndicator(floor)
		}
		time.Sleep(time.Millisecond * 40)
	}

}
