<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="dayTitle"
    width="450px"
  >
    <div class="day-detail">
      <div v-if="day?.todos?.length === 0" class="empty-state">
        <p>当天没有待办事项</p>
      </div>

      <div v-else class="todo-list">
        <div
          v-for="todo in day?.todos"
          :key="todo.id"
          class="todo-item"
          :class="{ completed: todo.isCompleted }"
        >
          <div class="todo-row-1">
            <el-tag 
              size="small" 
              :style="{ background: getTodoTypeColor(todo.type), color: '#fff', border: 'none' }"
            >
              {{ getTodoTypeLabel(todo.type) }}
            </el-tag>
            <span class="todo-title">{{ todo.title }}</span>
            <el-tag v-if="todo.isCompleted" size="small" type="success">已完成</el-tag>
          </div>
          <div class="todo-row-2">
            <span class="todo-time">{{ formatTime(todo.startDate) }} - {{ formatTime(todo.endDate) }}</span>
            <span class="todo-actions">
              <el-button size="small" text @click="handleToggleComplete(todo)">
                {{ todo.isCompleted ? '取消完成' : '完成' }}
              </el-button>
              <el-button size="small" text @click="$emit('edit', todo)">编辑</el-button>
              <el-button size="small" text type="danger" @click="handleDelete(todo)">删除</el-button>
            </span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">关闭</el-button>
      <el-button type="primary" @click="handleCreate">
        <el-icon><Plus /></el-icon>
        添加待办
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit, Delete, Plus } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type CalendarDay = models.CalendarDay
type Todo = models.Todo
type TodoType = { value: string; label: string; icon: string; color: string }

const props = defineProps<{
  visible: boolean
  day: CalendarDay | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  edit: [todo: Todo]
  create: [date: string]
}>()

const todoTypes = ref<TodoType[]>([])

const dayTitle = computed(() => {
  if (!props.day) return ''
  return dayjs(props.day.date).format('YYYY年MM月DD日') + ` (${props.day.lunar.monthName}${props.day.lunar.dayName})`
})

onMounted(async () => {
  try {
    todoTypes.value = await api.GetTodoTypes()
  } catch (error) {
    console.error('Failed to fetch todo types:', error)
  }
})

function getTodoTypeIcon(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.icon || '📋'
}

function getTodoTypeLabel(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.label || type
}

function getTodoTypeColor(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.color || '#999'
}

function formatTime(date: string): string {
  return dayjs(date).format('HH:mm')
}

function truncateContent(content: string): string {
  if (content.length > 100) {
    return content.slice(0, 100) + '...'
  }
  return content
}

async function handleToggleComplete(todo: Todo) {
  try {
    await api.MarkTodoCompleted(todo.id, !todo.isCompleted)
    todo.isCompleted = !todo.isCompleted
    ElMessage.success(todo.isCompleted ? '已完成' : '已取消完成')
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

async function handleDelete(todo: Todo) {
  try {
    await ElMessageBox.confirm(`确定要删除"${todo.title}"吗？`, '确认删除', {
      type: 'warning'
    })
    await api.DeleteTodo(todo.id)
    // 从列表中移除
    if (props.day?.todos) {
      const index = props.day.todos.findIndex(t => t.id === todo.id)
      if (index > -1) {
        props.day.todos.splice(index, 1)
      }
    }
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

function handleCreate() {
  if (props.day?.date) {
    emit('create', props.day.date)
  }
  emit('update:visible', false)
}
</script>

<style lang="scss" scoped>
.day-detail {
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  text-align: center;
  padding: 30px 20px;

  p {
    color: #909399;
    margin: 0;
  }
}

.todo-list {
  .todo-item {
    padding: 10px 12px;
    border: 1px solid #ebeef5;
    border-radius: 6px;
    margin-bottom: 8px;
    transition: all 0.2s;

    &:last-child {
      margin-bottom: 0;
    }

    &:hover {
      background: #f9fafc;
    }

    &.completed {
      opacity: 0.6;

      .todo-title {
        text-decoration: line-through;
        color: #909399;
      }
    }

    .todo-row-1 {
      display: flex;
      align-items: center;
      gap: 8px;

      .type-icon {
        font-size: 16px;
        flex-shrink: 0;
      }

      .todo-title {
        flex: 1;
        font-size: 14px;
        color: #303133;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .todo-row-2 {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-top: 6px;
      padding-left: 24px;

      .todo-time {
        font-size: 12px;
        color: #909399;
      }

      .todo-actions {
        display: flex;
        gap: 4px;

        .el-button {
          padding: 2px 6px;
          font-size: 12px;
        }
      }
    }
  }
}
</style>
