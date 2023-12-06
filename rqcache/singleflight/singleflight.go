package singleflight

import "sync"

// call 表示一个正在进行中或已完成的 Do 调用
type call struct {
	wg  sync.WaitGroup // 用于等待调用完成
	val interface{}    // 调用的结果值
	err error          // 调用过程中的错误信息
}

// Group 表示一类工作，并形成一个命名空间，其中可以对工作单元进行执行以进行重复抑制。
type Group struct {
	mu sync.Mutex // 保护 m
	m  map[string]*call
}

// Do 方法执行并返回给定函数的结果，确保对于给定的 key，在任何时刻只有一个执行正在进行。
// 如果出现重复调用，重复的调用者会等待原始调用完成，并接收相同的结果。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁，保护 g.m 的并发访问
	g.mu.Lock()
	// 如果 g.m 为 nil，进行初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 检查是否存在相同 key 的调用
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		// 如果存在相同 key 的调用，等待其完成，并返回结果
		c.wg.Wait()
		return c.val, c.err
	}
	// 如果不存在相同 key 的调用，创建一个新的调用并加入到 g.m 中
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 执行用户提供的函数 fn，并获取结果
	c.val, c.err = fn()
	// 减少 WaitGroup 的计数，表示调用已经完成
	c.wg.Done()

	// 再次加锁，删除已完成的调用信息
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	// 返回调用的结果
	return c.val, c.err
}
