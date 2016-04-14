package elevManager

import (
	. "../driver"
	. "../internalOrders"
	. "../message"
	"fmt"
	"os"
	"time"
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
	Internal [N_FLOORS]int
	Down     [N_FLOORS]int
	Up       [N_FLOORS]int
}

type LightCommand [3]int

func ElevManager(offline *bool, stateUpdateChan chan UDPMessage, requestStateUpdateChan chan bool, lightCommandChan chan LightCommand, NewNetworkOrderFromElevManagerChan chan OrderQueue, NewNetworkOrderToElevManagerChan chan OrderQueue) {

	var queue OrderQueue
	elevDir := up_dir
	var stateUpdate [2]int

	positionChan := make(chan int)
	commandChan := make(chan Command, 100)
	orderButtonChan := make(chan OrderQueue)
	hardwareErrorChan := make(chan bool)

	go elevatorController(commandChan)
	go elevPosition(positionChan)
	go checkOrderButton(orderButtonChan)

	initializeElevator(positionChan, requestStateUpdateChan)
	queue.Internal = ReadInternals()
	hardwareErrorTimer := time.AfterFunc(time.Second*10, func() { hardwareErrorChan <- true })
	hardwareErrorTimer.Stop()

	for {

		select {

		case orderButtonPushed := <-orderButtonChan:
			var order OrderQueue
			for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
				if (orderButtonPushed.Internal[floor] != queue.Internal[floor]) && (orderButtonPushed.Internal[floor] == 1) {
					queue.Internal[floor] = 1
					order.Internal[floor] = 1
					ButtonLamp(2, floor, 1)
					WriteInternalToFile(queue.Internal)
					NewNetworkOrderFromElevManagerChan <- order
					hardwareErrorTimer.Reset(time.Second * 30)
					break
				}
				if (orderButtonPushed.Up[floor] != queue.Up[floor]) && (orderButtonPushed.Up[floor] == 1) {
					order.Up[floor] = 1
					NewNetworkOrderFromElevManagerChan <- order
					break
				}
				if (orderButtonPushed.Down[floor] != queue.Down[floor]) && (orderButtonPushed.Down[floor] == 1) {
					order.Down[floor] = 1
					NewNetworkOrderFromElevManagerChan <- order
					break
				}

			}

		case neworder := <-NewNetworkOrderToElevManagerChan:
			for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
				if neworder.Up[floor] == 1 {
					queue.Up[floor] = 1
					ButtonLamp(0, floor, 1)
				}
				if neworder.Down[floor] == 1 {
					queue.Down[floor] = 1
					ButtonLamp(1, floor, 1)
				}
			}
			hardwareErrorTimer.Reset(time.Second * 30)

		case currentFloor := <-positionChan:
			stateUpdate[0], stateUpdate[1] = int(elevDir), currentFloor
			if currentFloor == 3 {
				elevDir = down_dir
			} else if currentFloor == 0 {
				elevDir = up_dir
			}

			if stopOnFloor(offline, hardwareErrorTimer, elevDir, currentFloor, &queue) {
				commandChan <- stop
				commandChan <- openDoor
				queueUpdate := AssembleMessageQueue(queue)
				stateUpdateChan <- UDPMessage{OrderQueue: queueUpdate, ElevatorStateUpdate: stateUpdate}
			}

			nextCmd := nextCommand(&elevDir, &queue, currentFloor)
			if nextCmd != stop {
				commandChan <- nextCmd
			}

		case light := <-lightCommandChan:
			ButtonLamp(Button_type(light[0]), light[1], light[2])

		case <-requestStateUpdateChan:
			queueUpdate := AssembleMessageQueue(queue)
			stateUpdateChan <- UDPMessage{OrderQueue: queueUpdate, ElevatorStateUpdate: stateUpdate}

		case <-hardwareErrorChan:
			ElevStart(0)
			fmt.Println("Hardware timeout")
			os.Exit(0)

		}
	}
}

func elevatorController(commandChan chan Command) {
	doorOpen := false
	doorTimeoutChan := make(chan bool)
	doorTimer := time.AfterFunc(time.Second*3, func() { doorTimeoutChan <- true })

	for {
		select {
		case command := <-commandChan:
			switch command {
			case stop:
				ElevStart(0)
				break
			case openDoor:
				doorOpen = true
				doorTimer.Reset(time.Second * 3)
				DoorOpen(1)
				break

			case goUp:
				if !doorOpen {
					ElevStart(1)
				}
				break
			case goDown:
				if !doorOpen {
					ElevStart(-1)
				}
				break
			default:

			}
		case <-doorTimeoutChan:
			DoorOpen(0)
			doorOpen = false
		}
	}
}

func nextCommand(elevDir *Direction, queue *OrderQueue, currentFloor int) Command {

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

func initializeElevator(positionChan chan int, requestStateUpdateChan chan bool) {
	HardwareInit()
	fmt.Println("Starting Elevator 3000...")
	ElevStart(1)
	fmt.Println("Initialized at floor", <-positionChan)
	ElevStart(0)
	requestStateUpdateChan <- true
}

func AssembleMessageQueue(queue OrderQueue) [12]int {
	var queueUpdate [12]int
	for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
		queueUpdate[floor] = queue.Internal[floor]
		queueUpdate[floor+DOWN_ROOT_POSITION] = queue.Down[floor]
		queueUpdate[floor+UP_ROOT_POSITION] = queue.Up[floor]
	}
	return queueUpdate
}

func DisassembleMessageQueue(msgQueue [12]int) OrderQueue {
	var queue OrderQueue
	for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
		queue.Internal[floor] = msgQueue[floor]
		queue.Down[floor] = msgQueue[(floor + DOWN_ROOT_POSITION)]
		queue.Up[floor] = msgQueue[(floor + UP_ROOT_POSITION)]
	}
	return queue
}

func removeFloorFromQueue(offline *bool, hardwareErrorTimer *time.Timer, currentFloor int, queue *OrderQueue) {
	queue.Internal[currentFloor] = 0
	queue.Up[currentFloor] = 0
	queue.Down[currentFloor] = 0
	ButtonLamp(2, currentFloor, 0)
	WriteInternalToFile(queue.Internal)
	emptyQueue := true
	for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
		if queue.Up[floor] == 0 && floor != BOTTOM_FLOOR && *offline {
			ButtonLamp(1, currentFloor, 0)
		}
		if queue.Down[floor] == 0 && floor != TOP_FLOOR && *offline {
			ButtonLamp(0, currentFloor, 0)
		}

		if queue.Internal[floor] == 1 || queue.Up[floor] == 1 || queue.Down[floor] == 1 {
			emptyQueue = false
		}
	}
	if emptyQueue {
		hardwareErrorTimer.Stop()
	}
}

func stopOnFloor(offline *bool, hardwareErrorTimer *time.Timer, elevDir Direction, currentFloor int, queue *OrderQueue) bool {
	if currentFloor == TOP_FLOOR && queue.Down[currentFloor] == 1 || currentFloor == BOTTOM_FLOOR && queue.Up[currentFloor] == 1 {
		removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
		return true
	}
	if queue.Internal[currentFloor] == 1 {
		removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
		return true
	}
	if elevDir == up_dir {
		if queue.Up[currentFloor] == 1 {
			removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
			return true
		}

	} else {
		if queue.Down[currentFloor] == 1 {
			removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
			return true
		}
	}

	if elevDir == up_dir {
		for floor := currentFloor + 1; floor < N_FLOORS; floor++ {
			if queue.Up[floor] == 1 || queue.Internal[floor] == 1 || queue.Down[floor] == 1 {
				return false
			} else if queue.Down[currentFloor] == 1 {
				removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
				return true
			}
		}
	} else {
		for floor := currentFloor - 1; floor > BOTTOM_FLOOR-1; floor-- {
			if queue.Up[floor] == 1 || queue.Internal[floor] == 1 || queue.Down[floor] == 1 {
				return false
			} else if queue.Up[currentFloor] == 1 {
				removeFloorFromQueue(offline, hardwareErrorTimer, currentFloor, queue)
				return true
			}
		}
	}
	return false
}

func checkOrderButton(orderButtonChan chan OrderQueue) {

	var prevbuttonsPressed OrderQueue

	for {
		var buttonsPressed OrderQueue
		for floor := BOTTOM_FLOOR; floor < N_FLOORS; floor++ {
			for button := 0; button < 3; button++ {
				if (floor == BOTTOM_FLOOR && button == 1) || (floor == TOP_FLOOR && button == 0) {

				} else {
					switch button {
					case 0:
						buttonVal := ButtonPushed(Button_type(button), floor)
						if buttonVal == 1 && prevbuttonsPressed.Up[floor] == 0 {
							buttonsPressed.Up[floor] = 1
							orderButtonChan <- buttonsPressed
						}
						prevbuttonsPressed.Up[floor] = buttonVal
						break
					case 1:
						buttonVal := ButtonPushed(Button_type(button), floor)
						if buttonVal == 1 && prevbuttonsPressed.Down[floor] == 0 {
							buttonsPressed.Down[floor] = 1
							orderButtonChan <- buttonsPressed
						}
						prevbuttonsPressed.Down[floor] = buttonVal
						break
					case 2:
						buttonVal := ButtonPushed(Button_type(button), floor)
						if buttonVal == 1 && prevbuttonsPressed.Internal[floor] == 0 {
							buttonsPressed.Internal[floor] = 1
							orderButtonChan <- buttonsPressed
						}
						prevbuttonsPressed.Internal[floor] = buttonVal
						break
					default:

					}
				}
			}
		}
	}
}

func elevPosition(positionChan chan int) {
	for {
		floor := FloorSensor()
		if floor != -1 {
			positionChan <- floor
			FloorIndicator(floor)
		}
		time.Sleep(time.Millisecond * 40)
	}
}
