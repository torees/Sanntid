package main
	
import (
	"fmt"
	"time"
	"runtime"
)




func thread1(guard chan int){
	
	
	for j:= 0; j <100000; j++{
		i:=<-guard
		i+=1
		guard <- i
	}
	
	
}

func thread2(guard chan int){
	for j:= 0; j <10000; j++{
		i:=<-guard
		i-=1
		guard <- i
	}
}

func main(){
	var i = 0
	runtime.GOMAXPROCS(runtime.NumCPU())
	guard := make(chan int,1)

	guard <- i
	go thread1(guard)
	go thread2(guard)
	
	time.Sleep(1000*time.Millisecond)
	
	fmt.Println("nubmer: ",<-guard)
}
