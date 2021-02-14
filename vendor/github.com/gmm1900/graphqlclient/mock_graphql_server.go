package graphqlclient

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
)

// MockGraphqlServer would start a real httpserver,
// that records requests in sequence and returns given mocked responses in sequence
type MockGraphqlServer struct {
	CapturedReqHeaders []http.Header // the request's header that mocked server receives
	CapturedReqBody []map[string]interface{} // the request (json unmarshalled) that mocked server receives
	MockedRespBody [][]byte // the response (before json marshalling) that mocked server should return upon receiving request
	URL string // URL to reach the server
	Close func() error // for shutting down
	idx int // the idx to the next response to return
}

func (s *MockGraphqlServer) Start(t *testing.T) {
	var wg sync.WaitGroup
	defer wg.Wait() // for waiting on mockSvr.Serve(ln), which runs on go routine, to complete

	mockSvr := http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// capture req headers
			s.CapturedReqHeaders = append(s.CapturedReqHeaders, r.Header)

			// capture req body
			reqBodyBytes, err := ioutil.ReadAll(r.Body)
			assert.Nil(t, err)
			var unmarshalledBody map[string]interface{}
			err = json.Unmarshal(reqBodyBytes, &unmarshalledBody)
			assert.Nil(t, err)
			s.CapturedReqBody = append(s.CapturedReqBody, unmarshalledBody)

			// write mocked response
			_, err = w.Write(s.MockedRespBody[s.idx])
			assert.Nil(t, err)

			s.idx++
		}),
	}

	ln, err := net.Listen("tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	s.URL = "http://" + ln.Addr().String()
	go func() {
		wg.Add(1)
		mockSvr.Serve(ln)
		wg.Done()
	}()

	s.Close = mockSvr.Close
}