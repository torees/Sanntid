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

func main() {
	//connect to port 20000 + n
	/*adress,err := net.ResolveUDPAddr("udp","129.241.187.23:20023")
	CheckForError(err)

	// listen at port
	connection,err := net.DialUDP("udp",nil,adress)
	CheckForError(err)
	//defer connection.Close()
	//fmt.Println(connection)

	*/
	go golisten()

	waitchan := make(chan int)
	<-waitchan
	/*for{
		message := []byte("Hello my old friend")
		_, err = connection.Write(message)
		if err != nil{
			fmt.Println("Error, could not send: ",err)
		}

		
		time.Sleep(time.Second *1)

		
	}*/

}

func golisten(){

	recvbuffer := make([]byte,1024)
	adr,err := net.ResolveUDPAddr("udp",":20023")
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
