package steamcommunity_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/suite"

	"alex-j-butler.com/steamcommunity"
)

type ClientTestSuite struct {
	suite.Suite
	Client              *steamcommunity.Client
	Server              *httptest.Server
	LastRequest         *http.Request
	LastRequestBody     string
	CurrentResponseFunc int
	ResponseFunc        []func(http.ResponseWriter, *http.Request)
}

type RewriteTransport struct {
	Transport http.RoundTripper
}

func (s *ClientTestSuite) SetupSuite() {
	s.Server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := ioutil.ReadAll(r.Body)
			s.LastRequestBody = string(body)
			s.LastRequest = r
			if s.ResponseFunc != nil {
				s.ResponseFunc[s.CurrentResponseFunc](w, r)
				if s.CurrentResponseFunc+1 < len(s.ResponseFunc) {
					s.CurrentResponseFunc++
				}
			}
		}),
	)
}

func (s *ClientTestSuite) TearDownSuite() {
	s.Server.Close()
}

func (s *ClientTestSuite) SetupTest() {
	s.ResponseFunc = nil
	s.CurrentResponseFunc = 0
	s.LastRequest = nil
}

func (t RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Test using HTTP instead of HTTPS.
	req.URL.Scheme = "http"

	return t.Transport.RoundTrip(req)
}
