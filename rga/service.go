//go:build !demo
// +build !demo

/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package rga

import (
	"context"
	"crypto/tls"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/drivers"
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/utils"
	"github.com/SSSOCPaulCote/gux"
	influx "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	e "github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type RGAService struct {
	BaseRGAService
	connection *drivers.RGAConnection
}

// A compile time check to make sure that RGAService fully implements the data.Service interface
var _ data.Service = (*RGAService)(nil)

// NewRGAService creates an instance of the RGAService struct. It also establishes a connection to the RGA device
func NewRGAService(
	logger *zerolog.Logger,
	rtdStore *gux.Store,
	_ *gux.Store,
	connection *drivers.RGAConnection,
	influxUrl string,
	influxToken string,
) *RGAService {
	var (
		wgL sync.WaitGroup
		wgR sync.WaitGroup
	)
	client := influx.NewClientWithOptions(influxUrl, influxToken, influx.DefaultOptions().SetTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	return &RGAService{
		BaseRGAService: BaseRGAService{
			rtdStateStore: rtdStore,
			Logger:        logger,
			QuitChan:      make(chan struct{}),
			CancelChan:    make(chan struct{}),
			name:          data.RgaName,
			wgListen:      wgL,
			wgRecord:      wgR,
			idb:           client,
		},
		connection: connection,
	}
}

// Start starts the RGA service. It does NOT start the data recording process
func (s *RGAService) Start() error {
	s.Logger.Info().Msg("Starting RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 0, 1); !ok {
		return errors.ErrServiceAlreadyStarted
	}
	s.wgListen.Add(1)
	go s.ListenForRTDSignal()
	s.Logger.Info().Msg("RGA Service successfully started.")
	return nil
}

// Stop stops the RGA service.
func (s *RGAService) Stop() error {
	s.Logger.Info().Msg("Stopping RGA Service...")
	if ok := atomic.CompareAndSwapInt32(&s.Running, 1, 0); !ok {
		return errors.ErrServiceAlreadyStopped
	}
	var stoppedRec bool
	if atomic.LoadInt32(&s.Recording) == 1 {
		err := s.stopRecording()
		if err != nil {
			return e.Wrap(err, "could not stop RGA service")
		}
		stoppedRec = true
	}
	close(s.CancelChan)
	if stoppedRec {
		s.wgRecord.Wait()
	}
	close(s.QuitChan)
	s.wgListen.Wait()
	s.connection.Close()
	s.Logger.Info().Msg("RGA Service successfully stopped.")
	return nil
}

// Name satisfies the fmtd.Service interface
func (s *RGAService) Name() string {
	return s.name
}

// startRecording starts data recording from the RGA device
func (s *RGAService) startRecording(pol_int int64, orgName string) error {
	if atomic.LoadInt32(&s.Recording) == 1 {
		return errors.ErrAlreadyRecording
	}
	if s.currentPressure == 0 || s.currentPressure > drivers.RGAMinimumPressure {
		return ErrPressureTooHigh
	}
	if pol_int < drivers.RGAMinPollingInterval && pol_int != 0 {
		return errors.ErrPollingIntervalTooSmall
	} else if pol_int == 0 { //No polling interval provided
		pol_int = drivers.RGAMinPollingInterval
	}
	// InitMsg
	err := s.connection.InitMsg()
	if err != nil {
		return e.Wrap(err, "unable to communicate with RGA")
	}
	// Setup Data
	_, err = s.connection.Control(utils.AppName, utils.AppVersion)
	if err != nil {
		return err
	}
	resp, err := s.connection.SensorState()
	if err != nil {
		return err
	}
	if resp.Fields["State"].Value.(string) != drivers.RGA_SENSOR_STATE_INUSE {
		return e.Wrap(resp.Fields["State"], "sensor not ready")
	}
	_, err = s.connection.AddBarchart("Bar1", 1, 200, drivers.RGA_PeakCenter, 5, 0, 0, 0)
	if err != nil {
		return e.Wrap(err, "could not add barchart")
	}
	_, err = s.connection.ScanAdd("Bar1")
	if err != nil {
		return e.Wrap(err, "could not add measurement to scan")
	}
	// Get bucket, create it if it doesn't exist
	orgAPI := s.idb.OrganizationsAPI()
	org, err := orgAPI.FindOrganizationByName(context.Background(), orgName)
	if err != nil {
		return err
	}
	bucketAPI := s.idb.BucketsAPI()
	buckets, err := bucketAPI.FindBucketsByOrgName(context.Background(), orgName)
	if err != nil {
		return err
	}
	var found bool
	for _, bucket := range *buckets {
		if bucket.Name == influxRGABucketName {
			found = true
			break
		}
	}
	if !found {
		_, err := bucketAPI.CreateBucketWithName(context.Background(), org, influxRGABucketName, domain.RetentionRule{EverySeconds: 0})
		if err != nil {
			return err
		}
	}
	writeAPI := s.idb.WriteAPI(orgName, influxRGABucketName)
	ticker := time.NewTicker(time.Duration(pol_int) * time.Second)
	// the actual data
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 0, 1); !ok {
		return errors.ErrAlreadyRecording
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

//record writes data from the RGA to a csv file and will pass it along it's Output channel
func (s *RGAService) record(writer api.WriteAPI) error {
	current_time := time.Now()
	dataField := make(map[int64]*fmtrpc.DataField)
	// Start scan
	_, err := s.connection.ScanResume(1)
	if err != nil {
		return e.Wrap(err, "could not resume scan")
	}
	for {
		resp, err := s.connection.ReadResponse()
		if err != nil {
			return e.Wrap(err, "could not read response")
		}
		if resp.ErrMsg.CommandName == drivers.MassReading {
			massPos := resp.Fields["MassPosition"].Value.(int64)
			dataField[massPos] = &fmtrpc.DataField{
				Name:  fmt.Sprintf("Mass %s", strconv.FormatInt(massPos, 10)),
				Value: resp.Fields["Value"].Value.(float64),
			}
			p := influx.NewPoint(
				"pressure",
				map[string]string{
					"mass": fmt.Sprintf("Mass %s", strconv.FormatInt(massPos, 10)),
				},
				map[string]interface{}{
					"pressure": resp.Fields["Value"].Value.(float64),
				},
				current_time,
			)
			// write asynchronously
			writer.WritePoint(p)
			if resp.Fields["MassPosition"].Value.(int64) == int64(200) {
				break
			}
		}
	}
	err = s.rtdStateStore.Dispatch(
		gux.Action{
			Type: "rga/update",
			Payload: fmtrpc.RealTimeData{
				Source:     s.name,
				IsScanning: true,
				Timestamp:  current_time.UnixMilli(),
				Data:       dataField,
			},
		},
	)
	if err != nil {
		return e.Wrap(err, "could not update state")
	}
	return nil
}

// stopRecording stops the data recording process
func (s *RGAService) stopRecording() error {
	if ok := atomic.CompareAndSwapInt32(&s.Recording, 1, 0); !ok {
		return errors.ErrAlreadyStoppedRecording
	}
	s.CancelChan <- struct{}{}
	_, err := s.connection.FilamentControl("Off")
	if err != nil {
		return e.Wrap(err, "could not safely turn off filament")
	}
	_, err = s.connection.Release()
	if err != nil {
		return e.Wrap(err, "could not safely release control of RGA")
	}
	return nil
}

//CheckIfBroadcasting listens for a signal from RTD service to either stop or start broadcasting data to it.
func (s *RGAService) ListenForRTDSignal() {
	defer s.wgListen.Done()
	signalChan, unsub := s.rtdStateStore.Subscribe(s.name)
	cleanUp := func() {
		unsub(s.rtdStateStore, s.name)
	}
	defer cleanUp()
	for {
		select {
		case msg := <-s.StateChangeChan:
			switch msg.Type {
			case data.RECORDING:
				if msg.State {
					err := s.startRecording(0, msg.Msg)
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not start recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: e.Wrap(err, "could not start recording")}
					} else {
						s.Logger.Info().Msg("Started recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: nil}
					}
				} else {
					s.Logger.Info().Msg("Stopping data recording...")
					err := s.stopRecording()
					if err != nil {
						s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: true, ErrMsg: e.Wrap(err, "could not stop recording")}
					} else {
						s.Logger.Info().Msg("Stopped recording.")
						s.StateChangeChan <- &data.StateChangeMsg{Type: data.RECORDING, State: false, ErrMsg: nil}
					}
				}
			}
		case <-signalChan:
			currentState := s.rtdStateStore.GetState()
			cState, ok := currentState.(data.InitialRtdState)
			if !ok {
				s.Logger.Error().Msg(fmt.Sprintf("Invalid type %v expected %v\nStopping recording...", reflect.TypeOf(currentState), reflect.TypeOf(data.InitialRtdState{})))
				err := s.stopRecording()
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
				}
			}
			if cState.RealTimeData.Data == nil {
				continue
			}
			s.currentPressure = cState.RealTimeData.Data[drivers.TelemetryPressureChannel].Value
			if s.currentPressure >= drivers.RGAMinimumPressure && atomic.LoadInt32(&s.Recording) == 1 {
				err := s.stopRecording()
				if err != nil {
					s.Logger.Error().Msg(fmt.Sprintf("Could not stop recording: %v", err))
				}
			}
		case <-s.QuitChan:
			return
		}
	}
}

// RegisterWithRTDService adds the RTD Service channels to the RGA Service Struct and incrememnts the number of registered data providers on the RTD
func (s *RGAService) RegisterWithRTDService(rtd *data.RTDService) {
	rtd.RegisterDataProvider(s.name)
	s.StateChangeChan = rtd.StateChangeChans[s.name]
}
