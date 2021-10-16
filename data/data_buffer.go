package data

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
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
	PollingInterval	int64
	IncomingChan	chan []byte
	OutgoingChan	chan *DataFrame
}

type DataBuffer struct {
	Running					int32 //used atomically
	DataProviders			map[string]DataProviderInfo
	Quit					chan struct{}

}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewDataBuffer() *DataBuffer {
	return &DataBuffer{
		DataProviders: make(map[string]DataProviderInfo),
		Quit: make(chan struct{}),
	}
}

// Start starts the Data Buffer service
func (b *DataBuffer) Start() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start Data Buffer Service.")
	}
	b.buffer(defaultBufferingPeriod)
	return nil
}

// Stop stops the Data buffer service and closes all channels
func (b *DataBuffer) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop Data Buffer Service.")
	}
	for _, p := range b.DataProviders {
		close(p.IncomingChan)
		close(p.OutgoingChan)
	}
	close(b.Quit)
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
func (b *DataBuffer) buffer(bufferPeriod int) {
	var wg sync.WaitGroup
	buf := make([]*DataFrame, bufferPeriod)
	for _, p := range b.DataProviders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < bufferPeriod; i++ {
				select {
				case rawData := <-p.IncomingChan:
					tmp, err := Decode(rawData)
					if err != nil {
						return // not sure what should happen here
					}
					buf[i] = tmp
				case <-b.Quit:
					return
				}
			}
			for i := 0; i < bufferPeriod; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					current_time := time.Now()
					poleTime := current_time.Add(time.Duration(p.PollingInterval) * time.Second)
					p.OutgoingChan <- buf[i]
					ct2 := time.Now()
					if ct2.Before(poleTime) {
						time.Sleep(poleTime.Sub(ct2))
					}
				}()	
			}
		}()
	}
	wg.Wait()
}