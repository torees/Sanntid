package main

import (
	"../driver"
	"fmt"
	"time"
)


type Direction int
type Command int

const N_FLOORS = 4
const BOTTOM_FLOOR = 0


const (
	stop Command =iota
	openDoor
	goUp
	goDown
)

const (
	up_dir Direction = iota
	down_dir
)

type orderQueue struct {
	internal [N_FLOORS]int
	down     [N_FLOORS]int
	up       [N_FLOORS]int
}

func main() {
	//variables
	
	

	fmt.Println("Starting Elevator 3000...")
	driver.ElevInit()

	//channels
	positionChan := make(chan int)
	
	queueChan := make(chan orderQueue)
	orderButtonChan := make(chan orderQueue)


	

	//go-routines
	go ElevPosition(positionChan)
	go CheckOrderButton(orderButtonChan)
	//LocalChan := make(chan orderQueue, 10)


	driver.ElevStart(1)
	<-positionChan
	driver.ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan)
	//Correct so far
	

	
	//fmt.Println("fstate", fstate)
	
	ElevManager(orderButtonChan,queueChan, positionChan)

}
func elevatorController(commandChan chan Command){
	for{
		command:=<-commandChan

		switch(command){
		case stop:
			driver.ElevStart(0)		
			break
		case openDoor:
			driver.DoorOpen(1)
			time.Sleep(time.Second*3)
			driver.DoorOpen(0)
			break
		case goUp:
			driver.ElevStart(1)
			break
		case goDown:
			driver.ElevStart(-1)
			break
		default:
			//fault tolerance?
		}
	}
}



func nextDirection(elevDir *Direction,queue *orderQueue,currentFloor int) Command{
		if *elevDir == up_dir{
			for i:=currentFloor+1 ; i < N_FLOORS ; i++{
				if( (queue.up[i] != 0 )|| (queue.internal[i] != 0 )|| (queue.down[i] != 0)){
					return goUp
				}
			}
			for i:=currentFloor-1 ; i >= BOTTOM_FLOOR ; i--{
				if(queue.up[i] != 0 || queue.internal[i] != 0 || queue.down[i] != 0){
					*elevDir = down_dir
					return goDown
				}

			}

		}
		if *elevDir == down_dir{
			for i:=currentFloor-1 ; i >=BOTTOM_FLOOR ; i--{
				if(queue.up[i] != 0 || queue.internal[i] != 0 || queue.down[i]!= 0){
					return goDown
				}

			}
			for i:=currentFloor+1 ; i < N_FLOORS ; i++{
				if( queue.up[i] != 0 || queue.internal[i] != 0 || queue.down[i]!= 0){
					*elevDir = up_dir
					return goUp
				}
			}
		}
		return stop 

		
	
}


func ElevManager(orderButtonChan chan orderQueue, queueChan chan orderQueue, positionChan chan int) {

	
	var queue orderQueue	
	commandChan:= make(chan Command,100)
	go elevatorController(commandChan)
	elevDir := up_dir

	for{

	select {
		case  orderButtonPushed := <- orderButtonChan:
			//if internal order, set light and update elevqueue(internal)
			//if(orderButtonPushed.internal){
				for i:=0 ; i<N_FLOORS ; i++{
					if (orderButtonPushed.internal[i] != queue.internal[i]) && (orderButtonPushed.internal[i] ==1){
						queue.internal[i]=1
						driver.ButtonLamp(2,i,1)
					}
					//dette erstattes senere av nettverkskommandoer:
					if (orderButtonPushed.up[i] != queue.up[i]) && (orderButtonPushed.up[i] ==1){
						queue.up[i]=1
						driver.ButtonLamp(0,i,1)
					}
					if (orderButtonPushed.down[i] != queue.down[i]) && (orderButtonPushed.down[i] ==1){
						queue.down[i]=1
						driver.ButtonLamp(1,i,1)
					}


				}

			//}			
			//toEventHandler <- orderButtonPushed

		//case neworder <- orderFromromEventHandler
			//update elevqueue with the new order
			//

		case currentFloor := <-positionChan:
			//new floor reached. 
			//if new floor in queue

			
			if stopOnFloor(elevDir,currentFloor, &queue) == true{
				commandChan <- stop
				commandChan <- openDoor

			}
			nextDir := nextDirection(&elevDir,&queue, currentFloor)
			
			if nextDir != stop{				
				commandChan <- nextDir		
			}
			
		}
	}

}
func removeFloorFromQueue(currentFloor int,queue *orderQueue){
	queue.internal[currentFloor]=0
	queue.up[currentFloor]=0
	queue.down[currentFloor]=0
	driver.ButtonLamp(0,currentFloor,0)
	driver.ButtonLamp(1,currentFloor,0)
	driver.ButtonLamp(2,currentFloor,0)
}


func stopOnFloor(elevDir Direction, currentFloor int, queue *orderQueue) bool{
	
	if queue.internal[currentFloor] == 1{
		removeFloorFromQueue(currentFloor,queue)
		return true
	}
	if elevDir ==up_dir{
		if queue.up[currentFloor] == 1{
			removeFloorFromQueue(currentFloor,queue)
			return true
		}

	}else{
		if queue.down[currentFloor] == 1{
			removeFloorFromQueue(currentFloor,queue)
			return true
		}
	}

	
	if elevDir ==up_dir{
		for i:=currentFloor+1 ; i<N_FLOORS ; i++{
			if queue.up[i] == 1 || queue.internal[i] == 1 || queue.down[i] == 1{				
				return false
			}
		}
	}else{
		for i:=currentFloor-1 ; i==BOTTOM_FLOOR ; i--{
			if queue.up[i] == 1 || queue.internal[i] == 1 || queue.down[i] == 1{				
				return false
			}
		}
	}
	return false

	 
}

func CheckOrderButton(orderButtonChan chan orderQueue) {

	var prevbuttonsPressed orderQueue
	var buttonsPressed orderQueue

	for {

		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < 3; button++ {
				//sjekker for manglende knapper i endeetasjer
				if (floor == 0 && button == 1) || (floor == 3 && button == 0) {

				}else {
			
					switch button {
					case 0:
						buttonsPressed.up[floor] =  driver.ButtonPushed(driver.Button_type(button), floor)
						break
					case 1:
						buttonsPressed.down[floor] =  driver.ButtonPushed(driver.Button_type(button), floor) 
						break
					
					case 2:
						buttonsPressed.internal[floor] =  driver.ButtonPushed(driver.Button_type(button), floor)

						break
					default:

						}
					}
				}
		}
		// Only send new if new button is pushed 
	if ((prevbuttonsPressed != buttonsPressed) ){
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
