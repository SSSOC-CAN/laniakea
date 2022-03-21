package main

import (
	"fmt"
	"reflect"
	"github.com/konimarti/opc"
	"github.com/SSSOC-CAN/fmtd/drivers"
)

var (
	flukeOPCServerName           = "Fluke.DAQ.OPC"
	flukeOPCServerHost           = "localhost"
)

func main() {
	tags, _ := drivers.GetAllTags()
	c, _ := opc.NewConnection(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
		tags,
	)
	fmt.Printf("OPC Connection type: %v", reflect.TypeOf(c))
	var _ drivers.DriverConnection = c
}