package rqcache

// 定义 ByteView 结构体，用于表示缓存值
type ByteView struct {
	b []byte // 存储真实的缓存值，选择 byte 类型是为了能够支持任意的数据类型的存储，例如字符串、图片等
}

// 返回缓存值的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回缓存值的拷贝
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 返回缓存值的字符串形式
func (v ByteView) String() string {
	return string(v.b)
}

// 拷贝缓存值
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
