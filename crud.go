package gosubscribe

import "fmt"

// Subscribe subscribes a user to a list of mappers.
func (user *User) Subscribe(mappers []Mapper) {
	for _, mapper := range mappers {
		DB.Table("subscriptions").Create(&Subscription{user.ID, mapper.ID})
	}
}

// Unsubscribe unsubscribes a user from a list of mappers.
func (user *User) Unsubscribe(mappers []Mapper) {
	for _, mapper := range mappers {
		DB.Delete(&Subscription{user.ID, mapper.ID})
	}
}

// Purge unsubscribes a user from all mappers.
func (user *User) Purge() {
	DB.Table("subscriptions").Where("user_id = ?", user.ID).Delete(Subscription{})
}

// ListSubscribed gets all mappers that a user is subscribed to.
func (user *User) ListSubscribed() []Mapper {
	var mappers []Mapper
	DB.Table("subscriptions").Joins(
		"inner join mappers on subscriptions.mapper_id = mappers.id",
	).Select("mappers.id, mappers.username").Find(&mappers)
	return mappers
}

// Count gets the number of users that are subscribed to a mapper.
func (mapper *Mapper) Count() uint {
	var count uint
	DB.Model(&Subscription{}).Where("mapper_id = ?", mapper.ID).Count(&count)
	return count
}

// Count gets the subscriber counts for a list of mappers.
func Count(mappers []Mapper) map[Mapper]uint {
	counts := make(map[Mapper]uint)
	for _, mapper := range mappers {
		counts[mapper] = mapper.Count()
	}
	return counts
}

// Top gets the n mappers with the most subscribers and their subscription counts.
func Top(n int) map[Mapper]uint {
	counts := make(map[Mapper]uint)
	var subs []Subscription
	// TODO: Figure out how to properly build this query.
	DB.Raw(fmt.Sprintf("SELECT mapper_id, COUNT(*) FROM subscriptions GROUP BY mapper_id ORDER BY COUNT DESC LIMIT %d", n)).Find(&subs)
	for _, sub := range subs {
		var mapper Mapper
		DB.Where("id = ?", sub.MapperID).First(&mapper)
		counts[mapper] = mapper.Count()
	}
	return counts
}
