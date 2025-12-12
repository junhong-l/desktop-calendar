# 📅 待办日历 - Windows 桌面应用# 待办日历 - Windows桌面应用



一个功能完善的 Windows 桌面日历待办应用，基于 Wails 框架开发，支持农历显示、Cron 定时提醒、桌面小部件、系统托盘等功能。一个功能完善的Windows桌面日历待办应用，支持农历显示、Cron定时提醒、桌面小部件等功能。



## ✨ 功能特性## 技术栈



### 📆 日历功能- **后端**: Go + Wails v2

- **月视图显示** - 直观的日历界面，同时显示公历和农历日期- **前端**: Vue 3 + Element Plus + TypeScript

- **悬停详情** - 鼠标悬停显示详细日期信息（周数、农历年月日、节日、节气等）- **数据库**: SQLite

- **待办预览** - 日期单元格直接显示当天待办事项- **打包**: Wails (生成Windows exe)

- **右键快捷操作** - 右键菜单快速创建待办

## 功能特性

### 📝 待办管理

- **8种待办类型** - 生日🎂、会议📅、纪念日💑、提醒⏰、任务📋、节日🎉、截止日期⏳、预约📌### 日历功能

- **Markdown 编辑** - 支持 Markdown 格式编写待办内容- 月视图显示，同时显示公历和农历日期

- **附件管理** - 支持附件上传，采用 AES-256 加密存储- 鼠标悬停显示详细日期信息（周数、农历年月日、节日等）

- **Cron 定时重复** - 支持 Cron 表达式设置重复提醒，自动解析显示未来执行时间- 右键菜单快速创建待办

- **高级筛选** - 支持按年月、类型、状态筛选和关键词搜索- 日期单元格显示当天待办预览



### 🎂 生日特殊功能### 待办管理

- **农历/公历选择** - 生日类型支持选择农历或公历日期- 8种待办类型：生日、会议、纪念日、提醒、任务、节日、截止日期、预约

- **隐藏年份** - 可只选择月日，隐藏具体年份- Markdown格式内容编辑

- 支持附件上传（AES加密存储）

### 🔔 通知提醒系统- Cron表达式定时重复，自动解析显示未来5次执行时间

- **多重提醒** - 支持提前提醒、开始提醒、结束提醒- 分页显示和搜索过滤

- **Windows Toast 通知** - 原生 Windows 通知，支持自定义声音

- **弹窗通知** - Outlook 风格右下角弹窗，始终置顶显示### 生日特殊功能

- **声音提醒** - 支持选择系统声音或导入自定义 WAV 声音- 支持农历/公历选择

- **自动提醒** - 打开软件时自动弹窗提醒今日和过期待办- 可隐藏年份，只选择月日



### 🖥️ 桌面小部件### 通知提醒

- **悬浮显示** - 独立的桌面小部件窗口，显示本周和逾期待办- 打开软件时自动弹窗提醒今日和过期待办

- **透明度可调** - 支持调节小部件透明度- 显示提醒次数 (x次/x次)

- **桌面嵌入** - 可设置为桌面级别，不遮挡其他窗口- 支持稍后提醒和标记完成

- **实时更新** - 待办状态变化实时同步

### 桌面小部件

### ⚙️ 系统功能- 显示本周和逾期未完成的待办

- **系统托盘** - 最小化到系统托盘，后台运行- 可调节透明度和位置

- **开机自启** - 支持设置开机自动启动- 实时更新

- **深色/浅色主题** - 支持主题切换（开发中）

- **数据本地存储** - 所有数据保存在本地，无需联网### 系统功能

- 开机自启动

## 🛠️ 技术栈- 最小化到系统托盘

- 深色/浅色主题

| 类别 | 技术 |

|------|------|## 项目结构

| **框架** | [Wails v2](https://wails.io/) - Go + Web 技术构建桌面应用 |

| **后端** | Go 1.21+ |```

| **前端** | Vue 3 + TypeScript + Vite |Todo/

| **UI 组件** | Element Plus |├── backend/                    # Go后端

| **状态管理** | Pinia |│   ├── internal/

| **数据库** | SQLite (本地存储) |│   │   ├── app/               # 应用核心逻辑

| **加密** | AES-256 (附件加密) |│   │   ├── database/          # 数据库操作

│   │   ├── models/            # 数据模型

## 📁 项目结构│   │   ├── notification/      # 通知管理

│   │   ├── tray/              # 系统托盘

```│   │   └── utils/             # 工具函数

Todo/│   ├── main.go                # 入口文件

├── build/                      # 构建输出│   ├── go.mod

│   ├── appicon.png            # 应用图标│   └── wails.json             # Wails配置

│   ├── icon.svg               # 图标源文件│

│   └── windows/├── frontend/                   # Vue前端

│       └── icon.ico           # Windows图标│   ├── src/

││   │   ├── components/        # 组件

├── internal/                   # Go 后端代码│   │   ├── views/             # 页面

│   ├── app/                   # 应用核心逻辑│   │   ├── stores/            # Pinia状态管理

│   │   ├── app.go            # 主应用│   │   ├── router/            # 路由

│   │   └── window_mode.go    # 窗口模式服务│   │   ├── wailsjs/           # Wails绑定

│   ├── database/              # 数据库操作│   │   └── styles/            # 样式

│   │   ├── database.go       # 数据库初始化│   ├── package.json

│   │   ├── todo_repository.go│   └── vite.config.ts

│   │   ├── settings_repository.go│

│   │   └── attachment_repository.go└── 需求.md                     # 需求文档

│   ├── models/                # 数据模型```

│   │   └── todo.go

│   ├── notification/          # 通知管理## 开发环境准备

│   │   ├── notifier.go       # 通知调度器

│   │   ├── toast_windows.go  # Windows Toast通知### 安装依赖

│   │   └── sound.go          # 声音管理

│   ├── tray/                  # 系统托盘1. 安装 Go 1.21+

│   │   ├── tray.go2. 安装 Node.js 18+

│   │   ├── tray_windows.go3. 安装 Wails CLI:

│   │   └── icon.ico   ```bash

│   └── utils/                 # 工具函数   go install github.com/wailsapp/wails/v2/cmd/wails@latest

│       ├── lunar.go          # 农历计算   ```

│       ├── cron.go           # Cron解析

│       ├── crypto.go         # 加密工具### 开发运行

│       ├── autostart.go      # 开机自启

│       └── window_windows.go # 窗口操作```bash

│# 进入后端目录

├── frontend/                   # Vue 前端代码cd backend

│   ├── src/

│   │   ├── views/            # 页面# 下载Go依赖

│   │   │   ├── CalendarView.vue      # 日历页go mod tidy

│   │   │   ├── TodosView.vue         # 待办列表

│   │   │   ├── HistoryView.vue       # 历史记录# 开发模式运行

│   │   │   ├── SettingsView.vue      # 设置页wails dev

│   │   │   ├── WidgetView.vue        # 桌面小部件```

│   │   │   └── NotificationPopupView.vue  # 通知弹窗

│   │   ├── components/       # 组件### 构建发布

│   │   │   ├── TodoFormDialog.vue    # 待办编辑弹窗

│   │   │   ├── DayDetailDialog.vue   # 日期详情弹窗```bash

│   │   │   ├── NotificationDialog.vue # 通知弹窗cd backend

│   │   │   └── HistoryDetailDialog.vue

│   │   ├── stores/           # Pinia 状态管理# 构建Windows exe

│   │   ├── router/           # 路由配置wails build

│   │   ├── wailsjs/          # Wails 自动生成的绑定

│   │   └── styles/           # 全局样式# 或构建带调试信息的版本

│   ├── index.htmlwails build -debug

│   ├── package.json```

│   ├── tsconfig.json

│   └── vite.config.ts构建完成后，可执行文件位于 `backend/build/bin/待办日历.exe`

│

├── main.go                     # 应用入口## 数据存储

├── wails.json                  # Wails 配置

├── go.mod- 数据库文件: `<程序目录>/data/todo_calendar.db`

├── go.sum- 附件存储: `<程序目录>/data/attachments/` (AES-256加密)

├── generate-icon.js            # 图标生成脚本

└── README.md## API接口

```

### 待办管理

## 🚀 快速开始- `CreateTodo(todo)` - 创建待办

- `UpdateTodo(todo)` - 更新待办

### 环境要求- `DeleteTodo(id)` - 删除待办

- `GetTodo(id)` - 获取单个待办

- **Go** 1.21 或更高版本- `GetTodoList(filter)` - 获取待办列表

- **Node.js** 18 或更高版本- `MarkTodoCompleted(id, completed)` - 标记完成状态

- **Wails CLI** v2

### 日历

### 安装 Wails CLI- `GetCalendarMonth(year, month)` - 获取月历数据

- `GetLunarDate(year, month, day)` - 获取农历信息

```bash- `ConvertLunarToSolar(...)` - 农历转公历

go install github.com/wailsapp/wails/v2/cmd/wails@latest

```### Cron

- `ParseCronExpression(expr)` - 解析Cron表达式

### 开发运行

### 附件

```bash- `UploadAttachment(...)` - 上传附件(加密)

# 克隆项目- `GetAttachment(id)` - 获取附件(解密)

git clone https://github.com/your-repo/todo-calendar.git- `DeleteAttachment(id)` - 删除附件

cd todo-calendar

### 设置

# 安装前端依赖- `GetSettings()` - 获取设置

cd frontend && npm install && cd ..- `UpdateSettings(settings)` - 更新设置



# 开发模式运行（热重载）## 许可证

wails dev

```MIT License


### 构建发布

```bash
# 构建 Windows 可执行文件
wails build

# 构建带调试工具的版本
wails build -debug
```

构建完成后，可执行文件位于 `build/bin/待办日历.exe`

### 更换图标

```bash
# 编辑 build/icon.svg 后运行
node generate-icon.js

# 重新构建
wails build
```

## 💾 数据存储

所有数据保存在程序目录下，便于备份和迁移：

```
<程序目录>/
├── data/
│   ├── todo_calendar.db      # SQLite 数据库
│   └── attachments/          # 附件存储（AES-256 加密）
└── sounds/                    # 自定义通知声音
```

## 🔌 API 接口

### 待办管理
| 方法 | 说明 |
|------|------|
| `CreateTodo(todo)` | 创建待办 |
| `UpdateTodo(todo)` | 更新待办 |
| `DeleteTodo(id)` | 删除待办 |
| `GetTodo(id)` | 获取单个待办 |
| `GetTodoList(filter)` | 获取待办列表（支持筛选分页） |
| `GetTodosByDate(date)` | 获取指定日期的待办 |
| `MarkTodoCompleted(id, completed)` | 标记完成状态 |

### 日历与农历
| 方法 | 说明 |
|------|------|
| `GetCalendarMonth(year, month)` | 获取月历数据 |
| `GetLunarDate(year, month, day)` | 获取农历信息 |
| `ConvertLunarToSolar(year, month, day, isLeap)` | 农历转公历 |

### Cron 表达式
| 方法 | 说明 |
|------|------|
| `ParseCronExpression(expr)` | 解析 Cron 表达式 |
| `GetNextCronTimes(expr, count)` | 获取未来执行时间 |

### 附件管理
| 方法 | 说明 |
|------|------|
| `UploadAttachment(todoId, name, data)` | 上传附件（自动加密） |
| `GetAttachment(id)` | 获取附件（自动解密） |
| `DeleteAttachment(id)` | 删除附件 |

### 系统设置
| 方法 | 说明 |
|------|------|
| `GetSettings()` | 获取设置 |
| `UpdateSettings(settings)` | 更新设置 |
| `OpenWidget()` | 打开桌面小部件 |
| `CloseWidget()` | 关闭桌面小部件 |

### 通知声音
| 方法 | 说明 |
|------|------|
| `GetAvailableSounds()` | 获取可用声音列表 |
| `PreviewSound(path)` | 预览声音 |
| `ImportSound()` | 导入自定义声音 |
| `DeleteSound(path)` | 删除自定义声音 |

## 📄 许可证

[MIT License](LICENSE)

---

**Made with ❤️ using [Wails](https://wails.io/)**
