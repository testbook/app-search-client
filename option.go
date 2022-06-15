package client

// Option represents a modification to the default behavior of the Client
type Option func(*client)

// WithCache sets the cache to use for the client
func WithCache(cache Cache) Option {
	return func(c *client) {
		c.cache = cache
	}
}
