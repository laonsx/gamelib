package task

import (
	"errors"
	"sync"
)

//Task 异步任务
type Task struct {
	workerPool chan worker
	mux        sync.Mutex
	wg         *sync.WaitGroup
	closed     bool
	quit       chan struct{}
	callback   func(v interface{})
}

//New 初始化
// n 启动goroutine数量，callback回调处理函数
func New(n int, callback func(v interface{})) *Task {

	t := new(Task)
	t.workerPool = make(chan worker, n)
	t.callback = callback
	t.quit = make(chan struct{})
	t.wg = new(sync.WaitGroup)

	for i := 0; i < n; i++ {

		w := newWorker(t)
		w.start()
	}

	return t
}

//Send 发送数据
func (t *Task) Send(v interface{}) error {

	if t.closed {

		return errors.New("closed")
	}

	t.wg.Add(1)

	w := t.get()
	w.send(v)

	return nil
}

func (t *Task) Close() {

	t.closed = true
	close(t.quit)
}

func (t *Task) Wait() {

	t.wg.Wait()
}

func (t *Task) get() worker {

	return <-t.workerPool
}

func (t *Task) put(w worker) {

	t.workerPool <- w
}

type worker struct {
	task    *Task
	msgChan chan interface{}
}

func newWorker(task *Task) worker {

	return worker{
		task:    task,
		msgChan: make(chan interface{}),
	}
}

func (w worker) start() {

	go func() {

		for {

			w.task.put(w)

			select {

			case msg := <-w.msgChan:

				w.task.callback(msg)
				w.done()

			case <-w.task.quit:

				return
			}
		}
	}()
}

func (w worker) send(v interface{}) {

	w.msgChan <- v
}

func (w worker) done() {

	w.task.wg.Done()
}
