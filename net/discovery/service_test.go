package discovery

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/curator-go/curator"
	"github.com/samuel/go-zookeeper/zk"
)

func getTestCluster(t *testing.T) *zk.TestCluster {
	zkCluster, err := zk.StartTestCluster(1, os.Stdout, os.Stderr)

	if err != nil {
		t.Fatal("cannot start testing zk cluster: ", err)
	}

	_, err = zkCluster.Connect(0)

	if err != nil {
		t.Fatal("cannot connect to test cluster:", err)
	}

	log.Println("zk running at: ", zkCluster.Servers[0].Port)

	return zkCluster
}

func getTestClient(t *testing.T, zkCluster *zk.TestCluster) curator.CuratorFramework {
	server := fmt.Sprintf("localhost:%d", zkCluster.Servers[0].Port)

	retryPolicy := curator.NewExponentialBackoffRetry(time.Second, 3, 15*time.Second)
	client := curator.NewClient(server, retryPolicy)

	if err := client.Start(); err != nil {
		t.Fatal("cannot start curator client: ", err)
	}
	return client
}

func TestRoundTrip(t *testing.T) {
	z := getTestCluster(t)
	c1 := getTestClient(t, z)
	c2 := getTestClient(t, z)
	c3 := getTestClient(t, z)

	defer z.Stop()

	base := fmt.Sprintf("/foobar/baz-%d", time.Now().UnixNano())

	s1 := NewServiceDiscovery(c1, base)
	s1.MaintainRegistrations()

	s2 := NewServiceDiscovery(c2, base)
	s3 := NewServiceDiscovery(c3, base)

	if err := s2.Watch(); err != nil {
		t.Fatal("error starting instance watch: ", err)
	}
	if err := s3.Watch(); err != nil {
		t.Fatal("error starting instance watch: ", err)
	}

	if err := s1.MaintainRegistrations(); err != nil {
		t.Fatal("error starting registration maintainer: ", err)
	}

	if len(s2.Services) > 0 || len(s3.Services) > 0 {
		t.Fatal("unknown reg? ", s2.Services, s3.Services)
	}

	reg1 := NewSimpleServiceInstance("baz", "a", 8080)
	if err := s1.Register(reg1); err != nil {
		t.Fatal("failed to register 1: ", err)
	}
	time.Sleep(10 * time.Millisecond)

	m, err := c1.GetChildren().ForPath(base)
	if err != nil {
		t.Fatal("error getting children:", err)
	} else if len(m) != 1 || m[0] != "baz" {
		t.Fatal("Service does not appear to be registered:", m)
	} else if len(s2.Services) != 1 {
		t.Fatal("s2 missing new service:", s2.Services)
	} else if len(s2.Services["baz"]) != 1 {
		t.Fatal("s2 missing new reg1:", s2.Services)
	} else if !reflect.DeepEqual(s2.Services, s3.Services) {
		t.Fatal("s2 != s3: ", s2, s3)
	}

	reg2 := NewSimpleServiceInstance("baz", "a", 8081)
	if err := s1.Register(reg2); err != nil {
		t.Fatal("error registering 2:", err)
	}
	time.Sleep(10 * time.Millisecond)

	if ls, err := c1.GetChildren().ForPath(s1.pathForName("baz")); err != nil {
		t.Fatal("error getting children:", err)
	} else if len(ls) != 2 {
		t.Fatal("2nd register had no effect:", ls)
	} else if len(s2.Services["baz"]) != 2 {
		t.Fatal("s2 missed new reg:", s2.Services)
	}

	if err := s1.Unregister(reg1); err != nil {
		t.Fatal("error unregistering 1:", err)
	}
	time.Sleep(10 * time.Millisecond)

	if ls, err := c1.GetChildren().ForPath(s1.pathForName("baz")); err != nil {
		t.Fatal("error getting children:", err)
	} else if len(ls) != 1 {
		t.Fatal("un-register had no effect:", ls)
	} else if len(s2.Services["baz"]) != 1 {
		t.Fatal("s2 missed unreg:", s2.Services)
	}

	if err := s1.Unregister(reg2); err != nil {
		t.Fatal("error unregistering:", err)
	}
	time.Sleep(10 * time.Millisecond)

	if ls, err := c1.GetChildren().ForPath(s1.pathForName("baz")); err != nil {
		t.Fatal("error getting children:", err)
	} else if len(ls) != 0 {
		t.Fatal("un-register had no effect:", ls)
	} else if len(s2.Services["baz"]) != 0 {
		t.Fatal("s2 missed unreg:", s2.Services)
	} else if !reflect.DeepEqual(s2.Services, s3.Services) {
		t.Fatal("s2 != s3: ", s2, s3)
	}

	qux := s2.Provider("qux")
	if q, _ := qux.GetInstance(); q != nil {
		t.Fatal("should not find qux yet")
	}

	reg3 := NewSimpleServiceInstance("qux", "b", 8080)
	if err := s1.Register(reg3); err != nil {
		t.Fatal("error registering:", err)
	}
	time.Sleep(10 * time.Millisecond)

	if ls, err := c1.GetChildren().ForPath(s1.pathForName("qux")); err != nil {
		t.Fatal("error getting children:", err)
	} else if len(ls) != 1 {
		t.Fatal("register had no effect:", ls)
	} else if len(s2.Services["qux"]) != 1 {
		t.Fatal("s2 missed reg:", s2.Services)
	} else if len(s2.Services["baz"]) > 0 {
		t.Fatal("s2 missed pick up bad reg:", s2.Services)
	} else if !reflect.DeepEqual(s2.Services, s3.Services) {
		t.Fatal("s2 != s3: ", s2, s3)
	}

	if q, _ := qux.GetInstance(); q == nil {
		t.Fatal("did not find qux")
	}

	s1.UnregisterAll()
	time.Sleep(10 * time.Millisecond)

}
