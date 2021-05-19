package channel

import (
	"github.com/grafov/m3u8"
)

type Source struct {
	PlayList []*m3u8.MediaSegment
}

func NewSource() *Source {
	return &Source{
		PlayList: fetchSegments(),
	}
}

func (s *Source) GetManifestByName(name ChannelName) *m3u8.MediaSegment {
	for _, item := range s.PlayList {
		if item.Title == string(name) {
			return item
		}
	}
	return nil
}
