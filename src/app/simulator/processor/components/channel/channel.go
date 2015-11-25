package channel

import (
	"app/logger"
)

const INFINITE = 1000

type Channel struct {
	channel  chan interface{}
	locks    chan bool
	isClosed bool
}

func New(capacity uint32) Channel {
	return Channel{
		channel:  make(chan interface{}, capacity),
		locks:    make(chan bool, capacity),
		isClosed: false,
	}
}

func (this Channel) Add(value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			logger.Collect("WARN: Recovered in %v", r)
		}
	}()
	if !this.isClosed {
		this.locks <- true
		this.channel <- value
	}
}

func (this Channel) Release() {
	<-this.locks
}

func (this Channel) Channel() chan interface{} {
	return this.channel
}

func (this Channel) Capacity() uint32 {
	return uint32(cap(this.channel))
}

func (this Channel) IsFull() bool {
	return len(this.channel) == cap(this.channel)
}

func (this Channel) Close() {
	this.isClosed = true
	close(this.channel)
	close(this.locks)
}
