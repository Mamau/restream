package stream

import "testing"

func TestAllStreams(t *testing.T) {
	mapStrm := map[string]*Stream{}
	test := InitStream()
	test.Name = "test"
	mapStrm["test"] = test
	test2 := InitStream()
	test2.Name = "test2"
	mapStrm["test2"] = test2

	live := Live{
		Streams: mapStrm,
	}
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
