<template>
  <div class="notification-popup">
    <div class="popup-header">
      <span class="popup-icon">{{ getIcon() }}</span>
      <span class="popup-type">{{ notifyType }}</span>
      <button class="close-btn" @click="closePopup">×</button>
    </div>
    <div class="popup-content" @click="viewDetail">
      <h3 class="popup-title">{{ title }}</h3>
      <p class="popup-message" v-if="message">{{ message }}</p>
      <p class="popup-time">{{ formatTimeRange() }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { EventsOn, Quit } from '@/wailsjs/runtime/runtime'
import * as api from '@/wailsjs/go/app/App'

const title = ref('待办提醒')
const message = ref('')
const todoId = ref(0)
const notifyType = ref('提醒')
const startTime = ref('')
const endTime = ref('')

function getIcon() {
  switch (notifyType.value) {
    case '提前提醒': return '⏰'
    case '开始提醒': return '🔔'
    case '结束提醒': return '✅'
    default: return '📅'
  }
}

function formatTimeRange() {
  if (!startTime.value) return ''
  
  // 格式化开始时间
  const startStr = formatDateTime(startTime.value)
  
  // 格式化结束时间
  if (endTime.value) {
    const endStr = formatDateTime(endTime.value)
    return `${startStr} - ${endStr}`
  }
  return startStr
}

// 将 "2006-01-02 15:04" 格式转换为 "YYYY年MM月DD日 HH:mm"
function formatDateTime(dateStr: string): string {
  if (!dateStr) return ''
  const [date, time] = dateStr.split(' ')
  if (!date) return ''
  const [year, month, day] = date.split('-')
  return `${year}年${month}月${day}日 ${time || ''}`
}

async function viewDetail() {
  if (todoId.value > 0) {
    try {
      await api.OpenMainWindowWithTodo(todoId.value)
    } catch (e) {
      console.error('Failed to open todo:', e)
    }
  }
  closePopup()
}

function closePopup() {
  Quit()
}

onMounted(() => {
  // 监听通知数据
  EventsOn('notification:show', (data: any) => {
    title.value = data.title || '待办提醒'
    message.value = data.message || ''
    todoId.value = data.todoId || 0
    notifyType.value = data.type || '提醒'
    startTime.value = data.startTime || ''
    endTime.value = data.endTime || ''
  })

  // 自动关闭（可配置）
  EventsOn('notification:autoclose', (seconds: number) => {
    if (seconds > 0) {
      setTimeout(closePopup, seconds * 1000)
    }
  })
})
</script>

<style lang="scss" scoped>
.notification-popup {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-family: 'Microsoft YaHei', sans-serif;

  .popup-header {
    display: flex;
    align-items: center;
    padding: 12px 14px;
    background: rgba(255, 255, 255, 0.15);
    
    .popup-icon {
      font-size: 20px;
      margin-right: 10px;
    }
    
    .popup-type {
      flex: 1;
      font-size: 13px;
      color: rgba(255, 255, 255, 0.95);
      font-weight: 500;
    }
    
    .close-btn {
      width: 28px;
      height: 28px;
      border: none;
      background: rgba(255, 255, 255, 0.25);
      color: white;
      border-radius: 50%;
      cursor: pointer;
      font-size: 18px;
      font-weight: bold;
      line-height: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      transition: all 0.2s;
      
      &:hover {
        background: rgba(255, 0, 0, 0.6);
        transform: scale(1.1);
      }
    }
  }
  
  .popup-content {
    flex: 1;
    padding: 12px 14px;
    color: white;
    cursor: pointer;
    transition: background 0.2s;
    
    &:hover {
      background: rgba(255, 255, 255, 0.1);
    }
    
    .popup-title {
      margin: 0 0 8px;
      font-size: 15px;
      font-weight: 600;
      line-height: 1.3;
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }
    
    .popup-message {
      margin: 0 0 8px;
      font-size: 13px;
      color: rgba(255, 255, 255, 0.85);
      overflow: hidden;
      text-overflow: ellipsis;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }
    
    .popup-time {
      margin: 0;
      font-size: 11px;
      color: rgba(255, 255, 255, 0.7);
    }
  }
}
</style>
