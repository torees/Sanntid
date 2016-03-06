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
	fmt.Println("hello world")
	driver.ElevInit()
	positionChan := make(chan int)
	nextTarget := make(chan int)
	queueChan := make(chan LocalQueue)
	go ElevPosition(positionChan)
	go CheckOrderButton(queueChan)
	driver.ElevStart(1)
	<-positionChan
	driver.ElevStart(0)
	fmt.Println("inlitialized")

	fstate := idle
	for {
		ElevManager(local, fstate, positionChan, nextTarget, queueChan)
	}

}
func ElevManager(local LocalQueue, fstate State, positionChan chan int, nextTarget chan int, queueChan chan LocalQueue) {

	var target int

	fmt.Println(<-positionChan)
	/*
		switch fstate {

		case idle:
			target = <-nextTarget
			position := <-positionChan
			dir := Sign(target - position)
			driver.ElevStart(driver.Elev_dir(dir))
			fstate = running
			break

		case running:
			target = <-nextTarget
			position := <-positionChan
			if position == target {
				driver.ElevStart(0)
				fstate = idle
				local.down[position] = 0
				local.internal[position] = 0
				break
			}
			fmt.Println(position)
			break

		case doorOpen:
			fstate = idle
			time.Sleep(time.Second * 1)
			break

		}*/

}

//if (driver.ButtonPushed(j,i)){
func CheckOrderButton(queueChan chan LocalQueue) {
	local := <-queueChan
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
	queueChan <- local
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

func OrderFromLocalQueue(local LocalQueue, nextTarget chan int, queueChan chan LocalQueue) {
	local = <-queueChan
	for i := 0; i < 4; i++ {
		if (local.up[i] != 0) || (local.internal[i] != 0) || (local.down[i] != 0) {
			nextTarget <- i
		}
	}
	queueChan <- local
}

func Lights(local LocalQueue) {

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
