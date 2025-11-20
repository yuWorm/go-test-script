package blueprint

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// AsyncTask 表示一个异步任务
type AsyncTask struct {
	ID         string
	ResultChan chan interface{}
	ErrorChan  chan error
	Ctx        context.Context
	Cancel     context.CancelFunc
	StartTime  time.Time
	Completed  bool
	mu         sync.RWMutex
}

// NewAsyncTask 创建一个新的异步任务
func NewAsyncTask(id string) *AsyncTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &AsyncTask{
		ID:         id,
		ResultChan: make(chan interface{}, 1),
		ErrorChan:  make(chan error, 1),
		Ctx:        ctx,
		Cancel:     cancel,
		StartTime:  time.Now(),
		Completed:  false,
	}
}

// SetResult 设置任务结果
func (at *AsyncTask) SetResult(result interface{}) {
	at.mu.Lock()
	defer at.mu.Unlock()

	if !at.Completed {
		at.ResultChan <- result
		at.Completed = true
		close(at.ResultChan)
	}
}

// SetError 设置任务错误
func (at *AsyncTask) SetError(err error) {
	at.mu.Lock()
	defer at.mu.Unlock()

	if !at.Completed {
		at.ErrorChan <- err
		at.Completed = true
		close(at.ErrorChan)
	}
}

// Wait 等待任务完成（带超时）
func (at *AsyncTask) Wait(timeout time.Duration) (interface{}, error) {
	if timeout > 0 {
		select {
		case result := <-at.ResultChan:
			return result, nil
		case err := <-at.ErrorChan:
			return nil, err
		case <-time.After(timeout):
			at.Cancel()
			return nil, fmt.Errorf("async task %s timed out after %v", at.ID, timeout)
		case <-at.Ctx.Done():
			return nil, fmt.Errorf("async task %s was cancelled", at.ID)
		}
	} else {
		// 无超时等待
		select {
		case result := <-at.ResultChan:
			return result, nil
		case err := <-at.ErrorChan:
			return nil, err
		case <-at.Ctx.Done():
			return nil, fmt.Errorf("async task %s was cancelled", at.ID)
		}
	}
}

// IsCompleted 检查任务是否已完成
func (at *AsyncTask) IsCompleted() bool {
	at.mu.RLock()
	defer at.mu.RUnlock()
	return at.Completed
}

// AsyncTaskManager 管理所有异步任务
type AsyncTaskManager struct {
	tasks map[string]*AsyncTask
	mu    sync.RWMutex
}

// NewAsyncTaskManager 创建异步任务管理器
func NewAsyncTaskManager() *AsyncTaskManager {
	return &AsyncTaskManager{
		tasks: make(map[string]*AsyncTask),
	}
}

// CreateTask 创建新任务
func (atm *AsyncTaskManager) CreateTask(id string) *AsyncTask {
	atm.mu.Lock()
	defer atm.mu.Unlock()

	task := NewAsyncTask(id)
	atm.tasks[id] = task
	return task
}

// GetTask 获取任务
func (atm *AsyncTaskManager) GetTask(id string) (*AsyncTask, error) {
	atm.mu.RLock()
	defer atm.mu.RUnlock()

	task, exists := atm.tasks[id]
	if !exists {
		return nil, fmt.Errorf("async task %s not found", id)
	}
	return task, nil
}

// RemoveTask 移除任务
func (atm *AsyncTaskManager) RemoveTask(id string) {
	atm.mu.Lock()
	defer atm.mu.Unlock()

	if task, exists := atm.tasks[id]; exists {
		task.Cancel()
		delete(atm.tasks, id)
	}
}

// GetAllTasks 获取所有任务
func (atm *AsyncTaskManager) GetAllTasks() []*AsyncTask {
	atm.mu.RLock()
	defer atm.mu.RUnlock()

	tasks := make([]*AsyncTask, 0, len(atm.tasks))
	for _, task := range atm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// CancelAll 取消所有任务
func (atm *AsyncTaskManager) CancelAll() {
	atm.mu.Lock()
	defer atm.mu.Unlock()

	for _, task := range atm.tasks {
		task.Cancel()
	}
	atm.tasks = make(map[string]*AsyncTask)
}
