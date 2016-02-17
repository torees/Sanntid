package main

import (
	"fmt"
	"strconv"
	"os/exec"
	"time"
	"net"
	
)



func main(){
	master := false
	
	pingPort := ":50001"
	countValuePort := ":50002"
	pingConn := ClientConnectUDP(pingPort)
	countConn := ClientConnectUDP(countValuePort)
	
	timeOut:= make(chan int)
	pingRecv:= make(chan int)


	
	//valueChan := make(chan int)
	
	go checkPing(pingPort,pingRecv,timeOut)
	i:= 0
	cmd:= exec.Command("gnome-terminal"," -x","go run /home/student/Documents/Sanntid/main.go")
	
	for !master{

		select{
			case <-timeOut:		
				fmt.Println("Dobby is slave no more! ")
				master = true
				
				
				//timeout
			case <- pingRecv:
				fmt.Println("tick")
				
			

			default:
				//fmt.Println("still slave")
				//i=<-valueChan


		}
	}


	go pingThread(pingConn)
	cmd.Output()
	for{
		fmt.Println(i)
		ClientSend(i,countConn)
		i += 1

	}

}



func checkCountVal(valueChan chan int,countValuePort string){
	ServAddr,_ := net.ResolveUDPAddr("udp",countValuePort)
	ServConn,_ := net.ListenUDP("udp",ServAddr)
	buf := make([]byte,1024)
	for{
		n,_,_ := ServConn.ReadFromUDP(buf)
		str := string(buf[:n])
		fmt.Println(str)
		/*intval := strconv.Atoi(str)
		valueChan <-intval*/
	}

}
	

func ClientConnectUDP(port string) *net.UDPConn{
	adress,_ :=net.ResolveUDPAddr("udp","129.241.187.20"+port)
	conn,_:= net.DialUDP("udp",nil,adress)

	return conn	
}


func ClientSend(CountValue int, conn *net.UDPConn){
	stringVal := strconv.Itoa(CountValue)
	msg := []byte(stringVal)
	_,_ = conn.Write(msg)
	time.Sleep(time.Second*1)
}

func pingThread(pingConn *net.UDPConn){
	p :=[]byte("1")
	for{
		_,_ = pingConn.Write(p)
		time.Sleep(time.Millisecond*250)
	}
}

func checkPing(pingPort string, pingRecv chan int,timeOut chan int){
	ServAddr,_ := net.ResolveUDPAddr("udp",pingPort)
	ServConn,_ := net.ListenUDP("udp",ServAddr)
	
	
	buf := make([]byte,1024)
	for{
		fmt.Println("Dobby is slave! ")
		ServConn.SetReadDeadline(time.Now().Add(time.Second*1))
		fmt.Println("Slaving away...")		
	
		_,_,err:=ServConn.ReadFromUDP(buf)
		fmt.Println("Dobby NOT slave! ")
		if err != nil{
			fmt.Println("error not nil")
			timeOut<-1

		}
		pingRecv <- 1
		//time.Sleep(time.Second*1)

	}
}





	