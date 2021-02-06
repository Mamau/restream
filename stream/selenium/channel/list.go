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
		{"11:00:00", "14:00:00"},
		{"18:00:00", "03:00:00"},
	},
	FIRST: {
		{"08:00:00", "10:00:00"},
	},
}
