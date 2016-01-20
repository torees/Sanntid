package main
	
import (
	"fmt"
	"time"
	"runtime"
)

var i = 0

func thread1(){
	for j:= 0; j <100000; j++{
		i+=1
	}
}

func thread2(){
	for j:= 0; j <99999; j++{
		i-=1
	}
}

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU())
	go thread1()
	go thread2()
	time.Sleep(1000*time.Millisecond)
	fmt.Println(i)
}
