package main
	
import (
	"fmt"	
	"net"
	"os"
	"time"
)



func CheckForError(err error){
	if err != nil {  //nil is default error from "net" package
		fmt.Println("Error in connecting to port:",err)
		os.Exit(0)
	}
}

func main() {
	//connect to port 30000
	adress,err := net.ResolveUDPAddr("udp",":30000")
	CheckForError(err)

	// listen at port
	connection,err := net.ListenUDP("udp",adress)
	CheckForError(err)
	defer connection.Close()

	recvbuffer := make([]byte,1024)

	for{
		n,adress,err :=  connection.ReadFromUDP(recvbuffer)
		fmt.Println("Receiving: ",string(recvbuffer[0:n]), " from ", adress)
		time.Sleep(time.Second *1)
		if err != nil{
			fmt.Println("Error: ",err)
		}
	}

}