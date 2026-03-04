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
	valueForItem := newCacheItem(key, value)

	if item, ok := l.items[key]; ok {
		l.updateItem(item, valueForItem)
		return true
	} else {
		l.addItem(valueForItem)
		l.removeLastIfExceededCapacity()
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

func (l *lruCache) addItem(valueForItem cacheItem) {
	newItem := l.queue.PushFront(valueForItem)
	l.items[valueForItem.key] = newItem
}

func (l *lruCache) updateItem(item *ListItem, valueForItem cacheItem) {
	item.Value = valueForItem
	l.queue.MoveToFront(item)
}

func (l *lruCache) removeLastIfExceededCapacity() {
	if l.isExceededCapacity() {
		back := l.queue.Back()
		l.queue.Remove(back)
		delete(l.items, back.Value.(cacheItem).key)
	}
}

func (l *lruCache) isExceededCapacity() bool {
	return l.queue.Len() > l.capacity
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}
