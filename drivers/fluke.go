/*
The MIT License (MIT)

Copyright © 2018 Kalkfabrik Netstal AG, <info@kfn.ch>

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the “Software”), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package drivers contains all the code to communicate with Fluke DAQ software over OPC DA
package drivers

import (
	"fmt"
	"github.com/konimarti/opc"
)

// GetAllTags returns a slice of all detected tags
func GetAllTags() ([]string, error) {
	b, err := opc.CreateBrowser(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
	)
	if err != nil {
		return []string{}, err
	}
	return opc.CollectTags(b), nil
}

func main() {
	b, _ := opc.CreateBrowser(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
	)
	//opc.PrettyPrint(b)
	listOTags := opc.CollectTags(b)
	
	c, _ := opc.NewConnection(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
		listOTags,
	)
	defer c.Close()
	reading := c.Read()
	fmt.Println(reading["Instrument 01.Module 3.Channel 304"].Value)

	for _, t := range listOTags {
		fmt.Printf("%s: %v\n", t, c.ReadItem(t))
	}
}

                  
