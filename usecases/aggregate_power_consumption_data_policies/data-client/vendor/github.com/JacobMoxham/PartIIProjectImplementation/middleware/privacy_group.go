package middleware

// PrivacyGroup a struct which contain a data structure of RequesterID's which we can Add to and Remove from
type PrivacyGroup struct {
	name    string
	members map[string]bool
}

func NewPrivacyGroup(name string) *PrivacyGroup {
	return &PrivacyGroup{
		name:    name,
		members: make(map[string]bool),
	}
}

func (pg *PrivacyGroup) Name() string {
	return pg.name
}

func (pg *PrivacyGroup) Add(id string) {
	pg.members[id] = true
}

func (pg *PrivacyGroup) AddMany(ids []string) {
	for _, id := range ids {
		pg.members[id] = true
	}
}

func (pg *PrivacyGroup) Remove(id string) error {
	_, ok := pg.members[id]
	if ok {
		delete(pg.members, id)
	}
	return nil
}

func (pg *PrivacyGroup) contains(id string) bool {
	in, ok := pg.members[id]
	return in && ok
}
