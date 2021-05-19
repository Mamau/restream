package channel

import (
	"bufio"
	"github.com/grafov/m3u8"
	"log"
	"net/http"
	"strings"
)

type ChannelName string

var playlists = []string{
	"https://iptvmaster.ru/russia.m3u",
	"https://webarmen.com/my/iptv/auto.nogrp.m3u",
}

const (
	TNT   ChannelName = "ТНТ"
	FIRST ChannelName = "Первый канал"
	MATCH ChannelName = "Матч Премьер"
)

var TimeTable = map[ChannelName][][]string{
	TNT: {
		{"11:00:00", "04:00:00"},
	},
	FIRST: {
		{"07:00:00", "13:00:00"},
	},
}

func fetchSegments() []*m3u8.MediaSegment {
	var segments []*m3u8.MediaSegment
	for _, url := range playlists {
		segments = append(segments, fetchPlaylist(url)...)
	}

	return segments
}

func fetchPlaylist(url string) []*m3u8.MediaSegment {
	var sl []*m3u8.MediaSegment
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}

	defer response.Body.Close()

	p, listType, err := m3u8.DecodeFrom(bufio.NewReader(response.Body), false)

	switch listType {
	case m3u8.MEDIA:
		mediapl := p.(*m3u8.MediaPlaylist)
		sl = cleanSegments(mediapl.Segments)
	}
	return sl
}

func cleanSegments(list []*m3u8.MediaSegment) []*m3u8.MediaSegment {
	var segments []*m3u8.MediaSegment
	for _, item := range list {
		if item == nil {
			continue
		}

		if strings.HasPrefix(item.Title, "[") {
			continue
		}

		segments = append(segments, item)
	}
	return segments
}
