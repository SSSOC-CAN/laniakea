// +build !demo

package telemetry

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
	"github.com/rs/zerolog"
)

// TelemetryService is a struct for holding all relevant attributes to interfacing with the DAQ
type TelemetryService struct {
	BaseTelemetryService
	connection *drivers.DAQConnection
}

// A compile time check to make sure that TelemetryService fully implements the data.Service interface
var _ data.Service = (*TelemetryService) (nil)


// NewTelemetryService creates a new Telemetry Service object which will use the appropriate drivers
func NewTelemetryService(
	logger *zerolog.Logger,
	store *state.Store,
	_ *state.Store,
	connection *drivers.DAQConnection,
	influxUrl string,
	influxToken string,
) *TelemetryService {
	var (
		wgL sync.WaitGroup
		wgR sync.WaitGroup
	)
	client := influx.NewClientWithOptions(influxUrl, influxToken, influx.DefaultOptions().SetTLSConfig(&tls.Config{InsecureSkipVerify : true}))
	return &TelemetryService{
		BaseTelemetryService: BaseTelemetryService{
			rtdStateStore:    store,
			Logger:       	  logger,
			QuitChan:     	  make(chan struct{}),
			CancelChan:   	  make(chan struct{}),
			name: 	      	  data.TelemetryName,
			wgListen:		  wgL,
			wgRecord:		  wgR,
			idb: 			  client,
		},
		connection:		  connection,
	}
}

// Start starts the service. Returns an error if any issues occur
func (s *TelemetryService) Start() error {
	s.Logger.Info().Msg("Starting Telemetry Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start Telemetry service: service already started.")
	}
	s.wgListen.Add(1)
	go s.ListenForRTDSignal()
	s.Logger.Info().Msg("Telemetry Service started.")
	return nil
}

// Stop stops the service. Returns an error if any issues occur
func (s *TelemetryService) Stop() error {
	s.Logger.Info().Msg("Stopping Telemetry Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop Telemetry service: service already stopped.")
	}
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return fmt.Errorf("Could not stop Telemetry service: %v", err)
		}
	}
	close(s.CancelChan)
	s.wgRecord.Wait()
	close(s.QuitChan)
	s.wgListen.Wait()
	s.connection.Close()
	s.Logger.Info().Msg("Telemetry Service stopped.")
	return nil
}

// StartRecording starts the recording process by creating a csv file and inserting the header row into the file and returns a quit channel and error message
func (s *TelemetryService) startRecording(pol_int int64, orgName, bucketName string) error {
	if atomic.LoadInt32(&s.Recording) == 1 {
		return ErrAlreadyRecording
	}
	if pol_int < drivers.MinTelemetryPollingInterval && pol_int != 0 {
		return fmt.Errorf("Inputted polling interval smaller than minimum value: %v", drivers.MinTelemetryPollingInterval)
	} else if pol_int == 0 { //No polling interval provided
		pol_int = drivers.TelemetryDefaultPollingInterval
	}
	// start connection
	err = s.connection.StartScanning()
	if err != nil {
		return err
	}
	// Get bucket, create it if it doesn't exist
	orgAPI := s.idb.OrganizationsAPI()
	org, err := orgAPI.FindOrganizationByName(context.Background(), orgName)
	if err != nil {
		return err
	}
	bucketAPI := s.idb.BucketsAPI()
	bucket, _ := bucketAPI.FindBucketByName(context.Background(), bucketName)
	if bucket == nil {
		_, err := bucketAPI.CreateBucketWithName(context.Background(), org, bucketName, domain.RetentionRule{EverySeconds: 0})
		if err != nil {
			return err
		}
	}
	writeAPI := s.idb.WriteAPI(orgName, bucketName)
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// Write polling interval to state
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "telemetry/polling_interval/update",
			Payload: pol_int,
		},
	)
	if err != nil {
		return fmt.Errorf("Could not update state: %v", err)
	}
	// the actual data
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return ErrAlreadyRecording
	}
	s.Logger.Info().Msg("Starting data recording...")
	s.wgRecord.Add(1)
	go func() {
		defer s.wgRecord.Done()
		for {
			select {
			case <-ticker.C:
				err = s.record(writeAPI)
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not write to influxdb: %v", err))
				}
			case <-s.CancelChan:
				ticker.Stop()
				writeAPI.Flush()
				s.idb.Close()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			case <-s.QuitChan:
				ticker.Stop()
				writeAPI.Flush()
				s.idb.Close()
				s.Logger.Info().Msg("Data recording stopped.")
				return
			}
		}
	}()
	return nil
}

// record records the live data from Telemetry and inserts it into a csv file and passes it to the RTD service
func (s *TelemetryService) record(writer api.WriteAPI) error {
	current_time := time.Now()
	dataField := make(map[int64]*fmtrpc.DataField)
	readings := s.connection.ReadItems()
	var (
		cumSum float64
		cnt	   int64
	)
	for i, reading := range readings {
		switch v := reading.Item.Value.(type) {
		case float64:
			dataField[int64(i)]= &fmtrpc.DataField{
				Name: reading.Name,
				Value: v,
			}
			if i != int(drivers.TelemetryPressureChannel) {
				cumSum += v
				cnt += 1
				p := influx.NewPoint(
					"temperature",
					map[string]string{
						"id": reading.Name,
					},
					map[string]interface{}{
						"temperature": v,
					},
					current_time,
				)
				// write asynchronously
				writer.WritePoint(p)
			} else {
				p := influx.NewPoint(
					"pressure",
					map[string]string{
						"id": reading.Name,
					},
					map[string]interface{}{
						"pressure": v,
					},
					current_time,
				)
				// write asynchronously
				writer.WritePoint(p)
			}
		case float32:
			dataField[int64(i)]= &fmtrpc.DataField{
				Name: reading.Name,
				Value: float64(v),
			}
			if i != int(drivers.TelemetryPressureChannel) {
				cumSum += float64(v)
				cnt += 1
				p := influx.NewPoint(
					"temperature",
					map[string]string{
						"id": reading.Name,
					},
					map[string]interface{}{
						"temperature": float64(v),
					},
					current_time,
				)
				// write asynchronously
				writer.WritePoint(p)
			} else {
				p := influx.NewPoint(
					"pressure",
					map[string]string{
						"id": reading.Name,
					},
					map[string]interface{}{
						"pressure": float64(v),
					},
					current_time,
				)
				// write asynchronously
				writer.WritePoint(p)
			}
		}

	}
	err := s.rtdStateStore.Dispatch(
		state.Action{
			Type: 	 "telemetry/update",
			Payload: data.InitialRtdState{
				RealTimeData: fmtrpc.RealTimeData{
					Source: s.name,
					IsScanning: true,
					Timestamp: current_time.UnixMilli(),
					Data: dataField,
				},
				AverageTemperature: cumSum/float64(cnt),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("Could not update state: %v", err)
	}
	return nil
}

// stopRecording sends an empty struct down the CancelChan to innitiate the stop recording process
func (s *TelemetryService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return ErrAlreadyStoppedRecording
	}
	s.CancelChan<-struct{}{}
	// Stop scanning
	err := s.connection.StopScanning()
	if err != nil {
		return err
	}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *TelemetryService) ListenForRTDSignal() {
	defer s.wgListen.Done()
	for {
		select {
		case msg := <-s.StateChangeChan:
			switch msg.Type {
			case data.RECORDING:
				if msg.State {
					split := strings.Split(msg.Msg, ":")
					n, err := strconv.ParseInt(split[0], 10, 64)
					if err != nil {
						n = drivers.TelemetryDefaultPollingInterval
					}
					err = s.startRecording(n, split[1], split[2])
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: fmt.Errorf("Could not start recording: %v", err)}
					} else {
						s.Logger.Info().Msg("Started recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil}
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

// RegisterWithRTDService adds the RTD Service channels to the Telemetry Service Struct and incrememnts the number of registered data providers on the RTD
func (s *TelemetryService) RegisterWithRTDService(rtd *data.RTDService) {
	rtd.RegisterDataProvider(s.name)
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}