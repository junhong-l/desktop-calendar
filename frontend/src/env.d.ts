/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface Window {
  go: {
    'todo-calendar/internal/app': {
      App: {
        CreateTodo: (todo: any) => Promise<number>
        UpdateTodo: (todo: any) => Promise<void>
        DeleteTodo: (id: number) => Promise<void>
        GetTodo: (id: number) => Promise<any>
        GetTodoList: (filter: any) => Promise<any>
        GetPendingTodos: () => Promise<any[]>
        GetWeekTodos: () => Promise<any>
        MarkTodoCompleted: (id: number, completed: boolean) => Promise<void>
        GetTodosByDate: (date: string) => Promise<any[]>
        GetTodosByMonth: (year: number, month: number) => Promise<any[]>
        GetCalendarMonth: (year: number, month: number) => Promise<any[]>
        GetLunarDate: (year: number, month: number, day: number) => Promise<any>
        ConvertLunarToSolar: (year: number, month: number, day: number, isLeap: boolean) => Promise<Date>
        ParseCronExpression: (expr: string) => Promise<any>
        UploadAttachment: (todoId: number, fileName: string, data: string, mimeType: string) => Promise<any>
        GetAttachment: (id: number) => Promise<string>
        GetAttachmentInfo: (id: number) => Promise<any>
        GetTodoAttachments: (todoId: number) => Promise<any[]>
        DeleteAttachment: (id: number) => Promise<void>
        GetSettings: () => Promise<any>
        UpdateSettings: (settings: any) => Promise<void>
        GetTodoTypes: () => Promise<any[]>
      }
    }
  }
}
