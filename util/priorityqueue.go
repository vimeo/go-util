package util

import (
    "sync"
)

// Thread-Safe Priority Queue.
// Priorities are integers only. Works best when the maximum priority is
// relatively small.
type PriorityQueue struct {
    queues []*Queue
    waiting []int
    max int
    waitLimit int
    top int
    total int
    mutex sync.Mutex
}

// Create a new empty PriorityQueue.
// waitLimit gives other queues priority over the top queue if they have been
// waiting for this many reads. A value of 0 allows unlimited reading from the
// top priority queue before reading from lower priority queues.
func NewPriorityQueueWithWaitLimit(maxPriority int, waitLimit int) *PriorityQueue {
    maxPriority = MaxInt(0, maxPriority)
    waitLimit = MaxInt(0, waitLimit)
    pq := new(PriorityQueue)
    pq.max = maxPriority
    pq.waitLimit = waitLimit
    pq.queues = make([]*Queue, maxPriority + 1)
    for i := range pq.queues {
        pq.queues[i] = NewQueue()
    }
    pq.waiting = make([]int, maxPriority + 1)
    return pq
}

// Create a new empty PriorityQueue.
func NewPriorityQueue(maxPriority int) *PriorityQueue {
    return NewPriorityQueueWithWaitLimit(maxPriority, 0)
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
        this.waiting[priority] = 0
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

func (this *PriorityQueue) nextQueue() int {
    if this.top == 0 || this.waitLimit == 0 {
        return this.top
    }

    next := this.top
    for i := this.top; i >= 0; i-- {
        if this.queues[i].Len() > 0 && this.waiting[i] >= this.waitLimit {
            next = i
            break
        }
    }

    for i, q := range this.queues {
        if i != next && q.Len() > 0 {
            this.waiting[i]++
        }
    }
    this.waiting[next] = 0

    return next
}

// Remove the first highest priority item from the PriorityQueue.
func (this *PriorityQueue) Remove() interface{} {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if this.total == 0 {
        return nil
    }

    next := this.nextQueue()
    ret  := this.queues[next].Remove()

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

    next := this.nextQueue()
    ret  := this.queues[next].RemoveWait()

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

    this.waiting[priority] = 0
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

    this.waiting[priority] = 0
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

// Discard all items in the PriorityQueue
func (this *PriorityQueue) Clear() {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    for i, q := range this.queues {
        q.Clear()
        this.waiting[i] = 0
    }
    this.top   = 0
    this.total = 0
}

// Close the PriorityQueue
func (this *PriorityQueue) Close() {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    for i, q := range this.queues {
        q.Close()
        this.waiting[i] = 0
    }
    this.top   = 0
    this.total = 0
}
