package channel

import (
	"github.com/grafov/m3u8"
	"sync"
)

var once sync.Once
var instance *Source

type Source struct {
	PlayList []*m3u8.MediaSegment
}

type Channel struct {
	Name string `json:"name"`
	Uri  string `json:"uri"`
}

func NewSource() *Source {
	once.Do(func() {
		instance = &Source{
			PlayList: fetchSegments(),
		}
	})
	return instance
}

func (s *Source) GetChannels() *[]Channel {
	var list []Channel
	for _, v := range s.PlayList {
		list = append(list, Channel{Name: v.Title, Uri: v.URI})
	}

	return &list
}

func (s *Source) GetManifestByName(name ChannelName) *m3u8.MediaSegment {
	for _, item := range s.PlayList {
		if item.Title == string(name) {
			return item
		}
	}
	return nil
}
