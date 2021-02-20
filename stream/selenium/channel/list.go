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
		{"07:00:00", "10:30:00"},
	},
}
