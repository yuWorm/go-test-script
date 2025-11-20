# 异步执行功能说明

## 概述

本项目实现了异步执行节点，支持在蓝图中创建异步任务并等待结果。

## 异步节点

### AsyncStart - 异步开始
- **功能**: 创建一个新的异步任务
- **输入**: 无
- **输出**:
  - `task_id` (string): 自动生成的唯一任务ID
  - `started` (bool): 是否成功启动

### AsyncEnd - 异步结束
- **功能**: 标记异步任务完成并设置结果
- **输入**:
  - `task_id` (string): 任务ID
  - `result` (any): 任务结果
- **输出**:
  - `completed` (bool): 是否完成
  - `task_id` (string): 任务ID
  - `result` (any): 任务结果

### Await - 等待结果
- **功能**: 等待异步任务完成并获取结果
- **输入**:
  - `task_id` (string): 要等待的任务ID
  - `timeout` (float): 超时时间（秒），默认30秒
- **输出**:
  - `result` (any): 任务结果
  - `completed` (bool): 是否完成

### Sleep - 休眠
- **功能**: 休眠指定时间（用于模拟耗时操作）
- **输入**:
  - `duration` (float): 休眠时间（毫秒）
- **输出**:
  - `done` (bool): 是否完成

### Parallel - 并行等待
- **功能**: 等待多个异步任务全部完成
- **输入**: 多个task_id (string)
- **输出**:
  - `result_0`, `result_1`, ... : 各任务的结果
  - `count` (float): 任务数量
  - `all_completed` (bool): 是否全部完成

## 工作原理

1. **AsyncStart** 创建一个异步任务对象，使用带缓冲的channel存储结果
2. **AsyncEnd** 将结果写入任务的ResultChan
3. **Await** 从ResultChan读取结果，支持超时

## 当前实现状态

### ✅ 已实现
- 异步任务管理器 (AsyncTaskManager)
- 异步任务对象 (AsyncTask) with channels
- AsyncStart, AsyncEnd, Await, Sleep, Parallel 执行器
- 超时支持
- 任务结果缓存

### ⚠️ 当前限制
- **顺序执行模式**: 当前实现中，所有节点按拓扑顺序在主线程中依次执行
- **非真正异步**: AsyncStart到AsyncEnd之间的节点不会在独立goroutine中执行
- **执行顺序依赖**: Await必须在AsyncEnd之后执行，需要通过连接保证依赖关系

### 🔄 未来改进
- **Goroutine集成**: 检测AsyncStart到AsyncEnd区域，在独立goroutine中执行
- **真正并行**: 支持多个异步区域同时执行
- **后台任务**: 支持不等待结果的后台异步任务

## 示例

### 简单异步任务

见 `test_simple_async.go` - 演示了完整的async创建、完成和等待流程。

执行顺序:
1. start
2. async_start (创建任务)
3. constant (生成值42)
4. async_end (设置任务结果为42)
5. await (等待并获取结果42)
6. end

关键点:
- async_end必须在await之前执行（通过连接依赖保证）
- 结果通过带缓冲channel传递，允许先设置后等待

### 测试

```bash
# 测试异步任务基本机制
go run test_async_task.go

# 测试简单异步蓝图
go run test_simple_async.go
```

## 技术实现

### AsyncTask 结构
```go
type AsyncTask struct {
    ID         string
    ResultChan chan interface{}  // 缓冲大小为1
    ErrorChan  chan error        // 缓冲大小为1
    Ctx        context.Context
    Cancel     context.CancelFunc
    StartTime  time.Time
    Completed  bool
    mu         sync.RWMutex
}
```

### 并发安全
- 使用 sync.RWMutex 保护任务状态
- Channel 用于goroutine间通信
- Context 支持取消操作

## 注意事项

1. **连接设计**: 在当前顺序执行模式下，必须通过连接确保AsyncEnd在Await之前执行
2. **超时设置**: Await的timeout参数应设置合理的超时时间
3. **任务ID**: 由AsyncStart自动生成，用户无需指定
4. **类型匹配**: 连接时注意引脚类型匹配，避免bool连接到float等情况

## 贡献者

如需改进异步执行功能，欢迎提交PR。优先改进方向：
1. 实现真正的goroutine-based异步执行
2. 添加更多异步控制节点（如WhenAll, WhenAny）
3. 改进错误处理和异常传播
