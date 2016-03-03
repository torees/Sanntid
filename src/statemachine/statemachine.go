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
	down
	up
	doorOpen
)

type LocalQueue struct {
	internal [4]int
	down     [4]int
	up       [4]int
}

func main() {
	fmt.Println("hello world")
	var local LocalQueue
	fmt.Println(local)
	for {
		local = CheckOrderButton(local)
		fmt.Println(local)
	}
}
func ElevatorManager() {
	fmt.Println("hello world")
	var local LocalQueue
	fmt.Println(local)
	local = CheckOrderButton(local)
	fmt.Println(local)

}

//if (driver.ButtonPushed(j,i)){
func CheckOrderButton(local LocalQueue) LocalQueue {

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
	return local
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

func OrderFromLocalQueue() {

}

func Lights() {

}
