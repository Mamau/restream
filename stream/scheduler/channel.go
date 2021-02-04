package scheduler

import (
	"github.com/mamau/restream/stream"
	"github.com/mamau/restream/stream/selenium"
	"github.com/mamau/restream/stream/selenium/channel"
	"log"
	"time"
)

type TimeTable struct {
	StartAt int64
	StopAt  int64
}

type Channel struct {
	*stream.Stream
	Channel    channel.Channel
	TimeTables []*TimeTable
}

func CreateChannel(chName channel.Channel) *Channel {
	ch := &Channel{
		Stream: stream.NewStream(),
	}
	ch.Stream.Name = string(chName)
	return ch
}

func (c *Channel) Schedule() {
	format := "15:04:05"
	startAt, err := time.Parse(format, "18:00:00")
	if err != nil {
		log.Fatalf("cant parse time %s\n", err.Error())
	}
	stopAt, err := time.Parse(format, "20:00:00")
	if err != nil {
		log.Fatalf("cant parse time %s\n", err.Error())
	}
	plan1 := &TimeTable{StartAt: startAt.Unix(), StopAt: stopAt.Unix()}

	c.TimeTables = []*TimeTable{plan1}
}

func (c *Channel) start() {
	manifest, err := selenium.GetManifest(channel.Channel(c.Stream.Name))
	if err != nil {
		log.Fatalf("cant start selenium %s\n", err.Error())
	}

	c.Stream.Manifest = manifest
	if err := c.Stream.Start(); err != nil {
		log.Fatalf("cant start stream %s\n", err.Error())
	}
}
