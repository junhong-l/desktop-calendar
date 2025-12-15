<template>
  <div class="history-view">
    <div class="page-header">
      <h2>历史记录</h2>
      <el-button @click="handleRefresh" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
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
        :data="completedTodos" 
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
              <div class="time-range">{{ formatTimeRange(row.startDate, row.endDate) }}</div>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column label="完成时间" width="160">
          <template #default="{ row }">
            {{ formatDateTime(row.completedAt) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="120" align="center">
          <template #default="{ row }">
            <el-button-group>
              <el-button size="small" @click="handleRestore(row)">
                <el-icon><RefreshLeft /></el-icon>
                恢复
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed, nextTick } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search, RefreshLeft, Refresh } from '@element-plus/icons-vue'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type Todo = models.Todo
type TodoType = { value: string; label: string; icon: string; color: string }

const completedTodos = ref<Todo[]>([])
const todoTypes = ref<TodoType[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)

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
      keyword: filter.keyword || '',
      year: filter.year || 0,
      month: filter.month || 0,
      types: filter.types.length > 0 ? filter.types : [],
      completed: true,
      page: page.value,
      pageSize: filter.pageSize
    })
    completedTodos.value = result?.todos || []
    total.value = result?.total || 0
  } catch (error) {
    console.error('Failed to fetch history:', error)
    completedTodos.value = []
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

function formatScheduledTime(date: string): string {
  if (!date) return '-'
  return dayjs(date).format('YYYY年MM月DD日 HH:mm')
}

// 格式化时间范围（开始时间 - 结束时间）
function formatTimeRange(startDate: string, endDate: string): string {
  if (!startDate) return '-'
  const start = dayjs(startDate)
  const end = endDate ? dayjs(endDate) : null
  
  const startStr = start.format('YYYY年MM月DD日 HH:mm')
  if (end) {
    const endStr = end.format('YYYY年MM月DD日 HH:mm')
    return `${startStr} - ${endStr}`
  }
  return startStr
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

function handleRefresh() {
  fetchHistory()
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

async function handleRestore(row: Todo) {
  try {
    await ElMessageBox.confirm(`确定要恢复"${row.title}"吗？`, '确认恢复')
    await api.MarkTodoCompleted(row.id, false)
    ElMessage.success('已恢复')
    fetchHistory()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('操作失败')
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
