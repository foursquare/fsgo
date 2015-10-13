package adminz

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/theevocater/go-atomicbool"
)

type Adminz struct {
	// represents the run state of the server
	running *atomicbool.AtomicBool

	// ticker that checks killfiles every 1 second
	killfileTicker *time.Ticker

	// list of killfilePaths to check
	killfilePaths []string

	// generates string to return to /servicez endpoint. should be json
	servicez func() interface{}

	// resume is called when the server is unkilled
	resume func() error

	// pause is called when the server is killed
	pause func() error
}

// Generates the standard set of killfiles. Pass these to Init()
func Killfiles(ports ...string) []string {
	// the number of ports + the "all" killfile
	var ret = make([]string, len(ports)+1)
	for i, port := range ports {
		ret[i] = fmt.Sprintf("/dev/shm/healthz/kill.%s", port)
	}
	ret[len(ports)] = "/dev/shm/healthz/kill.all"
	return ret
}

func New(pause func() error, resume func() error, servicez func() interface{}, killfilePaths []string) *Adminz {
	a := new(Adminz)

	a.pause = pause
	a.resume = resume
	a.servicez = servicez
	a.killfilePaths = killfilePaths

	go a.killfileLoop()
	a.killfileTicker = time.NewTicker(time.Second)
	a.running = atomicbool.New()
	http.HandleFunc("/healthz", a.healthzHandler)
	http.HandleFunc("/servicez", a.servicezHandler)
	log.Print("adminz registered")
	log.Print("Watching paths for killfile: ", killfilePaths)
	return a
}

func (a *Adminz) killed() bool {
	for _, killfile := range a.killfilePaths {
		file, err := os.Open(killfile)
		if file != nil && err == nil {
			file.Close()
			return true
		}
	}
	return false
}

func (a *Adminz) killfileLoop() {
	for _ = range a.killfileTicker.C {
		current := a.running.Get()
		next := !a.killed()
		if current == false && next == true {
			// If we are currently not running and the killfile is removed, call resume()
			a.resume()
			a.running.Set(next)
		} else if current == true && next == false {
			// If we are currently running and a killfile is dropped, call pause()
			a.pause()
			a.running.Set(next)
		}
		// If we hit neither of those, no state changed.
	}
}

func (a *Adminz) healthzHandler(w http.ResponseWriter, r *http.Request) {
	if a.running.Get() {
		w.Write(([]byte)("OK"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write(([]byte)("Service Unavailable"))
	}
}

func (a *Adminz) servicezHandler(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.Marshal(a.servicez())
	if err == nil {
		w.Header().Add("Content-Type", "text/json")
		// TODO I probably need to serialize reads to servicez as who knows what
		// people will put in that function
		w.Write(bytes)
	} else {
		http.Error(w, err.Error(), 500)
	}
}
