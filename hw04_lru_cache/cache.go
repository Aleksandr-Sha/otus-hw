package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func (l lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		item.Value = value
		l.queue.MoveToFront(item)
		return true
	} else {
		if l.queue.Len() < l.capacity {
			newItem := l.queue.PushFront(value)
			l.items[key] = newItem
		} else {
			l.queue.Remove(l.queue.Back())
			newItem := l.queue.PushFront(value)
			l.items[key] = newItem
		}

		return false
	}
}

func (l lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)
		return item.Value, true
	}

	return nil, false
}

// Подумать как сделать o(1)
func (l lruCache) Clear() {
	for key, item := range l.items {
		delete(l.items, key)
		l.queue.Remove(item)
	}
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
