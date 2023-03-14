package pkg

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpClientOK(t *testing.T) {
	assert := assert.New(t)
	expected := "dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, expected)
	}))
	defer svr.Close()
	client := &http.Client{}
	data, err := getJSON(context.TODO(), client, svr.URL, "")
	assert.Nil(err)
	assert.Equal([]byte(expected), data)
}

func TestHttpClientCode(t *testing.T) {
	assert := assert.New(t)
	expected := "dummy data"
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, expected)
	}))
	defer svr.Close()
	client := &http.Client{}
	data, err := getJSON(context.TODO(), client, svr.URL, "")
	assert.Error(err)
	assert.Equal([]byte{}, data)
}

func TestJetUrl(t *testing.T) {
	assert := assert.New(t)
	data := jetURL("test")
	assert.Equal("https://jetlend.ru/invest/api/test", data)
}
