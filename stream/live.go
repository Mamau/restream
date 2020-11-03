package stream

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

var once sync.Once

type Live struct {
	Streams map[string]Streamer
}

var instance *Live

func GetLive() *Live {
	once.Do(func() {
		instance = &Live{
			Streams: make(map[string]Streamer),
		}
	})

	return instance
}

func (l *Live) ScheduleStream(s Streamer) error {
	_, ok := l.Streams[s.GetName()]
	if ok {
		return errors.New(fmt.Sprintf("Stream %v already scheduled\n", s.GetName()))
	}
	zap.L().Info("stream scheduled", zap.String("stream", s.GetName()))

	l.Streams[s.GetName()] = s
	return nil
}

func (l *Live) GetScheduledStream(name string) (Streamer, error) {
	strm, ok := l.Streams[name]
	if !ok {
		return &ScheduledStream{}, errors.New(fmt.Sprintf("Not found stream by name: %v", name))
	}
	return strm, nil
}

func (l *Live) DeleteScheduledStream(name string) (Streamer, error) {
	strm, err := l.GetScheduledStream(name)
	if err == nil {
		delete(l.Streams, name)
		zap.L().Info("stream deleted", zap.String("stream", strm.GetName()))
		return strm, nil
	}
	return &ScheduledStream{}, errors.New(fmt.Sprintf("Stream with name %v not found", name))
}

func (l *Live) AllStreams() map[string]Streamer {
	return l.Streams
}

func (l *Live) SetStream(s Streamer) error {
	if _, err := l.GetStream(s.GetName()); err != nil {
		l.Streams[s.GetName()] = s
		return nil
	}
	return errors.New(fmt.Sprintf("Stream with name %v already exists", s.GetName()))
}

func (l *Live) GetStream(name string) (Streamer, error) {
	strm, ok := l.Streams[name]
	if !ok {
		return &Stream{}, errors.New(fmt.Sprintf("Not found stream by name: %v", name))
	}
	return strm, nil
}

func (l *Live) DeleteStream(name string) (Streamer, error) {
	strm, err := l.GetStream(name)
	if err == nil {
		delete(l.Streams, name)
		return strm, nil
	}
	return &Stream{}, errors.New(fmt.Sprintf("Stream with name %v not found", name))
}

func (l *Live) GetStreamByRequest(r *http.Request) (Streamer, error) {
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
