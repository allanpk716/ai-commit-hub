/**
 * 统一版本号格式化，确保只有一个 "v" 前缀
 */
export function formatVersion(version: string): string {
  if (!version) return ''
  // 移除所有前导的 v 或 V，然后统一添加小写 v
  const cleaned = version.replace(/^[vV]+/, '')
  return `v${cleaned}`
}
