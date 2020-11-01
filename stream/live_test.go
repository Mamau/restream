package stream

import (
	"testing"
)

func TestSetStream(t *testing.T) {
	streamName := "someStream"
	live := GetLive()
	strm := NewStream()
	strm.Name = streamName

	if err := live.SetStream(strm); err != nil {
		t.Error("Error setting stream")
	}

	if strm.Name != streamName {
		t.Errorf("Stream doesnt have stream with name %v\n", strm.Name)
	}
}

func TestGetStream(t *testing.T) {
	streamName := "someStream"
	strm := NewStream()
	strm.Name = streamName
	live := GetLive()

	err := live.SetStream(strm)
	if err != nil {
		t.Fatal(err.Error())
	}

	currentStream, err := live.GetStream(streamName)
	if err != nil {
		t.Fatal(err.Error())
	}
	if currentStream.Name != streamName {
		t.Errorf("stream %v doesn`t exist", streamName)
	}
}

func TestAllStreams(t *testing.T) {
	mapStrm := map[string]*Stream{}
	test := NewStream()
	test.Name = "test"
	mapStrm["test"] = test
	test2 := NewStream()
	test2.Name = "test2"
	mapStrm["test2"] = test2

	live := GetLive()
	for key, val := range live.AllStreams() {
		value, ok := mapStrm[key]
		if !ok {
			t.Errorf("Key %v not found", key)
		}
		if value.Name != val.Name {
			t.Errorf("Value %v at key %v is not same. %v != %v", val.Name, key, value.Name, val.Name)
		}
	}
}

func TestDeleteStream(t *testing.T) {
	streamName := "someStream"
	live := GetLive()
	strm := NewStream()
	strm.Name = streamName

	if err := live.SetStream(strm); err != nil {
		t.Error("Error setting stream")
	}

	deletedStream, err := live.DeleteStream(streamName)
	if err != nil {
		t.Fatal(err.Error())
	}
	if deletedStream.Name != streamName {
		t.Fatalf("Deleted stream %v not same", streamName)
	}

	_, err = live.GetStream(streamName)
	if err == nil {
		t.Errorf("Stream %v is not deleted", streamName)
	}
}
