// HTTP 客户端：适配 uni.request 与 axios 两种底层，注入 token、统一解包 R<T>
export interface RequestOptions {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: Record<string, unknown> | undefined
  header?: Record<string, string>
}

// 底层请求器：uni.request 或 axios 的薄包装，返回 { statusCode, data }
export type Requester = (options: RequestOptions) => Promise<{ statusCode: number; data: unknown }>

export interface ClientOptions {
  base: string // 例如 http://localhost:8080
  requester: Requester
  getToken: () => string | undefined
  onUnauthorized?: () => void
  header?: Record<string, string>
}

export class ApiError extends Error {
  code: number
  constructor(code: number, msg: string) {
    super(msg)
    this.code = code
  }
}

function buildQuery(params?: Record<string, unknown>): string {
  if (!params) return ''
  const qs = Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null && v !== '')
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`)
    .join('&')
  return qs ? `?${qs}` : ''
}

export interface ApiClient {
  get<T>(url: string, params?: Record<string, unknown>): Promise<T>
  post<T>(url: string, body?: Record<string, unknown>): Promise<T>
  put<T>(url: string, body?: Record<string, unknown>): Promise<T>
  del<T>(url: string, body?: Record<string, unknown>): Promise<T>
}

export function createClient(opts: ClientOptions): ApiClient {
  async function request<T>(method: RequestOptions['method'], url: string, data?: Record<string, unknown>, isGet = false): Promise<T> {
    const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(opts.header ?? {}) }
    const token = opts.getToken()
    if (token) headers.Authorization = `Bearer ${token}`
    const full = isGet ? `${opts.base}${url}${buildQuery(data)}` : `${opts.base}${url}`
    const res = await opts.requester({
      url: full,
      method,
      data: isGet ? undefined : data,
      header: headers,
    })
    if (res.statusCode === 401 || res.statusCode === 403) {
      opts.onUnauthorized?.()
      throw new ApiError(res.statusCode, '请先登录')
    }
    const body = res.data as { code?: number; msg?: string; data?: T }
    if (body && typeof body === 'object' && 'code' in body) {
      if (body.code === 0) return body.data as T
      if (body.code === 40100) {
        opts.onUnauthorized?.()
      }
      throw new ApiError(body.code ?? -1, body.msg ?? '请求失败')
    }
    if (res.statusCode >= 400) throw new ApiError(res.statusCode, `HTTP ${res.statusCode}`)
    return res.data as T
  }

  return {
    get: <T>(url: string, params?: Record<string, unknown>) => request<T>('GET', url, params, true),
    post: <T>(url: string, body?: Record<string, unknown>) => request<T>('POST', url, body),
    put: <T>(url: string, body?: Record<string, unknown>) => request<T>('PUT', url, body),
    del: <T>(url: string, body?: Record<string, unknown>) => request<T>('DELETE', url, body),
  }
}

// uni.request 适配器（小程序/H5 通用）
export function uniRequester(): Requester {
  return (options) =>
    new Promise((resolve, reject) => {
      uni.request({
        url: options.url,
        method: options.method as unknown as UniApp.RequestOptions['method'],
        data: options.data,
        header: options.header,
        success: (res) => resolve({ statusCode: res.statusCode, data: res.data }),
        fail: (err) => reject(new ApiError(-1, (err as { errMsg?: string }).errMsg ?? '网络异常')),
      })
    })
}

// axios 适配器（管理后台用）
export function axiosRequester(axios: { request: (config: Record<string, unknown>) => Promise<{ status: number; data: unknown }> }): Requester {
  return async (options) => {
    try {
      const res = await axios.request({
        url: options.url,
        method: options.method,
        data: options.data,
        headers: options.header,
      })
      return { statusCode: res.status, data: res.data }
    } catch (err: unknown) {
      const e = err as { response?: { status: number; data: unknown } }
      if (e.response) return { statusCode: e.response.status, data: e.response.data }
      throw new ApiError(-1, '网络异常')
    }
  }
}
