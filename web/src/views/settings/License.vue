<template>
  <div style="max-width: 720px; margin: 0 auto">
    <el-card shadow="never">
      <template #header>
        <div style="display: flex; align-items: center; gap: 10px">
          <span style="font-size: 16px; font-weight: 600">🔑 授权管理</span>
          <el-tag v-if="status" :type="status.activated ? (status.is_expired ? 'danger' : 'success') : 'warning'" size="small">
            {{ status.activated ? (status.is_expired ? '已过期' : '已激活') : '未激活' }}
          </el-tag>
        </div>
      </template>

      <!-- Loading -->
      <div v-if="!status" style="text-align: center; padding: 40px; color: #909399">加载中...</div>

      <template v-else>
        <!-- Machine Code -->
        <div style="margin-bottom: 24px">
          <div style="font-size: 13px; color: #909399; margin-bottom: 6px">本机机器码（请将此码发给授权方）</div>
          <div style="display: flex; align-items: center; gap: 10px">
            <code style="font-size: 16px; font-weight: 700; color: #303133; letter-spacing: 1px; background: #f5f7fa; padding: 8px 16px; border-radius: 6px; user-select: all">
              {{ status.machine_code }}
            </code>
            <el-button size="small" @click="copyCode">复制</el-button>
          </div>
        </div>

        <!-- Not activated: upload form -->
        <div v-if="!status.activated" style="border: 2px dashed #dcdfe6; border-radius: 8px; padding: 40px; text-align: center">
          <div style="font-size: 40px; margin-bottom: 12px">📄</div>
          <div style="font-size: 15px; font-weight: 600; margin-bottom: 6px">导入授权文件</div>
          <div style="font-size: 13px; color: #909399; margin-bottom: 20px">
            将授权方发给您的 license.json 授权文件拖入下方或点击选择
          </div>
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".json"
            :on-change="handleFileChange"
            drag
          >
            <el-icon style="font-size: 40px; color: #409eff; margin-bottom: 10px"><UploadFilled /></el-icon>
            <div style="font-size: 14px; color: #606266">将授权文件拖到此处，或<em style="color: #409eff">点击选择</em></div>
          </el-upload>
          <el-button v-if="selectedFile" type="primary" size="large" style="margin-top: 20px" :loading="activating" @click="submitActivate">
            立即激活
          </el-button>
        </div>

        <!-- Activated: show details -->
        <div v-else>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="客户名称">{{ status.customer }}</el-descriptions-item>
            <el-descriptions-item label="机器码">
              <code style="font-size: 12px; user-select: all">{{ status.machine_code }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="授权类型">
              {{ status.expires_at ? '期限授权' : '永久授权' }}
            </el-descriptions-item>
            <el-descriptions-item v-if="status.expires_at" label="有效期至">
              <span :style="{ color: status.is_expired ? '#f56c6c' : '#67c23a', fontWeight: 600 }">
                {{ status.expires_at }}
                <el-tag v-if="status.is_expired" type="danger" size="small" style="margin-left: 8px">已过期</el-tag>
                <el-tag v-else-if="status.days_left !== undefined && status.days_left <= 30" type="warning" size="small" style="margin-left: 8px">
                  剩余 {{ status.days_left }} 天
                </el-tag>
                <span v-else-if="status.days_left !== undefined" style="margin-left: 8px; color: #909399; font-weight: normal">
                  （剩余 {{ status.days_left }} 天）
                </span>
              </span>
            </el-descriptions-item>
            <el-descriptions-item label="授权功能">
              <el-tag v-for="f in status.features" :key="f" size="small" style="margin-right: 6px">
                {{ featureLabel(f) }}
              </el-tag>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </template>
    </el-card>

    <!-- Seller info -->
    <el-card shadow="never" style="margin-top: 16px">
      <div style="font-size: 13px; color: #909399; line-height: 1.8">
        <p style="margin: 0"><b>授权说明：</b></p>
        <p style="margin: 4px 0">• 首次使用请将上方的「本机机器码」发送给授权方</p>
        <p style="margin: 4px 0">• 收到授权文件后，在本页导入即可完成激活</p>
        <p style="margin: 4px 0">• 更换电脑或重装系统后需重新激活</p>
        <p style="margin: 4px 0">• 如有疑问请联系授权方获取支持</p>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getLicenseStatus, activateLicense } from '../../api/license'
import type { LicenseStatus } from '../../api/license'

const status = ref<LicenseStatus | null>(null)
const selectedFile = ref<File | null>(null)
const activating = ref(false)
const uploadRef = ref<any>(null)

const featureLabels: Record<string, string> = {
  all: '全部功能',
  customer: '客户管理',
  pos: '收银台',
  membership: '会员管理',
  appointment: '预约管理',
  reports: '报表中心',
  inventory: '库存管理',
  medical: '病历管理',
  documents: '证件档案',
  settings: '系统设置',
}

function featureLabel(f: string) {
  return featureLabels[f] || f
}

async function fetchStatus() {
  try {
    status.value = await getLicenseStatus()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '获取授权状态失败')
  }
}

function handleFileChange(_file: any) {
  selectedFile.value = _file.raw
}

async function submitActivate() {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择授权文件')
    return
  }
  activating.value = true
  try {
    const res = await activateLicense(selectedFile.value)
    ElMessage.success(res.message || '授权成功')
    selectedFile.value = null
    await fetchStatus()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '激活失败，请确认授权文件与当前机器码匹配')
  } finally {
    activating.value = false
  }
}

async function copyCode() {
  if (!status.value?.machine_code) return
  try {
    await navigator.clipboard.writeText(status.value.machine_code)
    ElMessage.success('已复制到剪贴板')
  } catch {
    // Fallback
    const ta = document.createElement('textarea')
    ta.value = status.value.machine_code
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    ElMessage.success('已复制到剪贴板')
  }
}

onMounted(fetchStatus)
</script>
