package main

import (
	"../driver"
	"fmt"
	"time"
)

type State int

const N_FLOORS = 4
const (
	idle State = iota
	running
	doorOpen
)

type LocalQueue struct {
	internal [4]int
	down     [4]int
	up       [4]int
}

func main() {
	var local LocalQueue
	fmt.Println("Starting Elevator 3000...")
	driver.ElevInit()
	positionChan := make(chan int)

	queueChan := make(chan LocalQueue)
	LocalChan := make(chan LocalQueue, 10)
	reset := make(chan int)
	go ElevPosition(positionChan)
	go CheckOrderButton(queueChan, LocalChan, reset)
	driver.ElevStart(1)
	<-positionChan
	driver.ElevStart(0)
	fmt.Println("Initialized at floor", <-positionChan)
	fstate := idle
	//fmt.Println("fstate", fstate)
	queueChan <- local
	ElevManager(local, fstate, positionChan, LocalChan, queueChan, reset)

}
func ElevManager(local LocalQueue, fstate State, positionChan chan int, LocalChan chan LocalQueue, queueChan chan LocalQueue, reset chan int) {

	var target int
	LocalChan <- local
	for {
		//<-LocalChan

		//fmt.Println("hei")

		select {
		case temp := <-queueChan:
			local = temp

		case <-LocalChan:
			target = OrderFromLocalQueue(local)

			//fmt.Println(<-queueChan)+
			fmt.Println("fstate", fstate)

			switch fstate {

			case idle:

				fmt.Println("Target: ", target)
				if target != -1 {
					position := <-positionChan
					dir := Sign(target - position)
					driver.ElevStart(driver.Elev_dir(dir))
					fstate = running
					break
				}

				break

			case running:
				fmt.Println("Running")
				//target = OrderFromLocalQueue(queueChan)
				position := <-positionChan
				if position == target {
					driver.ElevStart(0)
					fstate = idle
					local.up[position] = 0
					local.down[position] = 0
					local.internal[position] = 0
					reset <- 1
					break
				}

				break

			case doorOpen:
				fstate = idle
				time.Sleep(time.Second * 1)
				break

			}
			LocalChan <- local
		default:
			LocalChan <- local
		}
	}

}

//if (driver.ButtonPushed(j,i)){
func CheckOrderButton(queueChan chan LocalQueue, LocalChan chan LocalQueue, reset chan int) {
	local := <-queueChan
	prevque := local
	for {

		for i := 0; i < N_FLOORS; i++ {
			for j := 0; j < 3; j++ {
				//sjekker for manglende knapper i endeetasjer
				if (i == 0 && j == 1) || (i == 3 && j == 0) {

				} else {
					if driver.ButtonPushed(driver.Button_type(j), i) == 1 {
						switch j {
						case 1:
							local.down[i] = 1
							break
						case 0:
							local.up[i] = 1
							break
						case 2:
							local.internal[i] = 1
							break
						default:

						}
					}
				}
			}

		}
		select {
		case <-reset:
			local = <-LocalChan
			prevque = local
		default:
			if local != prevque {
				queueChan <- local
				prevque = local
			}
		}
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

func OrderFromLocalQueue(local LocalQueue) int {

	returnval := -1
	for i := 0; i < 4; i++ {
		if (local.up[i] != 0) || (local.internal[i] != 0) || (local.down[i] != 0) {
			returnval = i
		}
	}

	return returnval
}

func Lights(local LocalQueue) {
	for i := 0 ; i < N_FLOORS ; i++ {
	        if((local.up[i]==1) && (i < N_FLOORS-1){
	            driver.ButtonLamp(0,i,1)

	        }else if((upQueue[i]==0) && i < N_FLOORS-1){
	            driver.ButtonLamp(0,i,0);
	        }

	        if((downQueue[i]==1) && i > 0){
	            driver.ButtonLamp(1,i,1);
	        }else if((downQueue[i]==0) && i > 0){
	            driver.ButtonLamp(1,i,0);
	        }
	        if(commandQueue[i] == 1){
	            elev_set_button_lamp(2,i,1);
	        }
	        else if(commandQueue[i] == 0){
	            elev_set_button_lamp(2,i,0);
	        }
	    }

    switch(previousFloor){
        case 0:
            elev_set_floor_indicator(0);
            break;
        case 1:
            elev_set_floor_indicator(1);
            break;
        case 2:
            elev_set_floor_indicator(2);
            break;
        case 3:
            elev_set_floor_indicator(3);
            break;
        default:
            break;

    }

}

}

func Sign(val int) int {
	if val < 0 {
		return -1
	} else if val > 0 {
		return 1
	} else {
		return 0
	}

}
