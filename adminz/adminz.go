// Package adminz provides a simple set of adminz pages for administering
// a simple go server.
package adminz

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type Adminz struct {
	// keep track of killfile state
	running bool

	// ticker that checks killfiles every 1 second
	killfileTicker *time.Ticker

	// list of killfilePaths to check
	killfilePaths []string

	// defaults to 1 second
	checkInterval time.Duration

	// generates data to return to /servicez endpoint. marshalled into json.
	servicez func() interface{}

	// resume is called when the server is unkilled
	onresume func()

	// pause is called when the server is killed
	onpause func()

	// healthy returns true iff the server is ready to respond to requests
	healthy func() bool

	sync.Mutex
}

// Creates a new Adminz "builder". Not safe to use until Build() is called.
func New() *Adminz {
	return &Adminz{}
}

func (a *Adminz) Resume() {
	a.Lock()
	defer a.Unlock()
	a.doResume()
}

// Internal helper to resume if stopped. MUST be called while holding a.Lock. Use a.Resume if not.
func (a *Adminz) doResume() {
	if !a.running {
		if a.onresume != nil {
			a.onresume()
		}
		a.running = true
	}
}

// Resume is called when the server is unkilled
func (a *Adminz) OnResume(resume func()) *Adminz {
	a.onresume = resume
	return a
}

// Pause the server if it is running and return true, otherwise no-op and return false.
func (a *Adminz) Pause() bool {
	a.Lock()
	defer a.Unlock()
	return a.doPause()
}

// Internal helper to pause if running. MUST be called while holding a.Lock. Use a.Pause if not.
func (a *Adminz) doPause() bool {
	was := a.running
	if a.running {
		if a.onpause != nil {
			a.onpause()
		}
		a.running = false
	}
	return was
}

// pause is called when the server is killed
func (a *Adminz) OnPause(pause func()) *Adminz {
	a.onpause = pause
	return a
}

// healthy returns true iff the server is ready to respond to requests
func (a *Adminz) Healthy(healthy func() bool) *Adminz {
	a.healthy = healthy
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
func (a *Adminz) Build() *Adminz {
	http.HandleFunc("/healthz", a.healthzHandler)
	http.HandleFunc("/servicez", a.servicezHandler)
	http.HandleFunc("/gc", a.gcHandler)

	log.Print("adminz registered")

	if a.checkInterval == 0 {
		a.checkInterval = 1 * time.Second
	}
	a.killfileTicker = time.NewTicker(a.checkInterval)

	// start killfile checking loop
	if len(a.killfilePaths) > 0 {
		log.Print("Watching paths for killfile: ", a.killfilePaths)
		a.checkKillfiles()
		go a.killfileLoop()
	} else {
		log.Print("No killfiles being watched.")
	}

	return a
}

func (a *Adminz) Stop() {
	if a.killfileTicker != nil {
		a.killfileTicker.Stop()
	}
}

// Generates the standard set of killfiles. Pass these to KillfilePaths
func Killfiles(ports ...int) []string {
	// the number of ports + the "all" killfile
	log.Print(ports)
	var ret = make([]string, len(ports)+1)
	for i, port := range ports {
		ret[i] = fmt.Sprintf("/dev/shm/healthz/kill.%d", port)
	}
	ret[len(ports)] = "/dev/shm/healthz/kill.all"
	return ret
}

func (a *Adminz) checkKillfiles() bool {
	for _, killfile := range a.killfilePaths {
		file, err := os.Open(killfile)
		if file != nil && err == nil {
			file.Close()
			if a.Pause() {
				log.Println("paused due to ", killfile)
			}
			return true
		}
	}
	a.Resume()
	return false
}

func (a *Adminz) killfileLoop() {
	for _ = range a.killfileTicker.C {
		a.checkKillfiles()
	}
}

func (a *Adminz) healthzHandler(w http.ResponseWriter, r *http.Request) {
	// we are healthy iff:
	// we are not killed AND
	// a.healthy is unset (so we ignore it) OR
	// a.healthy() returns true
	var ret string

	if !a.running && (a.healthy == nil || a.healthy()) {
		ret = "OK"
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		ret = "Service Unavailable"
	}

	log.Print("Healthz returning ", ret)
	w.Write(([]byte)(ret))
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

func (a *Adminz) gcHandler(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats

	mb := uint64(1024 * 1024)
	runtime.ReadMemStats(&mem)
	fmt.Fprintln(w, "Before")
	fmt.Fprintln(w, "\tAlloc\t", mem.Alloc/mb)
	fmt.Fprintln(w, "\tTotalAlloc:\t", mem.TotalAlloc/mb)
	fmt.Fprintln(w, "\tHeapAlloc:\t", mem.HeapAlloc/mb)
	fmt.Fprintln(w, "\tHeapSys:\t", mem.HeapSys/mb)
	fmt.Fprintln(w, "\tSys:\t", mem.Sys/mb)

	a.Lock()
	was := a.doPause()
	runtime.GC()
	if was {
		a.doResume()
	}
	a.Unlock()

	runtime.ReadMemStats(&mem)
	fmt.Fprintln(w, "After")
	fmt.Fprintln(w, "\tAlloc\t", mem.Alloc/mb)
	fmt.Fprintln(w, "\tTotalAlloc:\t", mem.TotalAlloc/mb)
	fmt.Fprintln(w, "\tHeapAlloc:\t", mem.HeapAlloc/mb)
	fmt.Fprintln(w, "\tHeapSys:\t", mem.HeapSys/mb)
	fmt.Fprintln(w, "\tSys:\t", mem.Sys/mb)

	w.Write([]byte("OK"))
}
