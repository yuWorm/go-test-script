import axios from 'axios'
import type { Blueprint, ExecutionResult } from '../types/blueprint'

// API 基础 URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

/**
 * 执行蓝图
 */
export async function executeBlueprint(blueprint: Blueprint): Promise<ExecutionResult> {
  try {
    const response = await api.post<ExecutionResult>('/blueprint/execute', blueprint)
    return response.data
  } catch (error: any) {
    console.error('Failed to execute blueprint:', error)
    throw new Error(error.response?.data?.message || error.message || 'Failed to execute blueprint')
  }
}

/**
 * 验证蓝图
 */
export async function validateBlueprint(blueprint: Blueprint): Promise<{ valid: boolean; errors?: string[] }> {
  try {
    const response = await api.post('/blueprint/validate', blueprint)
    return response.data
  } catch (error: any) {
    console.error('Failed to validate blueprint:', error)
    throw new Error(error.response?.data?.message || error.message || 'Failed to validate blueprint')
  }
}

/**
 * 编译蓝图
 */
export async function compileBlueprint(blueprint: Blueprint): Promise<{ success: boolean; message?: string }> {
  try {
    const response = await api.post('/blueprint/compile', blueprint)
    return response.data
  } catch (error: any) {
    console.error('Failed to compile blueprint:', error)
    throw new Error(error.response?.data?.message || error.message || 'Failed to compile blueprint')
  }
}
