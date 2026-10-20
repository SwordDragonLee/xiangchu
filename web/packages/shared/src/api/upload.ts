// 上传：优先直传七牛（凭证由后端签发），后端未配置时降级报错提示
import type { ApiClient } from '../request'

export interface UploadToken {
  token: string
  key: string
  domain: string
  region: string
}

export function createUploadApi(c: ApiClient) {
  return {
    token: (ext: string) => c.get<UploadToken>('/upload/token', { ext }),
  }
}
