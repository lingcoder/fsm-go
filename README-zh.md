# FSM-Go: Go 语言轻量级有限状态机

FSM-Go 是一个轻量级、高性能、无状态的有限状态机 Go 实现，灵感来自阿里巴巴的 COLA 状态机组件。

## 特性

- 轻量级和无状态设计，提供高性能
- 使用 Go 泛型实现类型安全
- 流畅的 API 用于定义状态机
- 支持外部、内部状态转换
- 支持并行转换和批量转换
- 函数类型支持，简化条件和动作的定义
- 带有自定义逻辑的条件转换
- 转换过程中执行的动作
- 状态机监听器，用于监控状态变化
- 状态转换验证功能
- 线程安全，支持并发使用
- 支持状态机图表可视化

## 安装

```bash
go get github.com/lingcoder/fsm-go
```

## 使用方法

```go
package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/fsm-go"
)

// 定义状态
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
)

// 定义事件
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

// 定义上下文
type OrderContext struct {
	OrderID   string
	UserID    string
	Amount    float64
}

// 使用函数类型定义动作和条件
func main() {
	// 创建构建器
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()
	
	// 定义状态机 - 使用函数类型简化定义
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		When(func(ctx OrderContext) bool {
			return ctx.Amount > 0
		}).
		Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("订单 %s 支付了 %.2f 元\n", ctx.OrderID, ctx.Amount)
			return nil
		})
	
	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("订单 %s 已发货\n", ctx.OrderID)
			return nil
		})
	
	// 使用批量转换 - 从多个状态到一个状态
	builder.ExternalTransitions().
		FromAmong(OrderCreated, OrderPaid, OrderShipped).
		To(OrderCancelled).
		On(EventCancel).
		Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
			fmt.Printf("订单 %s 从 %s 状态被取消\n", ctx.OrderID, from)
			return nil
		})
	
	// 构建状态机
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("构建状态机失败: %v", err)
	}
	
	// 添加状态机监听器
	stateMachine.AddListener(&OrderStateListener{})
	
	// 使用状态机
	ctx := OrderContext{
		OrderID: "ORD-001",
		UserID:  "USR-001",
		Amount:  100.0,
	}
	
	// 验证转换是否可行
	if stateMachine.Verify(OrderCreated, EventPay) {
		fmt.Println("订单可以支付")
	}
	
	// 从 CREATED 转换到 PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, ctx)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}
	
	fmt.Printf("新状态: %v\n", newState)
}

// 定义状态机监听器
type OrderStateListener struct{}

func (l *OrderStateListener) BeforeTransition(from, to OrderState, event OrderEvent, ctx OrderContext) {
	fmt.Printf("准备从 %s 转换到 %s，事件: %s\n", from, to, event)
}

func (l *OrderStateListener) AfterTransition(from, to OrderState, event OrderEvent, ctx OrderContext, err error) {
	if err != nil {
		fmt.Printf("从 %s 转换到 %s 失败: %v\n", from, to, err)
	} else {
		fmt.Printf("成功从 %s 转换到 %s\n", from, to)
	}
}
```

## 核心概念

- **状态 (State)**: 表示业务流程中的特定状态
- **事件 (Event)**: 触发状态转换
- **转换 (Transition)**: 定义状态如何响应事件而变化
  - **外部转换 (External Transition)**: 不同状态之间的转换
  - **内部转换 (Internal Transition)**: 同一状态内的动作
  - **并行转换 (Parallel Transition)**: 一个事件触发到多个目标状态的转换
  - **批量转换 (Multiple Transition)**: 多个源状态到一个目标状态的转换
- **条件 (Condition)**: 决定是否应该发生转换的逻辑
- **动作 (Action)**: 转换发生时执行的逻辑
- **监听器 (Listener)**: 监控状态转换过程
- **状态机 (StateMachine)**: 管理状态和转换的核心组件

## 示例

查看 `examples` 目录获取更详细的示例：

- `examples/order`: 订单处理工作流
- `examples/workflow`: 审批工作流
- `examples/game`: 游戏状态管理

## 高级功能

### 函数类型支持

可以直接使用函数作为条件和动作，无需定义结构体：

```go
// 使用函数作为条件
.When(func(ctx OrderContext) bool {
    return ctx.Amount > 0
})

// 使用函数作为动作
.Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
    fmt.Printf("处理订单 %s\n", ctx.OrderID)
    return nil
})
```

### 并行转换

一个事件可以触发到多个目标状态的转换：

```go
builder.ExternalParallelTransition().
    From(OrderPaid).
    ToAmong(OrderShipped, OrderNotified).
    On(EventProcess).
    Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {
        fmt.Printf("处理订单: %s\n", ctx.OrderID)
        return nil
    })

// 触发并行转换
newStates, err := stateMachine.FireParallelEvent(OrderPaid, EventProcess, ctx)
```

### 批量转换

从多个源状态到一个目标状态的转换：

```go
builder.ExternalTransitions().
    FromAmong(OrderCreated, OrderPaid, OrderShipped).
    To(OrderCancelled).
    On(EventCancel).
    Perform(cancelAction)
```

### 状态机监听器

监控状态转换过程：

```go
// 定义监听器
type LoggingListener struct{}

func (l *LoggingListener) BeforeTransition(from, to OrderState, event OrderEvent, ctx OrderContext) {
    fmt.Printf("准备从 %s 转换到 %s\n", from, to)
}

func (l *LoggingListener) AfterTransition(from, to OrderState, event OrderEvent, ctx OrderContext, err error) {
    if err != nil {
        fmt.Printf("转换失败: %v\n", err)
    } else {
        fmt.Printf("转换成功\n")
    }
}

// 添加监听器
stateMachine.AddListener(&LoggingListener{})
```

### 转换验证

在执行转换前验证是否可行：

```go
if stateMachine.Verify(currentState, event) {
    // 可以执行转换
    newState, err := stateMachine.FireEvent(currentState, event, ctx)
} else {
    // 转换不可行
    fmt.Println("当前状态不能执行此操作")
}
```

## 性能

FSM-Go 设计注重高性能：

- 无状态设计最小化内存使用
- 高效的转换查找
- 线程安全，支持并发使用
- 测试套件中包含基准测试

## 实现细节

### 状态机接口

```go
type StateMachine[S comparable, E comparable, C any] interface {
	// FireEvent 触发基于当前状态和事件的状态转换
	// 返回新状态和可能发生的错误
	FireEvent(sourceState S, event E, ctx C) (S, error)

	// FireParallelEvent 触发并行状态转换
	// 返回新状态列表和可能发生的错误
	FireParallelEvent(sourceState S, event E, ctx C) ([]S, error)
	
	// Verify 检查给定状态和事件是否有有效的转换
	// 返回是否存在有效转换
	Verify(sourceState S, event E) bool

	// ShowStateMachine 返回状态机的字符串表示
	ShowStateMachine() string

	// GeneratePlantUML 返回状态机的 PlantUML 图表
	GeneratePlantUML() string
	
	// AddListener 添加状态机监听器
	AddListener(listener StateMachineListener[S, E, C])
	
	// RemoveListener 移除状态机监听器
	RemoveListener(listener StateMachineListener[S, E, C])
}
```

### 构建器 API

FSM-Go 使用流畅的构建器 API 来定义状态机：

```go
// 创建构建器
builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()

// 定义外部转换
builder.ExternalTransition().
    From(OrderCreated).  // 源状态
    To(OrderPaid).       // 目标状态
    On(EventPay).        // 触发事件
    When(func(ctx OrderContext) bool { return ctx.Amount > 0 }).  // 转换条件
    Perform(func(from, to OrderState, event OrderEvent, ctx OrderContext) error {  // 转换动作
        fmt.Printf("处理支付: %.2f\n", ctx.Amount)
        return nil
    })

// 定义并行转换
builder.ExternalParallelTransition().
    From(OrderPaid).
    ToAmong(OrderShipped, OrderNotified).
    On(EventProcess).
    Perform(processAction)

// 定义多源状态转换
builder.ExternalTransitions().
    FromAmong(OrderCreated, OrderPaid, OrderShipped).  // 多个源状态
    To(OrderCancelled).  // 目标状态
    On(EventCancel).     // 触发事件
    Perform(cancelAction)  // 转换动作

// 构建状态机
stateMachine, err := builder.Build("OrderStateMachine")
```

## 许可证

MIT
