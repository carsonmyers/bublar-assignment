package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/oklog/ulid"
	"go.uber.org/zap"
)

// APIClient extended http client for building a portal API
type APIClient struct {
	name    string
	logger  *zap.Logger
	entropy io.Reader
	config  *configure.APIConfig
}

var apiClient *APIClient

// API - get a portal provider client
func API() *APIClient {
	if apiClient != nil {
		return apiClient
	}

	config := configure.GetAPI()

	apiClient = &APIClient{
		config:  config,
		logger:  logger.GetLogger(),
		entropy: ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0),
	}

	return apiClient
}

// URL Build a full URL for a request
func (c *APIClient) URL(endpoint string) string {
	if len(endpoint) == 0 {
		log.Fatal("Empty endpoint")
	}

	return fmt.Sprintf("%s://%s:%d%s%s", c.config.Protocol, c.config.Host, c.config.Port, c.config.BasePath, endpoint)
}

// Request single request for a given client
type Request struct {
	Header   http.Header
	client   *http.Client
	name     string
	request  *http.Request
	err      error
	bytes    int
	logger   *zap.Logger
	response *http.Response
}

// NewRequest - start a new request
func (c *APIClient) NewRequest(method, url string, body interface{}) (*Request, *zap.Logger) {
	rID := ulid.MustNew(ulid.Timestamp(time.Now()), c.entropy).String()
	logger := c.logger.With(zap.String("requestID", rID))

	r := &Request{
		client: &http.Client{},
		name:   c.config.Name,
		logger: logger,
	}

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			r.err = err
			return r, logger
		}

		r.bytes = len(data)
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		r.err = err
		return r, logger
	}

	req.Header.Add("Request-ID", rID)
	if len(c.config.Session) > 0 {
		req.AddCookie(&http.Cookie{
			Name:  "AUTH",
			Value: c.config.Session,
		})
	}

	r.request = req
	r.Header = req.Header

	return r, logger
}

// Do - execute a request
func (r *Request) Do() (*http.Response, string, error) {
	if r.err != nil {
		r.logger.Error("Request error", zap.Error(r.err))
		return nil, "", r.err
	}

	r.logRequest(r.request.Method, r.request.URL.String(), r.bytes)

	start := time.Now()
	res, err := r.client.Do(r.request)
	d := time.Now().Sub(start)
	if err == nil {
		r.logResponse(res, d)
	} else {
		r.logResponseError(err, d)
		return nil, "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.logger.Error("Failed to read response body", zap.Error(err))
		return res, "", err
	}

	if len(data) != 0 {
		var pretty bytes.Buffer
		err = json.Indent(&pretty, data, "", "\t")
		if err != nil {
			r.logger.Error("Failed to indent response body", zap.String("data", string(data)), zap.Error(err))
			return res, string(data), err
		}

		return res, pretty.String(), nil
	}

	return res, "", nil
}

func (r *Request) logRequest(method, url string, bytes int) {
	r.logger.Info(fmt.Sprintf("<-- %s", r.name), zap.String("method", method), zap.String("url", url), zap.Int("bytes", bytes))
}

func (r *Request) logResponse(res *http.Response, d time.Duration) {
	r.logger.Info(fmt.Sprintf("--> %s %s", r.name, res.Status), zap.Int("status", res.StatusCode), zap.Int64("bytes", res.ContentLength), zap.Duration("duration", d))
}

func (r *Request) logResponseError(err error, d time.Duration) {
	r.logger.Error("Request failed", zap.Error(err), zap.Duration("duration", d))
}
