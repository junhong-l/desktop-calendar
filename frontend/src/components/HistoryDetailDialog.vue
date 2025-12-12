<template>
  <el-dialog
    v-model="dialogVisible"
    title="历史详情"
    width="750px"
    :close-on-click-modal="false"
  >
    <div class="history-detail" v-if="todo">
      <!-- 基本信息（只读） -->
      <el-form label-width="80px" label-position="left">
        <el-form-item label="类型">
          <el-tag :style="{ background: getTodoTypeColor(todo.type), color: '#fff', border: 'none' }">
            {{ getTodoTypeLabel(todo.type) }}
          </el-tag>
        </el-form-item>
        
        <el-form-item label="标题">
          <span class="readonly-text">{{ todo.title }}</span>
        </el-form-item>
        
        <el-form-item label="开始时间">
          <span class="readonly-text">{{ formatDateTime(todo.startDate) }}</span>
        </el-form-item>
        
        <el-form-item label="结束时间">
          <span class="readonly-text">{{ formatDateTime(todo.endDate) }}</span>
        </el-form-item>
        
        <el-form-item label="完成时间">
          <span class="readonly-text completed-time">{{ formatDateTime(todo.completedAt) }}</span>
        </el-form-item>
        
        <el-form-item label="循环">
          <span class="readonly-text">{{ todo.currentRepeat || 1 }}/{{ todo.repeatCount || 1 }}</span>
        </el-form-item>
        
        <!-- 内容（Markdown 编辑器） -->
        <el-form-item label="内容">
          <div class="md-editor-wrapper">
            <MdEditor
              v-model="editableContent"
              :preview="true"
              previewTheme="github"
              :toolbars="mdToolbars"
              style="height: 300px"
            />
          </div>
        </el-form-item>

        <!-- 附件列表 -->
        <el-form-item label="附件" v-if="attachments.length > 0">
          <div class="attachment-list">
            <div 
              v-for="attachment in attachments" 
              :key="attachment.id" 
              class="attachment-item"
            >
              <el-icon v-if="isImageFile(attachment.fileName)"><Picture /></el-icon>
              <el-icon v-else><Document /></el-icon>
              <span class="file-name">{{ attachment.fileName }}</span>
              <span class="file-size">({{ formatFileSize(attachment.fileSize) }})</span>
              <el-button 
                size="small" 
                text 
                type="primary" 
                @click="handleDownload(attachment)"
              >
                下载
              </el-button>
              <el-button 
                v-if="isImageFile(attachment.fileName)"
                size="small" 
                text 
                type="primary" 
                @click="handlePreview(attachment)"
              >
                预览
              </el-button>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="附件" v-else>
          <span class="readonly-text no-attachment">无附件</span>
        </el-form-item>
      </el-form>
    </div>

    <!-- 图片预览弹窗 -->
    <el-dialog
      v-model="previewVisible"
      title="图片预览"
      width="auto"
      :append-to-body="true"
    >
      <img :src="previewImageUrl" alt="预览图片" class="preview-image" />
    </el-dialog>
    
    <template #footer>
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture, Document } from '@element-plus/icons-vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type Todo = models.Todo
type Attachment = models.Attachment
type TodoType = { value: string; label: string; icon: string; color: string }

const props = defineProps<{
  visible: boolean
  todo: Todo | null
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const todoTypes = ref<TodoType[]>([])
const editableContent = ref('')
const attachments = ref<Attachment[]>([])
const saving = ref(false)
const previewVisible = ref(false)
const previewImageUrl = ref('')

const mdToolbars = [
  'bold', 'italic', 'strikeThrough', '-',
  'title', 'quote', 'unorderedList', 'orderedList', '-',
  'link', 'image', 'table', 'code', '-',
  'revoke', 'next', '-',
  'preview'
]

const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

watch(() => props.visible, async (val) => {
  if (val && props.todo) {
    await fetchTodoTypes()
    editableContent.value = props.todo.content || ''
    await fetchAttachments()
  }
})

async function fetchTodoTypes() {
  try {
    todoTypes.value = await api.GetTodoTypes()
  } catch (error) {
    console.error('Failed to fetch todo types:', error)
  }
}

async function fetchAttachments() {
  if (!props.todo) return
  try {
    attachments.value = await api.GetTodoAttachments(props.todo.id)
  } catch (error) {
    console.error('Failed to fetch attachments:', error)
    attachments.value = []
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

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function isImageFile(fileName: string): boolean {
  const ext = fileName.split('.').pop()?.toLowerCase() || ''
  return ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg'].includes(ext)
}

async function handleDownload(attachment: Attachment) {
  try {
    const result = await api.DownloadAttachment(attachment.id)
    if (result) {
      ElMessage.success('附件已保存')
    }
  } catch (error) {
    console.error('Failed to download:', error)
    ElMessage.error('下载失败')
  }
}

async function handlePreview(attachment: Attachment) {
  try {
    const base64 = await api.GetAttachment(attachment.id)
    if (base64) {
      const ext = attachment.fileName.split('.').pop()?.toLowerCase() || 'png'
      const mimeType = ext === 'png' ? 'image/png' : 
                       ext === 'gif' ? 'image/gif' : 
                       ext === 'webp' ? 'image/webp' : 'image/jpeg'
      previewImageUrl.value = `data:${mimeType};base64,${base64}`
      previewVisible.value = true
    }
  } catch (error) {
    console.error('Failed to preview:', error)
    ElMessage.error('预览失败')
  }
}

function handleCancel() {
  dialogVisible.value = false
}

async function handleSave() {
  if (!props.todo) return
  
  saving.value = true
  try {
    await api.UpdateTodo({
      ...props.todo,
      content: editableContent.value
    })
    ElMessage.success('保存成功')
    emit('saved')
    dialogVisible.value = false
  } catch (error) {
    console.error('Failed to save:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style lang="scss" scoped>
.history-detail {
  max-height: 70vh;
  overflow-y: auto;

  .readonly-text {
    color: #606266;
    font-size: 14px;
  }
  
  .completed-time {
    color: #67c23a;
    font-weight: 500;
  }

  .no-attachment {
    color: #909399;
  }
}

.md-editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}

.attachment-list {
  width: 100%;
  
  .attachment-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: #f5f7fa;
    border-radius: 4px;
    margin-bottom: 8px;
    
    &:last-child {
      margin-bottom: 0;
    }
    
    .el-icon {
      color: #409eff;
      font-size: 18px;
    }
    
    .file-name {
      flex: 1;
      color: #303133;
      font-size: 14px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    
    .file-size {
      color: #909399;
      font-size: 12px;
    }
  }
}

.preview-image {
  max-width: 80vw;
  max-height: 70vh;
  object-fit: contain;
}
</style>
