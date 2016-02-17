package main

import(
	"net"
	"time"	
)

func main(){
	sendPort := "50001"
	master(1,sendPort)
}


func master(init_val int,port string){
	addr,_ := net.ResolveUDPAddr("udp","129.241.187.20"+port)
	conn,_ := net.DialUDP("udp",nil,addr)

	conn.Close()

	buf := make([]byte,32)
	i := init_val
	for{
		buf[0] = byte(i)
		conn.Write(buf)
		time.Sleep(time.Millisecond*50)
		i += 1

	}
}





