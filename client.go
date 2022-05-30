package client

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type client struct {
	url        url.URL
	apiKey     string
	useragent  string
	isVerbose  bool
	httpClient *http.Client
	transport  *http.Transport
}

type HTTPConfig struct {
	// Addr should be of the form "http://host:port"
	// or "http://[ipv6-host%zone]:port".
	Addr string

	// APIKey is the required token used to authorize API calls.
	APIKey string

	// UserAgent is the http User Agent, defaults to "InfluxDBClient".
	UserAgent string

	// Log all the query requests and responses
	IsVerbose bool

	// Timeout for influxdb writes, defaults to no timeout.
	Timeout time.Duration
}

// Client is a client interface for writing & querying the database.
type Client interface {

	// Create or update documents (missing fields can be auto added
	Index(string, []interface{}) error

	// Query documents with search query service
	Search(string, *SearchQueryService) (*SearchQueryResponse, error)

	// Close releases any resources a Client may be using.
	Close() error
}

func NewHTTPClient(conf HTTPConfig) (Client, error) {
	if conf.UserAgent == "" {
		conf.UserAgent = "AppSearchClient"
	}

	u, err := url.Parse(conf.Addr)
	if err != nil {
		return nil, err
	} else if u.Scheme != "http" && u.Scheme != "https" {
		m := fmt.Sprintf("Unsupported protocol scheme: %s, your address"+
			" must start with http:// or https://", u.Scheme)
		return nil, errors.New(m)
	}

	tr := &http.Transport{}
	c := &client{
		url:       *u,
		apiKey:    conf.APIKey,
		useragent: conf.UserAgent,
		isVerbose: conf.IsVerbose,
		httpClient: &http.Client{
			Timeout:   conf.Timeout,
			Transport: tr,
		},
		transport: tr,
	}
	return c, nil
}

func (c *client) SetBearerAuth(r *http.Request) {
	r.Header.Set("Authorization", "Bearer "+c.apiKey)
}

// Upsert documents
// <https://www.elastic.co/guide/en/app-search/master/documents.html#documents-create>
func (c *client) Index(engine string, d []interface{}) error {
	u := c.url
	u.Path = path.Join(u.Path, "api/as/v1/engines", engine, "documents")

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.useragent)

	if c.apiKey != "" {
		c.SetBearerAuth(req)
	}

	r, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	var bir BulkIndexResponse
	err = json.NewDecoder(r.Body).Decode(&bir)
	if err != nil {
		return err
	}

	responseErr := bir.Errors()
	if r.StatusCode != http.StatusNoContent && r.StatusCode != http.StatusOK || len(responseErr) > 0 {
		var err = fmt.Errorf("%+v", bir)
		return err
	}

	return nil
}

// Query documents with filters
// <https://www.elastic.co/guide/en/app-search/current/filters.html#filters-nesting-filters>
func (c *client) Search(engine string, sqs *SearchQueryService) (sqr *SearchQueryResponse, err error) {
	u := c.url
	u.Path = path.Join(u.Path, "api/as/v1/engines", engine, "search")

	src, err := sqs.Source()
	if err != nil {
		return
	}

	b, err := json.Marshal(src)
	if err != nil {
		return
	}
	if c.isVerbose {
		log.Println("query", string(b))
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.useragent)

	if c.apiKey != "" {
		c.SetBearerAuth(req)
	}

	r, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&sqr)
	if err != nil {
		return
	}
	if c.isVerbose {
		marshBytes, _ := json.Marshal(sqr)
		log.Println("response", string(marshBytes))
	}

	if r.StatusCode != http.StatusNoContent && r.StatusCode != http.StatusOK || len(sqr.Errors) > 0 || len(sqr.Meta.Warnings) > 0 {
		err = fmt.Errorf("%+v", sqr)
	}
	return
}

// Close releases the client's resources.
func (c *client) Close() error {
	c.transport.CloseIdleConnections()
	return nil
}
