package discovery

import (
	"log"

	"github.com/curator-go/curator"
	"github.com/samuel/go-zookeeper/zk"
)

type TreeCache struct {
	*ServiceDiscovery

	existing map[string]map[string]*ServiceInstance

	serviceListChanges  chan bool
	instanceListChanges chan string
}

func NewTreeCache(s *ServiceDiscovery) *TreeCache {
	existing := make(map[string]map[string]*ServiceInstance)
	return &TreeCache{s, existing, make(chan bool, 10), make(chan string, 10)}
}

func (t *TreeCache) Start() {
	t.processServiceChanges()
	go t.processInstanceChanges()
}

func (t *TreeCache) processInstanceChanges() {
	for {
		s, ok := getMostRecentString(t.instanceListChanges)
		if !ok {
			break
		}
		t.readAndWatch(s, "restarting")
	}
	log.Println("Done watching for instance changes")
}

func (t *TreeCache) readAndWatch(service, verb string) {
	p := t.pathForName(service)
	w := curator.NewWatcher(func(e *zk.Event) { t.instanceListChanges <- service })
	children, err := t.client.GetChildren().UsingWatcher(w).ForPath(p)

	if err != nil {
		log.Printf("Error %s watch for %s: %s\n", verb, service, err)
	} else {
		t.readInstanceList(service, children)
	}
}

func (t *TreeCache) readInstanceList(s string, children []string) {
	instances := make([]*ServiceInstance, 0, len(children))

	existing, ok := t.existing[s]
	if !ok {
		existing = make(map[string]*ServiceInstance)
	}

	for _, id := range children {
		if e, found := existing[id]; found {
			instances = append(instances, e)
			continue
		}

		p := t.pathForInstance(s, id)
		data, err := t.client.GetData().ForPath(p)
		if err != nil {
			log.Printf("Error fetching instance info for %s-%s (%s): %s\n", s, id, p, err)
			continue
		}

		i, err := t.serializer.Deserialize(data)
		if err != nil {
			log.Printf("Error decoding instance info for %s-%s (%s): %s\n", s, id, p, err)
			continue
		}
		i.Id = id // Just in case, since we treat use path for caching.

		log.Printf("New instance for %s in %s (%s)\n", s, i.Spec(), id)
		existing[id] = i
		instances = append(instances, i)
	}
	t.existing[s] = existing
	t.Services[s] = instances
}

func (t *TreeCache) processServiceChanges() {
	watching := make(map[string]bool)

	go func() {
		for {
			_, cont := getMostRecentBool(t.serviceListChanges)
			if !cont {
				break
			}

			t.readServices(watching)
		}
		log.Println("Done watching for service changes")
	}()
	t.readServices(watching)
}

func (t *TreeCache) readServices(watching map[string]bool) {
	w := curator.NewWatcher(func(*zk.Event) { t.serviceListChanges <- true })

	children, err := t.client.GetChildren().UsingWatcher(w).ForPath(t.basePath)
	if err != nil {
		log.Println("Error reading service list: ", err)
		return
	}
	found := make(map[string]bool)
	for _, i := range children {
		found[i] = true
		if !watching[i] {
			t.readAndWatch(i, "starting")
			watching[i] = true
		}
	}
	for i, _ := range t.Services {
		if !found[i] {
			delete(t.Services, i)
		}
	}
}
