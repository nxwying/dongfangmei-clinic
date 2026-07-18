<template>
  <div style="max-width:900px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:14px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">角色管理</span>
        <div style="flex:1"/>
        <el-button type="primary" size="small" @click="openCreate">+ 新建角色</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="roles" v-loading="loading" stripe size="small" empty-text="暂无角色">
        <el-table-column label="角色名称" prop="name" min-width="100"/>
        <el-table-column label="说明" prop="description" min-width="140"/>
        <el-table-column label="已授权模块" min-width="200">
          <template #default="{row}">
            <el-tag v-for="k in (row.permissions||[])" :key="k" size="small" style="margin:1px 2px">{{ modLabel(k) }}</el-tag>
            <span v-if="!row.permissions?.length" style="color:#c0c4cc;font-size:12px">无模块权限</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.name!=='admin'" size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑角色':'新建角色'" width="600px">
      <el-form label-width="80px" size="small">
        <el-form-item label="角色名称" required>
          <el-input v-model="form.name" placeholder="如：咨询师、医生"/>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" placeholder="角色的简要说明"/>
        </el-form-item>
        <el-form-item label="模块权限">
          <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:4px">
            <div v-for="m in modules" :key="m.key" style="margin:2px 0">
              <el-checkbox :label="m.label" :checked="form.permissions.includes(m.key)" @change="togglePerm(m.key)"/>
            </div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const MODULES = [
  { key: 'dashboard', label: '工作台' },
  { key: 'customers', label: '客户管理' },
  { key: 'appointments', label: '预约管理' },
  { key: 'pos', label: '收银台' },
  { key: 'refund', label: '退款管理' },
  { key: 'members', label: '会员管理' },
  { key: 'followup', label: '回访管理' },
  { key: 'medical', label: '病历管理' },
  { key: 'inventory', label: '库存管理' },
  { key: 'documents', label: '证件档案' },
  { key: 'marketing', label: '营销工具' },
  { key: 'data', label: '数据中心' },
  { key: 'performance', label: '绩效中心' },
  { key: 'kpi', label: 'KPI目标' },
  { key: 'analysis', label: '运营分析' },
  { key: 'reports', label: '报表中心' },
  { key: 'expenses', label: '支出管理' },
  { key: 'staff', label: '员工管理' },
  { key: 'roles', label: '角色管理' },
  { key: 'items', label: '项目管理' },
  { key: 'packages', label: '套餐模板' },
  { key: 'commission', label: '提成规则' },
  { key: 'training', label: '培训认证' },
  { key: 'settings', label: '系统设置(备份/日志/授权)' },
]

const modules = MODULES
const loading = ref(false)
const roles = ref<any[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)
const form = ref({ name: '', description: '', permissions: [] as string[] })

function modLabel(key: string): string {
  const m = MODULES.find(x => x.key === key)
  return m?.label || key
}

function togglePerm(key: string) {
  const idx = form.value.permissions.indexOf(key)
  if (idx >= 0) form.value.permissions.splice(idx, 1)
  else form.value.permissions.push(key)
}

async function loadRoles() {
  loading.value = true
  try {
    const r = await api.get('/roles')
    roles.value = Array.isArray(r.data) ? r.data : []
  } catch { roles.value = [] }
  finally { loading.value = false }
}

function openCreate() {
  isEdit.value = false; editingId.value = null
  form.value = { name: '', description: '', permissions: [] }
  dialogVisible.value = true
}

function openEdit(row: any) {
  isEdit.value = true; editingId.value = row.id
  form.value = {
    name: row.name || '',
    description: row.description || '',
    permissions: row.permissions || [],
  }
  dialogVisible.value = true
}

async function submit() {
  if (!form.value.name) { ElMessage.warning('请输入角色名称'); return }
  saving.value = true
  try {
    if (isEdit.value && editingId.value) {
      await api.put(`/roles/${editingId.value}`, form.value)
      ElMessage.success('已更新')
    } else {
      await api.post('/roles', form.value)
      ElMessage.success('已创建')
    }
    dialogVisible.value = false
    await loadRoles()
  } catch (e: any) { ElMessage.error(e?.response?.data?.error || '操作失败') }
  finally { saving.value = false }
}

async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定删除角色「${row.name}」吗？`, '确认', { type: 'warning' })
    await api.delete(`/roles/${row.id}`)
    ElMessage.success('已删除')
    await loadRoles()
  } catch { }
}

onMounted(loadRoles)
</script>
