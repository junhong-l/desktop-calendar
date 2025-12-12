<template>
  <div class="history-view">
    <div class="page-header">
      <h2>历史记录</h2>
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

    <!-- 历史列表 -->
    <div class="history-list card">
      <el-table 
        :data="todos" 
        v-loading="loading" 
        style="width: 100%"
      >
        <el-table-column label="状态" width="80" align="center">
          <template #default="{ row }">
            <el-tag type="success" size="small">已完成</el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="标题" min-width="280">
          <template #default="{ row }">
            <div class="todo-title-cell">
              <div class="title-row">
                <el-tag size="small" :style="{ background: getTodoTypeColor(row.type), color: '#fff', border: 'none' }">
                  {{ getTodoTypeLabel(row.type) }}
                </el-tag>
                <span class="title completed">{{ row.title }}</span>
              </div>
              <div class="time-range">{{ formatDateTimeRange(row.startDate, row.endDate) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="完成时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.completedAt) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="180" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
                编辑
              </el-button>
              <el-button size="small" @click="handleRestore(row)">
                <el-icon><RefreshLeft /></el-icon>
                恢复
              </el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="filter.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSearch"
          @current-change="handleSearch"
        />
      </div>
    </div>

    <!-- 历史详情弹窗 -->
    <HistoryDetailDialog
      v-model:visible="detailVisible"
      :todo="selectedTodo"
      @saved="handleDetailSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, nextTick } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search, Delete, RefreshLeft, Edit } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'
import HistoryDetailDialog from '@/components/HistoryDetailDialog.vue'

type Todo = models.Todo
type TodoType = { value: string; label: string; icon: string; color: string }

const todos = ref<Todo[]>([])
const todoTypes = ref<TodoType[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)

const detailVisible = ref(false)
const selectedTodo = ref<Todo | null>(null)

const filter = reactive({
  keyword: '',
  year: null as number | null,
  month: null as number | null,
  types: [] as string[],
  pageSize: 10
})

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

async function fetchHistory() {
  loading.value = true
  try {
    const result = await api.GetTodoList({
      keyword: filter.keyword || undefined,
      year: filter.year || undefined,
      month: filter.month || undefined,
      types: filter.types.length > 0 ? filter.types : undefined,
      completed: true,
      page: page.value,
      pageSize: filter.pageSize
    })
    todos.value = result.todos || []
    total.value = result.total
  } catch (error) {
    console.error('Failed to fetch history:', error)
  } finally {
    loading.value = false
  }
}

function getTodoTypeLabel(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.label || type
}

function getTodoTypeColor(type: string): string {
  return todoTypes.value.find(t => t.value === type)?.color || '#999'
}

function formatDateTime(date: string): string {
  if (!date) return '-'
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

function formatDateTimeRange(startDate: string, endDate: string): string {
  const start = dayjs(startDate)
  const end = dayjs(endDate)
  return `${start.format('YYYY年MM月DD日 HH:mm')} - ${end.format('YYYY年MM月DD日 HH:mm')}`
}

function handleEdit(row: Todo) {
  selectedTodo.value = { ...row }
  detailVisible.value = true
}

function handleDetailSaved() {
  fetchHistory()
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

function handleSearch() {
  // 确保年份是数字（处理 allow-create 产生的字符串）
  if (filter.year && typeof filter.year === 'string') {
    const yearNum = parseInt(filter.year.replace(/\D/g, ''), 10)
    filter.year = isNaN(yearNum) ? null : yearNum
  }
  page.value = 1
  fetchHistory()
}

async function handleRestore(todo: Todo) {
  try {
    await ElMessageBox.confirm(`确定要恢复待办"${todo.title}"吗？`, '确认恢复')
    await api.MarkTodoCompleted(todo.id, false)
    ElMessage.success('已恢复')
    fetchHistory()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
    }
  }
}

async function handleDelete(todo: Todo) {
  try {
    await ElMessageBox.confirm(`确定要永久删除"${todo.title}"吗？`, '确认删除', {
      type: 'warning'
    })
    await api.DeleteTodo(todo.id)
    ElMessage.success('删除成功')
    fetchHistory()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

onMounted(() => {
  fetchTodoTypes()
  fetchHistory()
})
</script>

<style lang="scss" scoped>
.history-view {
  height: 100%;
}

.filter-bar {
  display: flex;
  gap: 15px;
  align-items: center;
  margin-bottom: 15px;
}

.history-list {
  .todo-title-cell {
    display: flex;
    flex-direction: column;
    gap: 4px;
    
    .title-row {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .title {
        font-size: 14px;
        color: #303133;
        
        &.completed {
          text-decoration: line-through;
          color: #909399;
        }
      }
    }
    
    .time-range {
      font-size: 12px;
      color: #909399;
    }
  }
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
  padding: 20px 0 0;
}
</style>
