package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	c, err := net.Dial("tcp", "localhost:6000")
	if err != nil {
		fmt.Println(err)
		return
	}
	message, _ := bufio.NewReader(c).ReadString('\n')
	fmt.Print("->: " + message)
	if strings.TrimSpace(string(message)) != "MKSRG" {
		fmt.Println("TCP client exiting...")
		return
	}
	defer c.Close()
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, text+"\n")

		message, _ := bufio.NewReader(c).ReadString('\n')
		fmt.Print("->: " + message)
		if strings.TrimSpace(string(text)) == "STOP" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
