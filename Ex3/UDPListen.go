package main
	
import (
	"fmt"	
	"net"
	"os"
)



func CheckForError(err error){
	if err != nil {  //nil is default error from "net" package
		fmt.Println("Error in connecting to port:",err)
		os.Exit(0)
	}
}

func main(){
	go golisten()


	
}

func golisten(){

	recvbuffer := make([]byte,1024)
	adr,err := net.ResolveUDPAddr("udp",":20003")
	if err != nil{
			fmt.Println("Error in resolve: ",err)
		}

	conn,err := net.ListenUDP("udp",adr)
	if err != nil{
			fmt.Println("Error in listen: ",err)
		}
	for{
	n,adr,err :=  conn.ReadFromUDP(recvbuffer)
		fmt.Println("Receiving: ",string(recvbuffer[0:n]), " from ", adr)
		
		if err != nil{
			fmt.Println("Error in read: ",err)
		}
	}
}