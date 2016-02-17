package main

import(
	"net"
	"time"
	"fmt"	
	"os/exec"
	"encoding/binary"
)



func main(){
	master := false
	var i uint64 = 0
	port := ":30005"
	
	udpConn,addr:= ClientConnectUDP(port)

	
	fmt.Println("Dobby is a nice slave...")
	buffer := make([]byte,8)


	for !(master){
		udpConn.SetReadDeadline(time.Now().Add(time.Second*2))
		n,_,err:=udpConn.ReadFromUDP(buffer)		
		
		
		if err == nil{				
			i= binary.BigEndian.Uint64(buffer[0:n])
			
		}else{
			fmt.Println("timeout")	
			master = true
			fmt.Println("Master has given Dobby a sock. Dobby is a free elf!...")
			time.Sleep(time.Second*1)
		}
	}
	udpConn.Close()
	startBackup()
	


	udpConn,_ = net.DialUDP("udp",nil,addr)
	for{
		fmt.Println(i)
		i += 1		
		ClientSend(i,udpConn,buffer)
		

		
		

	}
}

func startBackup(){
	Backup := exec.Command("gnome-terminal","-x", "sh", "-c", "go run phoenix.go")
	Backup.Run()
}



func ClientConnectUDP(port string) (*net.UDPConn,*net.UDPAddr){
	adress,_ :=net.ResolveUDPAddr("udp","129.241.187.20"+port)
	conn,_:= net.ListenUDP("udp",adress)

	return conn,adress
}


func ClientSend(i uint64, udpConn *net.UDPConn,buffer []byte){
	binary.BigEndian.PutUint64(buffer, i)
	_,_ = udpConn.Write(buffer)
	time.Sleep(time.Second*1)
}




