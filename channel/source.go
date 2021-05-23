package channel

type Source struct {
	Repository *SourceRepository
}

type Channel struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
	Uri  string `json:"uri"`
}

func NewSource() *Source {
	return &Source{
		Repository: NewSourceRepository(),
	}
}

func (s *Source) GetChannels() *[]Channel {
	r := s.Repository.GetChannels()
	return &r
}

func (s *Source) GetManifestByName(name string) *Channel {
	return s.Repository.GetManifestByName(name)
}
