package mpeg

import (
	"fmt"
	"strconv"
	"strings"
)

type media struct {
	chunkTimeName        int
	chunkAudioTimeName   int
	chunkNamePrefix      string
	chunkAudioNamePrefix string
	startNumber          uint64
}

func (m *media) GetMedia() string {
	chunkPath := fmt.Sprintf("%v_%v.mp4", m.chunkNamePrefix, fmt.Sprintf("%09d", m.chunkTimeName))
	m.chunkTimeName += int(m.startNumber)
	return chunkPath
}
func (m *media) GetAudioMedia() string {
	chunkPath := fmt.Sprintf("%v_%v.mp4", m.chunkAudioNamePrefix, fmt.Sprintf("%09d", m.chunkAudioTimeName))
	m.chunkAudioTimeName += int(m.startNumber)
	return chunkPath
}
func (m *media) DecrementMedia() {
	m.chunkTimeName -= int(m.startNumber)
}
func (m *media) DecrementAudioMedia() {
	m.chunkAudioTimeName -= int(m.startNumber)
}
func (m *media) SetAudioMedia(duration, timescale, startNumber uint64, diffTime int64, media string) {
	name := m.formula(duration, timescale, startNumber, diffTime, media)

	parsedMedia := strings.Split(name, "_")
	n := strings.Split(parsedMedia[len(parsedMedia)-1], ".")
	result, err := strconv.Atoi(n[0])
	if err != nil {
		fmt.Printf("cant parse chunk name %v, error: %v\n", parsedMedia, err)
	}
	res := strings.Split(parsedMedia[0], "/")
	m.chunkAudioNamePrefix = strings.Join(res[3:], "/")
	m.chunkAudioTimeName = result
}
func (m *media) SetMedia(duration, timescale, startNumber uint64, diffTime int64, media string) {
	name := m.formula(duration, timescale, startNumber, diffTime, media)

	parsedMedia := strings.Split(name, "_")
	n := strings.Split(parsedMedia[len(parsedMedia)-1], ".")
	result, err := strconv.Atoi(n[0])
	if err != nil {
		fmt.Printf("cant parse chunk name %v, error: %v\n", parsedMedia, err)
	}
	res := strings.Split(parsedMedia[0], "/")
	m.chunkNamePrefix = strings.Join(res[3:], "/")
	m.chunkTimeName = result
}
func (m *media) formula(duration, timescale, startNumber uint64, diffTime int64, media string) string {
	m.startNumber = startNumber
	formula := (int(diffTime) / (int(duration / timescale))) + int(startNumber)
	num := fmt.Sprintf(media, formula)
	res := strings.ReplaceAll(num, "$", "")
	preparedMedia := strings.ReplaceAll(res, "Number", "")
	return preparedMedia
}
