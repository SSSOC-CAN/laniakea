// +build mks,!demo

package drivers

import (
	"net"
	"github.com/SSSOC-CAN/fmtd/errors"
)

// TODO:SSSOCPaulCote - Add list of RGA commands
var (
	rgaIPAddr = "192.168.0.77"
	rgaPort = "10014"
	rgaServer = rgaIPAddr+":"+rgaPort
	RGAMinimumPressure float64 = 0.00005
	RGAMinPollingInterval int64 = 15
)

// ConnectToRGA establishes a conncetion with the RGA
func ConnectToRGA() (DriverConnectionErr, error) {
	c, err := net.Dial("tcp", rgaServer)
	if err != nil {
		return nil, err
	}
	cAssert, ok := c.(*net.TCPConn)
	if !ok {
		return nil, errors.ErrInvalidType
	}
	return &RGAConnection{cAssert}, nil
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
