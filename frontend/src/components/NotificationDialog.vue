<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    title="待办提醒"
    width="500px"
    :close-on-click-modal="false"
    class="notification-dialog"
  >
    <div v-if="notifications && notifications.length > 0" class="notification-content">
      <div class="notification-list">
        <div
          v-for="item in notifications"
          :key="item.todo.id"
          class="notification-item"
          @click="handleViewDetail(item)"
        >
          <el-checkbox
            v-model="selectedIds"
            :value="item.todo.id"
            @click.stop
          />
          <div class="item-content">
            <div class="item-row-1">
              <el-tag size="small" :style="{ background: getTodoTypeColor(item.todo.type), color: '#fff', border: 'none' }">
                {{ getTodoTypeLabel(item.todo.type) }}
              </el-tag>
              <span class="todo-title">{{ item.todo.title }}</span>
            </div>
            <div class="item-row-2">
              <span class="todo-time">{{ formatDateTime(item.todo.startDate) }}</span>
              <span v-if="item.totalCount > 1" class="cycle-info">
                第 {{ item.currentCount }} 次循环 / 共 {{ item.totalCount }} 次
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div v-else class="empty-state">
      <p>暂无待办提醒</p>
    </div>

    <template #footer>
      <el-button @click="handleSnooze">稍后提醒</el-button>
      <el-button type="success" :disabled="selectedIds.length === 0" @click="handleCompleteSelected">
        <el-icon><Check /></el-icon>
        标记完成 {{ selectedIds.length > 0 ? `(${selectedIds.length})` : '' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Check } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'

type TodoType = { value: string; label: string; icon: string; color: string }

interface NotificationData {
  todo: {
    id: number
    title: string
    content: string
    type: string
    startDate: string
    endDate: string
  }
  currentCount: number
  totalCount: number
  message: string
}

const props = defineProps<{
  visible: boolean
  notification: NotificationData | null
  notifications?: NotificationData[]
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  close: []
  viewDetail: [todo: any]
}>()

const todoTypes = ref<TodoType[]>([])
const selectedIds = ref<number[]>([])

// 兼容单个和多个通知
const notifications = ref<NotificationData[]>([])

watch(() => [props.notification, props.notifications, props.visible], () => {
  if (props.visible) {
    if (props.notifications && props.notifications.length > 0) {
      notifications.value = props.notifications
    } else if (props.notification) {
      notifications.value = [props.notification]
    } else {
      notifications.value = []
    }
    selectedIds.value = []
  }
}, { immediate: true })

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

function formatDateTime(date: string): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

function handleSnooze() {
  emit('update:visible', false)
  emit('close')
  ElMessage.info('将在稍后再次提醒')
}

async function handleCompleteSelected() {
  if (selectedIds.value.length === 0) return
  try {
    for (const id of selectedIds.value) {
      await api.MarkTodoCompleted(id, true)
    }
    ElMessage.success(`已标记 ${selectedIds.value.length} 项完成`)
    // 从列表中移除已完成的项
    notifications.value = notifications.value.filter(
      n => !selectedIds.value.includes(n.todo.id)
    )
    selectedIds.value = []
    if (notifications.value.length === 0) {
      emit('update:visible', false)
      emit('close')
    }
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

function handleViewDetail(item: NotificationData) {
  emit('viewDetail', item.todo)
  emit('update:visible', false)
}
</script>

<style lang="scss" scoped>
.notification-content {
  padding: 10px 0;

  .notification-list {
    max-height: 400px;
    overflow-y: auto;
  }

  .notification-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    .el-checkbox {
      margin-top: 2px;
    }

    .item-content {
      flex: 1;
      min-width: 0;

      .item-row-1 {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 4px;

        .todo-title {
          font-size: 15px;
          font-weight: 500;
          color: #303133;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      }

      .item-row-2 {
        display: flex;
        align-items: center;
        gap: 12px;
        font-size: 12px;
        color: #909399;

        .cycle-info {
          color: #E6A23C;
        }
      }
    }
  }
}

.empty-state {
  text-align: center;
  padding: 40px 0;
  color: #909399;
}

:deep(.notification-dialog) {
  .el-dialog__header {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    
    .el-dialog__title {
      color: #fff;
    }
    
    .el-dialog__headerbtn .el-dialog__close {
      color: #fff;
    }
  }
}
</style>
