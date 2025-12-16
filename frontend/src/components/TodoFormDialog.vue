<template>
  <el-dialog
    :model-value="visible"
    @update:model-value="$emit('update:visible', $event)"
    :title="isEdit ? '编辑待办' : '新建待办'"
    width="700px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="标题" prop="title">
        <el-input v-model="form.title" placeholder="请输入待办标题" maxlength="100" show-word-limit />
      </el-form-item>

      <el-form-item label="类型" prop="type">
        <el-radio-group v-model="form.type">
          <el-radio-button
            v-for="t in todoTypes"
            :key="t.value"
            :value="t.value"
          >
            {{ t.icon }} {{ t.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- 生日特殊选项 -->
      <template v-if="form.type === 'birthday'">
        <el-form-item label="日历类型">
          <el-radio-group v-model="form.isLunar">
            <el-radio :value="false">公历</el-radio>
            <el-radio :value="true">农历</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="隐藏年份">
          <el-switch v-model="form.hideYear" />
          <span class="form-hint">开启后只选择月和日</span>
        </el-form-item>

        <el-form-item label="日期" prop="startDate">
          <template v-if="form.hideYear">
            <el-date-picker
              v-model="form.startDate"
              type="date"
              format="MM-DD"
              value-format="YYYY-MM-DD"
              placeholder="选择月日"
            />
          </template>
          <template v-else>
            <el-date-picker
              v-model="form.startDate"
              type="date"
              format="YYYY-MM-DD"
              value-format="YYYY-MM-DD"
              placeholder="选择日期"
            />
          </template>
          <div v-if="form.isLunar && form.startDate" class="lunar-hint">
            农历: {{ lunarDateText }}
          </div>
        </el-form-item>
      </template>

      <!-- 非生日类型 -->
      <template v-else>
        <el-form-item label="开始时间" prop="startDate">
          <el-date-picker
            v-model="form.startDate"
            type="datetime"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss"
            placeholder="选择开始时间"
          />
        </el-form-item>

        <!-- Cron表达式 -->
        <el-form-item label="重复规则">
          <div class="cron-section">
            <el-input
              v-model="form.cronExpr"
              placeholder="Cron表达式，如: 0 9 * * 1 (每周一9点)"
              @change="onCronExprChange"
            >
              <template #prepend>
                <el-select v-model="cronPreset" style="width: 120px" @change="handleCronPreset">
                  <el-option label="不重复" value="none" />
                  <el-option label="每天" value="0 9 * * *" />
                  <el-option label="每周" value="0 9 * * 1" />
                  <el-option label="每月" value="0 9 1 * *" />
                  <el-option label="每年" value="0 9 1 1 *" />
                  <el-option label="工作日" value="0 9 * * 1-5" />
                </el-select>
              </template>
            </el-input>
            
            <div v-if="cronNextRuns.isValid && cronNextRuns.nextRuns.length" class="cron-preview">
              <div class="preview-title">接下来5次执行时间:</div>
              <div v-for="(run, index) in cronNextRuns.nextRuns" :key="index" class="preview-item">
                {{ formatDateTime(run) }}
              </div>
            </div>
            <div v-else-if="form.cronExpr && !cronNextRuns.isValid" class="cron-error">
              {{ cronNextRuns.error || '无效的Cron表达式' }}
            </div>
          </div>
        </el-form-item>

        <el-form-item label="结束时间" prop="endDate">
          <el-date-picker
            v-model="form.endDate"
            type="datetime"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss"
            placeholder="选择结束时间"
          />
        </el-form-item>

        <el-form-item v-if="form.cronExpr" label="循环终止时间">
          <el-date-picker
            v-model="form.repeatEndDate"
            type="datetime"
            format="YYYY-MM-DD HH:mm"
            value-format="YYYY-MM-DDTHH:mm:ss"
            placeholder="选择循环终止时间"
          />
          <span v-if="repeatCountPreview > 0" class="form-hint">将创建 {{ repeatCountPreview }} 条待办记录</span>
        </el-form-item>

        <el-form-item label="提醒设置">
          <div class="remind-settings">
            <el-checkbox v-model="form.remindAtStart">到点提醒</el-checkbox>
            <el-checkbox v-model="form.remindAtEnd">结束提醒</el-checkbox>
          </div>
        </el-form-item>

        <el-form-item label="提前提醒">
          <el-input-number 
            v-model="form.advanceRemind" 
            :min="0" 
            :max="1440" 
          />
          <span class="remind-unit">分钟</span>
          <span class="form-hint">0表示不提前提醒</span>
        </el-form-item>
      </template>

      <el-form-item label="内容">
        <div class="md-editor-wrapper" @paste="handlePaste">
          <MdEditor
            v-model="form.content"
            :preview="true"
            previewTheme="github"
            :toolbars="mdToolbars"
            style="height: 350px"
            @onUploadImg="handleUploadImg"
            :sanitize="sanitizeContent"
          />
          <div v-if="pastedImages.length > 0" class="pasted-images-hint">
            已粘贴 {{ pastedImages.length }} 张截图，将作为附件保存
          </div>
        </div>
      </el-form-item>

      <el-form-item label="附件">
        <el-upload
          class="attachment-upload"
          :auto-upload="false"
          :file-list="fileList"
          :on-change="handleFileChange"
          :on-remove="handleFileRemove"
          multiple
        >
          <el-button size="small" type="primary">
            <el-icon><Upload /></el-icon>
            添加附件
          </el-button>
          <template #tip>
            <div class="el-upload__tip">支持图片和文件，附件将加密存储</div>
          </template>
        </el-upload>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="$emit('update:visible', false)">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitting">
        {{ isEdit ? '保存' : '创建' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Upload } from '@element-plus/icons-vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import dayjs from 'dayjs'
import * as api from '@/wailsjs/go/app/App'
import { models } from '@/wailsjs/go/models'

type Todo = models.Todo
type CronNextRun = models.CronNextRun
type TodoType = { value: string; label: string; icon: string; color: string }

const props = defineProps<{
  visible: boolean
  todo?: Todo | null
  defaultDate?: string
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
  saved: []
}>()

const formRef = ref()
const todoTypes = ref<TodoType[]>([])
const submitting = ref(false)
const cronPreset = ref('none')
const cronNextRuns = ref<CronNextRun>({ expression: '', nextRuns: [], isValid: false })
const fileList = ref<any[]>([])
const lunarDateText = ref('')
const pastedImages = ref<File[]>([])
const repeatCountPreview = ref(0)

const mdToolbars = [
  'bold', 'italic', 'strikeThrough', '-',
  'title', 'quote', 'unorderedList', 'orderedList', '-',
  'link', 'image', 'table', 'code', '-',
  'revoke', 'next', '-',
  'preview'
]

// sanitize 函数：将 attachment:文件名 替换为实际的图片 URL
function sanitizeContent(html: string): string {
  // 替换 attachment:文件名 为实际的预览 URL
  let result = html
  for (const [fileName, dataUrl] of imagePreviewUrls.value.entries()) {
    // 替换 src="attachment:文件名" 格式
    result = result.replace(
      new RegExp(`src="attachment:${fileName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}"`, 'g'),
      `src="${dataUrl}"`
    )
  }
  return result
}

const form = reactive({
  id: 0,
  title: '',
  content: '',
  type: 'task',
  startDate: '',
  endDate: '',
  isLunar: false,
  hideYear: false,
  cronExpr: '',
  repeatType: 'none',  // 循环类型
  repeatEndDate: '',   // 循环终止时间
  advanceRemind: 15,
  remindAtStart: true,
  remindAtEnd: true
})

const isEdit = computed(() => props.todo && props.todo.id > 0)

const rules = {
  title: [{ required: true, message: '请输入标题', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  startDate: [{ required: true, message: '请选择日期', trigger: 'change' }],
  endDate: [{ required: true, message: '请选择结束时间', trigger: 'change' }]
}

// 监听visible变化，初始化表单
watch(() => props.visible, async (val) => {
  if (val) {
    // 清理旧的预览URL
    for (const url of imagePreviewUrls.value.values()) {
      if (url.startsWith('blob:')) {
        URL.revokeObjectURL(url)
      }
    }
    imagePreviewUrls.value.clear()
    fileList.value = []
    pastedImages.value = []

    await fetchTodoTypes()
    if (props.todo) {
      // 编辑模式：加载附件列表和预览URL
      try {
        const attachments = await api.GetTodoAttachments(props.todo.id)
        for (const attachment of attachments) {
          // 获取附件数据用于预览
          const dataUrl = await api.GetAttachmentAsDataURL(props.todo.id, attachment.fileName)
          imagePreviewUrls.value.set(attachment.fileName, dataUrl)
          
          // 将已有附件添加到列表显示（不带 raw，表示已上传）
          fileList.value.push({
            name: attachment.fileName,
            url: dataUrl,  // 用于显示缩略图
            status: 'success',
            uid: attachment.id
          })
        }
      } catch (error) {
        console.error('Failed to load attachments:', error)
      }

      Object.assign(form, {
        id: props.todo.id,
        title: props.todo.title,
        content: props.todo.content,  // 保持原始内容（attachment:文件名格式）
        type: props.todo.type,
        startDate: props.todo.startDate,
        endDate: props.todo.endDate,
        isLunar: props.todo.isLunar,
        hideYear: props.todo.hideYear,
        cronExpr: '',  // 编辑时不显示循环设置
        repeatType: 'none',
        repeatEndDate: '',
        advanceRemind: props.todo.advanceRemind ?? 15,
        remindAtStart: props.todo.remindAtStart ?? true,
        remindAtEnd: props.todo.remindAtEnd ?? false
      })
      cronPreset.value = 'none'
    } else {
      resetForm()
      if (props.defaultDate) {
        const today = dayjs().format('YYYY-MM-DD')
        if (props.defaultDate === today) {
          // 今天：使用下一个整点或半点作为开始时间
          const now = dayjs()
          const minutes = now.minute()
          let startTime: dayjs.Dayjs
          
          if (minutes === 0) {
            // 正好是整点，使用当前时间
            startTime = now.second(0).millisecond(0)
          } else if (minutes <= 30) {
            // 1-30分，进位到半点
            startTime = now.minute(30).second(0).millisecond(0)
          } else {
            // 31-59分，进位到下一个整点
            startTime = now.add(1, 'hour').minute(0).second(0).millisecond(0)
          }
          
          const endTime = startTime.add(1, 'hour')
          
          form.startDate = startTime.format('YYYY-MM-DDTHH:mm:ss')
          form.endDate = endTime.format('YYYY-MM-DDTHH:mm:ss')
        } else {
          // 其他日期：使用 09:00-10:00
          form.startDate = props.defaultDate + 'T09:00:00'
          form.endDate = props.defaultDate + 'T10:00:00'
        }
      }
    }
  }
})

// 监听日期变化，更新农历显示
watch(() => form.startDate, async () => {
  if (form.isLunar && form.startDate) {
    const date = dayjs(form.startDate)
    try {
      const lunar = await api.GetLunarDate(date.year(), date.month() + 1, date.date())
      lunarDateText.value = `${lunar.monthName}${lunar.dayName}`
    } catch (error) {
      lunarDateText.value = ''
    }
  }
})

async function fetchTodoTypes() {
  try {
    todoTypes.value = await api.GetTodoTypes()
  } catch (error) {
    console.error('Failed to fetch todo types:', error)
  }
}

function resetForm() {
  form.id = 0
  form.title = ''
  form.content = ''
  form.type = 'task'
  form.startDate = ''
  form.endDate = ''
  form.isLunar = false
  form.hideYear = false
  form.cronExpr = ''
  form.repeatType = 'none'
  form.repeatEndDate = ''
  form.advanceRemind = 15
  form.remindAtStart = true
  form.remindAtEnd = true
  cronPreset.value = 'none'
  cronNextRuns.value = { expression: '', nextRuns: [], isValid: false }
  repeatCountPreview.value = 0
  fileList.value = []
  pastedImages.value = []
  // 清理预览URL
  imagePreviewUrls.value.forEach(url => URL.revokeObjectURL(url))
  imagePreviewUrls.value.clear()
}

function handleCronPreset(value: string) {
  // "none" 表示不重复，实际 cronExpr 应为空
  form.cronExpr = value === 'none' ? '' : value
  form.repeatType = value === 'none' ? 'none' : 'custom'
  onCronExprChange()
}

async function onCronExprChange() {
  await parseCronExpr()
  // 更新循环次数预览
  await updateRepeatCountPreview()
}

async function updateRepeatCountPreview() {
  if (!form.startDate || !form.cronExpr || !form.repeatEndDate) {
    repeatCountPreview.value = 0
    return
  }
  
  try {
    const count = await api.CalculateRemindCount(form.startDate, form.cronExpr, form.repeatEndDate)
    repeatCountPreview.value = count > 0 ? count : 0
  } catch (error) {
    console.error('Failed to calculate repeat count:', error)
    repeatCountPreview.value = 0
  }
}

// 监听循环终止时间变化，更新预览
watch(() => form.repeatEndDate, () => {
  updateRepeatCountPreview()
})

watch(() => form.startDate, () => {
  if (form.cronExpr) {
    updateRepeatCountPreview()
  }
})

async function parseCronExpr() {
  if (!form.cronExpr) {
    cronNextRuns.value = { expression: '', nextRuns: [], isValid: false }
    return
  }
  try {
    cronNextRuns.value = await api.ParseCronExpression(form.cronExpr)
  } catch (error) {
    cronNextRuns.value = { expression: form.cronExpr, nextRuns: [], isValid: false, error: '解析失败' }
  }
}

function formatDateTime(date: string): string {
  return dayjs(date).format('YYYY-MM-DD HH:mm')
}

function handleFileChange(file: any, list: any[]) {
  fileList.value = list
}

async function handleFileRemove(file: any, list: any[]) {
  // 如果是已保存的附件（没有 raw 属性），需要从服务器删除
  if (!file.raw && file.uid && typeof file.uid === 'number') {
    try {
      await api.DeleteAttachment(file.uid)
      ElMessage.success('附件已删除')
    } catch (error) {
      console.error('Failed to delete attachment:', error)
      ElMessage.error('删除附件失败')
      return // 删除失败，不更新列表
    }
  }
  
  // 从预览URL映射中移除
  if (file.name) {
    const url = imagePreviewUrls.value.get(file.name)
    if (url && url.startsWith('blob:')) {
      URL.revokeObjectURL(url)
    }
    imagePreviewUrls.value.delete(file.name)
    
    // 从内容中移除对应的图片引用
    const attachmentPattern = `![${file.name}](attachment:${file.name})`
    form.content = form.content.replace(attachmentPattern, '')
    // 也处理可能存在的 blob URL 格式
    if (url) {
      form.content = form.content.replace(`![${file.name}](${url})`, '')
    }
  }
  
  fileList.value = list
}

// 存储图片的预览URL映射
const imagePreviewUrls = ref<Map<string, string>>(new Map())

// 处理粘贴事件（截图）
function handlePaste(event: ClipboardEvent) {
  const items = event.clipboardData?.items
  if (!items) return

  for (let i = 0; i < items.length; i++) {
    const item = items[i]
    if (item.type.startsWith('image/')) {
      const file = item.getAsFile()
      if (file) {
        // 生成文件名
        const timestamp = dayjs().format('YYYYMMDDHHmmss')
        const ext = file.type.split('/')[1] || 'png'
        const fileName = `screenshot_${timestamp}.${ext}`
        
        // 创建带有自定义名称的文件对象
        const namedFile = new File([file], fileName, { type: file.type })
        
        // 添加到粘贴图片列表
        pastedImages.value.push(namedFile)
        
        // 添加到附件列表
        fileList.value.push({
          name: fileName,
          raw: namedFile,
          status: 'ready',
          uid: Date.now() + i
        })
        
        // 创建预览URL
        const previewUrl = URL.createObjectURL(namedFile)
        imagePreviewUrls.value.set(fileName, previewUrl)
        
        // 在内容中插入图片，使用预览URL
        form.content += `\n![${fileName}](${previewUrl})\n`
        
        ElMessage.success(`已粘贴截图: ${fileName}`)
      }
    }
  }
}

// 处理编辑器内的图片上传
async function handleUploadImg(files: File[], callback: (urls: string[]) => void) {
  const urls: string[] = []
  
  for (const file of files) {
    // 添加到附件列表
    fileList.value.push({
      name: file.name,
      raw: file,
      status: 'ready',
      uid: Date.now()
    })
    
    // 创建预览URL
    const previewUrl = URL.createObjectURL(file)
    imagePreviewUrls.value.set(file.name, previewUrl)
    
    urls.push(previewUrl)
    ElMessage.success(`已添加图片: ${file.name}`)
  }
  
  callback(urls)
}

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  } catch (error) {
    return
  }

  submitting.value = true
  try {
    // 处理生日类型的结束日期
    let endDate = form.endDate
    if (form.type === 'birthday') {
      // 生日类型结束日期与开始日期相同
      endDate = form.startDate
    }

    // 将内容中的 blob: URL 替换为 attachment:文件名 格式（仅处理新粘贴的图片）
    let content = form.content
    for (const [fileName, url] of imagePreviewUrls.value.entries()) {
      // 只替换 blob: URL（新粘贴的图片），不替换 data: URL（已存在的附件）
      if (url.startsWith('blob:')) {
        content = content.replace(url, `attachment:${fileName}`)
      }
    }

    const todoData = {
      id: form.id,
      title: form.title,
      content: content,
      type: form.type,
      startDate: form.startDate,
      endDate: endDate,
      isLunar: form.isLunar,
      hideYear: form.hideYear,
      advanceRemind: form.advanceRemind,
      remindAtStart: form.remindAtStart,
      remindAtEnd: form.remindAtEnd,
      // 循环设置（仅新建时有效）
      repeatType: form.cronExpr ? 'custom' : 'none',
      cronExpr: form.cronExpr,
      repeatEndDate: form.repeatEndDate || null
    }

    let todoId: number
    if (isEdit.value) {
      await api.UpdateTodo(todoData as any)
      todoId = form.id
      ElMessage.success('更新成功')
    } else {
      todoId = await api.CreateTodo(todoData as any)
      if (form.cronExpr && form.repeatEndDate && repeatCountPreview.value > 1) {
        ElMessage.success(`已创建 ${repeatCountPreview.value} 条待办记录`)
      } else {
        ElMessage.success('创建成功')
      }
    }

    // 上传附件（等待所有上传完成）
    const uploadPromises: Promise<void>[] = []
    for (const file of fileList.value) {
      if (file.raw) {
        const promise = new Promise<void>((resolve) => {
          const reader = new FileReader()
          reader.onload = async (e) => {
            const base64 = (e.target?.result as string).split(',')[1]
            await api.UploadAttachment(todoId, file.name, base64, file.raw.type)
            resolve()
          }
          reader.readAsDataURL(file.raw)
        })
        uploadPromises.push(promise)
      }
    }
    await Promise.all(uploadPromises)

    emit('update:visible', false)
    emit('saved')
  } catch (error) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style lang="scss" scoped>
.form-hint {
  margin-left: 10px;
  font-size: 12px;
  color: #909399;
}

.lunar-hint {
  margin-top: 5px;
  font-size: 12px;
  color: #667eea;
}

.cron-section {
  width: 100%;

  .cron-preview {
    margin-top: 10px;
    padding: 10px;
    background: #f5f7fa;
    border-radius: 6px;

    .preview-title {
      font-size: 12px;
      color: #909399;
      margin-bottom: 5px;
    }

    .preview-item {
      font-size: 13px;
      color: #606266;
      padding: 2px 0;
    }
  }

  .cron-error {
    margin-top: 5px;
    font-size: 12px;
    color: #F56C6C;
  }
}

.md-editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  
  .pasted-images-hint {
    padding: 8px 12px;
    background: #f0f9ff;
    border-top: 1px solid #dcdfe6;
    font-size: 12px;
    color: #409eff;
  }
}

.attachment-upload {
  width: 100%;
}

.remind-settings {
  display: flex;
  align-items: center;
  
  .el-checkbox {
    margin-right: 24px;
  }
}

.remind-unit {
  margin-left: 8px;
  margin-right: 10px;
  font-size: 14px;
  color: #606266;
}
</style>
