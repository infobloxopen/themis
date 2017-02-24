package main

import (
	"fmt"
	"math"
	"sync"

	"github.com/infobloxopen/policy-box/pdp"
)

type Content struct {
	Id   string
	Data interface{}
}

type Queue struct {
	Lock          *sync.Mutex
	AutoIncrement int32
	Items         map[int32]interface{}
}

func NewQueue() *Queue {
	return &Queue{&sync.Mutex{}, -1, make(map[int32]interface{})}
}

func (q *Queue) rawResetAutoIncrement() bool {
	if q.AutoIncrement < math.MaxInt32 || len(q.Items) > 0 {
		return false
	}

	q.AutoIncrement = -1
	return true
}

func (q *Queue) Put(v interface{}) (int32, error) {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	if q.AutoIncrement >= math.MaxInt32 {
		return q.AutoIncrement,
			fmt.Errorf("Can't enqueue the policies as autoincrement has reached its maximum of %d", q.AutoIncrement)
	}

	q.AutoIncrement++
	q.Items[q.AutoIncrement] = v

	return q.AutoIncrement, nil
}

func (q *Queue) rawGet(id int32) ([]byte, error) {
	v, ok := q.Items[id]
	if !ok {
		return nil, fmt.Errorf("No data with id %d has been uploaded", id)
	}

	d, ok := v.([]byte)
	if !ok {
		return nil, fmt.Errorf("Expected bytes array with id %d but got %T", id, v)
	}

	return d, nil
}

func (q *Queue) Get(id int32) ([]byte, error) {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	return q.rawGet(id)
}

func (q *Queue) Replace(id int32, v interface{}) error {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	_, err := q.rawGet(id)
	if err != nil {
		return err
	}

	q.Items[id] = v
	return nil
}

func (q *Queue) rawGetPolicies(id int32) (pdp.EvaluableType, error) {
	v, ok := q.Items[id]
	if !ok {
		return nil, fmt.Errorf("No policies with id %d has been uploaded", id)
	}

	p, ok := v.(pdp.EvaluableType)
	if !ok {
		return nil, fmt.Errorf("Expected policy or policy set with id %d but got %T", id, v)
	}

	return p, nil
}

func (q *Queue) rawGetInclude(id int32) (Content, error) {
	v, ok := q.Items[id]
	if !ok {
		return Content{}, fmt.Errorf("No item with id %d has been uploaded", id)
	}

	c, ok := v.(Content)
	if !ok {
		return Content{}, fmt.Errorf("Expected content with id %d but got %T", id, v)
	}

	return c, nil
}

func (q *Queue) GetIncludes(ids []int32) (map[string]interface{}, error) {
	r := make(map[string]interface{})

	q.Lock.Lock()
	defer q.Lock.Unlock()

	for _, id := range ids {
		c, err := q.rawGetInclude(id)
		if err != nil {
			return nil, err
		}

		r[c.Id] = c.Data
	}

	return r, nil
}

func (q *Queue) PopIncludes(ids []int32) int {
	if ids == nil {
		return 0
	}

	q.Lock.Lock()
	defer q.Lock.Unlock()

	count := 0
	for _, id := range ids {
		_, err := q.rawGetInclude(id)
		if err != nil {
			continue
		}

		delete(q.Items, id)
		count++
	}

	return count
}
