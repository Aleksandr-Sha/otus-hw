package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len       int
	frontItem *ListItem
	backItem  *ListItem
}

func (l *list) Len() int {
	return l.len
}

func (l *list) Front() *ListItem {
	return l.frontItem
}

func (l *list) Back() *ListItem {
	return l.backItem
}

func (l *list) PushFront(v interface{}) *ListItem {
	newItem := &ListItem{Value: v, Next: nil, Prev: nil}

	if l.len == 0 {
		l.frontItem = newItem
		l.backItem = newItem
	} else {
		newItem.Next = l.frontItem
		l.frontItem.Prev = newItem
		l.frontItem = newItem
	}

	l.len++

	return newItem
}

func (l *list) PushBack(v interface{}) *ListItem {
	newItem := &ListItem{Value: v, Next: nil, Prev: nil}

	if l.len == 0 {
		l.frontItem = newItem
		l.backItem = newItem
	} else {
		newItem.Prev = l.backItem
		l.backItem.Next = newItem
		l.backItem = newItem
	}

	l.len++

	return newItem
}

func (l *list) Remove(i *ListItem) {
	if l.len == 1 {
		l.frontItem = nil
		l.backItem = nil
	} else if i == l.backItem {
		i.Prev.Next = nil
		l.backItem = i.Prev
	} else if i == l.frontItem {
		i.Next.Prev = nil
		l.frontItem = i.Next
	} else {
		i.Prev.Next = i.Next
		i.Next.Prev = i.Prev
	}

	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	if l.len == 1 || l.frontItem == i {
		return
	}

	if i == l.backItem {
		i.Prev.Next = nil
		l.backItem = i.Prev
	} else {
		i.Prev.Next = i.Next
		i.Next.Prev = i.Prev
	}

	i.Prev = nil
	i.Next = l.frontItem
	l.frontItem.Prev = i
	l.frontItem = i
}

func NewList() List {
	return new(list)
}
