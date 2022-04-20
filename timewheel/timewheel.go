package timewheel

import (
	"container/list"
	"log"
	"time"
)

type Job func(interface{})
type Task struct {
	delay  time.Duration
	data   interface{}
	circle int
}
type TimeWheel struct {
	interval   time.Duration
	slotNum    int
	slots      []*list.List
	currentPos int
	job        Job
	taskChan   chan Task
	stopChan   chan struct{}
}

func NewTimeWheel(interval time.Duration, slotNum int, job Job) *TimeWheel {
	tw := &TimeWheel{
		interval:   interval,
		slotNum:    slotNum,
		slots:      nil,
		currentPos: 0,
		job:        job,
		taskChan:   make(chan Task),
		stopChan:   make(chan struct{}),
	}
	tw.initSlots()
	return tw
}

func (tw *TimeWheel) initSlots() {
	tw.slots = make([]*list.List, tw.slotNum)
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

func (tw *TimeWheel) AddTimer(delay time.Duration, data interface{}) {
	tw.taskChan <- Task{
		delay:  delay,
		data:   data,
		circle: 0,
	}
}

func (tw *TimeWheel) Stop() {
	tw.stopChan <- struct{}{}
}

func (tw *TimeWheel) Start() {
	go tw.start()
}

func (tw *TimeWheel) start() {
	t := time.NewTicker(tw.interval)
	for {
		select {
		case <-t.C:
			tw.tickHandler()
		case task := <-tw.taskChan:
			tw.addTask(&task)
		case <-tw.stopChan:
			t.Stop()
			return
		}
	}
}

func (tw *TimeWheel) addTask(task *Task) {
	circle, pos := tw.calcDelayCircleAndPos(task)
	task.circle = circle
	tw.slots[pos].PushBack(task)
	log.Printf("addTask circle=%d, pos=%d\n", circle, pos)
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	log.Printf("tickHandler pos=%d\n", tw.currentPos)
	tw.runTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

func (tw *TimeWheel) calcDelayCircleAndPos(task *Task) (circle int, pos int) {
	circle = int(task.delay.Seconds()) / int(tw.interval.Seconds()) / tw.slotNum
	pos = (tw.currentPos + int(task.delay/tw.interval)) % tw.slotNum
	return
}

func (tw *TimeWheel) runTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.circle > 0 {
			task.circle--
			e = e.Next()
			continue
		}
		log.Printf("fetchTask circle=%d\n", task.circle)
		go tw.job(task.data)
		n := e.Next()
		l.Remove(e)
		e = n
	}
}
