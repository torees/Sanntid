package main

import(
	
	"time"
	
	"./driver"

)





func main(){
	driver.ElevatorInit()
	//for
		//listen on network
		//if master ping
			//continue
		//else
			//master=true
			
			//break
		//if backup
			//slave = true
	//<-networkAccessChannel
	for{
		driver.SetElevatorDir(-1)
		time.Sleep(time.Second*5)
		driver.SetElevatorDir(1)
		time.Sleep(time.Second*5)
	}

	//network.sendOrder(order typeorder )
}