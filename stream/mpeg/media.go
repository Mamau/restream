package mpeg

import (
	"fmt"
	"strconv"
	"strings"
)

type fileType string

const (
	video fileType = "video"
	audio          = "audio"
)

type chunkMpeg struct {
	Type            fileType
	chunkNamePrefix string
	chunkTimeName   int
	startNumber     uint64
}

type media struct {
	chunks      map[fileType]chunkMpeg
	startNumber uint64
}

func (m *media) GetMediaByType(t fileType) string {
	chunkPath := fmt.Sprintf("%v_%v.mp4", m.chunks[t].chunkNamePrefix, fmt.Sprintf("%09d", m.chunks[t].chunkTimeName))
	chnk := m.chunks[t]
	chnk.chunkTimeName += int(m.chunks[t].startNumber)
	m.chunks[t] = chnk
	return chunkPath
}
func (m *media) DecrementByType(t fileType) {
	chnk := m.chunks[t]
	chnk.chunkTimeName -= int(m.chunks[t].startNumber)
	m.chunks[t] = chnk
}
func (m *media) SetByType(t fileType, duration, timescale, startNumber uint64, diffTime int64, media string) {
	name := m.formula(duration, timescale, startNumber, diffTime, media)

	parsedMedia := strings.Split(name, "_")
	n := strings.Split(parsedMedia[len(parsedMedia)-1], ".")
	result, err := strconv.Atoi(n[0])
	if err != nil {
		fmt.Printf("cant parse chunk name %v, error: %v\n", parsedMedia, err)
	}
	res := strings.Split(parsedMedia[0], "/")
	chnk := chunkMpeg{Type: t, chunkTimeName: result, startNumber: startNumber, chunkNamePrefix: strings.Join(res[3:], "/")}
	m.chunks[t] = chnk
}
func (m *media) formula(duration, timescale, startNumber uint64, diffTime int64, media string) string {
	formula := (int(diffTime) / (int(duration / timescale))) + int(startNumber)
	num := fmt.Sprintf(media, formula)
	res := strings.ReplaceAll(num, "$", "")
	preparedMedia := strings.ReplaceAll(res, "Number", "")
	return preparedMedia
}
