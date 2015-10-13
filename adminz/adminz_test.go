package adminz

import (
	//"bytes"
	//"fmt"
	//"io/ioutil"
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
	a.Pause(func() error { /* do a thing */ return nil })
	a.Resume(func() error { /* do a thing */ return nil })
	a.Servicez(func() interface{} { return "{}" })
	a.Healthy(func() bool { return true })
	// If you don't add KillfilePaths, there will be no killfile checking.
	a.KillfilePaths(Killfiles("4000"))
	a.Build()
}

func TestKillfile(t *testing.T) {
	killfile := path.Join(os.TempDir(), "kill")
	pauseCounter := new(int)

	*pauseCounter = 0

	a := New()
	a.KillfilePaths([]string{killfile})
	a.Pause(func() error {
		*pauseCounter += 1
		return nil
	})
	a.Resume(func() error {
		*pauseCounter -= 1
		return nil
	})
	a.Build()
	defer a.Stop()

	assert.Equal(t, *pauseCounter, 0, "Pause shouldn't be called yet")

	assert.False(t, a.Killed.Get(), "Killfile shouldn't exist")
	k, err := os.Create(killfile)
	assert.Nil(t, err, "Unable to create killfile")
	defer k.Close()

	// Sleep for 2 seconds to ensure the ticker has run
	time.Sleep(time.Second * 2)
	assert.True(t, a.Killed.Get(), "Killfile missed")
	assert.Equal(t, *pauseCounter, 1, "Didn't call pause")

	time.Sleep(time.Second * 2)
	assert.Equal(t, *pauseCounter, 1, "Pause should only be called once")

	// Now remove and ensure we reset to running
	os.Remove(killfile)

	time.Sleep(time.Second * 2)
	assert.Equal(t, *pauseCounter, 0, "Resume should have been called")
	assert.False(t, a.Killed.Get(), "Killfile shouldn't exist")
}

// Can't run this until I figure out how to tear up and down http stuff.
// Otherwise I reregister the handlers.
//func TestBuildNoInputs(t *testing.T) {
//a := New()
//a.Build()
//defer a.Stop()
//}
