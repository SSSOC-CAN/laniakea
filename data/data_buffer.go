package data

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"github.com/rs/zerolog"
)

var (
	STRING_FIELD byte = 1
	FLOAT64_FIELD byte = 2
	defaultBufferingPeriod int = 3
)

type DataField struct {
	Name	string
	Value	float64
}

type DataFrame struct {
	IsScanning	bool
	Timestamp	string
	Data		map[int64]DataField
}

type DataProviderInfo struct {
	IncomingChan	chan []byte
	OutgoingChan	chan *DataFrame
}

type DataBuffer struct {
	Running					int32 //used atomically
	Logger					*zerolog.Logger
	DataProviders			map[string]DataProviderInfo
	Quit					chan struct{}
	NewProvider				chan string
	wg						sync.WaitGroup
}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewDataBuffer(log *zerolog.Logger) *DataBuffer {
	var (
		wg sync.WaitGroup
	)
	return &DataBuffer{
		DataProviders: make(map[string]DataProviderInfo),
		Quit: make(chan struct{}),
		NewProvider: make(chan string),
		Logger: log,
		wg: wg,
	}
}

// Start starts the Data Buffer service
func (b *DataBuffer) Start(ctx context.Context) error {
	b.Logger.Info().Msg("Starting Data Buffering service...")
	if ok := atomic.CompareAndSwapInt32(&b.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start Data Buffer Service.")
	}
	go b.buffer(ctx, defaultBufferingPeriod)
	b.Logger.Info().Msg("Data Buffering service successfully started.")
	return nil
}

// Stop stops the Data buffer service and closes all channels
func (b *DataBuffer) Stop() error {
	b.Logger.Info().Msg("Stopping Data Buffering service...")
	if ok := atomic.CompareAndSwapInt32(&b.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop Data Buffer Service.")
	}
	for _, p := range b.DataProviders {
		close(p.IncomingChan)
		close(p.OutgoingChan)
	}
	close(b.Quit)
	b.wg.Wait()
	b.Logger.Info().Msg("Data Buffering service stopped.")
	return nil
}

// SubscribeSingleStream returns the outgoing channel for the specified service name
func (b *DataBuffer) SubscribeSingleStream(name string) (chan *DataFrame, error) {
	if _, ok := b.DataProviders[name]; !ok {
		return nil, fmt.Errorf("No such data provider registered with data buffer service: %s", name)
	}
	return b.DataProviders[name].OutgoingChan, nil
}

// buffer creates goroutines for each incoming, outgoing data pair and decodes the incoming bytes into outgoing DataFrames
func (b *DataBuffer) buffer(ctx context.Context, bufferPeriod int) {
	for {
		select {
		case newProvider := <- b.NewProvider:
			b.Logger.Info().Msg(fmt.Sprintf("Incoming registration request: %s", newProvider))
			if _, ok := b.DataProviders[newProvider]; !ok {
				b.Logger.Warn().Msg(fmt.Sprintf("%s not registered with Data Buffer. Cannot start listening.", newProvider))
			} else {
				b.wg.Add(1)
				p := b.DataProviders[newProvider]
				go func(prov string, in chan []byte, out chan *DataFrame) {
					defer b.wg.Done()
					var (
						buf []*DataFrame
					)
					b.Logger.Info().Msg(fmt.Sprintf("Waiting for raw data from: %s...", prov))
					for {
						select {
						case rawData := <-in:
							b.Logger.Info().Msg(fmt.Sprintf("Received raw data from: %s", prov))
							tmp, err := Decode(rawData)
							if err == nil {
								buf = append(buf, tmp)
								if len(buf) < bufferPeriod {
									b.Logger.Info().Msg("Sending decoded data out.")
									out <- buf[0]
									buf = buf[1:] //pop
							} else {
								b.Logger.Warn().Msg(fmt.Sprintf("Could not decode %v: %v", rawData, err))
							}
						case <-ctx.Done():
							return
						case <- b.Quit:
							return
						}
					}
					b.Logger.Debug().Msg("Exited for loop")
				}(newProvider, p.IncomingChan, p.OutgoingChan)
			}
		case <-ctx.Done():
			return
		case <- b.Quit:
			return
		}
	}
}
