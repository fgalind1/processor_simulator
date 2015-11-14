package channel

type Channel struct {
	channel chan interface{}
	locks   chan bool
}

func New(capacity uint32) Channel {
	return Channel{
		channel: make(chan interface{}, capacity),
		locks:   make(chan bool, capacity),
	}
}

func (this Channel) Add(value interface{}) {
	this.locks <- true
	this.channel <- value
}

func (this Channel) Release() {
	<-this.locks
}

func (this Channel) Channel() chan interface{} {
	return this.channel
}
