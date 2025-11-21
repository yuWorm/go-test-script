package nodes

import (
	"github.com/go-blueprint-engine/blueprint"
)

// RegisterAllNodes 注册所有内置节点类型到注册表
func RegisterAllNodes(registry *blueprint.NodeRegistry) {
	// ==================== 算术运算节点（可执行） ====================
	registry.Register(blueprint.NodeTypeArithmetic, "add", NewArithmeticExecutor("add"))
	registry.Register(blueprint.NodeTypeArithmetic, "subtract", NewArithmeticExecutor("subtract"))
	registry.Register(blueprint.NodeTypeArithmetic, "multiply", NewArithmeticExecutor("multiply"))
	registry.Register(blueprint.NodeTypeArithmetic, "divide", NewArithmeticExecutor("divide"))
	registry.Register(blueprint.NodeTypeArithmetic, "modulo", NewArithmeticExecutor("modulo"))
	registry.Register(blueprint.NodeTypeArithmetic, "power", NewArithmeticExecutor("power"))

	// ==================== 比较运算节点（可执行） ====================
	registry.Register(blueprint.NodeTypeArithmetic, "equal", NewComparisonExecutor("equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "not_equal", NewComparisonExecutor("not_equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "greater", NewComparisonExecutor("greater"))
	registry.Register(blueprint.NodeTypeArithmetic, "greater_equal", NewComparisonExecutor("greater_equal"))
	registry.Register(blueprint.NodeTypeArithmetic, "less", NewComparisonExecutor("less"))
	registry.Register(blueprint.NodeTypeArithmetic, "less_equal", NewComparisonExecutor("less_equal"))

	// ==================== 逻辑运算节点（可执行） ====================
	registry.Register(blueprint.NodeTypeLogic, "and", NewLogicExecutor("and"))
	registry.Register(blueprint.NodeTypeLogic, "or", NewLogicExecutor("or"))
	registry.Register(blueprint.NodeTypeLogic, "not", NewLogicExecutor("not"))
	registry.Register(blueprint.NodeTypeLogic, "xor", NewLogicExecutor("xor"))

	// ==================== 控制流节点（自定义执行引脚） ====================
	registry.Register(blueprint.NodeTypeFlowControl, "branch", NewBranchExecutor())
	registry.Register(blueprint.NodeTypeFlowControl, "for_loop", NewForLoopExecutor())
	registry.Register(blueprint.NodeTypeFlowControl, "while_loop", NewWhileLoopExecutor())
	registry.Register(blueprint.NodeTypeFlowControl, "sequence", NewSequenceExecutor())

	// 兼容旧的控制流节点注册
	registry.Register(blueprint.NodeTypeCondition, "", NewConditionExecutor())
	registry.Register(blueprint.NodeTypeCondition, "if_else", NewIfElseExecutor())
	registry.Register(blueprint.NodeTypeCondition, "for_loop", NewForLoopExecutor())
	registry.Register(blueprint.NodeTypeCondition, "while_loop", NewWhileLoopExecutor())
	registry.Register(blueprint.NodeTypeCondition, "break", NewBreakExecutor())
	registry.Register(blueprint.NodeTypeCondition, "continue", NewContinueExecutor())
	registry.Register(blueprint.NodeTypeFunction, "branch", NewBranchExecutor())

	// ==================== 纯数据节点（无执行引脚） ====================
	registry.Register(blueprint.NodeTypeData, "constant", NewDataConstantExecutor())
	registry.Register(blueprint.NodeTypeData, "get_variable", NewDataGetVariableExecutor())
	registry.Register(blueprint.NodeTypeData, "make_float", NewMakeValueExecutor("float"))
	registry.Register(blueprint.NodeTypeData, "make_bool", NewMakeValueExecutor("bool"))
	registry.Register(blueprint.NodeTypeData, "make_string", NewMakeValueExecutor("string"))

	// 兼容旧的函数节点
	registry.Register(blueprint.NodeTypeFunction, "constant", NewConstantExecutor())
	registry.Register(blueprint.NodeTypeFunction, "get_variable", NewVariableGetExecutor())

	// ==================== 变量操作节点（可执行） ====================
	registry.Register(blueprint.NodeTypeVariable, "set_variable", NewVariableSetExecutor())
	registry.Register(blueprint.NodeTypeFunction, "set_variable", NewVariableSetExecutor())

	// ==================== 调试节点（可执行） ====================
	registry.Register(blueprint.NodeTypeDebug, "print", NewPrintExecutor())

	// ==================== 异步节点（可执行） ====================
	registry.Register(blueprint.NodeTypeAsync, "delay", NewDelayExecutor())
	registry.Register(blueprint.NodeTypeFunction, "sleep", NewSleepExecutor())

	// 兼容旧的异步节点
	registry.Register(blueprint.NodeTypeFunction, "async_start", NewAsyncStartExecutor())
	registry.Register(blueprint.NodeTypeFunction, "async_end", NewAsyncEndExecutor())
	registry.Register(blueprint.NodeTypeFunction, "await", NewAwaitExecutor())
	registry.Register(blueprint.NodeTypeFunction, "parallel", NewParallelExecutor())

	// ==================== 特殊节点 ====================
	registry.Register(blueprint.NodeTypeStart, "", NewStartExecutor())
	registry.Register(blueprint.NodeTypeEnd, "", NewEndExecutor())
}
