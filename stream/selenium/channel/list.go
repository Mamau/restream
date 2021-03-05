package channel

type Channel string

const (
	TNT   Channel = "tnt"
	FIRST Channel = "1tv"
	MATCH Channel = "match"
)

var ChUrls = map[Channel]string{
	TNT:   "https://tnt-online.ru/live",
	FIRST: "https://www.1tv.ru/live",
	MATCH: "https://matchtv.ru/on-air",
}

var TimeTable = map[Channel][][]string{
	TNT: {
		{"11:00:00", "04:00:00"},
	},
	FIRST: {
		{"07:00:00", "13:40:00"},
	},
}

var ChanManifestPatterns = map[Channel][]*Pattern{
	MATCH: {
		&Pattern{
			Scheme:  `https:\/\/live(.)+(\.m3u8)`,
			Attempt: 0,
		},
	},
	TNT: {
		&Pattern{
			Scheme:  `https:\/\/live(.)+(\.m3u8)`,
			Attempt: 0,
		},
		&Pattern{
			Scheme:  `https:\/\/matchtv(.)+(\.m3u8)`,
			Attempt: 0,
		},
		&Pattern{
			Scheme:  `https:\/\/bl.zxz.su(.)+(\.m3u8)`,
			Attempt: 0,
		},
	},
	FIRST: {
		&Pattern{
			Scheme:  `https:\/\/edge(.)+(\.mpd\?[a-z]{1}\=[0-9]+)`,
			Attempt: 0,
		},
		&Pattern{
			Scheme:  `https:\/\/cdn2.1internet.tv(.)+(\.mpd\?[a-z]{1}\=[0-9]+)`,
			Attempt: 0,
		},
		&Pattern{
			Scheme:  `https:\/\/[a-z\/0-9\._-]+(\.mpd\?[a-z]{1}\=[0-9]+)`,
			Attempt: 0,
		},
	},
}
