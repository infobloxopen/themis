package main

import (
	"math"
	"sync"

	"github.com/google/uuid"

	"github.com/infobloxopen/themis/pdp"
)

type Item struct {
	policy bool
	id     string

	fromTag *uuid.UUID
	toTag   *uuid.UUID

	p  *pdp.PolicyStorage
	pt *pdp.PolicyStorageTransaction

	c  *pdp.LocalContent
	ct *pdp.LocalContentStorageTransaction
}

type Queue struct {
	sync.Mutex

	idx   int32
	items map[int32]*Item
}

func NewQueue() *Queue {
	return &Queue{
		idx:   -1,
		items: make(map[int32]*Item)}
}

func NewPolicyItem(fromTag, toTag *uuid.UUID) *Item {
	return &Item{
		policy:  true,
		fromTag: fromTag,
		toTag:   toTag}
}

func NewContentItem(id string, fromTag, toTag *uuid.UUID) *Item {
	return &Item{
		policy:  false,
		id:      id,
		fromTag: fromTag,
		toTag:   toTag}
}

func (q *Queue) Push(item *Item) (int32, error) {
	q.Lock()
	defer q.Unlock()

	if q.idx >= math.MaxInt32 {
		return q.idx, newQueueOverflowError(q.idx)
	}

	q.idx++
	q.items[q.idx] = item

	return q.idx, nil
}

func (q *Queue) Pop(idx int32) (*Item, bool) {
	q.Lock()
	defer q.Unlock()

	item, ok := q.items[idx]
	if ok {
		delete(q.items, idx)
	}

	return item, ok
}
