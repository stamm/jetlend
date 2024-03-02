package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stamm/jetlend/pkg"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JetlendTestSuite struct {
	suite.Suite
}

func (suite *JetlendTestSuite) TestUrls() {
	r := require.New(suite.T())
	j := NewJetlend("", 1*time.Second)
	r.Equal("https://jetlend.ru/invest/api/handler", j.url("handler"))
	r.Equal("https://jetlend.ru/invest/api/portfolio/charts/expected_revenue?size=5", j.urlExpect(5))
}

func (suite *JetlendTestSuite) TestGetJson_WrongUrl() {
	r := require.New(suite.T())
	j := &Jetlend{Timeout: 5 * time.Second}
	ctx := context.Background()
	b, err := j.getJSON(ctx, http.DefaultClient, "http://foo.com/?foo\nbar", "sid")
	r.ErrorContains(err, "net/url: invalid control character in URL")
	r.EqualValues("", b)
}

func (suite *JetlendTestSuite) TestGetJson_EmptyUrl() {
	r := require.New(suite.T())
	j := &Jetlend{URL: "", Timeout: 5 * time.Second}
	ctx := context.Background()
	b, err := j.getJSON(ctx, http.DefaultClient, "", "sid")
	r.ErrorContains(err, "unsupported protocol scheme")
	r.EqualValues("", b)
}

func (suite *JetlendTestSuite) TestGetJson_WrongCode() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusBadGateway)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	r := require.New(suite.T())
	j := &Jetlend{URL: testServer.URL, Timeout: 5 * time.Second}
	ctx := context.Background()
	b, err := j.getJSON(ctx, http.DefaultClient, testServer.URL, "sid")
	r.ErrorContains(err, "response code 502: body")
	r.EqualValues("", b)
}

func (suite *JetlendTestSuite) TestGetJson_OK() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	r := require.New(suite.T())
	j := &Jetlend{URL: testServer.URL, Timeout: 5 * time.Second}
	ctx := context.Background()
	b, err := j.getJSON(ctx, http.DefaultClient, testServer.URL, "sid")
	r.NoError(err)
	r.EqualValues("body", b)
}

func (suite *JetlendTestSuite) TestExpectAmount_EmptySids() {
	r := require.New(suite.T())
	j := NewJetlend("", 5*time.Second)
	ctx := context.Background()
	b, err := j.ExpectAmount(ctx, pkg.Config{}, 5)
	r.ErrorContains(err, "no config")
	r.EqualValues("", b)
}
func (suite *JetlendTestSuite) TestExpectAmount_Error() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.RequestURI != "/portfolio/charts/expected_revenue?size=5" {
			return
		}
		res.WriteHeader(http.StatusBadGateway)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	r := require.New(suite.T())
	j := NewJetlend(testServer.URL, 5*time.Second)
	ctx := context.Background()
	b, err := j.ExpectAmount(ctx, pkg.Config{Sids: []string{"sid1"}}, 5)
	r.ErrorContains(err, "can't get json for portfolio/charts/expected_revenue: response code 502: body")
	r.EqualValues("", b)
}

func (suite *JetlendTestSuite) TestExpectAmount_OK() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		fmt.Printf("------------- %s", req.RequestURI)
		if req.RequestURI != "/portfolio/charts/expected_revenue?size=5" {
			return
		}
		res.WriteHeader(http.StatusOK)
		res.Write([]byte("body"))
	}))
	defer func() { testServer.Close() }()

	r := require.New(suite.T())
	j := NewJetlend(testServer.URL, 5*time.Second)
	ctx := context.Background()
	b, err := j.ExpectAmount(ctx, pkg.Config{Sids: []string{"sid1"}}, 5)
	r.NoError(err)
	r.EqualValues("body", b)
}

func TestJetlendTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(JetlendTestSuite))
}
