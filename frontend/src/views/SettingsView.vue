<template>
  <div class="settings-view">
    <div class="page-header">
      <h2>设置</h2>
    </div>

    <div class="settings-content card">
      <el-form :model="settings" label-width="150px" v-loading="loading">
        <el-divider content-position="left">系统设置</el-divider>
        
        <el-form-item label="开机自启动">
          <el-switch v-model="settings.enableAutoStart" />
          <span class="setting-hint">开机时自动启动待办日历</span>
        </el-form-item>

        <el-form-item label="最小化到托盘">
          <el-switch v-model="settings.minimizeToTray" />
          <span class="setting-hint">关闭窗口时最小化到系统托盘</span>
        </el-form-item>

        <el-divider content-position="left">桌面小部件</el-divider>
        
        <el-form-item label="小部件">
          <el-button 
            v-if="!widgetRunning"
            type="primary" 
            @click="openWidget" 
            :loading="openingWidget"
          >
            打开桌面小部件
          </el-button>
          <el-button 
            v-else
            type="danger" 
            @click="closeWidget" 
            :loading="closingWidget"
          >
            关闭桌面小部件
          </el-button>
          <span class="setting-hint">{{ widgetRunning ? '小部件正在运行' : '启动一个独立的小部件窗口' }}</span>
        </el-form-item>

        <el-divider content-position="left">通知设置</el-divider>
        
        <el-form-item label="通知声音">
          <el-switch v-model="settings.notificationSound" />
        </el-form-item>

        <el-form-item label="提示音" v-if="settings.notificationSound">
          <div class="sound-selector">
            <el-select 
              v-model="settings.notificationSoundFile" 
              placeholder="选择提示音"
              style="width: 220px"
            >
              <el-option-group label="默认">
                <el-option
                  v-for="sound in defaultSounds"
                  :key="sound.path || 'default'"
                  :label="sound.name"
                  :value="sound.path"
                />
              </el-option-group>
              <el-option-group label="系统声音" v-if="systemSounds.length > 0">
                <el-option
                  v-for="sound in systemSounds"
                  :key="sound.path"
                  :label="sound.name"
                  :value="sound.path"
                />
              </el-option-group>
              <el-option-group label="自定义声音" v-if="customSounds.length > 0">
                <el-option
                  v-for="sound in customSounds"
                  :key="sound.path"
                  :label="sound.name"
                  :value="sound.path"
                />
              </el-option-group>
            </el-select>
            <el-button 
              :icon="CaretRight" 
              circle 
              @click="previewSound"
              title="试听"
            />
            <el-button 
              type="primary" 
              :icon="Plus" 
              @click="importSound"
            >
              导入
            </el-button>
            <el-button 
              v-if="settings.notificationSoundFile && isCustomSound"
              type="danger" 
              :icon="Delete" 
              @click="deleteCurrentSound"
              title="删除当前声音"
            />
          </div>
        </el-form-item>

        <el-form-item label="通知显示时长">
          <el-select v-model="settings.notificationDuration">
            <el-option label="3秒" :value="3" />
            <el-option label="5秒" :value="5" />
            <el-option label="10秒" :value="10" />
            <el-option label="不自动关闭" :value="0" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveSettings" :loading="saving">
            保存设置
          </el-button>
          <el-button @click="resetSettings">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="about-section card">
      <el-divider content-position="left">关于</el-divider>
      <div class="about-content">
        <div class="app-info">
          <span class="app-icon">📅</span>
          <div>
            <h3>待办日历</h3>
            <p>版本 1.0.0</p>
          </div>
        </div>
        <p class="copyright">© 2025 待办日历. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CaretRight, Plus, Delete } from '@element-plus/icons-vue'
import { useSettingsStore } from '@/stores/settings'
import { models } from '@/wailsjs/go/models'
import * as api from '@/wailsjs/go/app/App'

type Settings = models.Settings

interface SoundInfo {
  name: string
  path: string
  isCustom: boolean
  isSystem: boolean
}

const settingsStore = useSettingsStore()
const loading = ref(false)
const saving = ref(false)
const openingWidget = ref(false)
const closingWidget = ref(false)
const widgetRunning = ref(false)
const availableSounds = ref<SoundInfo[]>([])

const settings = reactive<Settings>({
  id: 1,
  enableWidget: true,
  enableAutoStart: false,
  minimizeToTray: true,
  notificationSound: true,
  notificationSoundFile: 'default',
  notificationDuration: 5,
  widgetPosition: 'bottom-right',
  widgetOpacity: 90,
  theme: 'light'
})

// 检查当前选中的声音是否是自定义声音
const isCustomSound = computed(() => {
  const sound = availableSounds.value.find(s => s.path === settings.notificationSoundFile)
  return sound?.isCustom || false
})

// 分组声音列表
const defaultSounds = computed(() => 
  availableSounds.value.filter(s => !s.isCustom && !s.isSystem)
)

const systemSounds = computed(() => 
  availableSounds.value.filter(s => s.isSystem)
)

const customSounds = computed(() => 
  availableSounds.value.filter(s => s.isCustom)
)

// 检查小部件是否在运行
async function checkWidgetStatus() {
  try {
    widgetRunning.value = await api.IsWidgetRunning()
  } catch (error) {
    widgetRunning.value = false
  }
}

// 加载可用声音列表
async function loadSounds() {
  try {
    const sounds = await api.GetAvailableSounds()
    availableSounds.value = sounds
  } catch (error) {
    console.error('Failed to load sounds:', error)
  }
}

// 预览声音
async function previewSound() {
  try {
    await api.PreviewSound(settings.notificationSoundFile || '')
  } catch (error) {
    ElMessage.error('播放声音失败')
  }
}

// 导入自定义声音
async function importSound() {
  try {
    const newPath = await api.ImportSound()
    if (newPath) {
      await loadSounds()
      settings.notificationSoundFile = newPath
      ElMessage.success('声音导入成功')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '导入声音失败')
  }
}

// 删除当前选中的自定义声音
async function deleteCurrentSound() {
  if (!settings.notificationSoundFile || !isCustomSound.value) return

  try {
    await ElMessageBox.confirm('确定要删除这个自定义声音吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await api.DeleteSound(settings.notificationSoundFile)
    settings.notificationSoundFile = 'default'
    await loadSounds()
    ElMessage.success('声音已删除')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

async function fetchSettings() {
  loading.value = true
  try {
    await settingsStore.fetchSettings()
    if (settingsStore.settings) {
      Object.assign(settings, settingsStore.settings)
      // 确保 notificationSoundFile 有默认值
      if (!settings.notificationSoundFile) {
        settings.notificationSoundFile = 'default'
      }
    }
  } catch (error) {
    console.error('Failed to fetch settings:', error)
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    await settingsStore.updateSettings({ ...settings })
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

function resetSettings() {
  if (settingsStore.settings) {
    Object.assign(settings, settingsStore.settings)
  }
}

async function openWidget() {
  openingWidget.value = true
  try {
    await api.OpenWidget()
    ElMessage.success('小部件已启动')
    // 延迟检查状态，等待窗口创建
    setTimeout(checkWidgetStatus, 1000)
  } catch (error) {
    ElMessage.error('启动小部件失败')
    console.error('Failed to open widget:', error)
  } finally {
    openingWidget.value = false
  }
}

async function closeWidget() {
  closingWidget.value = true
  try {
    await api.CloseWidget()
    ElMessage.success('小部件已关闭')
    widgetRunning.value = false
  } catch (error) {
    ElMessage.error('关闭小部件失败')
    console.error('Failed to close widget:', error)
  } finally {
    closingWidget.value = false
  }
}

onMounted(() => {
  fetchSettings()
  checkWidgetStatus()
  loadSounds()
  // 定期检查小部件状态
  setInterval(checkWidgetStatus, 2000)
})
</script>

<style lang="scss" scoped>
.settings-view {
  max-width: 800px;
}

.settings-content {
  margin-bottom: 20px;

  .setting-hint {
    margin-left: 15px;
    font-size: 12px;
    color: #909399;
  }

  .opacity-value {
    margin-left: 15px;
    color: #606266;
  }

  .sound-selector {
    display: flex;
    align-items: center;
    gap: 10px;
  }
}

.about-section {
  .about-content {
    padding: 10px 0;
  }

  .app-info {
    display: flex;
    align-items: center;
    gap: 15px;
    margin-bottom: 20px;

    .app-icon {
      font-size: 48px;
    }

    h3 {
      margin: 0;
      font-size: 20px;
      color: #303133;
    }

    p {
      margin: 5px 0 0;
      color: #909399;
    }
  }

  .copyright {
    font-size: 12px;
    color: #909399;
  }
}
</style>
