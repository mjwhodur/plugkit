package client

type RawClientImpl interface {
	Handle(payload []byte)
}
type RawClient struct {
}
