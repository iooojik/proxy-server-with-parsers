package parsers

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestFull(t *testing.T) {
	client := NewGeoNodeParser()
	client.Run()
}

func TestUploadNParse(t *testing.T) {
	client := NewGeoNodeParser()
	proxies, err := client.parse()
	assert.NoError(t, err)
	assert.NotEmpty(t, proxies)
}

func TestGettingResponse(t *testing.T) {
	client := NewGeoNodeParser()
	proxies, err := client.getResponse()
	assert.NoError(t, err)
	assert.NotEmpty(t, proxies)
}

func TestParsingResponse(t *testing.T) {
	client := NewGeoNodeParser()
	response, err := client.getResponse()
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
	data, e := client.parseResponse(response)
	assert.NoError(t, e)
	assert.NotEmpty(t, data.Data)
}

func TestParsingJson(t *testing.T) {
	var (
		body io.ReadCloser
	)
	defer func() {
		if err := body.Close(); err != nil {
			assert.NoError(t, err)
		}
	}()
	path, err := os.Getwd()
	assert.NoError(t, err)
	data, err := os.ReadFile(path + "/Free_Proxy_List.json")
	assert.NoError(t, err)
	body = io.NopCloser(bytes.NewReader(data))
	client := NewGeoNodeParser()
	apiData, e := client.parseResponse(&http.Response{Body: body})
	assert.NoError(t, e)
	assert.True(t, len(apiData.Data) == 33)
}
