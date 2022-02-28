package main

// import (
// 	"fmt"
// 	"github.com/SSSOC-CAN/fmtd/drivers"
// 	"github.com/SSSOC-CAN/fmtd/utils"
// )

// func main() {
// 	c, err := drivers.ConnectToRGA()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	conn := &drivers.RGAConnection{c}
// 	defer conn.Close()
// 	err = conn.InitMsg()
// 	if err != nil {
// 		fmt.Printf("Unable to communicate with RGA: %v\n", err)
// 		return
// 	}
// 	// Setup Data
// 	_, err = conn.Control(utils.AppName, utils.AppVersion)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	resp, err := conn.SensorState()
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(resp.StringSlice())
// 	if resp.Fields["State"].Value.(string) != drivers.Rga_SENSOR_STATE_INUSE {
// 		fmt.Printf("Sensor not ready: %v\n", resp.Fields["State"])
// 		return
// 	}
// 	resp, err = conn.AddBarchart("Bar1", 1, 200, drivers.Rga_PeakCenter, 5, 0, 0, 0)
// 	if err != nil {
// 		fmt.Printf("Could not add Barchart: %v\n", err)
// 		return
// 	}
// 	fmt.Println(resp.StringSlice())
// 	resp, err = conn.AddPeakJump("PeakJump1", drivers.Rga_PeakCenter, 5, 0, 0, 0)
// 	if err != nil {
// 		fmt.Printf("Could not add PeakJump: %v\n", err)
// 		return
// 	}
// 	fmt.Println(resp.StringSlice())
// 	for i := 1; i < 6; i++ { // TODO:SSSOCPaulCote - Change this for test next week
// 		resp, err = conn.MeasurementAddMass(i)
// 		if err != nil {
// 			fmt.Printf("Could not add measurement mass: %v\n", err)
// 			return
// 		}
// 		fmt.Println(resp.StringSlice())
// 	}
// 	resp, err = conn.ScanAdd("Bar1")
// 	if err != nil {
// 		fmt.Printf("Could not add measurement to scan: %v\n", err)
// 		return
// 	}
// 	fmt.Println(resp.StringSlice())
// 	resp, err = conn.DetectorInfo(0)
// 	if err != nil {
// 		fmt.Printf("%v\n", err)
// 		return
// 	}
// 	fmt.Println(resp.StringSlice())
// }