package stream

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"net/http"
)

type Live struct {
	Streams map[string]*Stream
}

func (l *Live) AllStreams() map[string]*Stream {
	return l.Streams
}

func (l *Live) SetStream(s *Stream) error {
	if _, err := l.GetStream(s.Name); err != nil {
		l.Streams[s.Name] = s
		return nil
	}
	return errors.New(fmt.Sprintf("Stream with name %v already exists", s.Name))
}

func (l *Live) GetStream(name string) (*Stream, error) {
	strm, ok := l.Streams[name]
	if !ok {
		return &Stream{}, errors.New(fmt.Sprintf("Not found stream by name: %v", name))
	}
	return strm, nil
}

func (l *Live) DeleteStream(name string) (*Stream, error) {
	strm, err := l.GetStream(name)
	if err == nil {
		delete(l.Streams, name)
		return strm, nil
	}
	return &Stream{}, errors.New(fmt.Sprintf("Stream with name %v not found", name))
}

func (l *Live) GetStreamByRequest(r *http.Request) (*Stream, error) {
	type dataStream struct {
		Name string
	}
	var ds dataStream

	err := helpers.JsonRequestToMap(r, &ds)
	if err != nil {
		return &Stream{}, errors.New("error while parse request")
	}

	return l.GetStream(ds.Name)
}
