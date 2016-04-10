package internalOrders

import (
	"io/ioutil"
	"fmt"
	."../driver"
)

const filename = "orders.txt"

func ReadInternals() ([N_FLOORS]int){
	var order [N_FLOORS]int
	filebuffer,err := ioutil.ReadFile(filename)
	if err != nil{
		fmt.Println("Error in reading from file")
	}
	for floor:=0 ;floor < N_FLOORS ; floor++{
		order[floor] = int(filebuffer[floor])
		if order[floor] == 1{
			ButtonLamp(2, floor, 1)
		}
	}
	return order
}

func WriteInternalToFile(order [N_FLOORS]int){
	buf := make([]byte, N_FLOORS)
	for i := 0; i < N_FLOORS; i++ {
		buf[i] = byte(order[i])
	}
	err := ioutil.WriteFile(filename, buf, 0644)
	if err != nil{
		fmt.Println("Error in writing to file")
	}
}