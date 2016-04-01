package main

import (
	"../driver"
	"fmt"
	"time"
)

type State int
type Direction int
type Command int

const N_FLOORS = 4

const (
	idle State = iota
	running
	doorOpen
)
const (
	stop Command =iota
	open
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
	directionChan := make(chan int) //keeps an int that is the elevators primary direction.
	queueChan := make(chan orderQueue)
	orderButtonChan := make(chan orderQueue)


	//initialization procedure.
	driver.ElevStart(1)
	<-positionChan
	driver.ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan)
	//Correct so far

	//go-routines
	go ElevPosition(positionChan)
	go CheckOrderButton()
	//LocalChan := make(chan orderQueue, 10)

	

	
	//fmt.Println("fstate", fstate)
	
	ElevManager()

}
func elevatorController(commandChan chan Command){
	command 
}



func NextFloor(elevDir Direction,&queue orderQueue,currentFloor int){
		if elevDir == up_dir{
			for i:=0 ; i < N_FLOORS ; i++{
				if( queue.up[i] != 0 || queue.down)
			}

		}else {

			}																					//	FIX THIS
		
	
}


func ElevManager(orderButtonChan chan orderQueue, queueChan chan orderQueue ) {

	elevState:= idle
	var queue orderQueue
	runElevator := make(chan int)
	runElevator <- 1
	elevDir := up_dir

	for{

	select {
		case  orderButtonPushed := <- orderButtonChan:
			//if internal order, set light and update elevqueue(internal)
			if(orderButtonPushed.internal){
				for i:=0 ; i<N_FLOORS ; i++{
					if (orderButtonPushed.internal[i] != queue.internal[i]) && (orderButtonPushed.internal[i] ){
						queue.internal[i]=1
						driver.ButtonLamp(2,i,1)
					}


				}

			}			
			//toEventHandler <- orderButtonPushed

		//case neworder <- orderFromromEventHandler
			//update elevqueue with the new order
			//

		case currentFloor := <-positionChan:
			//new floor reached. 
			//if new floor in queue
			target := NextFloor(elevDir,&queue, currentFloor)
			if stopOnFloor(currentFloor, &queue){
				doChan <- stop

			}
			//if currentFloor == target
			// 	stop
			//	change state to doorOpne
			// 	set stoplight
			//	wait?
			//  queue.internal[currentfloor]=0
			//when door closes, pop this floor out of queue
		}
	}

}
func removeFloorFromQueue(currentFloor int){
	queue.internal[currentfloor]=0
	queue.up[currentfloor]=0
	queue.down[currentfloor]=0
}


func stopOnFloor(currentFloor int, &queue orderQueue) bool{
	//punkt 1
	if queue.internal[currentFloor]{
		removeFloorFromQueue(currentFloor)
		return true
	}
	//punkt 2
	if elevDir ==up_dir{
		if queue.up[currentFloor]{
			removeFloorFromQueue(currentFloor)
			return true
		}

	}else{
		if queue.down[currentFloor]{
			removeFloorFromQueue(currentFloor)
			return true
		}
	}

	//punkt 3. sjekke motsatte eksterne enn retningen
	if elevDir ==up_dir{
		for i:=currentFloor+1 ; i<N_FLOORS ; i++{
			if queue.up[i] != 1 || queue.internal[i] != 1 || queue.down[i] != 1{
				removeFloorFromQueue(currentFloor)
				return true
			}
		}
	}else{
		for i:=currentFloor-1 ; i==0 ; i--{
			if queue.up[i] != 1 || queue.internal[i] != 1 || queue.down[i] != 1{
				removeFloorFromQueue(currentFloor)
				return true
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

				} 
				else {
			
					switch button {
					case 0:
						buttonsPressed.up[floor] =  drive.ButtonPushed(button, floor)
						break
					case 1:
						buttonsPressed.down[floor] =  drive.ButtonPushed(button, floor) 
						break
					
					case 2:
						buttonsPressed.internal[floor] =  drive.ButtonPushed(button, floor)

						break
					default:

						}
					}
				}
		}
		// Only send new if new button is pushed 
	if ((prevbuttonsPressed != buttonsPressed) && (buttonsPressed == True)){
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
		}
		time.Sleep(time.Millisecond * 40)
	}

}

/*
func Lights(local orderQueue) {
	for i := 0 ; i < N_FLOORS ; i++ {
	        if((local.up[i]==1) && (i < N_FLOORS-1){
	            driver.ButtonLamp(0,i,1)

	        }else if((upQueue[i]==0) && i < N_FLOORS-1){
	            driver.ButtonLamp(0,i,0)
	        }

	        if((downQueue[i]==1) && i > 0){
	            driver.ButtonLamp(1,i,1)
	        }else if((downQueue[i]==0) && i > 0){
	            driver.ButtonLamp(1,i,0)
	        }
	        if(commandQueue[i] == 1){
	            elev_set_button_lamp(2,i,1)
	        }else if(commandQueue[i] == 0){
	            elev_set_button_lamp(2,i,0)
	        }
	    }

    switch(previousFloor){
        case 0:
            elev_set_floor_indicator(0)
            break;
        case 1:
            elev_set_floor_indicator(1)
            break;
        case 2:
            elev_set_floor_indicator(2)
            break;
        case 3:
            elev_set_floor_indicator(3)
            break;
        default:
            break;

    }
}*/

func Sign(val int) int {
	if val < 0 {
		return -1
	} else if val > 0 {
		return 1
	} else {
		return 0
	}

}
