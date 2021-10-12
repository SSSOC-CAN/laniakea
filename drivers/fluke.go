package main

import (
	//"bufio"
	"fmt"
	"net"
	"sync"
	// "os"
	// "strings"
)

var (
	flukeOutputfile = "C:\\Users\\Michael Graham\\Downloads\\H_Test.csv"
	mylocalIp = "192.168.0.202"
	flukeIp = "192.168.0.201"
	flukePort1 = "4369"
	flukePort2 = "3490"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(server_addr string) {
		defer wg.Done()
		_, err := net.Dial("tcp", server_addr)
		if err != nil {
			fmt.Println(err)
			return
		}
	}(flukeIp+":"+flukePort1)
	wg.Wait()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	// for {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Print(">> ")
	// 	text, _ := reader.ReadString('\n')
	// 	fmt.Fprintf(c, text+"\n")

	// 	message, _ := bufio.NewReader(c).ReadString('\n')
	// 	fmt.Print("->: " + message)
	// 	if strings.TrimSpace(string(text)) == "STOP" {
	// 		fmt.Println("TCP client exiting...")
	// 		return
	// 	}
	// }
	// l, err := net.Listen("tcp", mylocalIp+":"+flukePort1)
	// if err != nil {
	// 	fmt.Println("Error listening:", err.Error())
	// 	os.Exit(1)
	// }
	// defer l.Close()
	// fmt.Println("Listening on "+mylocalIp+":"+flukePort1)
	// for {
	// 	conn, err := l.Accept()
	// 	if err != nil {
	// 		fmt.Println("Error accepting: ", err.Error())
	// 		os.Exit(1)
	// 	}
	// 	go handleRequest(conn)
	// }
}

                  
