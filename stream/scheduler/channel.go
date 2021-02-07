package scheduler

import (
	"github.com/mamau/restream/stream"
	"github.com/mamau/restream/stream/selenium"
	"github.com/mamau/restream/stream/selenium/channel"
	"github.com/rk/go-cron"
	"time"
)

type TimeTable struct {
	StartAt time.Time
	StopAt  time.Time
}

type Channel struct {
	*stream.Stream
	Channel    channel.Channel
	TimeTables []*TimeTable
}

func CreateScheduledChannel(chName channel.Channel) *Channel {
	ch := &Channel{
		Stream: stream.NewStream(),
	}
	ch.Stream.Name = string(chName)
	ch.setTimeTable()
	ch.scheduleChannel()
	return ch
}

func (c *Channel) scheduleChannel() {
	for _, v := range c.TimeTables {
		cron.NewDailyJob(int8(v.StartAt.Hour()), int8(v.StartAt.Minute()), int8(v.StartAt.Second()), func(t time.Time) {
			c.Stream.SetDeadline(&v.StopAt)
			if err := c.Stream.StartViaSelenium(true); err != nil {
				c.Stream.Logger.FatalLogger.Fatalf("cant start via selenium %s, err: %s\n", c.Stream.Name, err.Error())
			}
		})
	}
}

func (c *Channel) setTimeTable() {
	var timeTable []*TimeTable

	periods := channel.TimeTable[channel.Channel(c.Name)]

	for _, v := range periods {
		timeTable = append(timeTable, c.createTimeTable(v[0], v[1]))
	}

	c.TimeTables = timeTable
}

func (c *Channel) createTimeTable(startAt, stopAt string) *TimeTable {
	t := time.Now()
	tt := time.Now()
	format := "15:04:05"

	start, err := time.Parse(format, startAt)
	if err != nil {
		c.Stream.Logger.FatalLogger.Fatalf("cant parse time %s\n", err.Error())
	}

	stop, err := time.Parse(format, stopAt)
	if err != nil {
		c.Stream.Logger.FatalLogger.Fatalf("cant parse time %s\n", err.Error())
	}

	if start.After(stop) {
		tt = tt.Add(time.Hour * 24)
	}

	start = time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), start.Second(), 0, time.UTC)
	stop = time.Date(tt.Year(), tt.Month(), tt.Day(), stop.Hour(), stop.Minute(), stop.Second(), 0, time.UTC)

	return &TimeTable{StartAt: start, StopAt: stop}
}

func (c *Channel) startStream() {
	manifest, err := selenium.GetManifest(channel.Channel(c.Stream.Name))
	if err != nil {
		c.Stream.Logger.FatalLogger.Fatalf("cant start selenium %s\n", err.Error())
	}

	c.Stream.Manifest = manifest
	if err := c.Stream.Start(); err != nil {
		c.Stream.Logger.FatalLogger.Fatalf("cant start stream %s\n", err.Error())
	}
}