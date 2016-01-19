package main
	
import (
	"fmt"
	"time"
	"runtime"
)

var i = 0

func thread1(){
	if <-mutexChannel == 1
	
	for j:= 0; j <10; j++{
		i+=1
	}
}

func thread2(){
	for j:= 0; j <10; j++{
		i-=1
	}
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU(runtime.NumCPU())
	mutexChannel := make(chan bool,1)

	go thread1(mutexChannel)
	go thread2(mutexChannel)
	time.Sleep(1000*time.Millisecond)
	fmt.Println(i)
}