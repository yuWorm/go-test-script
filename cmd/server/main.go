package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-blueprint-engine/blueprint"
	"github.com/go-blueprint-engine/nodes"
)

// 启用 CORS
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// 执行蓝图处理器
func handleExecuteBlueprint(registry *blueprint.NodeRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析请求体
		var bp blueprint.Blueprint
		if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		// 初始化蓝图内部字段
		bp.InitNodeMap()

		// 编译蓝图
		compiler := blueprint.NewCompiler(registry)
		if err := compiler.Compile(&bp); err != nil {
			http.Error(w, fmt.Sprintf("Failed to compile blueprint: %v", err), http.StatusBadRequest)
			return
		}

		// 执行蓝图
		executor := blueprint.NewExecutor(blueprint.DefaultExecutionOptions())
		result, err := executor.Execute(&bp, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to execute blueprint: %v", err), http.StatusInternalServerError)
			return
		}

		// 构建响应
		response := map[string]interface{}{
			"success":   result.Success,
			"duration":  result.Duration,
			"variables": result.Variables,
			"outputs":   result.Outputs,
		}

		if len(result.Errors) > 0 {
			errorStrs := make([]string, len(result.Errors))
			for i, err := range result.Errors {
				errorStrs[i] = err.Error()
			}
			response["errors"] = errorStrs
		}

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// 验证蓝图处理器
func handleValidateBlueprint(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var bp blueprint.Blueprint
	if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// 初始化内部字段
	bp.InitNodeMap()

	// 验证蓝图
	if err := bp.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid":  false,
			"errors": []string{err.Error()},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid": true,
	})
}

// 编译蓝图处理器
func handleCompileBlueprint(registry *blueprint.NodeRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var bp blueprint.Blueprint
		if err := json.NewDecoder(r.Body).Decode(&bp); err != nil {
			http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		// 初始化内部字段
		bp.InitNodeMap()

		// 编译蓝图
		compiler := blueprint.NewCompiler(registry)
		if err := compiler.Compile(&bp); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Blueprint compiled successfully",
		})
	}
}

// 健康检查
func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func main() {
	// 创建节点注册表
	registry := blueprint.NewNodeRegistry()
	nodes.RegisterAllNodes(registry)

	// 设置路由
	http.HandleFunc("/api/blueprint/execute", handleExecuteBlueprint(registry))
	http.HandleFunc("/api/blueprint/validate", handleValidateBlueprint)
	http.HandleFunc("/api/blueprint/compile", handleCompileBlueprint(registry))
	http.HandleFunc("/api/health", handleHealth)

	// 启动服务器
	port := ":8080"
	fmt.Printf("🚀 Blueprint API Server starting on http://localhost%s\n", port)
	fmt.Println("📍 Endpoints:")
	fmt.Println("   POST /api/blueprint/execute  - Execute a blueprint")
	fmt.Println("   POST /api/blueprint/validate - Validate a blueprint")
	fmt.Println("   POST /api/blueprint/compile  - Compile a blueprint")
	fmt.Println("   GET  /api/health             - Health check")
	fmt.Println("")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
