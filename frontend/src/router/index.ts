import { createRouter, createWebHashHistory, RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Calendar',
    component: () => import('@/views/CalendarView.vue'),
    meta: { title: '日历' }
  },
  {
    path: '/todos',
    name: 'Todos',
    component: () => import('@/views/TodosView.vue'),
    meta: { title: '待办事项' }
  },
  {
    path: '/history',
    name: 'History',
    component: () => import('@/views/HistoryView.vue'),
    meta: { title: '历史记录' }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/SettingsView.vue'),
    meta: { title: '设置' }
  },
  {
    path: '/widget',
    name: 'Widget',
    component: () => import('@/views/WidgetView.vue'),
    meta: { title: '桌面小部件', hideLayout: true }
  },
  {
    path: '/notification-popup',
    name: 'NotificationPopup',
    component: () => import('@/views/NotificationPopupView.vue'),
    meta: { title: '通知弹窗', hideLayout: true }
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
