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

func newCacheItem(key Key, value interface{}) cacheItem {
	return cacheItem{value, key}
}

func (l *lruCache) Set(key Key, value interface{}) bool {
	newValue := newCacheItem(key, value)

	if item, ok := l.items[key]; ok {
		item.Value = newValue
		l.queue.MoveToFront(item)
		return true
	} else {
		newItem := l.queue.PushFront(newValue)
		l.items[key] = newItem

		if l.queue.Len() > l.capacity {
			l.removeLast()
		}

		return false
	}
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)
		return item.Value.(cacheItem).value, true
	}

	return nil, false
}

func (l *lruCache) Clear() {
	clear(l.items)
	l.queue = NewList()
}

func (l *lruCache) removeLast() {
	back := l.queue.Back()
	l.queue.Remove(back)
	delete(l.items, back.Value.(cacheItem).key)
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
