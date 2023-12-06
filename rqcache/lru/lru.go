package lru

import "container/list"

// lru 包实现了使用最近最久未使用使用算法的缓存功能
type Cache struct {
	maxBytes  int64                         // Cache 最大容量(Byte)
	nbytes    int64                         // Cache 当前容量(Byte)
	ll        *list.List                    // 双向链表，用于存储缓存的键值对
	cache     map[string]*list.Element      // 用于存储键值对在双向链表中的节点地址
	OnEvicted func(key string, value Value) // 可选的，在清除条目时执行
}

// 定义双向链表节点所存储的对象
type entry struct {
	key   string
	value Value
}

// 定义 Value 接口，用于计算所存储的对象所占用的内存大小
type Value interface {
	Len() int
}

// 初始化 Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 向 Cache 中添加一个元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 从 Cache 中获取一个元素
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 从 Cache 中删除最近最久未使用的元素
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 获取 Cache 中元素的数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
