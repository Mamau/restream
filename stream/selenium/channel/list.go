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
