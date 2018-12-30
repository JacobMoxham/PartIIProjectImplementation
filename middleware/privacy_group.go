package middleware

// PrivacyGroup a struct which contain a data structure of ID's which we can add to and remove from
type PrivacyGroup struct {
	Name    string
	Members map[string]bool
}

// TODO: look into creating types with pointer type for structs
func (pg *PrivacyGroup) add(id string) {
	pg.Members[id] = true
}

func (pg *PrivacyGroup) remove(id string) error {
	// TODO: stop the map from getting too large
	pg.Members[id] = false
	return nil
}

func (pg *PrivacyGroup) contains(id string) bool {
	in, ok := pg.Members[id]
	return in && ok
}
