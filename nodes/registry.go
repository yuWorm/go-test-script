package nodes

import (
	"github.com/go-blueprint-engine/blueprint"
)

// RegisterAllNodes 注册所有内置节点类型到注册表
func RegisterAllNodes(registry *blueprint.NodeRegistry) {
	// 算术运算节点
	registry.Register(blueprint.NodeTypeArithmetic, "add", NewArithmeticExecutor("add"))
	registry.Register(blueprint.NodeTypeArithmetic, "subtract", NewArithmeticExecutor("subtract"))
	registry.Register(blueprint.NodeTypeArithmetic, "multiply", NewArithmeticExecutor("multiply"))
	registry.Register(blueprint.NodeTypeArithmetic, "divide", NewArithmeticExecutor("divide"))
	registry.Register(blueprint.NodeTypeArithmetic, "modulo", NewArithmeticExecutor("modulo"))
	registry.Register(blueprint.NodeTypeArithmetic, "power", NewArithmeticExecutor("power"))

	// 比较运算节点
	registry.Register(blueprint.NodeTypeArithmetic, "equal", NewComparisonExecutor("equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "not_equal", NewComparisonExecutor("not_equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "greater", NewComparisonExecutor("greater"))
	registry.Register(blueprint.NodeTypeArithmetic, "greater_equal", NewComparisonExecutor("greater_equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "less", NewComparisonExecutor("less"))
	registry.Register(blueprint.NodeTypeArithmetic, "less_equal", NewComparisonExecutor("less_equal"))

	// 逻辑运算节点
	registry.Register(blueprint.NodeTypeLogic, "and", NewLogicExecutor("and"))
	registry.Register(blueprint.NodeTypeLogic, "or", NewLogicExecutor("or"))
	registry.Register(blueprint.NodeTypeLogic, "not", NewLogicExecutor("not"))
	registry.Register(blueprint.NodeTypeLogic, "xor", NewLogicExecutor("xor"))

	// 控制流节点
	registry.Register(blueprint.NodeTypeCondition, "", NewConditionExecutor())
	registry.Register(blueprint.NodeTypeCondition, "if_else", NewIfElseExecutor())
	registry.Register(blueprint.NodeTypeCondition, "for_loop", NewForLoopExecutor())
	registry.Register(blueprint.NodeTypeCondition, "while_loop", NewWhileLoopExecutor())
	registry.Register(blueprint.NodeTypeCondition, "break", NewBreakExecutor())
	registry.Register(blueprint.NodeTypeCondition, "continue", NewContinueExecutor())

	// 执行流节点
	registry.Register(blueprint.NodeTypeFunction, "branch", NewBranchExecutor())

	// 特殊节点
	registry.Register(blueprint.NodeTypeStart, "", NewStartExecutor())
	registry.Register(blueprint.NodeTypeEnd, "", NewEndExecutor())

	// 函数节点（常量、变量）
	registry.Register(blueprint.NodeTypeFunction, "constant", NewConstantExecutor())
	registry.Register(blueprint.NodeTypeFunction, "get_variable", NewVariableGetExecutor())
	registry.Register(blueprint.NodeTypeFunction, "set_variable", NewVariableSetExecutor())
	registry.Register(blueprint.NodeTypeFunction, "sleep", NewSleepExecutor())

	// 异步执行节点
	registry.Register(blueprint.NodeTypeFunction, "async_start", NewAsyncStartExecutor())
	registry.Register(blueprint.NodeTypeFunction, "async_end", NewAsyncEndExecutor())
	registry.Register(blueprint.NodeTypeFunction, "await", NewAwaitExecutor())
	registry.Register(blueprint.NodeTypeFunction, "parallel", NewParallelExecutor())
}
