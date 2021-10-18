package data

type DataProvider interface {
	RegisterWithBufferService(*DataBuffer) error
	ServiceName() string
}