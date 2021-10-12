package drivers
// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"net"
// 	"os"
// 	"strings"
// )

// var (
// 	rgaIpPort = "192.168.0.77:10014"
// )

// func main() {
// 	reader := bufio.NewReader(os.Stdin)
// 	fmt.Print(">> ")
// 	text, _ := reader.ReadString('\n')
// 	c, err := net.Dial("tcp", "192.168.0.201:4369")
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Fprintf(c, text+"\n")
// 	message, _ := bufio.NewReader(c).ReadString('\n')
// 	fmt.Print("->: " + message)
// 	defer c.Close()
// 	for {
// 		reader := bufio.NewReader(os.Stdin)
// 		fmt.Print(">> ")
// 		text, _ := reader.ReadString('\n')
// 		fmt.Fprintf(c, text+"\n")

// 		message, _ := bufio.NewReader(c).ReadString('\n')
// 		fmt.Print("->: " + message)
// 		if strings.TrimSpace(string(text)) == "STOP" {
// 			fmt.Println("TCP client exiting...")
// 			return
// 		}
// 	}
// }
