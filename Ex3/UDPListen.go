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


	
}