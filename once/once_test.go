package once_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vimeo/go-util/once"
)

func TestThunder(t *testing.T) {
	var tgt uint32 = 2
	var calls uint32
	var wg sync.WaitGroup
	add := func() error {
		t.Log("called")
		time.Sleep(time.Millisecond)
		if atomic.AddUint32(&calls, 1) != tgt {
			return errors.New("errored")
		}
		return nil
	}
	o := once.New()

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			err := o.Do(context.Background(), add)
			t.Logf("%d:\t%v", i, err)
			wg.Done()
		}(i)
	}
	wg.Wait()

	if calls != tgt {
		t.Fatalf("calls = %d", calls)
	}
}

// Make sure that panicing actually returns an error.
func TestPanic(t *testing.T) {
	o := once.New()
	err := o.Do(context.Background(), func() error {
		panic("panic'd")
	})
	t.Log(err)
	if err == nil {
		t.Fatalf("wanted error, got %v", err)
	}
}
