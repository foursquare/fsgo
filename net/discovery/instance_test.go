package discovery

import (
	"bytes"
	"reflect"
	"testing"
)

func TestJsonRoundTrip(t *testing.T) {
	s := &JsonInstanceSerializer{}

	p80 := 80
	in := &ServiceInstance{"0", "a", "addr1", &p80, nil, nil, 3, DYNAMIC, nil}

	raw, err := s.Serialize(in)
	if err != nil {
		t.Fatal(err)
	}

	if out, err := s.Deserialize(raw); err != nil {
		t.Fatal(err)
	} else if !reflect.DeepEqual(in, out) {
		t.Fatalf("did not round trip:\n%v\n%v", *in, *out)
	}

	shouldBe := []byte(`{"name":"0","id":"a","address":"addr1","port":80,"sslPort":null,"payload":null,"registrationTimeUTC":3,"serviceType":"DYNAMIC","uriSpec":null}`)
	if !bytes.Equal(shouldBe, raw) {
		t.Fatalf("wire representation doesn't match:\n%s\n%s", raw, shouldBe)
	}
}
