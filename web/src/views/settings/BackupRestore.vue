<template>
  <div style="max-width:900px;margin:0 auto">
    <!-- Tabs -->
    <el-tabs v-model="activeTab">
      <el-tab-pane label="📋 备份管理" name="list">
        <el-card shadow="never">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between">
              <span style="font-weight:600">备份文件列表</span>
              <el-button type="primary" :loading="creating" @click="handleCreate" size="small">
                {{ creating?'正在备份...':'立即备份' }}
              </el-button>
            </div>
          </template>
          <el-table :data="backups" stripe border size="small" v-loading="loading" empty-text="暂无备份记录">
            <el-table-column label="文件名" min-width="180" prop="filename"/>
            <el-table-column label="文件大小" width="100">
              <template #default="{row}">{{ formatSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="类型" width="70">
              <template #default="{row}">
                <el-tag :type="row.backup_type==='auto'?'':'success'" size="small">
                  {{ row.backup_type==='auto'?'自动':'手动' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="70">
              <template #default="{row}">
                <el-tag :type="row.status==='success'?'success':'danger'" size="small">
                  {{ row.status==='success'?'成功':'失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="云端" width="80">
              <template #default="{row}">
                <el-tag v-if="row.cloud_status==='uploaded'" type="success" size="small">已上传</el-tag>
                <el-tag v-else-if="row.cloud_status==='failed'" type="danger" size="small">失败</el-tag>
                <span v-else style="color:#909399;font-size:12px">未上传</span>
              </template>
            </el-table-column>
            <el-table-column label="时间" width="150">
              <template #default="{row}">{{ row.created_at?.slice(0,16)||'--' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="195" fixed="right">
              <template #default="{row}">
                <el-button size="small" text type="primary" @click="downloadBackup(row)">下载</el-button>
                <el-button v-if="row.cloud_status!=='uploaded'" size="small" text type="warning" @click="handleCloudUpload(row)">上传云端</el-button>
                <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="backups.length===0 && !loading" style="text-align:center;padding:32px;color:#909399">
            还没有备份记录，点击「立即备份」创建第一个备份
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="⚙️ 自动备份" name="auto">
        <el-card shadow="never">
          <template #header><span style="font-weight:600">自动备份设置</span></template>
          <el-form label-width="120px" size="small">
            <el-form-item label="开启自动备份">
              <el-switch v-model="autoSettings.auto_backup_enabled" @change="saveSettings"/>
            </el-form-item>
            <el-form-item label="备份频率">
              <el-radio-group v-model="autoSettings.backup_interval" :disabled="!autoSettings.auto_backup_enabled" @change="saveSettings">
                <el-radio-button value="daily">每天</el-radio-button>
                <el-radio-button value="weekly">每周</el-radio-button>
                <el-radio-button value="monthly">每月</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="保留天数">
              <div style="display:flex;align-items:center;gap:8px">
                <el-input-number v-model="autoSettings.retention_days" :min="1" :max="365" size="small" style="width:160px" @change="saveSettings"/>
                <span style="color:#909399;font-size:12px">超过此天数的旧备份将自动删除</span>
              </div>
            </el-form-item>
          </el-form>
          <div style="margin-top:12px;padding:12px;background:#f0f9eb;border-radius:6px;font-size:13px;color:#67c23a">
            当前状态：{{ autoSettings.auto_backup_enabled ? '自动备份已开启（'+({daily:'每天',weekly:'每周',monthly:'每月'})[autoSettings.backup_interval]+'）' : '自动备份已关闭' }}
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="☁️ 云存储设置" name="cloud">
        <el-card shadow="never">
          <template #header><span style="font-weight:600">云存储配置</span></template>
          <p style="color:#909399;font-size:13px;margin-bottom:12px">
            支持阿里云 OSS 和腾讯云 COS（使用 S3 兼容协议）。备份完成后自动上传到云端。
          </p>
          <el-form label-width="140px" size="small">
            <el-form-item label="开启云上传">
              <el-switch v-model="autoSettings.cloud_upload_enabled" @change="saveSettings"/>
            </el-form-item>
            <el-form-item label="云服务商">
              <el-select v-model="autoSettings.provider" :disabled="!autoSettings.cloud_upload_enabled" @change="onProviderChange" style="width:200px">
                <el-option label="阿里云 OSS" value="oss"/>
                <el-option label="腾讯云 COS" value="cos"/>
                <el-option label="其他 S3 兼容" value="s3"/>
              </el-select>
            </el-form-item>
            <el-form-item label="Endpoint">
              <el-input v-model="autoSettings.endpoint" :disabled="!autoSettings.cloud_upload_enabled" placeholder="oss-cn-hangzhou.aliyuncs.com" style="width:300px" @change="saveSettings"/>
              <div style="color:#909399;font-size:11px;margin-left:8px">
                {{ providerHint }}
              </div>
            </el-form-item>
            <el-form-item label="Bucket">
              <el-input v-model="autoSettings.bucket" :disabled="!autoSettings.cloud_upload_enabled" placeholder="my-clinic-backup" style="width:300px" @change="saveSettings"/>
            </el-form-item>
            <el-form-item label="Region">
              <el-input v-model="autoSettings.region" :disabled="!autoSettings.cloud_upload_enabled" placeholder="cn-hangzhou" style="width:200px" @change="saveSettings"/>
            </el-form-item>
            <el-form-item label="AccessKey">
              <el-input v-model="autoSettings.access_key" :disabled="!autoSettings.cloud_upload_enabled" placeholder="LTAI..." style="width:300px" @change="saveSettings"/>
            </el-form-item>
            <el-form-item label="SecretKey">
              <el-input v-model="autoSettings.secret_key" type="password" :disabled="!autoSettings.cloud_upload_enabled" placeholder="..." style="width:300px" show-password @change="saveSettings"/>
            </el-form-item>
          </el-form>
          <div v-if="autoSettings.cloud_upload_enabled && autoSettings.endpoint && autoSettings.bucket" style="margin-top:12px;padding:12px;background:#ecf5ff;border-radius:6px;font-size:13px;color:#409eff">
            ✅ 配置已保存，备份完成后将自动上传到 <b>{{ autoSettings.bucket }}</b>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="📥 导入恢复" name="import">
        <el-card shadow="never" style="margin-bottom:12px">
          <template #header><span style="font-weight:600;color:#e6a23c">导入恢复</span></template>
          <p style="color:#e6a23c;margin-bottom:12px;font-size:13px">⚠ 导入将覆盖当前所有数据，请先导出备份再操作。</p>
          <el-upload ref="uploadRef" :auto-upload="false" :show-file-list="true" accept=".sql" :limit="1" :on-change="onFileChange">
            <template #trigger><el-button plain>选择文件</el-button></template>
          </el-upload>
          <div v-if="selectedFile" style="margin-top:12px"><el-button type="danger" :loading="importing" @click="handleImport">{{ importing?'正在恢复...':'开始恢复' }}</el-button></div>
        </el-card>
        <el-card shadow="never">
          <template #header><span style="font-weight:600;color:#f56c6c">初始化系统</span></template>
          <p style="color:#f56c6c;margin-bottom:12px;font-size:13px">⚠ 清除全部业务数据，仅保留员工账号和系统设置。建议先导出备份。</p>
          <el-button type="danger" :loading="resetting" @click="handleReset">{{ resetting?'正在初始化...':'初始化系统' }}</el-button>
        </el-card>
        <el-card shadow="never" style="margin-top:12px">
          <template #header><span style="font-weight:600">快速导出</span></template>
          <p style="color:#606266;margin-bottom:12px;font-size:14px">直接下载当前数据库的 SQL 文件，不进备份列表。</p>
          <el-button :loading="exporting" @click="handleExport">{{ exporting?'正在导出...':'导出 SQL' }}</el-button>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listBackups, createBackup, deleteBackup, downloadBackupFile, uploadToCloud, getBackupSettings, saveBackupSettings, exportBackup, importBackup, resetSystem } from '../../api/backup'

const activeTab = ref('list')
const loading = ref(false)
const backups = ref<any[]>([])
const creating = ref(false)
const exporting = ref(false)
const importing = ref(false)
const selectedFile = ref<File | null>(null)
const resetting = ref(false)

const autoSettings = reactive({
  auto_backup_enabled: false,
  backup_interval: 'daily' as string,
  retention_days: 30,
  cloud_upload_enabled: false,
  provider: 'oss' as string,
  endpoint: '',
  bucket: '',
  region: 'cn-hangzhou',
  access_key: '',
  secret_key: '',
})

const providerHint = computed(() => {
  if (!autoSettings.cloud_upload_enabled) return ''
  switch (autoSettings.provider) {
    case 'oss': return '例：oss-cn-hangzhou.aliyuncs.com'
    case 'cos': return '例：cos.ap-guangzhou.myqcloud.com'
    case 's3': return '输入 S3 兼容的 Endpoint'
    default: return ''
  }
})

function onProviderChange(val: string) {
  switch (val) {
    case 'oss':
      autoSettings.region = 'cn-hangzhou'
      saveSettings()
      break
    case 'cos':
      autoSettings.region = 'ap-guangzhou'
      saveSettings()
      break
  }
}

function formatSize(bytes: number) {
  if (!bytes) return '0 B'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024*1024) return (bytes/1024).toFixed(1) + ' KB'
  return (bytes/1024/1024).toFixed(1) + ' MB'
}

async function loadBackups() {
  loading.value = true
  try {
    const data = await listBackups()
    backups.value = Array.isArray(data) ? data : []
  } catch (e: any) {
    console.error('加载备份列表失败', e)
  } finally {
    loading.value = false
  }
}

async function loadSettings() {
  try {
    const data = await getBackupSettings()
    if (data) Object.assign(autoSettings, data)
  } catch (e: any) {
    console.error('加载设置失败', e)
  }
}

async function saveSettings() {
  try {
    await saveBackupSettings({ ...autoSettings })
  } catch (e: any) {
    console.error('保存设置失败', e)
  }
}

async function handleCreate() {
  creating.value = true
  try {
    await createBackup()
    ElMessage.success('备份成功')
    await loadBackups()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '备份失败')
  } finally {
    creating.value = false
  }
}

async function downloadBackup(row: any) {
  try {
    await downloadBackupFile(row.id, row.filename || `backup_${row.id}.sql`)
    ElMessage.success('文件已下载')
  } catch {
    ElMessage.error('下载失败')
  }
}

async function handleCloudUpload(row: any) {
  try {
    await uploadToCloud(row.id)
    ElMessage.success('已上传到云端')
    await loadBackups()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '上传失败')
  }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定删除备份文件 ${row.filename}？`, '确认删除', { type: 'warning' })
    await deleteBackup(row.id)
    ElMessage.success('已删除')
    await loadBackups()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

async function handleExport() {
  exporting.value = true
  try {
    await exportBackup()
    ElMessage.success('备份文件已下载')
  } catch {
    ElMessage.error('导出失败')
  } finally {
    exporting.value = false
  }
}

function onFileChange(uploadFile: any) { selectedFile.value = uploadFile.raw || null }

async function handleImport() {
  if (!selectedFile.value) return
  try {
    await ElMessageBox.confirm('导入将覆盖当前所有数据，此操作不可撤销。确定继续？', '危险操作', {
      confirmButtonText: '确认导入', cancelButtonText: '取消', type: 'warning'
    })
  } catch { return }
  importing.value = true
  try {
    const res = await importBackup(selectedFile.value)
    ElMessage.success(res?.message || '恢复成功')
    selectedFile.value = null
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '恢复失败')
  } finally {
    importing.value = false
  }
}

async function handleReset() {
  try {
    await ElMessageBox.prompt('此操作将清除所有业务数据且不可撤销。\n请输入 "确认初始化" 以继续：', '危险操作', {
      confirmButtonText: '确认初始化', cancelButtonText: '取消',
      inputPlaceholder: '请输入：确认初始化',
      inputValidator: (v: string) => v === '确认初始化' ? true : '请输入"确认初始化"',
      type: 'warning'
    })
  } catch { return }
  resetting.value = true
  try {
    await resetSystem()
    ElMessage.success('系统已初始化，所有业务数据已清除')
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '初始化失败')
  } finally {
    resetting.value = false
  }
}

onMounted(() => {
  loadBackups()
  loadSettings()
})
</script>
