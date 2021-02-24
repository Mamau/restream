package channel

import "sort"

type Pattern struct {
	Scheme  string
	Attempt int
}

type byAttempt []*Pattern

func GetPattern(list []*Pattern) *Pattern {
	sort.Sort(byAttempt(list))
	list[0].Attempt++
	return list[0]
}

func (s byAttempt) Len() int {
	return len(s)
}
func (s byAttempt) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byAttempt) Less(i, j int) bool {
	return s[i].Attempt <= s[j].Attempt
}
