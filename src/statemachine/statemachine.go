package statemachine

import (
	"../driver"
	"fmt"
	"time"
)

type Direction int
type Command int

const N_FLOORS = 4
const BOTTOM_FLOOR = 0
const TOP_FLOOR = 3

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
	Internal [N_FLOORS]int
	Down     [N_FLOORS]int
	Up       [N_FLOORS]int
}

func StateMachine(NewNetworkOrderFromSM chan OrderQueue, NewNetworkOrderToSM chan OrderQueue, stateUpdateFromSM chan [2]int) {
	//variables

	fmt.Println("Starting Elevator 3000...")
	driver.ElevInit()

	//channels
	positionChan := make(chan int)

	queueChan := make(chan OrderQueue)
	orderButtonChan := make(chan OrderQueue)

	//go-routines
	go ElevPosition(positionChan)
	go CheckOrderButton(orderButtonChan)
	//LocalChan := make(chan OrderQueue, 10)

	driver.ElevStart(1)
	<-positionChan
	driver.ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan)
	//Correct so far

	//fmt.Println("fstate", fstate)

	ElevManager(orderButtonChan, queueChan, positionChan, NewNetworkOrderFromSM, NewNetworkOrderToSM, stateUpdateFromSM)

}
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
	if *elevDir == up_dir {
		for i := currentFloor + 1; i < N_FLOORS; i++ {
			if (queue.Up[i] != 0) || (queue.Internal[i] != 0) || (queue.Down[i] != 0) {
				return goUp
			}
		}
		for i := currentFloor - 1; i >= BOTTOM_FLOOR; i-- {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				*elevDir = down_dir
				return goDown
			}

		}

	}
	if *elevDir == down_dir {
		for i := currentFloor - 1; i >= BOTTOM_FLOOR; i-- {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				return goDown
			}

		}
		for i := currentFloor + 1; i < N_FLOORS; i++ {
			if queue.Up[i] != 0 || queue.Internal[i] != 0 || queue.Down[i] != 0 {
				*elevDir = up_dir
				return goUp
			}
		}
	}
	return stop

}

func ElevManager(orderButtonChan chan OrderQueue, queueChan chan OrderQueue, positionChan chan int, NewNetworkOrderFromSM chan OrderQueue, NewNetworkOrderToSM chan OrderQueue, stateUpdateFromSM chan [2]int) {

	var queue OrderQueue

	commandChan := make(chan Command, 100)
	go elevatorController(commandChan)
	elevDir := up_dir
	defer driver.ElevStart(0)
	for {
		var order OrderQueue
		select {
		case orderButtonPushed := <-orderButtonChan:
			//if internal order, set light and update elevqueue(internal)
			//if(orderButtonPushed.internal){
			for i := 0; i < N_FLOORS; i++ {
				if (orderButtonPushed.Internal[i] != queue.Internal[i]) && (orderButtonPushed.Internal[i] == 1) {
					queue.Internal[i] = 1
					driver.ButtonLamp(2, i, 1)
				}
				//dette erstattes senere av nettverkskommandoer:
				if (orderButtonPushed.Up[i] != queue.Up[i]) && (orderButtonPushed.Up[i] == 1) {
					order.Up[i] = 1
				}
				if (orderButtonPushed.Down[i] != queue.Down[i]) && (orderButtonPushed.Down[i] == 1) {
					order.Down[i] = 1
				}

			}

			NewNetworkOrderFromSM <- order

		case neworder := <-NewNetworkOrderToSM:
			//update elevqueue with the new order
			for i := 0; i < N_FLOORS; i++ {
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
			//new floor reached.
			//if new floor in queue
			var stateUpdate [2]int

			if stopOnFloor(elevDir, currentFloor, &queue) == true {
				commandChan <- stop
				commandChan <- openDoor
				if currentFloor == 3 {
					elevDir = down_dir
				} else if currentFloor == 0 {
					elevDir = up_dir
				}
				stateUpdate[0], stateUpdate[1] = int(elevDir), currentFloor
				stateUpdateFromSM <- stateUpdate

			}
			nextDir := nextDirection(&elevDir, &queue, currentFloor)

			if nextDir != stop {
				commandChan <- nextDir
			}

		}
	}

}
func removeFloorFromQueue(currentFloor int, queue *OrderQueue) {
	queue.Internal[currentFloor] = 0
	queue.Up[currentFloor] = 0
	queue.Down[currentFloor] = 0
	driver.ButtonLamp(0, currentFloor, 0)
	driver.ButtonLamp(1, currentFloor, 0)
	driver.ButtonLamp(2, currentFloor, 0)
}

func stopOnFloor(elevDir Direction, currentFloor int, queue *OrderQueue) bool {
	//catch conercases in upper and lower floor
	if currentFloor == TOP_FLOOR && queue.Down[currentFloor] == 1 || currentFloor == BOTTOM_FLOOR && queue.Up[currentFloor] == 1 {
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
		for i := currentFloor + 1; i < N_FLOORS; i++ {
			if queue.Up[i] == 1 || queue.Internal[i] == 1 || queue.Down[i] == 1 {
				return false
			} else if queue.Down[currentFloor] == 1 {
				removeFloorFromQueue(currentFloor, queue)
				return true
			}
		}
	} else {
		for i := currentFloor - 1; i == BOTTOM_FLOOR; i-- {
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

		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < 3; button++ {
				//sjekker for manglende knapper i endeetasjer
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
		// Only send new if new button is pushed
		if prevbuttonsPressed != buttonsPressed {
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
