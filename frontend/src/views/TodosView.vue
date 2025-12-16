<template>
  <div class="todos-view">
    <div class="page-header">
      <h2>待办事项</h2>
      <div class="header-actions">
        <el-button @click="handleRefresh" :loading="todoStore.loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button type="primary" @click="handleCreate">
          <el-icon><Plus /></el-icon>
          新建待办
        </el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar card">
      <el-input
        v-model="filter.keyword"
        placeholder="搜索标题..."
        style="width: 200px"
        clearable
        @change="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      
      <el-select 
        v-model="filter.year" 
        placeholder="年份" 
        clearable 
        filterable
        allow-create
        default-first-option
        @change="handleSearch" 
        style="width: 120px"
      >
        <el-option v-for="y in yearOptions" :key="y" :label="`${y}年`" :value="y" />
      </el-select>
      
      <el-select 
        ref="monthSelectRef"
        v-model="filter.month" 
        placeholder="月份" 
        clearable 
        @change="handleSearch" 
        @visible-change="handleMonthDropdownVisible"
        style="width: 100px"
      >
        <el-option v-for="m in 12" :key="m" :label="`${m}月`" :value="m" />
      </el-select>
      
      <el-select v-model="filter.types" placeholder="类型" multiple clearable @change="handleSearch" style="width: 180px">
        <el-option v-for="t in todoTypes" :key="t.value" :label="t.label" :value="t.value">
          <span>{{ t.icon }} {{ t.label }}</span>
        </el-option>
      </el-select>
    </div>

    <!-- 待办列表 -->
    <div class="todo-list card">
      <el-table :data="todoStore.todos" v-loading="todoStore.loading" style="width: 100%">
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row)" size="small">
              {{ getStatusText(row) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="标题" min-width="250">
          <template #default="{ row }">
            <div class="todo-title-cell">
              <div class="title-row">
                <el-tag size="small" :style="{ background: getTodoTypeColor(row.type), color: '#fff', border: 'none' }">
                  {{ getTodoTypeLabel(row.type) }}
                </el-tag>
                <span class="title">{{ row.title }}</span>
              </div>
              <div class="time-range">{{ formatTimeRange(row.startDate, row.endDate) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="循环" width="80" align="center">
          <template #default="{ row }">
            <span class="repeat-info">{{ row.repeatIndex || 1 }}/{{ row.repeatTotal || 1 }}</span>
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="180" align="center">
          <template #default="{ row }">
            <el-button size="small" type="success" @click="handleComplete(row)">
              <el-icon><Check /></el-icon>
            </el-button>
            <el-button size="small" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
            </el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!todoStore.loading && todoStore.todos.length === 0" class="empty-state">
        <el-empty description="暂无待办事项" />
      </div>
      
      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="filter.page"
          v-model:page-size="filter.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="todoStore.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <!-- 待办表单弹窗 -->
    <TodoFormDialog
      v-model:visible="dialogVisible"
      :todo="editingTodo"
      @saved="handleSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Plus, Search, Edit, Delete, Check, Refresh } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import TodoFormDialog from '@/components/TodoFormDialog.vue'
import { useTodoStore } from '@/stores/todo'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type Todo = models.Todo
type TodoType = { value: string; label: string; icon: string; color: string }

const todoStore = useTodoStore()
const todoTypes = ref<TodoType[]>([])
const monthSelectRef = ref()

const filter = reactive({
  keyword: '',
  year: null as number | null,
  month: null as number | null,
  types: [] as string[],
  page: 1,
  pageSize: 10
})

const dialogVisible = ref(false)
const editingTodo = ref<Todo | null>(null)

const yearOptions = computed(() => {
  const years = []
  const currentYear = dayjs().year()
  // 前7年到后4年，共12年
  for (let i = currentYear - 7; i <= currentYear + 4; i++) {
    years.push(i)
  }
  return years
})

async function fetchTodoTypes() {
  try {
    todoTypes.value = await api.GetTodoTypes()
  } catch (error) {
    console.error('Failed to fetch todo types:', error)
  }
}

function getTodoTypeLabel(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.label || type
}

function getTodoTypeIcon(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.icon || '📋'
}

function getTodoTypeColor(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.color || '#999'
}

function formatDate(date: string): string {
  return dayjs(date).format('YYYY-MM-DD')
}

function formatDateTime(date: string): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

function formatDateTimeRange(startDate: string, endDate: string): string {
  const start = dayjs(startDate)
  const end = dayjs(endDate)
  return `${start.format('YYYY年MM月DD日 HH:mm')} - ${end.format('YYYY年MM月DD日 HH:mm')}`
}

function isOverdue(todo: Todo): boolean {
  return !todo.isCompleted && dayjs(todo.endDate).isBefore(dayjs())
}

// 格式化计划执行时间
function formatScheduledTime(time: string): string {
  if (!time) return ''
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

// 格式化时间范围（开始时间 - 结束时间）
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

// 获取状态文字：未开始 / 进行中 / 即将开始 / 已超时
function getStatusText(row: any): string {
  const now = dayjs()
  const startDate = dayjs(row.startDate)
  const endDate = dayjs(row.endDate)
  
  // 已过结束时间 - 超时
  if (endDate.isBefore(now)) {
    return '已超时'
  }
  // 已过开始时间，未到结束时间 - 进行中
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return '进行中'
  }
  // 开始时间在24小时内 - 即将开始
  if (startDate.diff(now, 'hour') < 24) {
    return '即将开始'
  }
  return '未开始'
}

// 获取状态标签类型
function getStatusType(row: any): 'info' | 'warning' | 'danger' | 'success' | '' {
  const now = dayjs()
  const startDate = dayjs(row.startDate)
  const endDate = dayjs(row.endDate)
  
  // 已过结束时间 - 红色
  if (endDate.isBefore(now)) {
    return 'danger'
  }
  // 进行中 - 绿色
  if (startDate.isBefore(now) && endDate.isAfter(now)) {
    return 'success'
  }
  // 即将开始 - 蓝色（primary）
  if (startDate.diff(now, 'hour') < 24) {
    return ''
  }
  return 'info'
}

function handleSearch() {
  filter.page = 1  // 重置到第一页
  fetchTodos()
}

function handleRefresh() {
  fetchTodos()
}

function handlePageChange(page: number) {
  filter.page = page
  fetchTodos()
}

function fetchTodos() {
  todoStore.fetchTodos({
    keyword: filter.keyword,
    year: filter.year || 0,
    month: filter.month || 0,
    types: filter.types,
    page: filter.page,
    pageSize: filter.pageSize
  })
}

function handleCreate() {
  editingTodo.value = null
  dialogVisible.value = true
}

function handleEdit(todo: Todo) {
  editingTodo.value = { ...todo }
  dialogVisible.value = true
}

async function handleDelete(todo: Todo) {
  try {
    await ElMessageBox.confirm(`确定要删除待办"${todo.title}"吗？`, '确认删除', {
      type: 'warning'
    })
    await todoStore.deleteTodo(todo.id)
    ElMessage.success('删除成功')
    fetchTodos() // 刷新列表
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

async function handleComplete(row: Todo) {
  try {
    await ElMessageBox.confirm(`确定要将"${row.title}"标记为已完成吗？`, '确认完成', {
      type: 'info'
    })
    await todoStore.markTodoCompleted(row.id, true)
    ElMessage.success('已标记完成，待办已移至历史记录')
    fetchTodos() // 刷新列表
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

function handleSaved() {
  fetchTodos()
}

// 月份下拉框展开时滚动到当前月份
function handleMonthDropdownVisible(visible: boolean) {
  if (visible) {
    nextTick(() => {
      const currentMonth = dayjs().month() + 1
      const dropdown = document.querySelector('.el-select-dropdown.is-multiple, .el-select-dropdown:not(.is-multiple)')
      if (dropdown) {
        const options = dropdown.querySelectorAll('.el-select-dropdown__item')
        if (options[currentMonth - 1]) {
          options[currentMonth - 1].scrollIntoView({ block: 'center' })
        }
      }
    })
  }
}

onMounted(() => {
  fetchTodoTypes()
  fetchTodos()
})
</script>

<style lang="scss" scoped>
.todos-view {
  height: 100%;
}

.filter-bar {
  display: flex;
  gap: 15px;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
}

.todo-list {
  .todo-title-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;

    .title-row {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .title {
        font-size: 15px;
        color: #303133;
      }
    }
    
    .time-range {
      font-size: 12px;
      color: #909399;
    }
  }
  
  .repeat-info {
    font-size: 13px;
    color: #606266;
  }
  
  .is-overdue {
    color: #F56C6C;
  }

  .no-repeat {
    color: #c0c4cc;
  }
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  padding: 20px 0 0;
}

.header-actions {
  display: flex;
  gap: 10px;
}
</style>
