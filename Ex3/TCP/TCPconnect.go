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

func main(){

	reply := make([]byte, 1024)
	tcpAddress, _ := net.ResolveTCPAddr("tcp","129.241.187.23:33546" )
	conn,err:= net.DialTCP("tcp",nil,tcpAddress )
	CheckForError(err)

	conn.Read(reply)

	fmt.Println("reply from server: ",string(reply))
	
	for{
		
		//text := "Connect to: 129.241.187.20:33546\x00"
		text := "halloem\x00"
		_, err := conn.Write([]byte(text))
		CheckForError(err)
		_, err = conn.Write([]byte(text))
		CheckForError(err)
		fmt.Println("forever looping!")
		

    	conn.Read(reply)
    	//CheckForError(err)

    	fmt.Println("reply: ",string(reply))

		time.Sleep(time.Second * 1)
	}

	//go goRead()
	//go goSend()

	
}


//func goSend(){
//
//}

//func goRead(){
//
//}