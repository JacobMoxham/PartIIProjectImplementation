package middleware

// PrivacyGroup a struct which contain a data structure of RequesterID's which we can add to and remove from
type PrivacyGroup struct {
	Name    string
	Members map[string]bool
}

func (pg *PrivacyGroup) add(id string) {
	pg.Members[id] = true
}

func (pg *PrivacyGroup) remove(id string) error {
	_, ok := pg.Members[id]
	if ok {
		delete(pg.Members, id)
	}
	return nil
}

func (pg *PrivacyGroup) contains(id string) bool {
	in, ok := pg.Members[id]
	return in && ok
}
