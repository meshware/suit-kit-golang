package task

import (
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Logger for debugging
var logger = log.New(os.Stdout, "TIMER: ", log.LstdFlags)

// TimeTask defines a task with a name and execution time.
type TimeTask interface {
	GetName() string
	GetTime() int64
	Run()
}

// DelegateTask is a concrete TimeTask implementation.
type DelegateTask struct {
	name     string
	time     int64
	runnable func()
}

func NewDelegateTask(name string, time int64, runnable func()) *DelegateTask {
	return &DelegateTask{name: name, time: time, runnable: runnable}
}

func (t *DelegateTask) GetName() string {
	return t.name
}

func (t *DelegateTask) GetTime() int64 {
	return t.time
}

func (t *DelegateTask) Run() {
	if t.runnable != nil {
		t.runnable()
	}
}

// Timeout defines methods for task lifecycle management.
type Timeout interface {
	IsExpired() bool
	IsCancelled() bool
	Cancel() bool
}

// Timer defines the scheduler's API.
type Timer interface {
	Add(name string, time int64, runnable func()) Timeout
	Delay(name string, delay int64, runnable func()) Timeout
	AddTask(task TimeTask) Timeout
}

// TimeSlot holds tasks for a specific time.
type TimeSlot struct {
	expiration int64
	tasks      []*TimeWork
	mu         sync.Mutex
}

func newTimeSlot() *TimeSlot {
	return &TimeSlot{
		expiration: -1,
		tasks:      make([]*TimeWork, 0),
	}
}

func (s *TimeSlot) add(timeWork *TimeWork, expire int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.expiration == -1 {
		s.expiration = expire
	}
	s.tasks = append(s.tasks, timeWork)
	return s.expiration == expire
}

func (s *TimeSlot) remove(timeWork *TimeWork) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, tw := range s.tasks {
		if tw == timeWork {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			break
		}
	}
	if len(s.tasks) == 0 {
		s.expiration = -1
	}
}

func (s *TimeSlot) flush(consumer func(*TimeWork)) {
	s.mu.Lock()
	tasks := s.tasks
	s.tasks = make([]*TimeWork, 0)
	s.expiration = -1
	s.mu.Unlock()
	for _, tw := range tasks {
		consumer(tw)
	}
}

// TimeWheel manages a circular array of time slots.
type TimeWheel struct {
	tickTime  int64
	ticks     int
	duration  int64
	now       int64
	index     int
	timeSlots []*TimeSlot
	queue     chan *TimeSlot
}

func newTimeWheel(tickTime int64, ticks int, now int64) *TimeWheel {
	tw := &TimeWheel{
		tickTime:  tickTime,
		ticks:     ticks,
		duration:  tickTime * int64(ticks),
		now:       now - (now % tickTime),
		index:     0,
		timeSlots: make([]*TimeSlot, ticks),
		queue:     make(chan *TimeSlot, ticks),
	}
	for i := 0; i < ticks; i++ {
		tw.timeSlots[i] = newTimeSlot()
	}
	return tw
}

func (tw *TimeWheel) getLeastOneTick(t int64) int64 {
	result := time.Now().UnixMilli() + tw.tickTime
	if t > result {
		return t
	}
	return result
}

func (tw *TimeWheel) add(timeWork *TimeWork) bool {
	currentTime := time.Now().UnixMilli()
	timeDiff := timeWork.time - currentTime
	logger.Printf("Adding task %s: scheduled=%d, current=%d, diff=%d", timeWork.name, timeWork.time, currentTime, timeDiff)
	if timeDiff <= 0 {
		logger.Printf("Task %s is already expired, scheduling immediately", timeWork.name)
		return false // Task is already due
	}
	slotTime := timeWork.time - (timeWork.time % tw.tickTime)
	count := (slotTime - tw.now) / tw.tickTime
	if count < 0 || count >= int64(tw.ticks) {
		logger.Printf("Task %s time %d is beyond wheel duration, rejecting", timeWork.name, timeWork.time)
		return false
	}
	slotIndex := (int(count) + tw.index) % tw.ticks
	timeSlot := tw.timeSlots[slotIndex]
	if timeSlot.add(timeWork, slotTime) {
		logger.Printf("Task %s added to slot %d with expiration %d", timeWork.name, slotIndex, slotTime)
		tw.queue <- timeSlot
	}
	return true
}

func (tw *TimeWheel) advance(timestamp int64) {
	if timestamp >= tw.now+tw.tickTime {
		tw.now = timestamp - (timestamp % tw.tickTime)
		tw.index = int((tw.now / tw.tickTime) % int64(tw.ticks))
		//logger.Printf("Advancing time wheel to %d, index %d", tw.now, tw.index)
	}
}

// TimeWork represents a scheduled task.
type TimeWork struct {
	name        string
	time        int64
	runnable    func()
	afterRun    func(*TimeWork)
	afterCancel func(*TimeWork)
	timeSlot    *TimeSlot
	state       int32 // 0: INIT, 1: CANCELLED, 2: EXPIRED
}

const (
	INIT      = 0
	CANCELLED = 1
	EXPIRED   = 2
)

func newTimeWork(name string, time int64, runnable func(), afterRun, afterCancel func(*TimeWork)) *TimeWork {
	return &TimeWork{
		name:        name,
		time:        time,
		runnable:    runnable,
		afterRun:    afterRun,
		afterCancel: afterCancel,
		state:       INIT,
	}
}

func (tw *TimeWork) Run() {
	currentTime := time.Now().UnixMilli()
	if tw.time > currentTime {
		logger.Printf("Task %s not yet due: scheduled=%d, current=%d, delaying", tw.name, tw.time, currentTime)
		time.Sleep(time.Duration(tw.time-currentTime) * time.Millisecond)
	}
	if atomic.CompareAndSwapInt32(&tw.state, INIT, EXPIRED) {
		logger.Printf("Executing task %s at %d", tw.name, time.Now().UnixMilli())
		tw.runnable()
		if tw.afterRun != nil {
			tw.afterRun(tw)
		}
	} else {
		logger.Printf("Task %s skipped: state=%d", tw.name, tw.state)
	}
}

func (tw *TimeWork) IsExpired() bool {
	return atomic.LoadInt32(&tw.state) == EXPIRED
}

func (tw *TimeWork) IsCancelled() bool {
	return atomic.LoadInt32(&tw.state) == CANCELLED
}

func (tw *TimeWork) Cancel() bool {
	if atomic.CompareAndSwapInt32(&tw.state, INIT, CANCELLED) {
		logger.Printf("Cancelled task %s", tw.name)
		if tw.afterCancel != nil {
			tw.afterCancel(tw)
		}
		return true
	}
	return false
}

func (tw *TimeWork) remove() {
	if tw.timeSlot != nil {
		tw.timeSlot.remove(tw)
		tw.timeSlot = nil
	}
}

// TimeScheduler implements the Timer interface.
type TimeScheduler struct {
	prefix        string
	workerThreads int
	timeWheel     *TimeWheel
	flying        chan *TimeWork
	working       chan *TimeWork
	cancels       chan *TimeWork
	tasks         int64
	started       int32
	ctx           context.Context
	cancelFunc    context.CancelFunc
}

func NewTimeScheduler(prefix string, tickTime int64, ticks, workerThreads int) *TimeScheduler {
	if tickTime <= 0 || ticks <= 0 || workerThreads <= 0 {
		panic("invalid parameters")
	}
	if prefix == "" {
		prefix = "timer"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &TimeScheduler{
		prefix:        prefix,
		workerThreads: workerThreads,
		timeWheel:     newTimeWheel(tickTime, ticks, time.Now().UnixMilli()),
		flying:        make(chan *TimeWork, 1000),
		working:       make(chan *TimeWork, 1000),
		cancels:       make(chan *TimeWork, 1000),
		ctx:           ctx,
		cancelFunc:    cancel,
	}
}

func (ts *TimeScheduler) Start() {
	if atomic.CompareAndSwapInt32(&ts.started, 0, 1) {
		logger.Println("Starting scheduler")
		go ts.processQueue()
		for i := 0; i < ts.workerThreads; i++ {
			go ts.processWorking()
		}
	}
}

func (ts *TimeScheduler) Close() {
	if atomic.CompareAndSwapInt32(&ts.started, 1, 0) {
		logger.Println("Closing scheduler")
		ts.cancel()
	}
}

func (ts *TimeScheduler) Add(name string, time int64, runnable func()) Timeout {
	if runnable == nil {
		return nil
	}
	tw := newTimeWork(name, time, runnable, ts.afterRun, ts.afterCancel)
	return ts.add(tw)
}

func (ts *TimeScheduler) Delay(name string, delay int64, runnable func()) Timeout {
	if runnable == nil {
		return nil
	}
	t := time.Now().UnixMilli() + delay
	logger.Printf("Scheduling delay task %s: delay=%d, scheduled=%d", name, delay, t)
	tw := newTimeWork(name, t, runnable, ts.afterRun, ts.afterCancel)
	return ts.add(tw)
}

func (ts *TimeScheduler) AddTask(task TimeTask) Timeout {
	if task == nil {
		return nil
	}
	t := task.GetTime()
	if _, ok := task.(*DelayTask); ok {
		t = time.Now().UnixMilli() + t
	}
	tw := newTimeWork(task.GetName(), t, task.Run, ts.afterRun, ts.afterCancel)
	return ts.add(tw)
}

func (ts *TimeScheduler) add(timeWork *TimeWork) Timeout {
	atomic.AddInt64(&ts.tasks, 1)
	ts.flying <- timeWork
	logger.Printf("Task %s queued for scheduling", timeWork.name)
	return timeWork
}

func (ts *TimeScheduler) afterRun(tw *TimeWork) {
	atomic.AddInt64(&ts.tasks, -1)
}

func (ts *TimeScheduler) afterCancel(tw *TimeWork) {
	atomic.AddInt64(&ts.tasks, -1)
	ts.cancels <- tw
}

func (ts *TimeScheduler) cancel() {
	for {
		select {
		case tw := <-ts.cancels:
			logger.Printf("Processing cancellation for task %s", tw.name)
			tw.remove()
		default:
			return
		}
	}
}

func (ts *TimeScheduler) supply() {
	for i := 0; i < 100000; i++ {
		select {
		case tw := <-ts.flying:
			if !tw.IsCancelled() {
				if !ts.timeWheel.add(tw) {
					logger.Printf("Task %s is due, adding to working queue", tw.name)
					ts.working <- tw
				}
			} else {
				logger.Printf("Task %s already cancelled, skipping", tw.name)
			}
		default:
			return
		}
	}
}

func (ts *TimeScheduler) processQueue() {
	for {
		select {
		case <-ts.ctx.Done():
			logger.Println("Queue processor shutting down")
			return
		case slot := <-ts.timeWheel.queue:
			currentTime := time.Now().UnixMilli()
			if slot.expiration <= currentTime {
				logger.Printf("Processing slot with expiration %d at %d", slot.expiration, currentTime)
				ts.cancel()
				ts.supply()
				slot.flush(func(tw *TimeWork) {
					if !tw.IsCancelled() {
						ts.working <- tw
					} else {
						logger.Printf("Skipping cancelled task %s in slot", tw.name)
					}
				})
				ts.timeWheel.advance(currentTime)
			} else {
				// Requeue slot if not yet expired
				select {
				case ts.timeWheel.queue <- slot:
				default:
				}
			}
		default:
			ts.cancel()
			ts.supply()
			time.Sleep(time.Duration(ts.timeWheel.tickTime) * time.Millisecond)
			ts.timeWheel.advance(time.Now().UnixMilli())
		}
	}
}

func (ts *TimeScheduler) processWorking() {
	for {
		select {
		case <-ts.ctx.Done():
			logger.Println("Worker shutting down")
			return
		case tw := <-ts.working:
			if tw != nil && !tw.IsCancelled() {
				tw.Run()
			}
		}
	}
}

// DelayTask is a TimeTask for delayed execution.
type DelayTask struct {
	DelegateTask
}

func NewDelayTask(name string, delay int64, runnable func()) *DelayTask {
	return &DelayTask{DelegateTask{name: name, time: delay, runnable: runnable}}
}
