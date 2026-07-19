<template>
  <div style="max-width:700px;margin:0 auto">
    <el-card shadow="never">
      <template #header>
        <div style="display:flex;align-items:center;gap:10px">
          <span style="font-size:16px;font-weight:600">🔑 授权管理</span>
          <el-tag v-if="status" :type="status.activated?(status.is_expired?'danger':'success'):'warning'" size="small">
            {{ status.activated?(status.is_expired?'已过期':'已激活'):'未激活' }}
          </el-tag>
        </div>
      </template>

      <div v-if="!status" style="text-align:center;padding:40px;color:#909399">加载中...</div>
      <template v-else>
        <!-- 机器码 -->
        <div style="margin-bottom:24px">
          <div style="font-size:13px;color:#909399;margin-bottom:6px">本机机器码（发给授权方）</div>
          <div style="display:flex;align-items:center;gap:10px">
            <code style="font-size:16px;font-weight:700;color:#303133;letter-spacing:1px;background:#f5f7fa;padding:8px 16px;border-radius:6px;user-select:all">
              {{ status.machine_code }}
            </code>
            <el-button size="small" @click="copyCode">复制</el-button>
          </div>
        </div>

        <!-- 未激活：上传 -->
        <div v-if="!status.activated" style="border:2px dashed #dcdfe6;border-radius:8px;padding:40px;text-align:center">
          <div style="font-size:40px;margin-bottom:12px">📄</div>
          <div style="font-size:15px;font-weight:600;margin-bottom:6px">导入授权文件</div>
          <div style="font-size:13px;color:#909399;margin-bottom:20px">将授权方给您的 .license 文件拖入下方</div>
          <el-upload ref="uploadRef" :auto-upload="false" :limit="1" accept=".json,.license" :on-change="handleFile" drag>
            <el-icon style="font-size:40px;color:#409eff;margin-bottom:10px"><UploadFilled /></el-icon>
            <div style="font-size:14px;color:#606266">拖到此处，或<em style="color:#409eff">点击选择</em></div>
          </el-upload>
          <el-button v-if="selectedFile" type="primary" size="large" style="margin-top:20px" :loading="activating" @click="doActivate">
            立即激活
          </el-button>
        </div>

        <!-- 已激活 -->
        <div v-else>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="客户">{{ status.customer }}</el-descriptions-item>
            <el-descriptions-item label="机器码">
              <code style="font-size:12px;user-select:all">{{ status.machine_code }}</code>
            </el-descriptions-item>
            <el-descriptions-item label="授权类型">
              {{ status.expires_at ? '期限授权' : '永久授权' }}
            </el-descriptions-item>
            <el-descriptions-item v-if="status.expires_at" label="有效期至">
              <span :style="{color:status.is_expired?'#f56c6c':'#67c23a',fontWeight:600}">
                {{ status.expires_at }}
                <el-tag v-if="status.is_expired" type="danger" size="small" style="margin-left:8px">已过期</el-tag>
                <el-tag v-else-if="status.days_left<=30" type="warning" size="small" style="margin-left:8px">
                  剩余{{ status.days_left }}天
                </el-tag>
              </span>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </template>
    </el-card>

    <el-card shadow="never" style="margin-top:16px">
      <div style="font-size:13px;color:#909399;line-height:1.8">
        <p style="margin:0"><b>说明：</b></p>
        <p style="margin:4px 0">• 首次使用请将机器码发给授权方</p>
        <p style="margin:4px 0">• 收到 .license 授权文件后导入即可激活</p>
        <p style="margin:4px 0">• 每台电脑需单独授权</p>
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

async function fetch() {
  try {
    status.value = await getLicenseStatus()
  } catch { ElMessage.error('获取授权状态失败') }
}

function handleFile(f: any) { selectedFile.value = f.raw }

async function doActivate() {
  if (!selectedFile.value) { ElMessage.warning('请选择授权文件'); return }
  activating.value = true
  try {
    const r = await activateLicense(selectedFile.value)
    ElMessage.success(r.message || '激活成功')
    selectedFile.value = null
    await fetch()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '激活失败')
  } finally { activating.value = false }
}

async function copyCode() {
  if (!status.value?.machine_code) return
  try {
    await navigator.clipboard.writeText(status.value.machine_code)
    ElMessage.success('已复制')
  } catch {
    const ta = document.createElement('textarea')
    ta.value = status.value.machine_code
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    ElMessage.success('已复制')
  }
}

onMounted(fetch)
</script>
