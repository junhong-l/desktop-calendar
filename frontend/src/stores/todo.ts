import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '@/wailsjs/go/app/App'

export const useTodoStore = defineStore('todo', () => {
  const todos = ref<any[]>([])
  const pendingTodos = ref<any[]>([]) // 待处理的待办列表
  const loading = ref(false)
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const totalPages = ref(0)

  // 获取待办列表
  async function fetchTodos(filter: any = {}) {
    loading.value = true
    try {
      const result: any = await api.GetTodoList({
        keyword: filter.keyword || '',
        year: filter.year || 0,
        month: filter.month || 0,
        types: filter.types || [],
        completed: filter.completed,
        page: filter.page || page.value,
        pageSize: filter.pageSize || pageSize.value
      })
      todos.value = result.todos || []
      total.value = result.total
      page.value = result.page
      pageSize.value = result.pageSize
      totalPages.value = result.totalPages
    } catch (error) {
      console.error('Failed to fetch todos:', error)
    } finally {
      loading.value = false
    }
  }

  // 获取未完成的待办列表
  async function fetchPendingTodos() {
    loading.value = true
    try {
      const result: any[] = await api.GetPendingTodos()
      pendingTodos.value = result || []
      total.value = pendingTodos.value.length
    } catch (error) {
      console.error('Failed to fetch pending todos:', error)
      pendingTodos.value = []
    } finally {
      loading.value = false
    }
  }

  // 创建待办
  async function createTodo(todo: any): Promise<number> {
    const id = await api.CreateTodo(todo as any)
    await fetchPendingTodos()
    return id
  }

  // 更新待办
  async function updateTodo(todo: any): Promise<void> {
    await api.UpdateTodo(todo as any)
    await fetchPendingTodos()
  }

  // 删除待办
  async function deleteTodo(id: number): Promise<void> {
    await api.DeleteTodo(id)
    await fetchPendingTodos()
  }

  // 标记待办完成
  async function markTodoCompleted(id: number, completed: boolean): Promise<void> {
    await api.MarkTodoCompleted(id, completed)
    await fetchPendingTodos()
  }

  // 获取单个待办
  async function getTodo(id: number): Promise<any> {
    return api.GetTodo(id)
  }

  return {
    todos,
    pendingTodos,
    loading,
    total,
    page,
    pageSize,
    totalPages,
    fetchTodos,
    fetchPendingTodos,
    createTodo,
    updateTodo,
    deleteTodo,
    markTodoCompleted,
    getTodo
  }
})
