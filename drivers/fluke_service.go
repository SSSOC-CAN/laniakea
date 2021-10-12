package drivers

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync/atomic"
	"time"
	"github.com/konimarti/opc"
	"github.com/rs/zerolog"
)

var (
	customerChannelString = "customer channel "
	coldfingerSup = "Coldfinger sup"
	coldfingerRet = "Coldfinger ret"
	coldfingerFin = "Coldfinger fin"
	platenIn = "Platen_TTin"
	platenOut = "Platen_TTout"
	couponString = "Coupon#"
	customerSupply = "CustomerSup"
	voltageString = "Voltage"
	tcStr = "TC#"
	rearShroudSupStr = "Rear Shroud (R.S. Supply)"
	rearShroudUpStr = "Rear Shroud (R.S. Upright)"
	rearShroudRetStr = "Rear Shroud (R.S. Return)"
	coldfingerStr = "Coldfinger"
	platenLeftRearStr = "Platen (L.S. Rear)"
	platenLeftFrontStr = "Platen (L.S. Front)"
	platenRightFronStr = "Platen (R.S. Front)"
	platenRightRearStr = "Platen (R.S. Rear)"
	platenRetStr = "Platen (Return S-bend)"
	mainSupRearStr = "Main (Supply man Rear)"
	mainSupFrontStr = "Main (Supply man Front)"
	mainRetFrontStr = "Main (Return man Front)"
	mainRetRearStr = "Main (Return man Rear)"
	frontDoorSupStr = "Front Door (D.S. Supply)"
	frontDoorRetStr = "Front Door (D.S. Return)"
	frontDoorSkinStr = "Front Door Skin"
	rearSkinStr = "Rear Skin (bell)"
	mainShroudRearStr = "Rear of Main Shroud"
	mainShroudFrontStr = "Front of Main Shroud"
	platenSupStr = "Platen (Supply S-bend)"
	computedOne = "CustSup_Current"
	pressureStr = "Pressure_Test"
	unusedStr = "Unused"
	customerModulePattern = "Instrument 01.Module [1-2]"
	customerChannelPattern = "Channel ([0-9]{2}([0-9]))"
	channelStr = "Channel"
	defaultTagMap = func(tags []string) (map[string]string, err) {
		rM, err := regexp.Compile(customerModulePattern)
		if err != nil {
			return nil, err
		}
		tagMap := make(map[string]string)
		for i, t := range tags {
			if rM.MatchString(t) { // first 2 modules are customer channels
				rC, err := regexp.Compile(customerChannelPattern)
				if err != nil {
					return nil, err
				}
				s := rC.FindStringSubmatch(t)
				channelNum, err := strconv.Atoi(s[1])
				if err != nil {
					return nil, err
				}
				number, err := strconv.Atoi(s[2])
				if err != nil {
					return nil, err
				}
				if channelNum < 200 {
					tagMap[customerChannelString+s[2]] = t
				} else {
					tagMap[customerChannelString+strconv.Itoa(number+19)] = t
				}
			}
			switch {
			case i < 40: // first 40 channels are customer channels
				tagMap[customerChannelString+strconv.Itoa(i+1)] = t
			case i == 42:
				tagMap[coldfingerSup] = t
			case i == 43:
				tagMap[coldfingerRet] = t
			case i == 44:
				tagMap[coldfingerFin] = t
			case i == 65:
				tagMap[platenIn] = t
			case i == 66:
				tagMap[platenOut] = t
			case i == 80:
				tagMap[customerSupply+" - "+voltageString] = t
			case i < 94 && i > 80:
				tagMap[couponString+strconv.Itoa(i-79)+" - "+voltageString] = t
			case i == 94:
				tagMap[couponString+strconv.Itoa(1)+" - "+voltageString] = t
			case i == 100:
				tagMap[rearShroudSupStr] = t
			case i == 101:
				tagMap[rearShroudUpStr] = t
			case i == 102:
				tagMap[rearShroudRetStr] = t
			case i == 103:
				tagMap[coldfingerStr] = t
			case i == 104:
				tagMap[platenLeftRearStr] = t
			case i == 105:
				tagMap[platenLeftFrontStr] = t
			case i == 106:
				tagMap[platenRightFronStr] = t
			case i == 107:
				tagMap[platenRightRearStr] = t
			case i == 108:
				tagMap[platenRetStr] = t
			case i == 109:
				tagMap[mainSupRearStr] = t
			case i == 110:
				tagMap[mainSupFrontStr] = t
			case i == 111:
				tagMap[mainRetFrontStr] = t
			case i == 112:
				tagMap[mainRetRearStr] = t
			case i == 113:
				tagMap[frontDoorSupStr] = t
			case i == 114:
				tagMap[frontDoorRetStr] = t
			case i == 115:
				tagMap[frontDoorSkinStr] = t
			case i == 116:
				tagMap[rearSkinStr] = t
			case i == 117:
				tagMap[mainShroudRearStr] = t
			case i == 118:
				tagMap[mainShroudFrontStr] = t
			case i == 119:
				tagMap[platenSupStr] = t
			case i == 120:
				tagMap[computedOne] = t
			case i == 121:
				tagMap[pressureStr] = t
			case i > 121 && i < 135:
				tagMap[couponString+strconv.Itoa(i-119)+" - Current"] = t
			case i == 135:
				tagMap[couponString+strconv.Itoa(1)+" - Current"] = t
			default:
				tagMap[unusedStr] = t
			}
		}
		return tagMap, nil
	}
	flukeOPCServerName = "Fluke.DAQ.OPC"
	flukeOPCServerHost = "localhost"
)

// FlukeService is a struct for holding all relevant attributes to interfacing with the Fluke DAQ
type FlukeService struct {
	Active 		int32 // atomic
	Stopping	int32 // atomic
	Logger		*zerolog.Logger
	tags		[]string
	tagMap		map[string]string
	connection	opc.Connection
}

// NewFlukeService creates a new Fluke Service object which will use the drivers for the Fluke DAQ software
func NewFlukeService(logger *zerolog.Logger) (*FlukeService, error) {
	tags, err := GetAllTags()
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve all tags: %v", err)
	}
	tMap, err := defaultTagMap(tags)
	if err != nil {
		return nil, fmt.Errorf("Could not create default tag map: %v", err)
	}
	return &FlukeService{
		Logger: logger,
		tags: tags,
		tagMap: ,
	}, nil
}

// Start starts the service. Returns an error if any issues occur
func (s *FlukeService) Start() error {
	s.Logger.Info().Msg("Starting Fuke...")
	atomic.StoreInt32(&s.Active, 1)
	c, err := opc.NewConnection(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
		s.tags,
	)
	if err != nil {
		return err
	}
	s.connection = c
	return nil
}

// Stop stops the service. Returns an error if any issues occur
func (s *FlukeService) Stop() error {
	s.Logger.Info().Msg("Stopping Daemon...")
	atomic.StoreInt32(&s.Stopping, 1)
	s.connection.Close()
	return nil
}

func (s *FlukeService) StartRecording() error {
	current_time := time.Now()
	file, err := os.Create(fmt.Sprintf("%02d-%02d-%d-fluke.csv", current_time.Day(), current_time.Month(), current_time.Year()))
	if err != nil {
		return fmt.Errorf("Could not create file %s: %v", file, err)
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	// headers
	headerData := []string{"Timestamp"}
	for name, _ := range s.tagMap {
		if name != unusedStr {
			headerData = append(headerData, name)
		}
	}
	err = writer.Write(headerData)
	if err != nil {
		return err
	}
	data := []string{fmt.Sprintf("%02d:%02d:%02d", current_time.Hour(), current_time.Minute(), current_time.Second())}
	for name, tag := range s.tagMap {
		if name != unusedStr {
			data = append(data, fmt.Sprintf("%g", s.connection.ReadItem(tag).Value))
		}
	}
	err = writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}