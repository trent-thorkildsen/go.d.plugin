package nginx

import (
	"github.com/netdata/go.d.plugin/modules"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/netdata/go.d.plugin/pkg/web"

	"github.com/stretchr/testify/assert"
)

var (
	okStatus, _      = ioutil.ReadFile("testdata/status.txt")
	invalidStatus, _ = ioutil.ReadFile("testdata/status-invalid.txt")
)

func TestNginx_Cleanup(t *testing.T) {
	New().Cleanup()
}

func TestNew(t *testing.T) {
	assert.Implements(t, (*modules.Module)(nil), New())
}

func TestNginx_Init(t *testing.T) {
	mod := New()

	assert.True(t, mod.Init())
	assert.NotNil(t, mod.request)
	assert.NotNil(t, mod.client)
}

func TestNginx_Check(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stub_status" {
					_, _ = w.Write(okStatus)
					return
				}
			}))

	defer ts.Close()

	mod := New()

	mod.HTTP.Request = web.Request{URL: ts.URL + "/stub_status"}
	require.True(t, mod.Init())
	assert.True(t, mod.Check())
}

func TestNginx_CheckNG(t *testing.T) {
	mod := New()

	mod.HTTP.Request = web.Request{URL: "http://127.0.0.1:38001/stub_status"}
	require.True(t, mod.Init())
	assert.False(t, mod.Check())
}

func TestNginx_Charts(t *testing.T) {
	assert.NotNil(t, New().Charts())
}

func TestNginx_Collect(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stub_status" {
					_, _ = w.Write(okStatus)
					return
				}
			}))
	defer ts.Close()

	mod := New()
	mod.HTTP.Request = web.Request{URL: ts.URL + "/stub_status"}

	assert.True(t, mod.Init())

	metrics := mod.Collect()
	assert.NotNil(t, metrics)

	expected := map[string]int64{
		"active":   1,
		"accepts":  36,
		"handled":  36,
		"requests": 126,
		"reading":  0,
		"writing":  1,
		"waiting":  0,
	}

	assert.Equal(t, expected, metrics)
}

func TestNginx_InvalidData(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/stub_status" {
					_, _ = w.Write(invalidStatus)
					return
				}
			}))
	defer ts.Close()

	mod := New()
	mod.HTTP.Request = web.Request{URL: ts.URL + "/stub_status"}

	require.True(t, mod.Init())
	assert.False(t, mod.Check())
}

func TestNginx_404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	mod := New()
	mod.HTTP.Request = web.Request{URL: ts.URL + "/stub_status"}

	require.True(t, mod.Init())
	assert.False(t, mod.Check())
}