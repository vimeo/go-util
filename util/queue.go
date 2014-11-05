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
