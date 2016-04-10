package elevManager

import (
	. "../driver"
	. "../internalOrders"
	"../message"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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
				//fault tolerance?
			}
		case <-doorTimeoutChan:
			DoorOpen(0)
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
	<-positionChan
	ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan+1)
	requestStateUpdateChan <- true

}

func ElevManager(offline *bool,requestStateUpdateChan chan bool, lightCommandChan chan LightCommand, NewNetworkOrderFromSM chan OrderQueue, NewNetworkOrderToSM chan OrderQueue, stateUpdateFromSM chan message.UDPMessage) {

	var queue OrderQueue

	//channels
	positionChan := make(chan int)
	commandChan := make(chan Command, 100)
	orderButtonChan := make(chan OrderQueue)
	watchDogChan := make(chan bool)

	//goroutines
	go elevatorController(commandChan)
	go ElevPosition(positionChan)
	go CheckOrderButton(orderButtonChan)

	initializeElevator(positionChan, requestStateUpdateChan)
	queue.Internal = ReadInternals()
	elevWatchDog := time.AfterFunc(time.Second*10, func() { watchDogChan <- true })
	elevWatchDog.Stop()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	elevDir := up_dir
	//previousFloor := <-positionChan
	defer ElevStart(0)
	var stateUpdate [2]int
	for {
		var order OrderQueue
		select {
		case orderButtonPushed := <-orderButtonChan:
			for i := 0; i < N_FLOORS; i++ {
				if (orderButtonPushed.Internal[i] != queue.Internal[i]) && (orderButtonPushed.Internal[i] == 1) {
					queue.Internal[i] = 1
					order.Internal[i] = 1
					ButtonLamp(2, i, 1)
					NewNetworkOrderFromSM <- order
					elevWatchDog.Reset(time.Second * 30)
					break
				}
				if (orderButtonPushed.Up[i] != queue.Up[i]) && (orderButtonPushed.Up[i] == 1) {
					order.Up[i] = 1
					//fmt.Println("sending order")
					NewNetworkOrderFromSM <- order
					break
				}
				if (orderButtonPushed.Down[i] != queue.Down[i]) && (orderButtonPushed.Down[i] == 1) {
					order.Down[i] = 1
					//fmt.Println("sending order")
					NewNetworkOrderFromSM <- order
					break
				}

			}
			WriteInternals(queue.Internal)

		case neworder := <-NewNetworkOrderToSM:
			for i := 0; i < N_FLOORS; i++ {
				if neworder.Up[i] == 1 {
					queue.Up[i] = 1
					ButtonLamp(0, i, 1)
				}
				if neworder.Down[i] == 1 {
					queue.Down[i] = 1
					ButtonLamp(1, i, 1)
				}
			}
			elevWatchDog.Reset(time.Second * 30)
			break

		case currentFloor := <-positionChan:
			stateUpdate[0], stateUpdate[1] = int(elevDir), currentFloor

			/*if currentFloor != previousFloor {
				elevWatchDog.Reset(time.Second * 10)
				previousFloor = currentFloor
			}*/
			//elevWatchDog.Reset(time.Second * 10)

			if stopOnFloor(offline, elevWatchDog,elevDir, currentFloor, &queue) == true {
				commandChan <- stop
				commandChan <- openDoor
				var localqueue [12]int
				for i := 0; i < N_FLOORS; i++ {
					localqueue[i] = queue.Internal[i]
					localqueue[i+4] = queue.Down[i]
					localqueue[i+8] = queue.Up[i]
				}
				stateUpdateFromSM <- message.UDPMessage{OrderQueue: localqueue, ElevatorStateUpdate: stateUpdate}

			}
			nextDir := nextDirection(&elevDir, &queue, currentFloor)

			if nextDir != stop {
				commandChan <- nextDir
			}

		case light := <-lightCommandChan:
			//fmt.Println("new light command", light)
			//fmt.Println("turning light off")
			ButtonLamp(Button_type(light[0]), light[1], light[2])
			//time.Sleep(time.Microsecond * 10)

		case <-requestStateUpdateChan:
			var localqueue [12]int
			for i := 0; i < N_FLOORS; i++ {
				localqueue[i] = queue.Internal[i]
				localqueue[i+4] = queue.Down[i]
				localqueue[i+8] = queue.Up[i]
			}
			stateUpdateFromSM <- message.UDPMessage{OrderQueue: localqueue, ElevatorStateUpdate: stateUpdate}

		case <-watchDogChan:
			Backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run eventhandler.go")
			Backup.Run()
			fmt.Println("Hardware timeout")
			os.Exit(0)

		// flytt til eventhandler
		case <-signalChan:
			ElevStart(0)
			fmt.Println("Software killed")
			os.Exit(0)

		}

	}

}
func removeFloorFromQueue(offline *bool,elevWatchDog *time.Timer, currentFloor int, queue *OrderQueue) {
	queue.Internal[currentFloor] = 0
	queue.Up[currentFloor] = 0
	queue.Down[currentFloor] = 0
	ButtonLamp(2, currentFloor, 0)
	WriteInternals(queue.Internal)
	i:=0
	for floor := 0; floor < N_FLOORS; floor++ {
		if queue.Up[floor] == 0 && floor != BOTTOM_FLOOR  && *offline{
				ButtonLamp(1,currentFloor,0)
			}else if queue.Up[floor] ==1{
				i++
			}
		if queue.Down[floor] == 0 && floor != TOP_FLOOR && *offline{
			ButtonLamp(0,floor,0)
			}else if queue.Up[floor] ==1{
				i++
			}
	
		if queue.Internal[floor] ==1{
				i++
			}
		}
	if i==0{
			fmt.Println("stopping watchdog")
			elevWatchDog.Stop()
		}

}

func stopOnFloor(offline * bool, elevWatchDog *time.Timer, elevDir Direction, currentFloor int, queue *OrderQueue) bool {
	if currentFloor == TOP_FLOOR && queue.Down[currentFloor] == 1 || currentFloor == BOTTOM_FLOOR && queue.Up[currentFloor] == 1 {
		removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
		return true
	}
	if queue.Internal[currentFloor] == 1 {
			removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
		return true
	}
	if elevDir == up_dir {
		if queue.Up[currentFloor] == 1 {
			removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
			return true
		}

	} else {
		if queue.Down[currentFloor] == 1 {
			removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
			return true
		}
	}

	if elevDir == up_dir {
		for i := currentFloor + 1; i < N_FLOORS; i++ {
			if queue.Up[i] == 1 || queue.Internal[i] == 1 || queue.Down[i] == 1 {
				return false
			} else if queue.Down[currentFloor] == 1 {
			removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
				return true
			}
		}
	} else {
		for i := currentFloor - 1; i > BOTTOM_FLOOR-1; i-- {
			if queue.Up[i] == 1 || queue.Internal[i] == 1 || queue.Down[i] == 1 {
				return false
			} else if queue.Up[currentFloor] == 1 {
			removeFloorFromQueue(offline,elevWatchDog, currentFloor, queue)
				return true
			}
		}
	}
	return false

}

func CheckOrderButton(orderButtonChan chan OrderQueue) {

	var prevbuttonsPressed OrderQueue

	for {
		var buttonsPressed OrderQueue
		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < 3; button++ {
				if (floor == 0 && button == 1) || (floor == 3 && button == 0) {

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

func ElevPosition(positionChan chan int) {
	for {
		floor := FloorSensor()
		if floor != -1 {
			positionChan <- floor
			FloorIndicator(floor)
		}
		time.Sleep(time.Millisecond * 40)
	}

}
