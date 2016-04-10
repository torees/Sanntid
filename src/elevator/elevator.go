package elevator

import (
	"../elevManager"
	. "../driver"
	"math"

)



const MAX_ORDER_COST = 100


type Elevator struct {
	Queue        elevManager.OrderQueue
	Direction    int
	CurrentFloor int
	IP           string
}


func (elev Elevator) Cost(order elevManager.OrderQueue) (int, string) {
	const dirCost = 2
	const distCost = 4
	const numOrderCost = 6
	cost := 5

	distanceCost := (elev.CurrentFloor - elev.findOrderFloor(order)) * distCost
	directionCost := 0

	if distanceCost < 0 {
		directionCost = dirCost
		distanceCost = int(math.Abs(float64(distanceCost)))
	}

	cost = elev.numOrdersInQueue()*numOrderCost + distanceCost + directionCost
	return cost, elev.IP
}

func (elev Elevator) findOrderFloor(order elevManager.OrderQueue) int {
	for i := 0; i < N_FLOORS; i++ {
		if order.Up[i] == 1 || order.Down[i] == 1 {
			return i
		}
	}
	return -1
}

func (elev Elevator) numOrdersInQueue() int {
	numOrders := 0
	for i := 0; i < N_FLOORS; i++ {
		if elev.Queue.Up[i] == 1 {
			numOrders += 1
		}
		if elev.Queue.Down[i] == 1 {
			numOrders += 1
		}
		if elev.Queue.Internal[i] == 1 {
			numOrders += 1
		}
	}
	return numOrders
}

func (elev Elevator) NewOrder(order elevManager.OrderQueue) bool {

	for floor := 0; floor < N_FLOORS; floor++ {
		if elev.Queue.Up[floor] == 1 && order.Up[floor] == 1 {
			return false
		}
		if elev.Queue.Down[floor] == 1 && order.Down[floor] == 1 {
			return false
		}

	}
	return true
}
