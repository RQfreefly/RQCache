RQCache 是基于 Go 语言的分布式缓存系统，支持特性有：

- 单机缓存和基于 HTTP 的分布式缓存；
- 最近最少访问（Least Recently Used, LRU）缓存策略；
- 使用 Go 锁机制防止缓存击穿；
- 使用一致性哈希选择节点，实现负载均衡；
- 使用 protobuf 优化节点间二进制通信。