package discovery

import (
	"sync"
	"testing"
)

func countBools(c chan bool, wg *sync.WaitGroup, i *int) {
	getMostRecentBool(c)
	*i++
	wg.Done()
}

func countStrings(c chan string, wg *sync.WaitGroup, i *int) {
	getMostRecentString(c)
	*i++
	wg.Done()
}

func TestGetMostRecentBool(t *testing.T) {
	wg := new(sync.WaitGroup)
	expected := 0

	c := make(chan bool, 5)
	i := 0

	// Sending an event at a time (and waiting between), we'll see each counted.
	go countBools(c, wg, &i)
	wg.Add(1)
	c <- true
	wg.Wait()

	go countBools(c, wg, &i)
	wg.Add(1)
	c <- true
	wg.Wait()

	if expected = 2; i != expected {
		t.Fatalf("wrong number of things: %d vs %d", i, expected)
	}

	i = 0
	// now we'll see it fold 3 items down to 1.
	wg.Add(1)
	c <- true
	c <- true
	c <- true
	go countBools(c, wg, &i)
	wg.Wait()
	if expected = 1; i != expected {
		t.Fatalf("wrong number of things: %d vs %d", i, expected)
	}
}

func TestGetMostRecentString(t *testing.T) {
	wg := new(sync.WaitGroup)
	expected := 0

	c := make(chan string, 5)
	i := 0

	// Sending an event at a time (and waiting between), we'll see each counted.
	go countStrings(c, wg, &i)
	wg.Add(1)
	c <- "1"
	wg.Wait()

	go countStrings(c, wg, &i)
	wg.Add(1)
	c <- "1"
	wg.Wait()

	if expected = 2; i != expected {
		t.Fatalf("wrong number of things: %d vs %d", i, expected)
	}

	i = 0
	// now we'll see it fold 3 items down to 1.
	wg.Add(1)
	c <- "1"
	c <- "1"
	c <- "1"
	go countStrings(c, wg, &i)
	wg.Wait()
	if expected = 1; i != expected {
		t.Fatalf("wrong number of things: %d vs %d", i, expected)
	}
}
