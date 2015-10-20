// Package adminz provides a simple set of adminz pages for administering
// a simple go server.
package adminz

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/foursquare/fsgo/concurrent/atomicbool"
)

func Flag() *int {
	return flag.Int("admin-port", 0, "port for http admin (defaults to default http servemux)")
}

type Adminz struct {
	// keep track of killfile state
	Killed *atomicbool.AtomicBool

	// ticker that checks killfiles every 1 second
	killfileTicker *time.Ticker

	// list of killfilePaths to check
	killfilePaths []string

	// defaults to 1 second
	checkInterval time.Duration

	// generates data to return to /servicez endpoint. marshalled into json.
	servicez func() interface{}

	beforeShutdown func()

	// resume is called when the server is unkilled
	resume func()

	// pause is called when the server is killed
	pause func()

	// healthy returns true iff the server is ready to respond to requests
	healthy func() bool
}

// Creates a new Adminz "builder". Not safe to use until Build() is called.
func New() *Adminz {
	return &Adminz{
		Killed: atomicbool.New(),
	}
}

// Resume is called when the server is unkilled
func (a *Adminz) Resume(resume func()) *Adminz {
	a.resume = resume
	return a
}

// pause is called when the server is killed
func (a *Adminz) Pause(pause func()) *Adminz {
	a.pause = pause
	return a
}

// healthy returns true iff the server is ready to respond to requests
func (a *Adminz) Healthy(healthy func() bool) *Adminz {
	a.healthy = healthy
	return a
}

// function to run before exiting when a shutdown is requested over http admin.
func (a *Adminz) BeforeShutdown(f func()) *Adminz {
	a.beforeShutdown = f
	return a
}

// servicez generates data to return to /servicez endpoint. marshalled into
// json.
func (a *Adminz) Servicez(servicez func() interface{}) *Adminz {
	a.servicez = servicez
	return a
}

// Sets the list of killfilePaths to check.
func (a *Adminz) KillfilePaths(killfilePaths []string) *Adminz {
	a.killfilePaths = killfilePaths
	return a
}

// Sets frequency the killfile is checked. defaults every second
func (a *Adminz) KillfileInterval(interval time.Duration) *Adminz {
	a.checkInterval = interval
	return a
}

// Build initializes handlers and starts killfile checking. Make sure to
// remember to call this!
func (a *Adminz) Build(mux *http.ServeMux) *Adminz {
	if a.checkInterval == 0 {
		a.checkInterval = 1 * time.Second
	}
	a.killfileTicker = time.NewTicker(a.checkInterval)

	// start killfile checking loop
	if len(a.killfilePaths) > 0 {
		go a.killfileLoop()
	} else {
		log.Print("Not checking killfiles.")
	}

	if mux == nil {
		mux = http.DefaultServeMux
	}

	mux.HandleFunc("/healthz", a.healthzHandler)
	mux.HandleFunc("/servicez", a.servicezHandler)

	mux.HandleFunc("/stopstopstop", a.gracefulShutdownHandler)
	mux.HandleFunc("/abortabortabort", a.fastShutdownHandler)

	log.Print("adminz registered")
	log.Print("Watching paths for killfile: ", a.killfilePaths)
	return a
}

func (a *Adminz) Listen(port *int) {
	mux := http.NewServeMux()
	a.Build(mux)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), mux); err != nil {
		log.Fatal("Error starting admin listener:", err)
	}
}

func (a *Adminz) Stop() {
	if a.killfileTicker != nil {
		a.killfileTicker.Stop()
	}
}

// Generates the standard set of killfiles. Pass these to KillfilePaths
func Killfiles(ports ...string) []string {
	// the number of ports + the "all" killfile
	var ret = make([]string, len(ports)+1)
	for i, port := range ports {
		ret[i] = fmt.Sprintf("/dev/shm/healthz/kill.%s", port)
	}
	ret[len(ports)] = "/dev/shm/healthz/kill.all"
	return ret
}

func (a *Adminz) checkKillfiles() bool {
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
		current := a.Killed.Get()
		next := a.checkKillfiles()
		if current == false && next == true {
			// If we are currently running and a killfile is dropped, call pause()
			if a.pause != nil {
				a.pause()
			}
			a.Killed.Set(next)
		} else if current == true && next == false {
			// If we are currently not running and the killfile is removed, call resume()
			if a.resume != nil {
				a.resume()
			}
			a.Killed.Set(next)
		}
		// If we hit neither of those, no state changed.
	}
}

func (a *Adminz) fastShutdownHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Fast shutdown requested, shutting down now.")
	go os.Exit(0)
}

func (a *Adminz) gracefulShutdownHandler(w http.ResponseWriter, r *http.Request) {
	// BUG(davidt): Not protected against concurrent shutdown calls.
	if a.beforeShutdown != nil {
		go func() {
			log.Println("Graceful shutdown starting...")
			a.beforeShutdown()
			log.Println("Graceful complete, exiting.")
			os.Exit(0)
		}()
	}
	w.Write([]byte("OK"))
}

func (a *Adminz) healthzHandler(w http.ResponseWriter, r *http.Request) {
	// we are healthy iff:
	// we are not killed AND
	// a.healthy is unset (so we ignore it) OR
	// a.healthy() returns true
	var ret string
	if !a.Killed.Get() && (a.healthy == nil || a.healthy()) {
		ret = "OK"
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		ret = "Service Unavailable"
	}
	log.Print("Healthz returning ", ret)
	w.Write([]byte(ret))
}

func (a *Adminz) servicezHandler(w http.ResponseWriter, r *http.Request) {
	if a.servicez == nil {
		return
	}

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
