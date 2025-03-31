package goroutine_pool

import (
	"context"
	"errors"
	"fmt"
	"go-fission-activity/activity/web/middleware/logTracing"
	"sync"
)

// GoroutinePool 结构体包含一个等待组（sync.WaitGroup）和一个通道（chan struct{}）来控制并发数
type GoroutinePool struct {
	wg sync.WaitGroup
	Ch chan struct{}
}

// NewGoroutinePool 创建一个新的GoroutinePool实例，最大并发数由maxGoroutines参数指定
func NewGoroutinePool(maxGoroutines int) *GoroutinePool {
	return &GoroutinePool{
		Ch: make(chan struct{}, maxGoroutines),
	}
}

// Execute 启动一个新的goroutine，如果通道满了，则等待
func (p *GoroutinePool) Execute(f func(param interface{}), param interface{}) {
	p.wg.Add(1)
	p.Ch <- struct{}{} // 占用一个槽位
	go func() {
		// defer 异常处理
		defer func() {
			if e := recover(); e != nil {
				logTracing.LogErrorPrintf(context.Background(), errors.New(fmt.Sprintf("方法[%s]，发生panic异常", "协程")), logTracing.ErrorLogFmt, e)
				return
			}
		}()
		defer p.wg.Done()
		defer func() { <-p.Ch }()
		f(param) // 执行传入的函数
	}()
}

// Wait 等待所有goroutine执行完毕
func (p *GoroutinePool) Wait() {
	p.wg.Wait()
}
