# FSM-Go: A Lightweight Finite State Machine for Go

FSM-Go is a lightweight, high-performance, stateless finite state machine implementation in Go, inspired by Alibaba's COLA state machine component.

[中文文档](#中文文档)

## Features

- Lightweight and stateless design for high performance
- Type-safe implementation using Go generics
- Fluent API for defining state machines
- Support for external, internal, and parallel transitions
- Conditional transitions with custom logic
- Actions that execute during transitions
- Thread-safe for concurrent use
- Visualization support for state machine diagrams

## Installation

```bash
go get github.com/lingcoder/fsm-go
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/fsm-go"
)

// Define states
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
)

// Define events
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

// Define context
type OrderContext struct {
	OrderID   string
	UserID    string
	Amount    float64
}

// Define action
type OrderAction struct{}

func (a *OrderAction) Execute(from OrderState, to OrderState, event OrderEvent, ctx OrderContext) error {
	fmt.Printf("Order %s transitioning from %s to %s on event %s\n", 
		ctx.OrderID, from, to, event)
	return nil
}

// Define condition
type OrderCondition struct{}

func (c *OrderCondition) IsSatisfied(ctx OrderContext) bool {
	return true
}

func main() {
	// Create a builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()
	
	// Define the state machine
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	// Build the state machine
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("Failed to build state machine: %v", err)
	}
	
	// Use the state machine
	ctx := OrderContext{
		OrderID: "ORD-001",
		UserID:  "USR-001",
		Amount:  100.0,
	}
	
	// Transition from CREATED to PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, ctx)
	if err != nil {
		log.Fatalf("Failed to transition: %v", err)
	}
	
	fmt.Printf("New state: %v\n", newState)
}
```

## Core Concepts

- **State**: Represents a specific state in your business process
- **Event**: Triggers state transitions
- **Transition**: Defines how states change in response to events
  - **External Transition**: Transition between different states
  - **Internal Transition**: Actions within the same state
- **Condition**: Logic that determines if a transition should occur
- **Action**: Logic executed when a transition occurs
- **StateMachine**: The core component that manages states and transitions

## Examples

Check the `examples` directory for more detailed examples:

- `examples/order`: Order processing workflow
- `examples/workflow`: Approval workflow
- `examples/game`: Game state management

## Performance

FSM-Go is designed for high performance:

- Stateless design minimizes memory usage
- Efficient transition lookup
- Thread-safe for concurrent use
- Benchmarks included in the test suite

## License

MIT

---

# 中文文档

FSM-Go 是一个轻量级、高性能、无状态的有限状态机 Go 实现，灵感来自阿里巴巴的 COLA 状态机组件。

## 特性

- 轻量级和无状态设计，提供高性能
- 使用 Go 泛型实现类型安全
- 流畅的 API 用于定义状态机
- 支持外部、内部和并行状态转换
- 带有自定义逻辑的条件转换
- 转换过程中执行的动作
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

// 定义动作
type OrderAction struct{}

func (a *OrderAction) Execute(from OrderState, to OrderState, event OrderEvent, ctx OrderContext) error {
	fmt.Printf("订单 %s 从 %s 状态转换到 %s 状态，触发事件: %s\n", 
		ctx.OrderID, from, to, event)
	return nil
}

// 定义条件
type OrderCondition struct{}

func (c *OrderCondition) IsSatisfied(ctx OrderContext) bool {
	return true
}

func main() {
	// 创建构建器
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderContext]()
	
	// 定义状态机
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		When(&OrderCondition{}).
		Perform(&OrderAction{})
	
	// 构建状态机
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("构建状态机失败: %v", err)
	}
	
	// 使用状态机
	ctx := OrderContext{
		OrderID: "ORD-001",
		UserID:  "USR-001",
		Amount:  100.0,
	}
	
	// 从 CREATED 转换到 PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, ctx)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}
	
	fmt.Printf("新状态: %v\n", newState)
}
```

## 核心概念

- **状态 (State)**: 表示业务流程中的特定状态
- **事件 (Event)**: 触发状态转换
- **转换 (Transition)**: 定义状态如何响应事件而变化
  - **外部转换 (External Transition)**: 不同状态之间的转换
  - **内部转换 (Internal Transition)**: 同一状态内的动作
- **条件 (Condition)**: 决定是否应该发生转换的逻辑
- **动作 (Action)**: 转换发生时执行的逻辑
- **状态机 (StateMachine)**: 管理状态和转换的核心组件

## 示例

查看 `examples` 目录获取更详细的示例：

- `examples/order`: 订单处理工作流
- `examples/workflow`: 审批工作流
- `examples/game`: 游戏状态管理

## 性能

FSM-Go 设计注重高性能：

- 无状态设计最小化内存使用
- 高效的转换查找
- 线程安全，支持并发使用
- 测试套件中包含基准测试

## 许可证

MIT
