package gcplocal_test

import (
	"fmt"
	PubSub "gcplocal/test"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {
	out := make(chan string, 3)
	go PubSub.SubTopic(3, out)

	PubSub.PubTopic("msg1")
	PubSub.PubTopic("msg2")
	PubSub.PubTopic("msg3")

	got := []string{}
	timeout := time.After(3 * time.Second)

	for i := 0; i < 3; i++ {
		select {
		case m := <-out:
			got = append(got, m)
		case <-timeout:
			t.Fatalf("timeout waiting for messages, got: %v", got)
		}
	}

	want := []string{"msg1", "msg2", "msg3"}
	sort.Strings(got)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("messages mismatch: got %v, want %v", got, want)
	} else {
		fmt.Printf("Passed Pub/Sub, sent from pub -> received on sub\n\n")
	}
}
