package adminz

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleAdminz_Build() {
	// To set up the adminz pages, first call New, then add whichever handlers
	// you need, then call build.
	a := New()
	a.OnPause(func() { /* do a thing */ })
	a.OnResume(func() { /* do a thing */ })
	a.Servicez(func() interface{} { return "{}" })
	a.Healthy(func() bool { return true })
	// If you don't add KillfilePaths, there will be no killfile checking.
	a.KillfilePaths(Killfiles(4000))
	a.Start()
}

func TestKillfile(t *testing.T) {
	killfile := path.Join(os.TempDir(), "kill")
	os.Remove(killfile)

	ok := "OK"
	notOK := "Service Unavailable"

	checkInterval := 50 * time.Millisecond

	pauseCounter := new(int)

	*pauseCounter = 1

	ts, a := newTestAdminz()
	defer ts.Close()

	url := ts.URL + "/healthz"

	a.KillfilePaths([]string{killfile})
	a.KillfileInterval(checkInterval)
	a.OnPause(func() {
		*pauseCounter += 1
	})
	a.OnResume(func() {
		*pauseCounter -= 1
	})
	a.Start()
	defer a.Stop()

	assert.Equal(t, *pauseCounter, 0, "Pause shouldn't be called yet")
	assert.Equal(t, ok, string(readAllURL(t, url)), "Server should be healthy")

	assert.True(t, a.running, "Killfile shouldn't exist")
	k, err := os.Create(killfile)
	assert.Nil(t, err, "Unable to create killfile")
	defer k.Close()

	// Sleep for 2 seconds to ensure the ticker has run
	time.Sleep(checkInterval * 2)
	assert.False(t, a.running, "Killfile missed")
	assert.Equal(t, *pauseCounter, 1, "Didn't call pause")
	assert.Equal(t, notOK, string(readAllURL(t, url)), "Server should not be healthy")

	time.Sleep(checkInterval * 2)
	assert.Equal(t, *pauseCounter, 1, "Pause should only be called once")

	// Now remove and ensure we reset to running
	os.Remove(killfile)

	time.Sleep(checkInterval * 2)
	assert.Equal(t, *pauseCounter, 0, "Resume should have been called")
	assert.True(t, a.running, "Killfile shouldn't exist")
	assert.Equal(t, ok, string(readAllURL(t, url)), "Server should be healthy")
}

func TestServicez(t *testing.T) {
	servicez := []byte("{\"hello\":5}")

	// create an unmarshalled version of above
	servicezMap := make(map[string]int)
	servicezMap["hello"] = 5
	err := json.Unmarshal(servicez, &servicezMap)
	assert.Nil(t, err)

	// fake out an http server, add trivial servicez
	ts, a := newTestAdminz()
	defer ts.Close()
	a.Servicez(func() interface{} {
		return servicezMap
	})
	a.Start()
	defer a.Stop()

	assert.Equal(t, string(servicez), string(readAllURL(t, ts.URL+"/servicez")))
}

// This test must be last as it uses the DefaultServeMux
func TestStartNoInputs(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	a := New()
	a.ServeMux(mux)
	a.Start()
	defer a.Stop()
}

func readAllURL(t *testing.T, url string) []byte {
	res, err := http.Get(url)
	assert.Nil(t, err)

	ret, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	assert.Nil(t, err)
	return ret
}

func newTestAdminz() (ts *httptest.Server, a *Adminz) {
	mux := http.NewServeMux()
	ts = httptest.NewServer(mux)
	a = New()
	a.ServeMux(mux)
	return
}
