package channel

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/grafov/m3u8"
	"github.com/mamau/restream/helpers"
	"github.com/mamau/restream/storage"
	"strings"
	"time"
)

const ChannelsKey = "channels"

var ctx = context.Background()

type SourceRepository struct {
	redis *redis.Client
}

func NewSourceRepository() *SourceRepository {
	return &SourceRepository{
		redis: storage.NewRedis(),
	}
}

func (s *SourceRepository) GetManifestByName(name string) *Channel {
	list := s.GetChannels()

	for _, item := range list {
		fmt.Println(item.Slug)
		if item.Slug == strings.ToLower(name) {
			return &item
		}
	}
	return nil
}

func (s *SourceRepository) GetChannels() []Channel {
	var listOfChannels []Channel
	list, _ := s.redis.Get(ctx, ChannelsKey).Result()

	if list == "" {
		data := fetchSegments()
		b, errMarshal := json.Marshal(data)
		if errMarshal != nil {
			storage.GetLogger().Fatal(errMarshal)
		}
		list = string(b)
		s.redis.Set(ctx, ChannelsKey, b, 24*time.Hour)
	}

	var segment []m3u8.MediaSegment
	err := json.Unmarshal([]byte(list), &segment)
	if err != nil {
		storage.GetLogger().Fatal(err)
	}

	for _, v := range segment {
		listOfChannels = append(listOfChannels, Channel{Name: v.Title, Uri: v.URI, Slug: helpers.CyrillicToLatin(v.Title)})
	}

	return listOfChannels
}
