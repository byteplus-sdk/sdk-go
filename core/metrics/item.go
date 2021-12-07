package metrics

import (
	"fmt"
	"time"
)

const defaultItemExpireDuration = 10 * time.Second

type Item struct {
	tags string
	value float64
	expireTime time.Time
}

func NewItem(tags string, value float64) *Item {
	return NewItemWithExpire(tags, value, defaultItemExpireDuration)
}

func NewItemWithExpire(tags string, value float64, expireDuration time.Duration) *Item {
	realExpireDuration := expireDuration
	if expireDuration <= 0 {
		realExpireDuration = defaultItemExpireDuration
	}
	return &Item{
		tags:       tags,
		value:      value,
		expireTime: time.Now().Add(realExpireDuration),
	}
}

func (i *Item) toString() string {
	return fmt.Sprintf("Item{\"tags\":%s,\"Value\":%v}", i.tags, i.value)
}
