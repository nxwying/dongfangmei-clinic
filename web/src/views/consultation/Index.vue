<template>
  <div>
    <!-- Stats bar -->
    <el-row :gutter="12" style="margin-bottom:16px">
      <el-col :span="4" v-for="card in statCards" :key="card.key">
        <el-card shadow="never" style="text-align:center">
          <div :style="{color:card.color,fontSize:'24px',fontWeight:700}">{{ card.value }}</div>
          <div style="font-size:13px;color:#909399;margin-top:4px">{{ card.label }}</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Filter bar -->
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;gap:12px;flex-wrap:wrap;align-items:center">
        <el-select v-model="filter.intention_level" placeholder="意向等级" clearable style="width:120px" @change="loadList">
          <el-option label="A级 高意向" value="A" />
          <el-option label="B级 中意向" value="B" />
          <el-option label="C级 低意向" value="C" />
          <el-option label="D级 无意向" value="D" />
        </el-select>
        <el-select v-model="filter.status" placeholder="状态" clearable style="width:120px" @change="loadList">
          <el-option label="跟进中" value="pending" />
          <el-option label="已成交" value="won" />
          <el-option label="已流失" value="lost" />
        </el-select>
        <el-input v-model="filter.q" placeholder="搜索客户" clearable style="width:200px" @clear="loadList" @keyup.enter="loadList" />
        <el-button type="primary" @click="openCreate">新增跟单</el-button>
      </div>
    </el-card>

    <!-- List -->
    <el-card shadow="never">
      <el-table :data="list" stripe size="small" style="width:100%">
        <el-table-column label="客户" min-width="120">
          <template #default="{row}">
            <span style="font-weight:600">{{ row.customer?.name || '-' }}</span>
            <div style="font-size:12px;color:#909399">{{ row.customer?.phone }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="contact_date" label="沟通日期" width="110" />
        <el-table-column prop="contact_method" label="方式" width="80">
          <template #default="{row}">{{ methodLabel(row.contact_method) }}</template>
        </el-table-column>
        <el-table-column label="意向" width="70" align="center">
          <template #default="{row}">
            <el-tag :type="levelTag(row.intention_level)" size="small" effect="dark">{{ row.intention_level || 'C' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="沟通内容" min-width="200" show-overflow-tooltip />
        <el-table-column prop="interested_items" label="感兴趣项目" min-width="120" show-overflow-tooltip />
        <el-table-column label="预计金额" width="100" align="right">
          <template #default="{row}">¥{{ (row.estimated_amount||0).toFixed(0) }}</template>
        </el-table-column>
        <el-table-column prop="next_contact_date" label="下次跟进" width="110">
          <template #default="{row}">
            <span v-if="row.next_contact_date" :style="isOverdue(row.next_contact_date) ? 'color:#f56c6c;font-weight:600' : ''">
              {{ row.next_contact_date }}
            </span>
            <span v-else style="color:#c0c4cc">-</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80" align="center">
          <template #default="{row}">
            <el-tag :type="statusTag(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{row}">
            <el-button text size="small" @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status==='pending'" text size="small" type="success" @click="markWon(row)">成交</el-button>
            <el-button text size="small" type="danger" @click="removeItem(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top:16px;text-align:right">
        <el-pagination v-model:current-page="page" :page-size="20" :total="total" layout="total, prev, pager, next" @current-change="loadList" />
      </div>
    </el-card>

    <!-- Create/Edit dialog -->
    <el-dialog v-model="dialogVisible" :title="editing ? '编辑跟单' : '新增跟单'" width="680px" append-to-body>
      <el-form :model="form" label-width="100px" size="default">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="客户" required>
              <el-select v-model="form.customer_id" filterable remote :remote-method="searchCustomer" placeholder="搜索客户" style="width:100%">
                <el-option v-for="c in custOpts" :key="c.id" :label="c.name+' ('+c.phone+')'" :value="c.id" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="沟通日期">
              <el-date-picker v-model="form.contact_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item label="沟通方式">
              <el-select v-model="form.contact_method" style="width:100%">
                <el-option label="微信" value="wechat" />
                <el-option label="电话" value="phone" />
                <el-option label="到店" value="in-store" />
                <el-option label="视频" value="video" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="意向等级">
              <el-select v-model="form.intention_level" style="width:100%">
                <el-option label="A级 高意向" value="A" />
                <el-option label="B级 中意向" value="B" />
                <el-option label="C级 低意向" value="C" />
                <el-option label="D级 无意向" value="D" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="预计金额">
              <el-input-number v-model="form.estimated_amount" :min="0" :controls="false" style="width:100%">
                <template #prepend>¥</template>
              </el-input-number>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="感兴趣项目">
          <el-input v-model="form.interested_items" placeholder="如：热玛吉、玻尿酸隆鼻" />
        </el-form-item>
        <el-form-item label="沟通内容">
          <el-input v-model="form.content" type="textarea" :rows="3" placeholder="记录本次沟通的要点" />
        </el-form-item>
        <el-form-item label="客户顾虑">
          <el-input v-model="form.customer_concern" type="textarea" :rows="2" placeholder="价格、效果、安全等顾虑" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="下次跟进">
              <el-date-picker v-model="form.next_contact_date" type="date" value-format="YYYY-MM-DD" style="width:100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="下一步行动">
              <el-input v-model="form.next_action" placeholder="如：发案例图、邀约到店" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="save">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const list = ref([])
const total = ref(0)
const page = ref(1)
const loading = ref(false)
const dialogVisible = ref(false)
const editing = ref(false)
const saving = ref(false)
const custOpts = ref([])
const stats = ref({})

const filter = reactive({ intention_level: '', status: '', q: '' })
const form = reactive({
  id: null, customer_id: null, contact_date: '', contact_method: 'wechat',
  content: '', customer_concern: '', intention_level: 'C', interested_items: '',
  estimated_amount: 0, status: 'pending', next_contact_date: '', next_action: ''
})

const statCards = computed(() => [
  { key:'A', label:'A级意向', value: countByLevel('A'), color:'#f56c6c' },
  { key:'B', label:'B级意向', value: countByLevel('B'), color:'#e6a23c' },
  { key:'today', label:'今日跟进', value: stats.value.today_followup || 0, color:'#409eff' },
  { key:'overdue', label:'逾期未跟', value: stats.value.overdue || 0, color:'#f56c6c' },
  { key:'won', label:'已成交', value: stats.value.won || 0, color:'#67c23a' },
  { key:'rate', label:'成交率', value: (stats.value.win_rate || 0).toFixed(0) + '%', color:'#909399' },
])

function countByLevel(l) {
  const arr = stats.value.by_level || []
  const found = arr.find(s => s.level === l)
  return found ? found.count : 0
}

function methodLabel(m) {
  return { wechat:'微信', phone:'电话', 'in-store':'到店', video:'视频' }[m] || m || ''
}
function levelTag(l) {
  return { A:'danger', B:'warning', C:'info', D:'' }[l] || 'info'
}
function statusTag(s) {
  return { pending:'warning', won:'success', lost:'danger' }[s] || 'info'
}
function statusLabel(s) {
  return { pending:'跟进中', won:'已成交', lost:'已流失' }[s] || s
}
function isOverdue(d) {
  if (!d) return false
  return d < new Date().toISOString().substring(0,10)
}

async function loadList() {
  loading.value = true
  try {
    const params = { page: page.value, page_size: 20 }
    if (filter.intention_level) params.intention_level = filter.intention_level
    if (filter.status) params.status = filter.status
    const res = await api.get('/consultations', { params })
    list.value = res.data?.data || []
    total.value = res.data?.total || 0
  } catch(e) {}
  loading.value = false
}

async function loadStats() {
  try {
    const res = await api.get('/consultations/stats')
    stats.value = res.data || {}
  } catch(e) {}
}

async function searchCustomer(q) {
  try {
    const res = await api.get('/customers', { params: { q, page_size: 20 } })
    custOpts.value = res.data?.data || res.data || []
  } catch(e) {}
}

function openCreate() {
  editing.value = false
  Object.assign(form, {
    id: null, customer_id: null, contact_date: new Date().toISOString().substring(0,10),
    contact_method: 'wechat', content: '', customer_concern: '', intention_level: 'C',
    interested_items: '', estimated_amount: 0, status: 'pending', next_contact_date: '', next_action: ''
  })
  dialogVisible.value = true
}

function openEdit(row) {
  editing.value = true
  Object.assign(form, row)
  if (row.customer) {
    custOpts.value = [row.customer]
  }
  dialogVisible.value = true
}

async function save() {
  if (!form.customer_id) { ElMessage.warning('请选择客户'); return }
  saving.value = true
  try {
    if (editing.value) {
      await api.put(`/consultations/${form.id}`, form)
    } else {
      await api.post('/consultations', form)
    }
    ElMessage.success('保存成功')
    dialogVisible.value = false
    loadList()
    loadStats()
  } catch(e) {
    ElMessage.error('保存失败')
  }
  saving.value = false
}

async function markWon(row) {
  try {
    await ElMessageBox.confirm(`确认「${row.customer?.name}」已成交？`, '确认成交')
    await api.put(`/consultations/${row.id}`, { ...row, status: 'won' })
    ElMessage.success('已标记成交')
    loadList()
    loadStats()
  } catch(e) {}
}

async function removeItem(row) {
  try {
    await ElMessageBox.confirm('确认删除此跟单记录？', '确认')
    await api.delete(`/consultations/${row.id}`)
    ElMessage.success('已删除')
    loadList()
    loadStats()
  } catch(e) {}
}

onMounted(() => {
  loadList()
  loadStats()
})
</script>
