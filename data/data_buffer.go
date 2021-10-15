package data

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	STRING_FIELD = 1
	FLOAT64_FIELD = 2
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

type DataSubscription struct {
	DataStream	chan *DataFrame
	CancelChan	chan struct{}
}

type DataBuffer struct {
	Running					int32 //used atomically
	IncomingDataChannels	map[string]chan []byte
	OutgoingDataChannels	map[string]chan *DataFrame
	Quit					chan struct{}

}

//NewDataBuffer returns an instantiated DataBuffer struct
func NewDataBuffer() *DataBuffer {
	return &DataBuffer{
		IncomingDataChannels: make(map[string]chan []byte),
		OutgoingDataChannels: make(map[string]chan *DataFrame),
		Quit: make(chan struct{}),
	}
}

// Start starts the Data Buffer service
func (b *DataBuffer) Start() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 0, 1); !ok {
		return fmt.Errorf("Could not start Data Buffer Service.")
	}
	return nil
}

// Stop stops the Data buffer service and closes all channels
func (b *DataBuffer) Stop() error {
	if ok := atomic.CompareAndSwapInt32(&b.Running, 1, 0); !ok {
		return fmt.Errorf("Could not stop Data Buffer Service.")
	}
	for _, inChan := range b.IncomingDataChannels {
		close(inChan)
	}
	for _, outChan := range b.OutgoingDataChannels {
		close(outChan)
	}
	close(b.Quit)
	return nil
}

// RegisterDataProvider creates a new Incoming data channel, appends it to the map of incoming data channels and also gives the channel to the data provider. Additionally, it creates a new outgoing channel and appends that to the map of outgoing channels
func (b *DataBuffer) RegisterDataProvider(dataProvider *DataProvider) error {
	if _, ok := b.IncomingDataChannels[dataProvider.ServiceName()]; ok {
		return fmt.Errorf("Data provider already registered with Data Buffer.")
	}
	newIncomingChan := make(chan []byte)
	newOutgoingChan := make(chan *DataFrame)
	dataProvider.RegisterWithBufferService(newIncomingChan)
	b.IncomingDataChannels[dataProvider.ServiceName()] = newOutgoingChan
}

// SubscribeSingleStream returns the outgoing channel for the specified service name
func (b *DataBuffer) SubscribeSingleStream(name string) (chan *DataFrame, error) {
	if outChan, ok := b.OutgoingDataChannels[name]; !ok {
		nil, return fmt.Errorf("No such channel exists: %s", name)
	}
	return b.OutgoingDataChannels[name], nil
}

func (b *DataBuffer) buffer() {
	var wg sync.WaitGroup
	for name, inChan := range b.IncomingDataChannels {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.OutgoingDataChannels[name] <- Decode(inChan)
		}
	}
}