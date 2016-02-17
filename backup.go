package main

import (
	"fmt"
	"os/exec"
	"time"
	"net"
	"strconv"
	
)


func main(){
	listenPort := ":50001"
	addr,_ := net.ResolveUDPAddr("udp",listenPort)

	listenConn,_ := net.ListenUDP("udp",addr)

	backupVal := backup(listenConn)

	listenConn.Close()

	addr,_ = net.ResolveUDPAddr("udp","129.241.187.20"+listenPort)

	sendConn,_ := net.DialUDP("udp",nil,addr)

	master(backupVal,sendConn)
	sendConn.Close()
}


func master(init_val int, conn *net.UDPConn){
	Backup := exec.Command("gnome-terminal","-x","go", "run", "backup.go")
	Backup.Run()

	buf := make([]byte,1)
	i := init_val
	for {
		fmt.Println(i)
		buf[0]=byte(i)
		conn.Write(buf)
		time.Sleep(time.Millisecond*1000)
		i += 1
	}
}

func backup(conn *net.UDPConn) int{
	someChan := make(chan int, 1)
	backupVal := 0
	go UDPlisten(someChan,conn)
	for{
		select{
			case backupVal = <- someChan:
				time.Sleep(time.Millisecond*200)
				break
			case <- time.After(time.Second*1):
				return backupVal

		}
	}
}

func UDPlisten( pingChan chan int, conn *net.UDPConn){
	buf := make([]byte,1024)

	for {
		conn.ReadFromUDP(buf[:])
		str := string(buf)
		
		intval,_ := strconv.Atoi(str)
		pingChan <- intval
	}
}