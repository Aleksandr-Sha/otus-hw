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

type cacheItem struct {
	value interface{}
	key   Key
}

func (l lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		item.Value = cacheItem{value: value, key: key}
		l.queue.MoveToFront(item)
		return true
	} else {
		valueForItem := cacheItem{value: value, key: key}

		if l.queue.Len() < l.capacity {
			newItem := l.queue.PushFront(valueForItem)
			l.items[key] = newItem
		} else {
			back := l.queue.Back()

			l.queue.Remove(l.queue.Back())
			newItem := l.queue.PushFront(valueForItem)
			l.items[key] = newItem

			delete(l.items, back.Value.(cacheItem).key)
		}

		return false
	}
}

func (l lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)
		return item.Value.(cacheItem).value, true
	}

	return nil, false
}

func (l lruCache) Clear() {
	clear(l.items)
	l.queue = NewList()
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
