/// <reference types="vite/client" />

import type { ClientOptions } from '@xc/shared/request'

// uni-app 全局声明：仅供 @xc/shared 中 uniRequester 的类型检查使用（admin 端不会实际调用）
declare global {
  const uni: {
    request(options: {
      url: string
      method?: string
      data?: unknown
      header?: Record<string, string>
      success?: (res: { statusCode: number; data: unknown }) => void
      fail?: (err: { errmsg?: string }) => void
    }): void
  }

  namespace UniApp {
    interface RequestOptions {
      url: string
      method?: string
      data?: unknown
      header?: Record<string, string>
      success?: (res: { statusCode: number; data: unknown }) => void
      fail?: (err: { errmsg?: string }) => void
    }
  }
}

// 扩充 @xc/shared/request 的 ClientOptions（header 字段，shared 为只读源码，这里做声明合并）
declare module '@xc/shared/request' {
  interface ClientOptions {
    header?: Record<string, string>
  }
}

export {}
