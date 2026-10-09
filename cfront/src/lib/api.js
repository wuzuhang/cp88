const base = (import.meta.env.VITE_API_BASE || '').replace(/\/$/, '')

export async function request(path, options = {}) {
  const response = await fetch(`${base}/api/v1${path}`, {
    ...options,
    headers: { 'Content-Type': 'application/json', ...options.headers },
    signal: options.signal || AbortSignal.timeout(15000),
  })
  let result
  try { result = await response.json() } catch { throw new Error('服务暂不可用，请检查后端连接后重试') }
  if (!response.ok || result.code !== 200) throw new Error(result.msg || '请求失败，请稍后重试')
  return result.data
}

// 首次加载最多 1500 期，覆盖 1000 期样本并留出历史对照空间。
export async function fetchRecords(path='/records') {
  const first = await request(`${path}?limit=500`)
  const pages = []
  for (let offset=500; offset<Math.min(first.total,1500); offset+=500) {
    pages.push(request(`${path}?limit=500&offset=${offset}`))
  }
  const remaining = await Promise.all(pages)
  return { ...first, items: [first, ...remaining].flatMap(page => page.items) }
}
export const analyzeBalls = (reds, blue) => request('/analysis', { method: 'POST', body: JSON.stringify({ red_balls: [...reds].sort((a,b) => a-b), blue_ball: blue }) })
export const analyze3D = numbers => request('/fucai3d/analysis', { method:'POST', body:JSON.stringify({numbers}) })
