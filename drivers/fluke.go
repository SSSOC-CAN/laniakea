// +build windows,386,fluke,!demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10

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
	"sort"
	"strconv"

	"github.com/konimarti/opc"
)

type Tag struct {
	name string
	tag  string
}

var (
	TelemetryDefaultPollingInterval int64 = 10
	MinTelemetryPollingInterval     int64 = 5
	TelemetryPressureChannel        int64 = 81
	flukeOPCServerName                    = "Fluke.DAQ.OPC"
	flukeOPCServerHost                    = "localhost"
	customerChannelString = "customer channel "
	coldfingerSup         = "Coldfinger sup"
	coldfingerRet         = "Coldfinger ret"
	coldfingerFin         = "Coldfinger fin"
	platenIn              = "Platen_TTin"
	platenOut             = "Platen_TTout"
	couponString          = "Coupon#"
	customerSupply        = "CustomerSup"
	voltageString         = "Voltage"
	rearShroudSupStr      = "Rear Shroud (R.S. Supply)"
	rearShroudUpStr       = "Rear Shroud (R.S. Upright)"
	rearShroudRetStr      = "Rear Shroud (R.S. Return)"
	coldfingerStr         = "Coldfinger"
	platenLeftRearStr     = "Platen (L.S. Rear)"
	platenLeftFrontStr    = "Platen (L.S. Front)"
	platenRightFrontStr   = "Platen (R.S. Front)"
	platenRightRearStr    = "Platen (R.S. Rear)"
	platenRetStr          = "Platen (Return S-bend)"
	mainSupRearStr        = "Main (Supply man Rear)"
	mainSupFrontStr       = "Main (Supply man Front)"
	mainRetFrontStr       = "Main (Return man Front)"
	mainRetRearStr        = "Main (Return man Rear)"
	frontDoorSupStr       = "Front Door (D.S. Supply)"
	frontDoorRetStr       = "Front Door (D.S. Return)"
	frontDoorSkinStr      = "Front Door Skin"
	rearSkinStr           = "Rear Skin (bell)"
	mainShroudRearStr     = "Rear of Main Shroud"
	mainShroudFrontStr    = "Front of Main Shroud"
	platenSupStr          = "Platen (Supply S-bend)"
	computedOne           = "CustSup_Current"
	pressureStr           = "Pressure_Test"
	defaultTagMap         = func(tags []string) map[int]Tag {
		tagMap := make(map[int]Tag)
		for i, t := range tags {
			var str string
			switch {
			case i == 0:
				str = "Scan"
			case i < 41 && i > 0: // first 40 channels are customer channels

				str = customerChannelString + strconv.Itoa(i)
			case i == 43:

				str = coldfingerSup
			case i == 44:

				str = coldfingerRet
			case i == 45:

				str = coldfingerFin
			case i == 66:

				str = platenIn
			case i == 67:

				str = platenOut
			case i == 81:

				str = customerSupply + " - " + voltageString
			case i < 95 && i > 81:

				str = couponString + strconv.Itoa(i-80) + " - " + voltageString
			case i == 95:

				str = couponString + strconv.Itoa(1) + " - " + voltageString
			case i == 101:

				str = rearShroudSupStr
			case i == 102:

				str = rearShroudUpStr
			case i == 103:

				str = rearShroudRetStr
			case i == 104:

				str = coldfingerStr
			case i == 105:

				str = platenLeftRearStr
			case i == 106:

				str = platenLeftFrontStr
			case i == 107:

				str = platenRightFrontStr
			case i == 108:

				str = platenRightRearStr
			case i == 109:

				str = platenRetStr
			case i == 110:

				str = mainSupRearStr
			case i == 111:

				str = mainSupFrontStr
			case i == 112:

				str = mainRetFrontStr
			case i == 113:

				str = mainRetRearStr
			case i == 114:

				str = frontDoorSupStr
			case i == 115:

				str = frontDoorRetStr
			case i == 116:

				str = frontDoorSkinStr
			case i == 117:

				str = rearSkinStr
			case i == 118:

				str = mainShroudRearStr
			case i == 119:

				str = mainShroudFrontStr
			case i == 120:

				str = platenSupStr
			case i == 121:

				str = computedOne
			case i == 122:

				str = pressureStr
			case i > 122 && i < 136:

				str = couponString + strconv.Itoa(i-121) + " - Current"
			case i == 136:
				str = couponString + strconv.Itoa(1) + " - Current"
			}
			if str != "" {
				tagMap[i] = Tag{name: str, tag: t}
			}
		}
		return tagMap
	}
)

type DAQConnection struct {
	opc.Connection
	Tags	[]string
	TagMap	map[int]Tag
}

var _ DriverConnection = (*DAQConnection) (nil)
var _ DriverConnection = DAQConnection{}

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

// ConnectToDAQ establishes a connection with the OPC server of the Fluke DAQ software and the FMTD
func ConnectToDAQ() (DriverConnection, error) {
	tags, err := GetAllTags()
	if err != nil {
		return nil, err
	}
	c, err := opc.NewConnection(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
		tags,
	)
	if err != nil {
		return nil, err
	}
	return &DAQConnection{
		Connection: c,
		Tags: tags,
		TagMap: defaultTagMap(tags),
	}, nil
}

// StartScanning starts the scanning process on the DAQ
func (d *DAQConnection) StartScanning() error {
	err := d.Write(d.TagMap[0].tag, true)
	if err != nil {
		return err
	}
	return nil
}

// StopScanning stops the scanning process on the DAQ
func (d *DAQConnection) StopScanning() error {
	err := d.Write(d.TagMap[0].tag, false)
	if err != nil {
		return err
	}
	return nil
}

// GetTagMapNames returns a slice of all the TagMap names
func (d *DAQConnection) GetTagMapNames() []string {
	idxs := make([]int, 0, len(d.TagMap))
	for idx := range d.TagMap {
		idxs = append(idxs, idx)
	}
	sort.Ints(idxs)
	names := make([]string, 0, len(idxs)-1)
	for _, i := range idxs {
		if i != 0 {
			names = append(names, d.TagMap[i].name)
		}
	}
	return names
}

type Reading struct {
	Item	opc.Item
	Name	string
}

// ReadItems returns a slice of all readings
func (d *DAQConnection) ReadItems() []Reading {
	idxs := make([]int, 0, len(d.TagMap))
	for idx := range d.TagMap {
		idxs = append(idxs, idx)
	}
	sort.Ints(idxs)
	readings := make([]Reading, 0, len(idxs)-1)
	for _, i := range idxs {
		if i != 0 {
			readings = append(readings, Reading{
				Item: d.ReadItem(d.TagMap[i].tag),
				Name: d.TagMap[i].name,
			})
		}
	}
	return readings
}