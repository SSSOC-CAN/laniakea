package drivers

import (
	"net"
)

var (
	rgaIPAddr = "192.168.0.77"
	rgaPort = "10014"
	rgaServer =rgaIPAddr+":"+rgaPort
)

//connectToRGA establishes a conncetion with the RGA
func connectToRGA() (*net.Conn, error) {
	c, err := net.Dial("tcp", rgaServer)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

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
