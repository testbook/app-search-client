package client

type Cache interface {
	Set(key []byte, value interface{}) (err error)
	Get(key []byte, value interface{}) (exists bool)
}
