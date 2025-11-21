package nodes

import (
	"fmt"
	"github.com/go-blueprint-engine/blueprint"
)

// IfElseExecutor If/Else 条件执行器
type IfElseExecutor struct{}

// NewIfElseExecutor 创建 If/Else 执行器
func NewIfElseExecutor() *IfElseExecutor {
	return &IfElseExecutor{}
}

// Execute 执行 If/Else 条件
func (e *IfElseExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	condition, err := toBool(inputs["condition"])
	if err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	outputs := make(map[string]interface{})
	outputs["condition"] = condition

	// If 分支
	if condition {
		outputs["then_exec"] = true
		outputs["else_exec"] = false
		// 传递 then 分支的值
		if thenValue, exists := inputs["then_value"]; exists {
			outputs["result"] = thenValue
		}
	} else {
		// Else 分支
		outputs["then_exec"] = false
		outputs["else_exec"] = true
		// 传递 else 分支的值
		if elseValue, exists := inputs["else_value"]; exists {
			outputs["result"] = elseValue
		}
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *IfElseExecutor) Validate(node *blueprint.Node) error {
	hasCondition := false
	for _, pin := range node.InputPins {
		if pin.Name == "condition" {
			hasCondition = true
			break
		}
	}

	if !hasCondition {
		return fmt.Errorf("if/else node must have 'condition' input pin")
	}

	return nil
}

// ForLoopExecutor For 循环执行器
type ForLoopExecutor struct{}

// NewForLoopExecutor 创建 For 循环执行器
func NewForLoopExecutor() *ForLoopExecutor {
	return &ForLoopExecutor{}
}

// Execute 执行 For 循环
func (e *ForLoopExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	start, err := toFloat64(inputs["start"])
	if err != nil {
		return nil, fmt.Errorf("invalid start: %w", err)
	}

	end, err := toFloat64(inputs["end"])
	if err != nil {
		return nil, fmt.Errorf("invalid end: %w", err)
	}

	step := 1.0
	if stepVal, exists := inputs["step"]; exists {
		step, err = toFloat64(stepVal)
		if err != nil {
			return nil, fmt.Errorf("invalid step: %w", err)
		}
		if step == 0 {
			return nil, fmt.Errorf("step cannot be zero")
		}
	}

	outputs := make(map[string]interface{})

	// 执行循环体（简化版本，返回最后一次迭代的值）
	var lastIndex float64
	loopCount := 0
	maxIterations := 10000 // 防止无限循环

	if step > 0 {
		for i := start; i < end; i += step {
			lastIndex = i
			loopCount++
			if loopCount > maxIterations {
				return nil, fmt.Errorf("loop exceeded maximum iterations (%d)", maxIterations)
			}
		}
	} else {
		for i := start; i > end; i += step {
			lastIndex = i
			loopCount++
			if loopCount > maxIterations {
				return nil, fmt.Errorf("loop exceeded maximum iterations (%d)", maxIterations)
			}
		}
	}

	outputs["index"] = lastIndex
	outputs["count"] = float64(loopCount)
	outputs["completed"] = true

	return outputs, nil
}

// Validate 验证节点配置
func (e *ForLoopExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// WhileLoopExecutor While 循环执行器
type WhileLoopExecutor struct{}

// NewWhileLoopExecutor 创建 While 循环执行器
func NewWhileLoopExecutor() *WhileLoopExecutor {
	return &WhileLoopExecutor{}
}

// Execute 执行 While 循环
func (e *WhileLoopExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	condition, err := toBool(inputs["condition"])
	if err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	outputs := make(map[string]interface{})

	// 简化版本：只检查初始条件
	// 在完整实现中，需要支持循环体内重新计算条件
	iterations := 0
	maxIterations := 10000

	for condition && iterations < maxIterations {
		iterations++
		// 在实际实现中，这里需要执行循环体并更新条件
		// 目前简化为固定迭代次数
		if iterations >= 100 {
			break
		}
	}

	outputs["iterations"] = float64(iterations)
	outputs["completed"] = iterations < maxIterations

	return outputs, nil
}

// Validate 验证节点配置
func (e *WhileLoopExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// BreakExecutor Break 执行器
type BreakExecutor struct{}

// NewBreakExecutor 创建 Break 执行器
func NewBreakExecutor() *BreakExecutor {
	return &BreakExecutor{}
}

// Execute 执行 Break
func (e *BreakExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"break": true,
	}, nil
}

// Validate 验证节点配置
func (e *BreakExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// ContinueExecutor Continue 执行器
type ContinueExecutor struct{}

// NewContinueExecutor 创建 Continue 执行器
func NewContinueExecutor() *ContinueExecutor {
	return &ContinueExecutor{}
}

// Execute 执行 Continue
func (e *ContinueExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	return map[string]interface{}{
		"continue": true,
	}, nil
}

// Validate 验证节点配置
func (e *ContinueExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// BranchExecutor Branch 分支执行器（专用于执行流）
type BranchExecutor struct{}

// NewBranchExecutor 创建 Branch 执行器
func NewBranchExecutor() *BranchExecutor {
	return &BranchExecutor{}
}

// Execute 执行 Branch - 根据条件激活不同的执行引脚
func (e *BranchExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	condition, err := toBool(inputs["condition"])
	if err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	outputs := make(map[string]interface{})

	// 根据条件设置执行输出引脚的激活状态
	if condition {
		outputs["true_exec"] = true
		outputs["false_exec"] = false
	} else {
		outputs["true_exec"] = false
		outputs["false_exec"] = true
	}

	return outputs, nil
}

// Validate 验证节点配置
func (e *BranchExecutor) Validate(node *blueprint.Node) error {
	hasCondition := false
	for _, pin := range node.InputPins {
		if pin.Name == "condition" {
			hasCondition = true
			break
		}
	}

	if !hasCondition {
		return fmt.Errorf("branch node must have 'condition' input pin")
	}

	return nil
}

// ForkExecutor Fork 节点执行器 - 并行分叉
type ForkExecutor struct{}

// NewForkExecutor 创建 Fork 执行器
func NewForkExecutor() *ForkExecutor {
	return &ForkExecutor{}
}

// Execute 执行 Fork - 启动并行分支
func (e *ForkExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})

	// 生成 join_id 供 Join 节点使用
	joinID := fmt.Sprintf("fork_%d", inputs["__node_id"])
	if id, ok := inputs["join_id"]; ok {
		if strID, ok := id.(string); ok && strID != "" {
			joinID = strID
		}
	}

	outputs["join_id"] = joinID
	outputs["started"] = true

	return outputs, nil
}

// Validate 验证节点配置
func (e *ForkExecutor) Validate(node *blueprint.Node) error {
	return nil
}

// JoinExecutor Join 节点执行器 - 等待汇合
type JoinExecutor struct{}

// NewJoinExecutor 创建 Join 执行器
func NewJoinExecutor() *JoinExecutor {
	return &JoinExecutor{}
}

// Execute 执行 Join - 等待所有分支完成（实际等待逻辑在 executor.go 中）
func (e *JoinExecutor) Execute(ctx *blueprint.ExecutionContext, inputs map[string]interface{}) (map[string]interface{}, error) {
	outputs := make(map[string]interface{})
	outputs["completed"] = true
	return outputs, nil
}

// Validate 验证节点配置
func (e *JoinExecutor) Validate(node *blueprint.Node) error {
	return nil
}

