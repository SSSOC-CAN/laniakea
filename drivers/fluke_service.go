package drivers

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/konimarti/opc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Tag struct {
	name string
	tag  string
}

var (
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
	fmtrpc.UnimplementedDataCollectorServer
	Active     int32 // atomic
	Stopping   int32 // atomic
	Recording  int32 // atomic
	Logger     *zerolog.Logger
	tags       []string
	tagMap     map[int]Tag
	connection opc.Connection
	QuitChan   chan struct{}
	outputDir  string
}

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
		outputDir: outputDir,
	}, nil
}

// RegisterWithGrpcServer registers the gRPC server to the unlocker service
func (s *FlukeService) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	fmtrpc.RegisterDataCollectorServer(grpcServer, s)
	return nil
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
	s.Logger.Info().Msg("Fluke Service started.")
	return nil
}

// Stop stops the service. Returns an error if any issues occur
func (s *FlukeService) Stop() error {
	s.Logger.Info().Msg("Stopping Fluke Service...")
	atomic.StoreInt32(&s.Stopping, 1)
	close(s.QuitChan)
	// Stop scanning
	err := s.connection.Write(s.tagMap[0].tag, false)
	if err != nil {
		return err
	}
	s.connection.Close()
	s.Logger.Info().Msg("Fluke Service stopped.")
	return nil
}

// StartRecording starts the recording process by creating a csv file and inserting the header row into the file and returns a quit channel and error message
func (s *FlukeService) startRecording(pol_int int64) error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return fmt.Errorf("Could not start recording. Data recording already started")
	}
	if pol_int < minPollingInterval && pol_int != 0 {
		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", minPollingInterval)
	} else if pol_int == 0 { //No polling interval provided
		pol_int = DefaultPollingInterval
	}
	current_time := time.Now()
	file_name := fmt.Sprintf("%s/%02d-%02d-%d-fluke.csv", s.outputDir, current_time.Day(), current_time.Month(), current_time.Year())
	file_name = utils.UniqueFileName(file_name)
	file, err := os.Create(file_name)
	if err != nil {
		return fmt.Errorf("Could not create file %s: %v", file, err)
	}
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

// record records the live data from Fluke and inserts it into a csv file
func (s *FlukeService) record(writer *csv.Writer, idxs []int) error {
	current_time := time.Now()
	data := []string{fmt.Sprintf("%02d:%02d:%02d", current_time.Hour(), current_time.Minute(), current_time.Second())}
	for _, i := range idxs {
		if i != 0 {
			data = append(data, fmt.Sprintf("%g", s.connection.ReadItem(s.tagMap[i].tag).Value))
		}
	}
	err := writer.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// StartRecording is called by gRPC client and CLI to begin the data recording process with Fluke
func (s *FlukeService) StartRecording(ctx context.Context, req *fmtrpc.RecordRequest) (*fmtrpc.RecordResponse, error) {
	err := s.startRecording(req.PollingInterval)
	if err != nil {
		return &fmtrpc.RecordResponse{
			Msg: fmt.Sprintf("Could not start data recording: %v", err),
		}, err
	}
	return &fmtrpc.RecordResponse{
		Msg: "Data recording successfully started.",
	}, nil
}

// StopRecording is called by gRPC client and CLI to end data recording process
func (s *FlukeService) StopRecording(ctx context.Context, req *fmtrpc.StopRecRequest) (*fmtrpc.StopRecResponse, error) {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return &fmtrpc.StopRecResponse{
			Msg: "Could not stop data recording. Data recording already stopped.",
		}, fmt.Errorf("Could not stop data recording. Data recording already stopped.")
	}
	go func() {
		s.QuitChan <- struct{}{}
	}()
	return &fmtrpc.StopRecResponse{
		Msg: "Data recording successfully stopped.",
	}, nil
}
