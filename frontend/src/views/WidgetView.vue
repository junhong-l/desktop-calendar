<template>
  <div class="widget-view">
    <div class="widget-header">
      <div class="header-left">
        <span class="widget-title">📅 第{{ currentWeek }}周待办</span>
      </div>
      <div class="header-right">
        <span class="widget-date">{{ currentDate }}</span>
        <span class="refresh-btn" @click.stop="fetchData" title="刷新">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6"/>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
        </span>
      </div>
    </div>

    <div class="widget-content">
      <!-- 本周待办 -->
      <div v-if="weekTodos.length > 0" class="todo-section">
        <div class="todo-list">
          <div 
            v-for="todo in weekTodos" 
            :key="todo.id"
            class="todo-item"
            :style="{ 
              background: getTodoStatusBg(todo), 
              color: getTodoStatusColor(todo),
              border: getTodoStatusBorder(todo)
            }"
            @click="handleTodoClick(todo)"
          >
            <span class="status-tag" :style="{ background: getStatusTagColor(todo) }">
              {{ getStatusText(todo) }}
            </span>
            <div class="todo-row-1">
              <span 
                class="complete-btn" 
                @click.stop="handleComplete(todo)"
                title="标记完成"
              >✓</span>
              <el-tag 
                size="small" 
                :color="getTodoTypeColor(todo.type)"
                effect="dark"
                class="type-tag"
              >
                {{ getTodoTypeLabel(todo.type) }}
              </el-tag>
              <span class="todo-title">{{ todo.title }}</span>
            </div>
            <div class="todo-row-2">
              {{ formatTimeRange(todo.startDate, todo.endDate) }}
            </div>
          </div>
        </div>
      </div>

      <!-- 空状态 -->
      <div v-if="weekTodos.length === 0" class="empty-state">
        <span class="empty-icon">🎉</span>
        <p>本周没有待办事项</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import weekOfYear from 'dayjs/plugin/weekOfYear'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

// 配置 dayjs
dayjs.locale('zh-cn')
dayjs.extend(weekOfYear)

type Todo = models.Todo
type WeekTodosResult = models.WeekTodosResult
type TodoType = { value: string; label: string; icon: string; color: string }

const weekTodosData = ref<WeekTodosResult | null>(null)
const todoTypes = ref<TodoType[]>([])

// 当前日期：MM月DD日 星期X
const currentDate = computed(() => dayjs().format('MM月DD日 dddd'))

// 当前是第几周
const currentWeek = computed(() => dayjs().week())

// 合并本周待办和逾期待办，只显示未完成的
const weekTodos = computed(() => {
  const overdue = weekTodosData.value?.overdue || []
  const todos = weekTodosData.value?.todos || []
  return [...overdue, ...todos]
})

async function fetchData() {
  try {
    const [todosResult, types] = await Promise.all([
      api.GetWeekTodosNew(),
      api.GetTodoTypes()
    ])
    weekTodosData.value = todosResult
    todoTypes.value = types
  } catch (error) {
    console.error('Failed to fetch data:', error)
  }
}

function getTodoTypeLabel(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.label || type
}

function getTodoTypeColor(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.color || '#999'
}

// 根据待办状态获取背景色
function getTodoStatusBg(todo: Todo): string {
  const now = dayjs()
  const startDate = dayjs(todo.startDate)
  const endDate = dayjs(todo.endDate)
  
  // 超时（已过结束时间）- 浅红色背景
  if (endDate.isBefore(now)) {
    return '#fef0f0'
  }
  
  // 进行中（已过开始时间，未到结束时间）- 浅绿色背景
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return '#f0f9eb'
  }
  
  // 即将开始（8小时内）- 浅蓝色背景
  if (startDate.diff(now, 'hour') < 8) {
    return '#ecf5ff'
  }
  
  // 未开始 - 浅灰色
  return '#f5f5f5'
}

// 根据待办状态获取文字颜色
function getTodoStatusColor(todo: Todo): string {
  const now = dayjs()
  const endDate = dayjs(todo.endDate)
  
  // 超时 - 红色文字
  if (endDate.isBefore(now)) {
    return '#c45656'
  }
  
  // 其他 - 深色文字
  return '#333333'
}

// 根据待办状态获取边框颜色
function getTodoStatusBorder(todo: Todo): string {
  const now = dayjs()
  const startDate = dayjs(todo.startDate)
  const endDate = dayjs(todo.endDate)
  
  // 超时 - 红色边框
  if (endDate.isBefore(now)) {
    return '1px solid #f56c6c'
  }
  
  // 进行中 - 绿色边框
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return '1px solid #67c23a'
  }
  
  // 即将开始（8小时内）- 蓝色边框
  if (startDate.diff(now, 'hour') < 8) {
    return '1px solid #409eff'
  }
  
  // 未开始 - 灰色边框
  return '1px solid #dcdfe6'
}

// 获取状态文字
function getStatusText(todo: Todo): string {
  const now = dayjs()
  const startDate = dayjs(todo.startDate)
  const endDate = dayjs(todo.endDate)
  
  if (endDate.isBefore(now)) {
    return '已超时'
  }
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return '进行中'
  }
  if (startDate.diff(now, 'hour') < 8) {
    return '即将开始'
  }
  return '未开始'
}

// 获取状态标签颜色
function getStatusTagColor(todo: Todo): string {
  const now = dayjs()
  const startDate = dayjs(todo.startDate)
  const endDate = dayjs(todo.endDate)
  
  if (endDate.isBefore(now)) {
    return '#f56c6c'  // 红色 - 已超时
  }
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return '#67c23a'  // 绿色 - 进行中
  }
  if (startDate.diff(now, 'hour') < 8) {
    return '#409eff'  // 蓝色 - 即将开始
  }
  return '#909399'  // 灰色 - 未开始
}

function formatScheduledTime(date: string): string {
  return dayjs(date).format('YYYY年MM月DD日 HH:mm')
}

// 格式化时间范围
function formatTimeRange(startDate: string, endDate: string): string {
  if (!startDate) return ''
  const start = dayjs(startDate)
  const end = endDate ? dayjs(endDate) : null
  
  const startStr = start.format('YYYY年MM月DD日 HH:mm')
  if (end) {
    const endStr = end.format('YYYY年MM月DD日 HH:mm')
    return `${startStr} - ${endStr}`
  }
  return startStr
}

// 点击待办打开主软件并显示详情
async function handleTodoClick(todo: Todo) {
  try {
    await api.OpenMainWindowWithTodo(todo.id)
  } catch (error) {
    console.error('Failed to open main window:', error)
  }
}

// 标记待办完成
async function handleComplete(todo: Todo) {
  try {
    await api.MarkTodoCompleted(todo.id, true)
    await fetchData()
  } catch (error) {
    console.error('Failed to complete todo:', error)
  }
}

onMounted(() => {
  fetchData()
  // 每分钟刷新一次
  setInterval(fetchData, 60000)
})
</script>

<style lang="scss" scoped>
.widget-view {
  width: 100%;
  height: 100vh;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.widget-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  cursor: grab;
  user-select: none;
  --wails-draggable: drag;

  &:active {
    cursor: grabbing;
  }

  .header-left {
    display: flex;
    align-items: center;
  }

  .widget-title {
    font-size: 16px;
    font-weight: 600;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .refresh-btn {
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    padding: 5px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 6px;
    transition: all 0.3s;
    
    svg {
      width: 16px;
      height: 16px;
    }
    
    &:hover {
      transform: rotate(180deg);
      background: rgba(255, 255, 255, 0.3);
    }
  }

  .widget-date {
    font-size: 12px;
    opacity: 0.9;
  }
}

.widget-content {
  padding: 15px;
  flex: 1;
  overflow-y: auto;
}

.todo-section {
  margin-bottom: 15px;
}

.todo-list {
  .todo-item {
    position: relative;
    padding: 12px 15px;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
    margin-bottom: 8px;

    &:hover {
      filter: brightness(0.95);
    }

    .status-tag {
      position: absolute;
      top: 0;
      right: 0;
      padding: 2px 8px;
      font-size: 10px;
      color: #fff;
      border-radius: 0 8px 0 8px;
    }

    .todo-row-1 {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 6px;

      .complete-btn {
        width: 20px;
        height: 20px;
        border-radius: 50%;
        border: 2px solid #67c23a;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 12px;
        color: transparent;
        cursor: pointer;
        transition: all 0.2s;
        flex-shrink: 0;

        &:hover {
          background: #67c23a;
          color: #fff;
        }
      }

      .type-tag {
        border: none;
        flex-shrink: 0;
      }

      .todo-title {
        font-size: 14px;
        color: #303133;
        font-weight: 500;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .instance-index {
        font-size: 12px;
        color: #909399;
        flex-shrink: 0;
      }
    }

    .todo-row-2 {
      font-size: 12px;
      color: #909399;
      margin-left: 30px;
    }
  }
}

.empty-state {
  text-align: center;
  padding: 40px 20px;

  .empty-icon {
    font-size: 48px;
    display: block;
    margin-bottom: 10px;
  }

  p {
    color: #909399;
    margin: 0;
  }
}
</style>
