package main

import (
	"sync"
	"time"
)

type rollingWindowCounter struct {
	Size uint //窗口大小，单位为秒
	Number uint //窗口切分数
	Max uint //限流大小
	Bucket []uint //小窗口计数
	StartTM int64 //开始时间
	Current int //当前小窗口
	mu sync.Mutex
}

func NewRollingWindow(size, number, max uint) *rollingWindowCounter {
	return &rollingWindowCounter{
		Size: size,
		Number: number,
		Max: max,
		Bucket: make([]uint, number),
		StartTM: time.Now().Unix(),
	}

}

func (r *rollingWindowCounter) tryAcquire() bool {
	tmNow := time.Now().Unix()
	offset := uint(tmNow - r.StartTM)
	if offset < r.Size {
		return true

	}

	number := (offset-r.Size) / r.bucketSize()
	r.Roll(number)
	if r.count() > r.Max {
		return false

	}

	curIndx := r.getCurrentBucket()
	r.Bucket[curIndx]++

	return true

}

func (r *rollingWindowCounter) Roll(number uint) {
	if number == 0 {
		return

	}

	var rollNum uint
	if number < r.Number {
		rollNum = number

	}

	for i := 0; i<int(rollNum); i++{
		r.Current = (r.Current+1) % int(r.Number)
		r.Bucket[r.Current] = 0

	}

	r.StartTM = r.StartTM + int64(number * (r.Size/r.Number))
}

//小窗口的大小
func (r *rollingWindowCounter) bucketSize() uint {
	return r.Size / r.Number

}

func (r *rollingWindowCounter) count() uint {
	var count uint = 0
	for i := 0; i<int(r.Number); i++ {
		count += r.Bucket[i]
	}

	return  count
}

func (r *rollingWindowCounter) getCurrentBucket() uint {
	tmNow := time.Now().Unix()
	return uint(tmNow / int64(r.bucketSize()) % int64(r.Number))

}

func main(){
	r := NewRollingWindow(10, 5, 5)
	r.tryAcquire()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		for i:=0; i<= 100; i++ {
			for j:=0; j<1; j++ {
				if r.tryAcquire() {
				}
			}
			//time.Sleep(time.Millisecond * 1429)
			time.Sleep(time.Second *6)
		}
	}()
	go func() {
		defer func() {
			wg.Done()
		}()
		for i:=0; i<= 100; i++ {
			for j:=0; j<1; j++ {
				if r.tryAcquire() {
				}
			}
			time.Sleep(time.Millisecond * 1429)
			//time.Sleep(time.Second *6)
		}
	}()
	wg.Wait()
}
