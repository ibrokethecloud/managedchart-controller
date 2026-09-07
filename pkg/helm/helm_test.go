package helm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/testify/v2/require"
)

func Test_DownloadIndex(t *testing.T) {
	assert := require.New(t)
	handler := http.FileServer(http.Dir("testdata"))
	server := httptest.NewServer(handler)
	defer server.Close()
	idxFile, err := DownloadIndex(server.URL)
	assert.NoError(err)
	assert.NotNil(idxFile)
}

func Test_ReadChart(t *testing.T) {
	assert := require.New(t)
	handler := http.FileServer(http.Dir("testdata"))
	server := httptest.NewServer(handler)
	defer server.Close()
	chart, err := ReadChart(server.URL, "harvester-cloud-provider", "0.2.14")
	assert.NoError(err)
	assert.NotNil(chart)
}
