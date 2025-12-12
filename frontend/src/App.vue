<template>
  <!-- 小部件模式：只显示小部件视图 -->
  <div v-if="isWidgetMode" id="widget-container">
    <WidgetView />
  </div>
  
  <!-- 通知弹窗模式：只显示通知弹窗视图 -->
  <div v-else-if="isNotificationPopupMode" id="notification-popup-container">
    <NotificationPopupView />
  </div>
  
  <!-- 正常模式：显示完整应用 -->
  <div v-else id="app-container">
    <el-container class="main-container">
      <el-aside width="220px" class="app-aside">
        <div class="logo">
          <span class="logo-icon">📅</span>
          <span class="logo-text">待办日历</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          class="app-menu"
          router
        >
          <el-menu-item index="/">
            <el-icon><Calendar /></el-icon>
            <span>日历</span>
          </el-menu-item>
          <el-menu-item index="/todos">
            <el-icon><List /></el-icon>
            <span>待办事项</span>
          </el-menu-item>
          <el-menu-item index="/history">
            <el-icon><Clock /></el-icon>
            <span>历史记录</span>
          </el-menu-item>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <span>设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="app-main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>

    <!-- 通知弹窗 -->
    <NotificationDialog 
      v-model:visible="notificationVisible"
      :notification="currentNotification"
      @close="handleNotificationClose"
      @viewDetail="handleViewDetail"
    />

    <!-- 待办编辑弹窗 -->
    <TodoFormDialog
      v-model:visible="todoFormVisible"
      :todo="selectedTodo"
      @saved="handleTodoSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Calendar, List, Clock, Setting } from '@element-plus/icons-vue'
import NotificationDialog from '@/components/NotificationDialog.vue'
import TodoFormDialog from '@/components/TodoFormDialog.vue'
import WidgetView from '@/views/WidgetView.vue'
import NotificationPopupView from '@/views/NotificationPopupView.vue'
import { useNotificationStore } from '@/stores/notification'
import { EventsOn, EventsOff } from '@/wailsjs/runtime/runtime'
import * as api from '@/wailsjs/go/app/App'
import { GetMode } from '@/wailsjs/go/app/WindowModeService'

const route = useRoute()
const notificationStore = useNotificationStore()

// 检测是否是小部件模式（通过窗口大小或 URL hash 判断）
const isWidgetMode = ref(false)
// 检测是否是通知弹窗模式
const isNotificationPopupMode = ref(false)

// 检测小部件模式 - 使用后端提供的模式服务
async function checkWidgetMode() {
  try {
    const mode = await GetMode()
    if (mode === 'notification') {
      isNotificationPopupMode.value = true
      isWidgetMode.value = false
      return
    } else if (mode === 'widget') {
      isWidgetMode.value = true
      isNotificationPopupMode.value = false
      return
    }
    // main 模式
    isWidgetMode.value = false
    isNotificationPopupMode.value = false
  } catch (error) {
    // 如果后端服务不可用，使用传统的检测方式
    // 先检测是否是通知弹窗模式（通过 URL hash 判断）
    if (window.location.hash === '#/notification-popup' || 
        window.location.hash.startsWith('#/notification-popup?')) {
      isNotificationPopupMode.value = true
      isWidgetMode.value = false
      return
    }
    
    // 小部件窗口宽度为 340px
    isWidgetMode.value = window.innerWidth <= 400 || 
      window.location.hash === '#/widget' || 
      window.location.hash.startsWith('#/widget?')
  }
}

const activeMenu = computed(() => route.path)
const notificationVisible = ref(false)
const currentNotification = ref<any>(null)
const todoFormVisible = ref(false)
const selectedTodo = ref<any>(null)

// 检查IPC文件中的待办ID
let ipcCheckInterval: number | null = null

async function checkIPCTodo() {
  try {
    const todoId = await api.CheckIPCTodo()
    if (todoId > 0) {
      const todo = await api.GetTodo(todoId)
      if (todo) {
        selectedTodo.value = todo
        todoFormVisible.value = true
      }
    }
  } catch (error) {
    // 忽略错误
  }
}

// 监听后端通知事件
onMounted(async () => {
  // 检测小部件模式
  await checkWidgetMode()
  
  // 监听hash变化以支持模式切换（但异步模式下不使用）
  // window.addEventListener('hashchange', checkWidgetMode)
  
  // 小部件模式或通知弹窗模式不检查通知
  if (isWidgetMode.value || isNotificationPopupMode.value) return
  
  // 启动IPC检查定时器（每500ms检查一次）
  ipcCheckInterval = window.setInterval(checkIPCTodo, 500)
  
  EventsOn('todo:notification', (data: any) => {
    currentNotification.value = data
    notificationVisible.value = true
    notificationStore.addNotification(data)
  })

  // 监听从小部件打开待办详情的事件
  EventsOn('open:todo', async (todoIdStr: string) => {
    try {
      const todoId = parseInt(todoIdStr, 10)
      if (todoId > 0) {
        const todo = await api.GetTodo(todoId)
        if (todo) {
          selectedTodo.value = todo
          todoFormVisible.value = true
        }
      }
    } catch (error) {
      console.error('Failed to open todo:', error)
    }
  })
})

onUnmounted(() => {
  EventsOff('todo:notification')
  EventsOff('open:todo')
  if (ipcCheckInterval) {
    clearInterval(ipcCheckInterval)
  }
})

const handleNotificationClose = () => {
  notificationVisible.value = false
}

const handleViewDetail = (todo: any) => {
  selectedTodo.value = todo
  todoFormVisible.value = true
}

const handleTodoSaved = () => {
  todoFormVisible.value = false
  selectedTodo.value = null
}
</script>

<style lang="scss">
#widget-container {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: transparent;
}

#app-container {
  width: 100vw;
  height: 100vh;
  overflow: hidden;
}

.main-container {
  height: 100%;
}

.app-aside {
  background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
  padding: 20px 0;
  
  .logo {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 10px 20px 30px;
    
    .logo-icon {
      font-size: 28px;
      margin-right: 10px;
    }
    
    .logo-text {
      font-size: 20px;
      font-weight: bold;
      color: #fff;
    }
  }
  
  .app-menu {
    background: transparent;
    border: none;
    
    .el-menu-item {
      color: rgba(255, 255, 255, 0.8);
      margin: 5px 10px;
      border-radius: 8px;
      
      &:hover {
        background: rgba(255, 255, 255, 0.1);
        color: #fff;
      }
      
      &.is-active {
        background: rgba(255, 255, 255, 0.2);
        color: #fff;
      }
    }
  }
}

.app-main {
  background: linear-gradient(180deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
