package scheduler

import (
	"fmt"
	"github.com/mamau/restream/channel"
	"time"

	"github.com/mamau/restream/storage"
	"github.com/mamau/restream/stream"
	"github.com/rk/go-cron"
)

type TimeTable struct {
	StartAt time.Time
	StopAt  time.Time
}

type Channel struct {
	*stream.Stream
	Channel    channel.ChannelName
	TimeTables []*TimeTable
}

func CreateScheduledChannel(chName channel.ChannelName) *Channel {
	ch := &Channel{
		Stream: stream.NewStream(),
	}
	ch.Name = string(chName)
	ch.setTimeTable()
	ch.scheduleChannel()
	return ch
}

func (c *Channel) scheduleChannel() {
	for _, v := range c.TimeTables {
		cron.NewDailyJob(int8(v.StartAt.Hour()), int8(v.StartAt.Minute()), int8(v.StartAt.Second()), func(t time.Time) {
			fmt.Println("Start by cron")
			c.StartByIPTV()
		})

		cron.NewDailyJob(int8(v.StopAt.Hour()), int8(v.StopAt.Minute()), int8(v.StopAt.Second()), func(t time.Time) {
			fmt.Println("Stop by cron")
			c.Stop()
		})

		if time.Now().After(v.StartAt) && time.Now().Before(v.StopAt) {
			fmt.Printf("start %s immediately\n", c.Name)
			c.StartByIPTV()
		}
	}
}

func (c *Channel) setTimeTable() {
	var timeTable []*TimeTable

	periods := channel.TimeTable[channel.ChannelName(c.Name)]
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
		storage.GetLogger().Fatal(err)
	}

	stop, err := time.Parse(format, stopAt)
	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	if start.After(stop) {
		tt = tt.Add(time.Hour * 24)
	}

	start = time.Date(t.Year(), t.Month(), t.Day(), start.Hour(), start.Minute(), start.Second(), 0, time.Local)
	stop = time.Date(tt.Year(), tt.Month(), tt.Day(), stop.Hour(), stop.Minute(), stop.Second(), 0, time.Local)

	return &TimeTable{StartAt: start, StopAt: stop}
}
