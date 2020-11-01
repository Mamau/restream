package stream

import (
	"errors"
	"fmt"
	"github.com/mamau/restream/helpers"
	"net/http"
	"sync"
)

var once sync.Once

type Live struct {
	Streams          map[string]*Stream
	ScheduledStreams map[string]*ScheduledStream
}

var instance *Live

func GetLive() *Live {
	once.Do(func() {
		instance = &Live{
			Streams:          make(map[string]*Stream),
			ScheduledStreams: make(map[string]*ScheduledStream),
		}
	})

	return instance
}

func (l *Live) ScheduleStream(s *ScheduledStream) error {
	_, ok := l.ScheduledStreams[s.Name]
	if ok {
		return errors.New(fmt.Sprintf("Stream %v already scheduled at %v\n", s.Name, s.StartAt))
	}

	//t := time.Now()
	//start := s.StartAt - t.Unix()
	//stop := s.StopAt - t.Unix()
	//format := "15_04_05_02_01_2006"
	//
	//killStreamWithDelay := stop + 10
	//startAt := time.Unix(s.StartAt, 10).Format(format)
	//stopAt := time.Unix(s.StopAt, 10).Format(format)
	//
	//zap.L().Info("stream scheduled download",
	//	zap.String("start", startAt),
	//	zap.String("stop", stopAt),
	//)
	//
	//s.Name = fmt.Sprintf("%v-%v-%v", startAt, stopAt, s.Name)
	//s.streamDuration = time.Duration(stop) * time.Second

	l.ScheduledStreams[s.Name] = s
	//s.ScheduleDownload(start, killStreamWithDelay)
	return nil
}

func (l *Live) AllStreams() map[string]*Stream {
	return l.Streams
}

func (l *Live) AllScheduledStreams() map[string]*ScheduledStream {
	return l.ScheduledStreams
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
