package adlist

type ListNode struct {
	prev  *ListNode
	next  *ListNode
	value interface{}
}

func (ln *ListNode) PrevNode() *ListNode {
	return ln.prev
}

func (ln *ListNode) NextNode() *ListNode {
	return ln.next
}

func (ln *ListNode) NodeValue() interface{} {
	return ln.value
}

type List struct {
	head  *ListNode
	tail  *ListNode
	dup   func(interface{}) interface{}
	match func(interface{}, interface{}) bool
	len   uint32
}

type ListIter struct {
	next      *ListNode
	startHead bool
}

func (l *List) Length() uint32 {
	return l.len
}

func (l *List) First() *ListNode {
	return l.head
}

func (l *List) Last() *ListNode {
	return l.tail
}

// Empty list, remove all the elements from the list without destroying the list itself.
func (l *List) Empty() {
	l.head = nil
	l.tail = nil
	l.len = 0
}

func (l *List) AddNodeHead(value interface{}) {
	ln := &ListNode{
		value: value,
		next:  nil,
		prev:  nil,
	}
	if l.len == 0 {
		l.head = ln
		l.tail = ln
	} else {
		ln.next = l.head
		l.head.prev = ln
		l.head = ln
	}
	l.len++
}

func (l *List) AddNodeTail(value interface{}) {
	ln := &ListNode{
		value: value,
		next:  nil,
		prev:  nil,
	}
	if l.len == 0 {
		l.head = ln
		l.tail = ln
	} else {
		ln.prev = l.tail
		l.tail.prev = ln
		l.tail = ln
	}
	l.len++
}

func (l *List) DelNode(ln *ListNode) {
	if ln.prev != nil {
		ln.prev.next = ln.next
	} else {
		l.head = ln.next
	}
	if ln.next != nil {
		ln.next.prev = ln.prev
	} else {
		l.tail = ln.prev
	}
	l.len--
}

func (l *List) Index(idx int32) *ListNode {
	var n *ListNode
	if idx < 0 {
		idx = (-idx) - 1
		n = l.tail
		for idx > 0 && n != nil {
			n = n.prev
			idx--
		}
	} else {
		n = l.head
		for idx > 0 && n != nil {
			n = n.next
			idx--
		}
	}
	return n
}

func Create() *List {
	return &List{
		head: nil,
		tail: nil,
		len:  0,
		dup:  nil,
	}
}

func (l *List) Iterator(startHead bool) *ListIter {
	iter := &ListIter{
		startHead: startHead,
	}
	if startHead {
		iter.next = l.head
	} else {
		iter.next = l.tail
	}
	return iter
}

func (l *List) SearchKey(key interface{}) *ListNode {
	var iter *ListIter = l.Iterator(true)
	for ln := iter.Next(); ln != nil; ln = iter.Next() {
		if l.match != nil {
			if l.match(ln.value, key) {
				return ln
			}
		} else {
			if key == ln.value {
				return ln
			}
		}
	}
	return nil
}

func (li *ListIter) Next() *ListNode {
	ln := li.next
	if ln != nil {
		if li.startHead {
			li.next = ln.next
		} else {
			li.next = ln.prev
		}
	}
	return ln
}
