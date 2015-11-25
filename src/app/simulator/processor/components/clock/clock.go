package clock

import (
	"time"

	"app/logger"
)

type Clock struct {
	*clock
}

type clock struct {
	period      time.Duration
	ticker      *time.Ticker
	cycles      uint32
	paused      bool
	stopped     bool
	duration    time.Duration
	finishEvent func() bool
	tick        chan bool

	startTime time.Time
}

func New(period time.Duration, finishEvent func() bool) *Clock {
	return &Clock{
		&clock{
			period:      period,
			paused:      false,
			stopped:     false,
			cycles:      0,
			duration:    0,
			tick:        make(chan bool, 1),
			finishEvent: finishEvent,
		},
	}
}

func (this *Clock) Run() {
	this.clock.startTime = time.Now()
	this.clock.ticker = time.NewTicker(this.clock.period)

	for tickTime := range this.clock.ticker.C {
		// If event finish equals to true, then finish
		if this.finishEvent() {
			logger.Collect(" => Stopping clock...")
			this.Stop()
			break
		}

		// If paused wait until next tick
		if !this.clock.paused {

			// Update clock status
			this.clock.cycles += 1
			this.clock.duration = tickTime.Sub(this.clock.startTime)
			this.clock.tick <- true

			logger.Collect("\n-------- Cycle: %04d ------- (%04d ms)", this.clock.cycles, this.DurationMs())
		}
	}
}

func (this *Clock) Cycles() uint32 {
	return this.clock.cycles
}

func (this *Clock) DurationMs() uint32 {
	return uint32(this.clock.duration / 1000000)
}

func (this *Clock) Finished() bool {
	return this.clock.stopped
}

func (this *Clock) Stop() {
	this.clock.stopped = true
	this.clock.ticker.Stop()
	this.clock.tick <- true
}

func (this *Clock) Pause() {
	this.clock.paused = true
}

func (this *Clock) Continue() {
	this.clock.paused = false
}

func (this *Clock) Wait() {
	if !this.Finished() {
		<-this.clock.tick
	}
}
