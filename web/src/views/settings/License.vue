<template>
  <div style="padding: 20px; max-width: 800px">
    <h2 style="margin-bottom: 24px; font-size: 18px; color: #303133">授权管理</h2>

    <el-card shadow="never" style="margin-bottom: 20px">
      <template #header><span>系统激活状态</span></template>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="激活状态">
          <el-tag v-if="status?.activated" type="success">已激活</el-tag>
          <el-tag v-else type="danger">未激活</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="客户数量">
          {{ status?.customer_count || 0 }} / {{ status?.customer_limit || '不限' }}
        </el-descriptions-item>
        <el-descriptions-item label="病历数量">
          {{ status?.record_count || 0 }} / {{ status?.record_limit || '不限' }}
        </el-descriptions-item>
        <el-descriptions-item label="机器码">
          <code style="font-size: 13px; background: #f5f7fa; padding: 2px 8px; border-radius: 3px">{{ status?.machine_code || '—' }}</code>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card v-if="!status?.activated" shadow="never">
      <template #header><span>激活系统</span></template>
      <el-alert v-if="status && !status.activated" title="当前系统未激活，部分功能受限" type="warning" show-icon :closable="false" style="margin-bottom: 16px" />
      <el-form label-width="100px">
        <el-form-item label="机器码">
          <el-input :model-value="status?.machine_code" readonly>
            <template #append>
              <el-button @click="copyCode">复制</el-button>
            </template>
          </el-input>
          <div style="font-size: 12px; color: #909399; margin-top: 4px">请将机器码发给软件供应商获取解锁码</div>
        </el-form-item>
        <el-form-item label="解锁码">
          <el-input v-model="unlockCode" placeholder="输入供应商提供的解锁码" @keyup.enter="doActivate" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="activating" @click="doActivate">立即激活</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card v-else shadow="never">
      <template #header><span>授权信息</span></template>
      <el-result icon="success" title="系统已激活" sub-title="所有功能均可正常使用">
        <template #extra>
          <el-button type="danger" plain @click="showDeactivate = true">卸载授权</el-button>
        </template>
      </el-result>
    </el-card>

    <el-dialog v-model="showDeactivate" title="确认卸载授权" width="360px" append-to-body>
      <p>卸载授权后系统将回到未激活状态，确定要继续吗？</p>
      <template #footer>
        <el-button @click="showDeactivate = false">取消</el-button>
        <el-button type="danger" :loading="deactivating" @click="doDeactivate">确认卸载</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
import { ElMessage } from 'element-plus'

const status = ref<any>(null)
const unlockCode = ref('')
const activating = ref(false)
const deactivating = ref(false)
const showDeactivate = ref(false)

async function checkStatus() {
  try {
    const res = await api.get('/license/status')
    status.value = res.data
  } catch {}
}

async function doActivate() {
  if (!unlockCode.value.trim()) return
  activating.value = true
  try {
    const res = await api.post('/license/activate', { code: unlockCode.value.trim() })
    ElMessage.success(res.data.message || '激活成功')
    unlockCode.value = ''
    await checkStatus()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '激活失败')
  } finally {
    activating.value = false
  }
}

async function doDeactivate() {
  deactivating.value = true
  try {
    await api.post('/license/deactivate')
    showDeactivate.value = false
    ElMessage.success('授权已卸载')
    await checkStatus()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '卸载失败')
  } finally {
    deactivating.value = false
  }
}

function copyCode() {
  if (status.value?.machine_code) {
    navigator.clipboard.writeText(status.value.machine_code)
    ElMessage.success('已复制机器码')
  }
}

onMounted(() => checkStatus())
</script>
