<template>
  <div>
    <!-- Toolbar -->
    <el-card shadow="never" style="margin-bottom: 16px">
      <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap">
        <el-input
          v-model="keyword"
          placeholder="搜索姓名或手机号"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
          @clear="handleSearch"
        />
        <el-button type="primary" @click="handleSearch">搜索</el-button>
        <el-radio-group v-model="statusFilter" @change="fetchData">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="potential">潜在客户</el-radio-button>
          <el-radio-button value="active">活跃客户</el-radio-button>
          <el-radio-button value="lost">流失客户</el-radio-button>
        </el-radio-group>
        <div style="flex: 1" />
        <el-button type="primary" @click="dialogVisible = true">新建客户</el-button>
      </div>
    </el-card>

    <!-- Table -->
    <el-card shadow="never">
      <el-table
        :data="customers"
        v-loading="loading"
        stripe
        style="width: 100%"
        @row-click="handleRowClick"
      >
        <el-table-column prop="name" label="姓名" width="120" />
        <el-table-column prop="phone" label="手机号" width="140" />
        <el-table-column label="会员等级" width="120">
          <template #default="{ row }">
            <el-tag v-if="row.membership?.level" :type="levelTagType(row.membership.level)" size="small">
              {{ levelLabel(row.membership.level) }}
            </el-tag>
            <span v-else style="color: #c0c4cc; font-size: 13px">--</span>
          </template>
        </el-table-column>
        <el-table-column prop="source" label="来源" width="120">
          <template #default="{ row }">
            {{ sourceLabel(row.source) }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click.stop="router.push(`/customers/${row.id}`)">
              查看
            </el-button>
            <el-button type="danger" size="small" plain @click.stop="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="display: flex; justify-content: center; margin-top: 20px">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="fetchData"
          @current-change="fetchData"
        />
      </div>
    </el-card>

    <!-- Create dialog -->
    <el-dialog v-model="dialogVisible" title="新建客户" width="520px" @close="resetForm">
      <el-form ref="formRef" :model="form" label-width="80px" :rules="rules">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" placeholder="请输入手机号" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="form.gender" placeholder="请选择性别" style="width: 100%">
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input v-model="form.id_card" placeholder="18 位身份证号" maxlength="18" @input="onIdCardInput" />
        </el-form-item>
        <el-form-item label="生日" prop="birthday">
          <el-date-picker
            v-model="form.birthday"
            type="date"
            placeholder="选择生日"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-select v-model="form.source" placeholder="请选择来源" style="width: 100%">
            <el-option label="到店" value="walk_in" />
            <el-option label="转介绍" value="referral" />
            <el-option label="小红书" value="xiaohongshu" />
            <el-option label="微信" value="wechat" />
            <el-option label="抖音" value="douyin" />
            <el-option label="大众点评" value="dianping" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="备注（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCustomers, createCustomer, deleteCustomer } from '../../api/customer'

const router = useRouter()

// ── Filters ──────────────────────────────────────────────
const keyword = ref('')
const statusFilter = ref('')
const loading = ref(false)
const customers = ref<any[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

async function fetchData() {
  loading.value = true
  try {
    const params: any = {
      page: page.value,
      page_size: pageSize.value,
    }
    if (keyword.value) params.keyword = keyword.value
    if (statusFilter.value) params.status = statusFilter.value
   const res = await getCustomers(params)
    customers.value = res?.data ?? []
    total.value = res?.total ?? 0
  } catch {
    ElMessage.error('加载客户列表失败')
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  fetchData()
}

function handleRowClick(row: any) {
  router.push(`/customers/${row.id}`)
}

// ── Status helpers ──────────────────────────────────────
function statusTagType(status: string): 'info' | 'success' | 'danger' | '' {
  const map: Record<string, 'info' | 'success' | 'danger' | ''> = {
    potential: 'info',
    active: 'success',
    lost: 'danger',
  }
  return map[status] ?? 'info'
}

function statusLabel(status: string): string {
  const map: Record<string, string> = {
    potential: '潜在客户',
    active: '活跃客户',
    lost: '流失客户',
  }
  return map[status] ?? status
}

function levelLabel(level?: string): string {
  const map: Record<string, string> = {
    regular: '普通会员',
    silver: '白银会员',
    gold: '黄金会员',
    platinum: '铂金会员',
    diamond: '钻石会员',
  }
  return level ? map[level] || level : '--'
}

function levelTagType(level?: string): 'info' | 'success' | 'warning' | 'danger' | '' {
  const map: Record<string, 'info' | 'success' | 'warning' | 'danger' | ''> = {
    regular: 'info',
    silver: '',
    gold: 'warning',
    platinum: '',
    diamond: 'danger',
  }
  return level ? map[level] || 'info' : 'info'
}

function sourceLabel(source: string): string {
  const map: Record<string, string> = {
    walk_in: '到店',
    referral: '转介绍',
    xiaohongshu: '小红书',
    wechat: '微信',
    douyin: '抖音',
    dianping: '大众点评',
    other: '其他',
  }
  return map[source] ?? source
}

// ── Create dialog ───────────────────────────────────────
const dialogVisible = ref(false)
const submitting = ref(false)
const formRef = ref<any>(null)

const form = ref({
  name: '',
  phone: '',
  gender: 0,
  birthday: '',
  source: '',
  remark: '',
})

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1\d{10}$/, message: '请输入有效的手机号', trigger: 'blur' },
  ],
}

function resetForm() {
  formRef.value?.resetFields()
  form.value = { name: '', phone: '', gender: 0, birthday: '', source: '', remark: '' }
}

function onIdCardInput() {
  const v = form.value.id_card
  if (v && v.length === 18) {
    form.value.birthday = v.substring(6, 10) + '-' + v.substring(10, 12) + '-' + v.substring(12, 14)
    const d = parseInt(v.charAt(16))
    form.value.gender = d % 2 === 1 ? 1 : 2
  }
}

async function submitForm() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  submitting.value = true
  try {
    await createCustomer(form.value)
    ElMessage.success('客户创建成功')
    dialogVisible.value = false
    await fetchData()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '创建失败')
  } finally {
    submitting.value = false
  }
}

// ── Delete ──────────────────────────────────────────────
async function handleDelete(row: any) {
  try {
    await ElMessageBox.confirm(`确定要删除客户「${row.name}」吗？`, '确认删除', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
    await deleteCustomer(row.id)
    ElMessage.success('删除成功')
    await fetchData()
  } catch {
    // cancelled or error
  }
}

onMounted(() => {
  fetchData()
})
</script>
