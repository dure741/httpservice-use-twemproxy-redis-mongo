package gocelery

import (
	"errors"
	"fmt"
	"time"
	//log "github.com/Sirupsen/logrus"
)

// CeleryClient provides API for sending celery tasks
type CeleryClient struct {
	broker  CeleryBroker
	backend CeleryBackend
	worker  *CeleryWorker
}

// CeleryBroker is interface for celery broker database
type CeleryBroker interface {
	SendCeleryMessage(*CeleryMessage) error
	GetTaskMessage() (*TaskMessage, error) // must be non-blocking
	Reconnect() error
}

// CeleryBackend is interface for celery backend database
type CeleryBackend interface {
	GetResult(string) (*ResultMessage, error) // must be non-blocking
	SetResult(taskID string, result *ResultMessage) error
}

// NewCeleryClient creates new celery client
func NewCeleryClient(broker CeleryBroker, backend CeleryBackend, numWorkers int) (*CeleryClient, error) {
	return &CeleryClient{
		broker,
		backend,
		NewCeleryWorker(broker, backend, numWorkers),
	}, nil
}

// Register task
func (cc *CeleryClient) Register(name string, task interface{}) {
	cc.worker.Register(name, task)
}

// StartWorker starts celery workers
func (cc *CeleryClient) StartWorker() {
	cc.worker.StartWorker()
}

// StopWorker stops celery workers
func (cc *CeleryClient) StopWorker() {
	cc.worker.StopWorker()
}

// Delay gets asynchronous result
func (cc *CeleryClient) Delay(task string, args ...interface{}) (*AsyncResult, error) {
	celeryTask := getTaskMessage(task)
	celeryTask.Args = args
	return cc.delay(celeryTask)
}

// DelayKwargs gets asynchronous results with argument map
func (cc *CeleryClient) DelayKwargs(task string, args map[string]interface{}) (*AsyncResult, error) {
	celeryTask := getTaskMessage(task)
	celeryTask.Kwargs = args
	return cc.delay(celeryTask)
}

func (cc *CeleryClient) delay(task *TaskMessage) (*AsyncResult, error) {
	defer releaseTaskMessage(task)
	encodedMessage, err := task.Encode()
	if err != nil {
		return nil, err
	}
	celeryMessage := getCeleryMessage(encodedMessage)
	defer releaseCeleryMessage(celeryMessage)
	err = cc.broker.SendCeleryMessage(celeryMessage)
	if err != nil {
		if err = cc.broker.Reconnect(); err != nil {
			return nil, err
		}
		err = cc.broker.SendCeleryMessage(celeryMessage)
		return nil, err
	}
	return &AsyncResult{
		taskID:  task.ID,
		backend: cc.backend,
	}, nil
}

// CeleryTask is an interface that represents actual task
// Passing CeleryTask interface instead of function pointer
// avoids reflection and may have performance gain.
// ResultMessage must be obtained using GetResultMessage()
type CeleryTask interface {

	// ParseKwargs - define a method to parse kwargs
	ParseKwargs(map[string]interface{}) error

	// RunTask - define a method to run
	RunTask() (interface{}, error)
}

// AsyncResult is pending result
type AsyncResult struct {
	taskID  string
	backend CeleryBackend
	result  *ResultMessage
}

// Get gets actual result from redis
// It blocks for period of time set by timeout and return error if unavailable
func (ar *AsyncResult) Get(timeout time.Duration) (interface{}, error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			err := fmt.Errorf("%v timeout getting result for %s", timeout, ar.taskID)
			return nil, err
		case <-ticker.C:
			val, err := ar.AsyncGet()
			if err != nil {
				continue
			}
			return val, nil
		}
	}
}

////
// status
const (
	ST_TASK_OK      = 0
	ST_TASK_FAILED  = 1
	ST_TASK_TIMEOUT = 2
)

var (
	ERR_NilObj = errors.New("KEY NOT FOUND")
)

func (ar *AsyncResult) GetRet(timeout time.Duration) (interface{}, int, error) {
	ticker := time.NewTicker(50 * time.Millisecond)
	timeoutChan := time.After(timeout)
	for {
		select {
		case <-timeoutChan:
			//log.Errorf("%v timeout getting result for %s", timeout, ar.taskID)
			return nil, ST_TASK_TIMEOUT, nil
		case <-ticker.C:
			val, st, err := ar.AsyncGetRet()
			//if err != nil {
			//	continue
			//}
			if err != nil {
				if err == ERR_NilObj {
					//log.Warnf("not found")
					continue
				} else {
					//log.Errorf("!!!!!!!!!!!!err:%v", err)
					return nil, ST_TASK_FAILED, err
				}
			}
			return val, st, nil
		}
	}
}

// AsyncGet gets actual result from redis and returns nil if not available
func (ar *AsyncResult) AsyncGet() (interface{}, error) {
	if ar.result != nil {
		return ar.result.Result, nil
	}
	// process

	val, err := ar.backend.GetResult(ar.taskID)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, err
	}
	if val.Status != "SUCCESS" {
		return nil, fmt.Errorf("error response status %v", val)
	}
	ar.result = val
	return val.Result, nil
}

func (ar *AsyncResult) AsyncGetRet() (interface{}, int, error) {
	if ar.result != nil {
		return ar.result.Result, ST_TASK_OK, nil
	}
	// process

	val, err := ar.backend.GetResult(ar.taskID)
	if err != nil {
		return nil, ST_TASK_FAILED, err
	}
	if val == nil {
		return nil, ST_TASK_FAILED, err
	}
	if val.Status != "SUCCESS" {
		return val.Result, ST_TASK_FAILED, nil
	}
	ar.result = val
	return val.Result, ST_TASK_OK, nil
}

// Ready checks if actual result is ready
func (ar *AsyncResult) Ready() (bool, error) {
	if ar.result != nil {
		return true, nil
	}
	val, err := ar.backend.GetResult(ar.taskID)
	if err != nil {
		return false, err
	}
	ar.result = val
	return (val != nil), nil
}
