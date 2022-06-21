package client

type Cache interface {
	Set(key []byte, value interface{})
	Get(key []byte, value interface{}) (exists bool)
}
