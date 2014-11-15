package util

import (
    "container/list"
    "sync"
)

// Simple Thread-Safe FIFO Queue
type Queue struct {
    list *list.List
    mutex sync.Mutex
}

// Create a new empty Queue
func NewQueue() *Queue {
    q := new(Queue)
    q.list = list.New()
    return q
}

// Add an item to the end of the Queue
func (this *Queue) Add(v interface{}) {
    this.mutex.Lock()
    this.list.PushBack(v)
    this.mutex.Unlock()
}

// Remove the first item in the Queue
func (this *Queue) Remove() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    e := this.list.Front()
    if e == nil {
        return nil
    }
    return this.list.Remove(e)
}

// Get the number of items in the Queue
func (this *Queue) Len() int {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    return this.list.Len()
}

// Discard all items in the Queue
func (this *Queue) Clear() {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    this.list.Init()
}

const (
    LimitStrategyReject = iota
    LimitStrategyCycle
)

type LimitQueue struct {
    Queue
    maxItems int
    strategy int
}

// Create a new empty LimitQueue
func NewLimitQueue(maxItems int, strategy int) *LimitQueue {
    // validate params
    if maxItems <= 0 {
        return nil
    }
    switch strategy {
    case LimitStrategyReject, LimitStrategyCycle:
    default:
        return nil
    }

    q := new(LimitQueue)
    q.list     = list.New()
    q.maxItems = maxItems
    q.strategy = strategy

    return q
}

// Add an item to the end of the Queue.
// If the queue is full, use the limit strategy to determine whether to reject
// the new item or to remove the oldest item to make room.
// For LimitStrategyReject, returns whether the item was added.
// For LimitStrategyCycle, returns whether the queue was NOT full.
func (this *LimitQueue) Add(v interface{}) bool {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if this.list.Len() >= this.maxItems && this.strategy == LimitStrategyReject {
        return false
    }
    space := true
    for this.list.Len() >= this.maxItems {
        space = false
        e := this.list.Front()
        if e == nil {
            panic("empty queue. this is a bug.")
        }
        this.list.Remove(e)
    }
    this.list.PushBack(v)
    return space
}
