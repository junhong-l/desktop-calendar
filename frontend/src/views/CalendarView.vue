<template>
  <div class="calendar-view">
    <div class="calendar-header">
      <div class="header-left">
        <el-button-group>
          <el-button @click="prevMonth">
            <el-icon><ArrowLeft /></el-icon>
          </el-button>
          <el-button @click="nextMonth">
            <el-icon><ArrowRight /></el-icon>
          </el-button>
        </el-button-group>
        <el-button @click="goToToday" class="today-btn">今天</el-button>
      </div>
      <h2 class="current-month">{{ currentYear }}年{{ currentMonth }}月</h2>
      <div class="header-right">
        <el-select v-model="currentYear" @change="fetchCalendar" style="width: 100px; margin-right: 10px;">
          <el-option v-for="year in yearOptions" :key="year" :label="`${year}年`" :value="year" />
        </el-select>
        <el-select v-model="currentMonth" @change="fetchCalendar" style="width: 80px;">
          <el-option v-for="month in 12" :key="month" :label="`${month}月`" :value="month" />
        </el-select>
      </div>
    </div>

    <div class="calendar-body">
      <!-- 星期标题 -->
      <div class="week-header">
        <div v-for="day in weekDays" :key="day" class="week-day">{{ day }}</div>
      </div>

      <!-- 日历格子 -->
      <div class="calendar-grid" :style="{ gridTemplateRows: `repeat(${calendarRows}, 1fr)` }">
        <div
          v-for="day in calendarDays"
          :key="day.date"
          class="calendar-cell"
          :class="{
            'is-today': day.isToday,
            'is-other-month': !day.isCurrentMonth,
            'has-todo': day.todoCount > 0
          }"
          @click="handleDayClick(day)"
          @contextmenu.prevent="handleContextMenu($event, day)"
          @mouseenter="handleMouseEnter($event, day)"
          @mouseleave="handleMouseLeave"
        >
          <div class="cell-header">
            <span class="solar-day">{{ day.day }}</span>
            <span class="lunar-day">{{ getLunarDisplay(day.lunar) }}</span>
          </div>
          <div v-if="day.todoCount > 0" class="todo-indicators">
            <div 
              v-for="(todo, index) in day.todos.slice(0, 3)" 
              :key="todo.id"
              class="todo-indicator"
              :class="[`todo-type-${todo.type}`, { 'is-completed': todo.isCompleted }]"
            >
              <span class="todo-dot" :style="{ background: todo.isCompleted ? '#c0c4cc' : getTodoTypeColor(todo.type) }"></span>
              <span class="todo-title" :class="{ 'completed': todo.isCompleted }">{{ todo.title }}</span>
            </div>
            <div v-if="day.todoCount > 3" class="more-todos">
              +{{ day.todoCount - 3 }} 更多
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Hover提示框 -->
    <div 
      v-if="hoverInfo && hoverVisible"
      class="day-hover-tooltip"
      :style="{ left: hoverPosition.x + 'px', top: hoverPosition.y + 'px' }"
    >
      <div class="tooltip-header">
        <div class="tooltip-date">{{ hoverInfo.solar.year }}年{{ hoverInfo.solar.month }}月{{ hoverInfo.solar.day }}日</div>
        <div class="tooltip-week">第{{ hoverDay?.weekNumber }}周 · {{ getWeekDayName(hoverDay?.date) }}</div>
      </div>
      <div class="tooltip-lunar">
        <span>农历: {{ hoverInfo.lunar.yearName }} {{ hoverInfo.lunar.animal }}年</span>
        <span>{{ hoverInfo.lunar.monthName }}{{ hoverInfo.lunar.dayName }}</span>
      </div>
      <div v-if="hoverInfo.festivals && hoverInfo.festivals.length" class="tooltip-festivals">
        <el-tag v-for="f in hoverInfo.festivals" :key="f" size="small" type="warning">{{ f }}</el-tag>
      </div>
    </div>

    <!-- 右键菜单 -->
    <div 
      v-if="contextMenuVisible"
      class="context-menu"
      :style="{ left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px' }"
    >
      <div class="menu-item" @click="handleCreateTodo">
        <el-icon><Plus /></el-icon>
        <span>新建待办</span>
      </div>
      <div class="menu-item" @click="handleViewDay">
        <el-icon><View /></el-icon>
        <span>查看详情</span>
      </div>
    </div>

    <!-- 待办创建/编辑弹窗 -->
    <TodoFormDialog
      v-model:visible="todoDialogVisible"
      :todo="editingTodo"
      :default-date="defaultDate"
      @saved="handleTodoSaved"
    />

    <!-- 日期详情弹窗 -->
    <DayDetailDialog
      v-model:visible="dayDetailVisible"
      :day="selectedDay"
      @edit="handleEditTodo"
      @create="handleCreateFromDay"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { ArrowLeft, ArrowRight, Plus, View } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import TodoFormDialog from '@/components/TodoFormDialog.vue'
import DayDetailDialog from '@/components/DayDetailDialog.vue'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type CalendarDay = models.CalendarDay
type Todo = models.Todo
type TodoType = { value: string; label: string; icon: string; color: string }

const weekDays = ['一', '二', '三', '四', '五', '六', '日']

const currentYear = ref(dayjs().year())
const currentMonth = ref(dayjs().month() + 1)
const calendarDays = ref<CalendarDay[]>([])
const todoTypes = ref<TodoType[]>([])

// Hover相关
const hoverVisible = ref(false)
const hoverDay = ref<CalendarDay | null>(null)
const hoverInfo = ref<any>(null)
const hoverPosition = ref({ x: 0, y: 0 })

// 右键菜单
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })
const contextMenuDay = ref<CalendarDay | null>(null)

// 弹窗
const todoDialogVisible = ref(false)
const editingTodo = ref<Todo | null>(null)
const dayDetailVisible = ref(false)
const selectedDay = ref<CalendarDay | null>(null)
const defaultDate = ref('')

const yearOptions = computed(() => {
  const years = []
  const nowYear = dayjs().year()
  for (let i = 1950; i <= nowYear + 50; i++) {
    years.push(i)
  }
  return years
})

// 计算日历行数
const calendarRows = computed(() => {
  return Math.ceil(calendarDays.value.length / 7)
})

// 获取日历数据
async function fetchCalendar() {
  try {
    const data = await api.GetCalendarMonth(currentYear.value, currentMonth.value)
    calendarDays.value = data || []
  } catch (error) {
    console.error('Failed to fetch calendar:', error)
    calendarDays.value = []
  }
}

// 获取待办类型
async function fetchTodoTypes() {
  try {
    todoTypes.value = await api.GetTodoTypes()
  } catch (error) {
    console.error('Failed to fetch todo types:', error)
  }
}

function getTodoTypeColor(type: string): string {
  const typeInfo = todoTypes.value.find(t => t.value === type)
  return typeInfo?.color || '#999'
}

// 获取农历显示文字：显示月份+日期
function getLunarDisplay(lunar: any): string {
  if (!lunar) return ''
  // 初一显示月份名（如"冬月"、"腊月"）
  if (lunar.day === 1) {
    return lunar.monthName || ''
  }
  // 其他日子显示月份+日期（如"冬月初二"、"冬月十五"）
  return (lunar.monthName || '') + (lunar.dayName || '')
}

function prevMonth() {
  if (currentMonth.value === 1) {
    currentYear.value--
    currentMonth.value = 12
  } else {
    currentMonth.value--
  }
  fetchCalendar()
}

function nextMonth() {
  if (currentMonth.value === 12) {
    currentYear.value++
    currentMonth.value = 1
  } else {
    currentMonth.value++
  }
  fetchCalendar()
}

function goToToday() {
  currentYear.value = dayjs().year()
  currentMonth.value = dayjs().month() + 1
  fetchCalendar()
}

function getWeekDayName(dateStr: string | undefined): string {
  if (!dateStr) return ''
  const weekNames = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
  return weekNames[dayjs(dateStr).day()]
}

// 鼠标悬停
async function handleMouseEnter(event: MouseEvent, day: CalendarDay) {
  hoverDay.value = day
  const date = dayjs(day.date)
  try {
    const info = await api.GetLunarDate(date.year(), date.month() + 1, date.date())
    hoverInfo.value = {
      solar: { year: date.year(), month: date.month() + 1, day: date.date() },
      lunar: info,
      festivals: [] // 可以从后端获取节日信息
    }
    hoverPosition.value = {
      x: event.clientX + 10,
      y: event.clientY + 10
    }
    hoverVisible.value = true
  } catch (error) {
    console.error('Failed to get lunar date:', error)
  }
}

function handleMouseLeave() {
  hoverVisible.value = false
}

// 右键菜单
function handleContextMenu(event: MouseEvent, day: CalendarDay) {
  contextMenuDay.value = day
  contextMenuPosition.value = { x: event.clientX, y: event.clientY }
  contextMenuVisible.value = true
}

// 点击日期
function handleDayClick(day: CalendarDay) {
  if (day.todoCount > 0) {
    selectedDay.value = day
    dayDetailVisible.value = true
  } else {
    defaultDate.value = dayjs(day.date).format('YYYY-MM-DD')
    editingTodo.value = null
    todoDialogVisible.value = true
  }
}

function handleCreateTodo() {
  if (contextMenuDay.value) {
    defaultDate.value = dayjs(contextMenuDay.value.date).format('YYYY-MM-DD')
    editingTodo.value = null
    todoDialogVisible.value = true
  }
  contextMenuVisible.value = false
}

function handleViewDay() {
  if (contextMenuDay.value) {
    selectedDay.value = contextMenuDay.value
    dayDetailVisible.value = true
  }
  contextMenuVisible.value = false
}

function handleEditTodo(todo: Todo) {
  editingTodo.value = todo
  todoDialogVisible.value = true
}

function handleCreateFromDay(date: string) {
  // 设置默认日期并打开创建弹窗
  defaultDate.value = date
  editingTodo.value = null
  todoDialogVisible.value = true
}

async function handleTodoSaved() {
  await fetchCalendar()
  // 如果日期详情弹窗还开着，需要更新其中的数据
  if (dayDetailVisible.value && selectedDay.value) {
    const updatedDay = calendarDays.value.find(d => d.date === selectedDay.value?.date)
    if (updatedDay) {
      selectedDay.value = updatedDay
    }
  }
}

// 点击其他地方关闭菜单
function handleDocumentClick() {
  contextMenuVisible.value = false
}

onMounted(async () => {
  await fetchCalendar()
  fetchTodoTypes()
  document.addEventListener('click', handleDocumentClick)
})

onUnmounted(() => {
  document.removeEventListener('click', handleDocumentClick)
})
</script>

<style lang="scss" scoped>
.calendar-view {
  height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
}

.calendar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background: #fff;
  border-radius: 12px;
  margin-bottom: 15px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);

  .current-month {
    font-size: 22px;
    font-weight: 600;
    color: #303133;
  }

  .today-btn {
    margin-left: 15px;
  }
}

.calendar-body {
  flex: 1;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.week-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  background: #409eff;
  
  .week-day {
    text-align: center;
    padding: 12px;
    color: #fff;
    font-weight: 500;
  }
}

.calendar-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  // grid-template-rows 由内联样式动态设置
}

.calendar-cell {
  border: 1px solid #e8e8e8;
  padding: 8px;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #fff;

  &:hover {
    background: #f5f7fa;
  }

  &.is-today {
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    
    .solar-day {
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      color: #fff;
      border-radius: 50%;
      width: 28px;
      height: 28px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }

  &.is-other-month {
    background: #f0f0f0;
    
    .solar-day {
      color: #b0b0b0;
    }
    
    .lunar-day {
      color: #c0c0c0;
    }
    
    .todo-item {
      opacity: 0.5;
      color: #999;
    }
  }

  .cell-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 5px;

    .solar-day {
      font-size: 16px;
      font-weight: 500;
      color: #303133;
    }

    .lunar-day {
      font-size: 12px;
      color: #909399;
    }
  }

  .todo-indicators {
    flex: 1;
    overflow: hidden;
    min-width: 0;
  }

  .todo-indicator {
    display: flex;
    align-items: center;
    gap: 5px;
    padding: 2px 0;
    font-size: 12px;
    min-width: 0;

    .todo-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      flex-shrink: 0;
    }

    .todo-title {
      flex: 1;
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      color: #606266;

      &.completed {
        text-decoration: line-through;
        color: #c0c4cc;
      }
    }
  }

  &.is-completed {
    opacity: 0.6;
  }

  .more-todos {
    font-size: 11px;
    color: #909399;
    padding: 2px 0;
  }
}

.day-hover-tooltip {
  position: fixed;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  padding: 12px 15px;
  z-index: 1000;
  min-width: 200px;

  .tooltip-header {
    margin-bottom: 8px;
    
    .tooltip-date {
      font-size: 14px;
      font-weight: 600;
      color: #303133;
    }
    
    .tooltip-week {
      font-size: 12px;
      color: #909399;
    }
  }

  .tooltip-lunar {
    font-size: 13px;
    color: #606266;
    margin-bottom: 8px;
    
    span {
      display: block;
    }
  }

  .tooltip-festivals {
    display: flex;
    flex-wrap: wrap;
    gap: 5px;
  }
}

.context-menu {
  position: fixed;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  overflow: hidden;

  .menu-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 20px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }
  }
}
</style>
