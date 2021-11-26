package drivers

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/konimarti/opc"
	"github.com/rs/zerolog"
)

type Tag struct {
	name string
	tag  string
}

var (
	pressureRead int = 122
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
	flukeOPCServerName           = "Fluke.DAQ.OPC"
	flukeOPCServerHost           = "localhost"
	DefaultPollingInterval int64 = 10
	minPollingInterval     int64 = 5
)

// FlukeService is a struct for holding all relevant attributes to interfacing with the Fluke DAQ
type FlukeService struct {
	Active     				int32 // atomic
	Stopping   				int32 // atomic
	Recording  				int32 // atomic
	Broadcasting			int32 // atomic
	Logger     				*zerolog.Logger
	name					string
	tags       				[]string
	tagMap     				map[int]Tag
	connection 				opc.Connection
	QuitChan   				chan struct{}
	CancelChan				chan struct{}
	outputDir  				string
	BuffedChan				chan *fmtrpc.RealTimeData
	StateChangeChan			chan *data.StateChangeMsg
	PressureChan			chan float64
	IsRGAReady				bool
	filepath				string
}

// A compile time check to make sure that FlukeService fully implements the data.Service interface
var _ data.Service = (*FlukeService) (nil)


// NewFlukeService creates a new Fluke Service object which will use the drivers for the Fluke DAQ software
func NewFlukeService(logger *zerolog.Logger, outputDir string) (*FlukeService, error) {
	tags, err := GetAllTags()
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve all tags: %v", err)
	}
	return &FlukeService{
		Logger:    logger,
		tags:      tags,
		tagMap:    defaultTagMap(tags),
		QuitChan:  make(chan struct{}),
		CancelChan: make(chan struct{}),
		PressureChan: make(chan float64),
		outputDir: outputDir,
		name: 	   data.FlukeName,
	}, nil
}

// Start starts the service. Returns an error if any issues occur
func (s *FlukeService) Start() error {
	s.Logger.Info().Msg("Starting Fluke Service...")
	atomic.StoreInt32(&s.Active, 1)
	c, err := opc.NewConnection(
		flukeOPCServerName,
		[]string{flukeOPCServerHost},
		s.tags,
	)
	if err != nil {
		return err
	}
	// Start scanning
	err = c.Write(s.tagMap[0].tag, true)
	if err != nil {
		return err
	}
	s.connection = c
	go s.ListenForRTDSignal()
	s.Logger.Info().Msg("Fluke Service started.")
	return nil
}

// Stop stops the service. Returns an error if any issues occur
func (s *FlukeService) Stop() error {
	s.Logger.Info().Msg("Stopping Fluke Service...")
	atomic.StoreInt32(&s.Stopping, 1)
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return fmt.Errorf("Could not stop Fluke service: %v", err)
		}
	}
	close(s.CancelChan)
	close(s.QuitChan)
	// Stop scanning
	err := s.connection.Write(s.tagMap[0].tag, false)
	if err != nil {
		return err
	}
	s.connection.Close()
	close(s.PressureChan)
	s.Logger.Info().Msg("Fluke Service stopped.")
	return nil
}

// Name satisfies the fmtd.Service interface
func (s *FlukeService) Name() string {
	return s.name
}

// StartRecording starts the recording process by creating a csv file and inserting the header row into the file and returns a quit channel and error message
func (s *FlukeService) startRecording(pol_int int64) error {
	if pol_int < minPollingInterval && pol_int != 0 {
		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", minPollingInterval)
	} else if pol_int == 0 { //No polling interval provided
		pol_int = DefaultPollingInterval
	}
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return fmt.Errorf("Could not start recording. Data recording already started")
	}
	current_time := time.Now()
	file_name := fmt.Sprintf("%s/%02d-%02d-%d-fluke.csv", s.outputDir, current_time.Day(), current_time.Month(), current_time.Year())
	file_name = utils.UniqueFileName(file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("Could not create file %v: %v", file, err)
	}
	s.filepath = file_name
	writer := csv.NewWriter(file)
	// headers
	headerData := []string{"Timestamp"}
	idxs := make([]int, 0, len(s.tagMap))
	for idx := range s.tagMap {
		idxs = append(idxs, idx)
	}
	sort.Ints(idxs)
	for _, i := range idxs {
		if i != 0 {
			headerData = append(headerData, s.tagMap[i].name)
		}
	}
	err = writer.Write(headerData)
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// the actual data
	s.Logger.Info().Msg("Starting data recording...")
	go func() {
		for {
			select {
			case <-ticker.C:
				err = s.record(writer, idxs)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not write to %s: %v", file_name, err))
				}
			case <-s.CancelChan:
				ticker.Stop()
				file.Close()
				writer.Flush()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			case <-s.QuitChan:
				ticker.Stop()
				file.Close()
				writer.Flush()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

// record records the live data from Fluke and inserts it into a csv file and passes it to the RTD service
func (s *FlukeService) record(writer *csv.Writer, idxs []int) error {
	current_time := time.Now()
	current_time_str := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d", current_time.Day(), current_time.Month(), current_time.Year(), current_time.Hour(), current_time.Minute(), current_time.Second())
	dataString := []string{current_time_str}
	dataField := make(map[int64]*fmtrpc.DataField)
	for _, i := range idxs {
		reading := s.connection.ReadItem(s.tagMap[i].tag)
		if i != 0 {
			if atomic.LoadInt32(&s.Broadcasting) == 1 {
				switch v := reading.Value.(type) {
				case float64:
					dataField[int64(i)]= &fmtrpc.DataField{
						Name: s.tagMap[i].name,
						Value: v,
					}
				case float32:
					dataField[int64(i)]= &fmtrpc.DataField{
						Name: s.tagMap[i].name,
						Value: float64(v),
					}
				}
			}
			dataString = append(dataString, fmt.Sprintf("%g", reading.Value))
			if s.IsRGAReady {
				if i == pressureRead {
					s.PressureChan <- reading.Value.(float64)
				}
			}
		}
	}
	err := writer.Write(dataString)
	if err != nil {
		return err
	}
	if atomic.LoadInt32(&s.Broadcasting) == 1 {
		dataFrame := &fmtrpc.RealTimeData{
			Source: s.name,
			IsScanning: true,
			Timestamp: current_time.UnixMilli(),
			Data: dataField,
		}
		s.BuffedChan <- dataFrame // may need to go into a goroutine
	}
	return nil
}

func (s *FlukeService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return fmt.Errorf("Could not stop data recording. Data recording already stopped.")
	}
	s.CancelChan<-struct{}{}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *FlukeService) ListenForRTDSignal() {
	for {
		select {
		case msg := <-s.StateChangeChan:
			switch msg.Type {
			case data.BROADCASTING:
				if msg.State {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 0, 1); !ok {
						s.Logger.Warn().Msg("Could not start broadcasting to RTD Service: Already broadcasting")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: false, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				} else {
					if ok := atomic.CompareAndSwapInt32(&s.Broadcasting, 1, 0); !ok {
						s.Logger.Warn().Msg("Could not stop broadcasting to RTD Service: Already stopped broadcasting")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.BROADCASTING, State: true, ErrMsg: fmt.Errorf("Could not change broadcasting state.")}
					} else {
						s.StateChangeChan <- &data.StateChangeMsg{Type: msg.Type, State: msg.State, ErrMsg: nil}
					}
				}
			case data.RECORDING:
				if msg.State {
					err := s.startRecording(DefaultPollingInterval) // TODO:SSSOCPaulCote - RecordingState data will include polling interval
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Started recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil, Msg: s.filepath}
					}
				} else {
					s.Logger.Info().Msg("Stopping data recording...")
					err := s.stopRecording()
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: fmt.Errorf("Could not stop recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Stopped recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: nil}
					}
				}
			}
		case <-s.QuitChan:
			return
		}
	}
}

// RegisterWithRTDService adds the RTD Service channels to the Fluke Service Struct and incrememnts the number of registered data providers on the RTD
func (s *FlukeService) RegisterWithRTDService(rtd *data.RTDService) {
	rtd.RegisterDataProvider(s.name)
	s.BuffedChan = rtd.DataProviderChan
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}