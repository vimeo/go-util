package util

import (
    "sync"
)

// Thread-Safe Priority Queue.
// Priorities are integers only. Works best when the maximum priority is
// relatively small.
type PriorityQueue struct {
    queues []*Queue
    max int
    top int
    total int
    mutex sync.Mutex
}

// Create a new empty PriorityQueue.
func NewPriorityQueue(maxPriority int) *PriorityQueue {
    if maxPriority < 0 {
        maxPriority = 0
    }
    pq := new(PriorityQueue)
    pq.max = maxPriority
    pq.queues = make([]*Queue, maxPriority + 1)
    for i := range pq.queues {
        pq.queues[i] = NewQueue()
    }
    return pq
}

// Add an item to the PriorityQueue with a specified priority.
// If the priority is out-of-range it is clipped.
func (this *PriorityQueue) Add(v interface{}, priority int) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    priority = ClipInt(priority, 0, this.max)

    this.queues[priority].Add(v)
    this.total++
    if priority > this.top {
        this.top = priority
    }
}

func (this *PriorityQueue) updateRemoval() {
    this.total--
    if this.total == 0 {
        this.top = 0
    } else {
        for this.top > 0 && this.queues[this.top].Len() == 0 {
            this.top--
        }
    }
}

// Remove the first highest priority item from the PriorityQueue.
func (this *PriorityQueue) Remove() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if this.total == 0 {
        return nil
    }

    ret := this.queues[this.top].Remove()

    this.updateRemoval()

    return ret
}

// Remove the first highest priority item from the PriorityQueue.
// Blocks until an item is available.
func (this *PriorityQueue) RemoveWait() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if this.total == 0 {
        return nil
    }

    ret := this.queues[this.top].RemoveWait()

    this.updateRemoval()

    return ret
}

// Remove the first item of a specific priority from the PriorityQueue.
func (this *PriorityQueue) RemoveP(priority int) interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if priority < 0 || priority > this.max {
        return nil
    }

    if this.total == 0 {
        return nil
    }

    ret := this.queues[priority].Remove()

    this.updateRemoval()

    return ret
}

// Remove the first item of a specific priority from the PriorityQueue.
// Blocks until an item is available.
func (this *PriorityQueue) RemovePWait(priority int) interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if priority < 0 || priority > this.max {
        return nil
    }

    if this.total == 0 {
        return nil
    }

    ret := this.queues[priority].RemoveWait()

    this.updateRemoval()

    return ret
}

// Retrieve (but do not remove) the first highest priority item from the PriorityQueue.
func (this *PriorityQueue) Peek() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    return this.queues[this.top].Peek()
}

// Retrieve (but do not remove) the first highest priority item from the PriorityQueue.
// Blocks until an item is available.
func (this *PriorityQueue) PeekWait() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    return this.queues[this.top].PeekWait()
}

// Retrieve (but do not remove) the first item of a specific priority from the PriorityQueue.
func (this *PriorityQueue) PeekP(priority int) interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if priority < 0 || priority > this.max {
        return nil
    }

    return this.queues[priority].Peek()
}

// Retrieve (but do not remove) the first item of a specific priority from the PriorityQueue.
// Blocks until an item is available.
func (this *PriorityQueue) PeekPWait(priority int) interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if priority < 0 || priority > this.max {
        return nil
    }

    return this.queues[priority].PeekWait()
}

// Get the number of items in the PriorityQueue.
func (this *PriorityQueue) Len() int {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    return this.total
}

// Get the number of items of each priority in the PriorityQueue.
func (this *PriorityQueue) Lens() map[int]int {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    ret := make(map[int]int)

    for i, q := range this.queues {
        if q.Len() > 0 {
            ret[i] = q.Len()
        }
    }

    return ret
}

// Discard all items in the Queue
func (this *PriorityQueue) Clear() {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    for _, q := range this.queues {
        q.Clear()
    }
    this.top   = 0
    this.total = 0
}
