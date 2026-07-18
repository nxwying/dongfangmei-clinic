<template>
  <div style="max-width:1200px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:14px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <el-date-picker v-model="startDate" type="date" placeholder="开始日期" value-format="YYYY-MM-DD" style="width:150px" @change="loadData"/>
        <span style="color:#909399">至</span>
        <el-date-picker v-model="endDate" type="date" placeholder="结束日期" value-format="YYYY-MM-DD" style="width:150px" @change="loadData"/>
        <el-radio-group v-model="statusFilter" @change="loadData" size="small">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="booked">已预约</el-radio-button>
          <el-radio-button value="checked_in">已到店</el-radio-button>
          <el-radio-button value="completed">已完成</el-radio-button>
          <el-radio-button value="cancelled">已取消</el-radio-button>
        </el-radio-group>
        <span style="flex:1"/>
        <el-button size="small" @click="viewTab=viewTab==='calendar'?'list':'calendar'">{{ viewTab==='calendar'?'📋 列表':'📅 周视图' }}</el-button>
        <el-button size="small" text @click="printSchedule">🖨️ 打印</el-button>
        <el-button type="primary" size="small" @click="openCreate">+ 新建预约</el-button>
      </div>
    </el-card>

    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399">总预约</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.checkedIn }}</div><div style="font-size:12px;color:#909399">已到店</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.completed }}</div><div style="font-size:12px;color:#909399">已完成</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.cancelled }}</div><div style="font-size:12px;color:#909399">已取消</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700" :style="{color:stats.rate>=80?'#67c23a':stats.rate>=60?'#e6a23c':'#f56c6c'}">{{ stats.rate }}%</div><div style="font-size:12px;color:#909399">到店率</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never">
      <el-table :data="appointments" v-loading="loading" stripe size="small" empty-text="暂无预约记录">
        <el-table-column label="日期" width="100"><template #default="{row}">{{ formatDate(row.date)||row.appointment_date||'—' }}</template></el-table-column>
        <el-table-column label="时间" width="80"><template #default="{row}">{{ row.time_slot||'—' }}</template></el-table-column>
        <el-table-column label="结束" width="70"><template #default="{row}">{{ calcEndTimeStr(row.time_slot,row.duration) }}</template></el-table-column>
        <el-table-column label="客户" min-width="100"><template #default="{row}">{{ row.customer?.name||row.customer_name||'—' }}</template></el-table-column>
        <el-table-column label="电话" width="110"><template #default="{row}">{{ row.customer?.phone||'—' }}</template></el-table-column>
        <el-table-column label="项目" min-width="120"><template #default="{row}">{{ row.items||row.item_name||'—' }}</template></el-table-column>
        <el-table-column label="咨询师" width="80"><template #default="{row}">{{ row.consultant_name||'—' }}</template></el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="{row}">
            <el-tag :type="row.status==='booked'?'warning':row.status==='checked_in'?'primary':row.status==='completed'?'success':'info'" size="small">
              {{ row.status==='booked'?'已预约':row.status==='checked_in'?'已到店':row.status==='completed'?'已完成':'已取消' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{row}">
            <el-button v-if="row.status==='booked'" size="small" type="warning" @click="handleCheckIn(row)">到店</el-button>
            <el-button v-if="row.status==='checked_in'" size="small" type="success" @click="handleComplete(row)">完成</el-button>
            <el-button v-if="row.status==='booked'||row.status==='checked_in'" size="small" text @click="openEdit(row)">编辑</el-button>
            <el-button v-if="row.status==='booked'||row.status==='checked_in'" size="small" text type="danger" @click="handleCancel(row)">取消</el-button>
            <span v-else style="color:#c0c4cc;font-size:12px">—</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑预约':'新建预约'" width="520px" destroy-on-close>
      <el-form ref="formRef" :model="form" label-width="80px" :rules="rules">
        <el-form-item label="客户" prop="customer_id">
          <el-select v-model="form.customer_id" filterable clearable placeholder="请选择客户" style="width:100%" :loading="custLoading">
            <el-option v-for="c in custList" :key="c.id" :label="`${c.name} (${c.phone||''})`" :value="c.id"/>
          </el-select>
        </el-form-item>
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="日期" prop="appointment_date">
            <el-date-picker v-model="form.appointment_date" type="date" value-format="YYYY-MM-DD" style="width:100%"/>
          </el-form-item></el-col>
          <el-col :span="8"><el-form-item label="时间" prop="time_slot">
            <el-select v-model="form.time_slot" placeholder="选择时间" style="width:100%" @change="calcEndTime">
              <el-option v-for="t in timeSlots" :key="t" :label="t" :value="t"/>
            </el-select>
          </el-form-item></el-col>
          <el-col :span="4"><el-form-item label="时长(分)">
            <el-select v-model="form.duration" style="width:100%" @change="calcEndTime">
              <el-option label="15分" :value="15"/>
              <el-option label="30分" :value="30"/>
              <el-option label="45分" :value="45"/>
              <el-option label="60分" :value="60"/>
              <el-option label="90分" :value="90"/>
              <el-option label="120分" :value="120"/>
            </el-select>
          </el-form-item></el-col>
          <el-col :span="4"><el-form-item label="结束时间">
            <el-input :model-value="formEndTime" disabled style="width:100%"/>
          </el-form-item></el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="咨询师">
            <el-select v-model="form.consultant_id" placeholder="选择" filterable style="width:100%">
              <el-option v-for="s in staffList" :key="s.id" :label="s.real_name" :value="s.id"/>
            </el-select>
          </el-form-item></el-col>
          <el-col :span="12"><el-form-item label="医生">
            <el-select v-model="form.doctor_id" placeholder="选择" filterable style="width:100%">
              <el-option v-for="s in docList" :key="s.id" :label="s.real_name" :value="s.id"/>
            </el-select>
          </el-form-item></el-col>
        </el-row>
        <el-form-item label="项目"><el-input v-model="form.items" placeholder="如：水光针、光子嫩肤"/></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2"/></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const viewTab = ref('list')
const calendarWeekOffset = ref(0)
const today = () => new Date().toISOString().slice(0,10)
const monthStart = () => {const d=new Date();d.setDate(1);return d.toISOString().slice(0,10)}
const monthEnd = () => {const d=new Date();d.setMonth(d.getMonth()+1);d.setDate(0);return d.toISOString().slice(0,10)}

const startDate = ref(monthStart())
const endDate = ref(monthEnd())
const statusFilter = ref('')
const loading = ref(false)
const appointments = ref<any[]>([])
const timeSlots = ['09:00','09:30','10:00','10:30','11:00','11:30','13:00','13:30','14:00','14:30','15:00','15:30','16:00','16:30','17:00','17:30']

const stats = computed(() => {
  const all = appointments.value
  const total = all.length
  const checkedIn = all.filter((a:any) => a.status === 'checked_in').length
  const completed = all.filter((a:any) => a.status === 'completed').length
  const cancelled = all.filter((a:any) => a.status === 'cancelled').length
  const arrived = checkedIn + completed
  return { total, checkedIn, completed, cancelled, rate: total > 0 ? Math.round(arrived / total * 100) : 0 }
})

async function loadData() {
  loading.value = true
  try {
    const params: any = { start_date: startDate.value, end_date: endDate.value }
    if (statusFilter.value) params.status = statusFilter.value
    const r = await api.get('/appointments', { params })
    appointments.value = Array.isArray(r.data) ? r.data : []
  } catch (e) { appointments.value = []; console.error('加载预约失败', e) }
  finally { loading.value = false }
}

// Customer data
const custList = ref<any[]>([])
const custLoading = ref(false)
async function loadCustomers() {
  custLoading.value = true
  try {
    const r = await api.get('/customers', { params: { page_size: 200 } })
    custList.value = r.data?.data ?? (Array.isArray(r.data) ? r.data : [])
  } catch { custList.value = [] }
  finally { custLoading.value = false }
}

// Staff data
const staffList = ref<any[]>([])
const docList = ref<any[]>([])
async function loadStaff() {
  try {
    const r = await api.get('/users')
    const users = Array.isArray(r.data) ? r.data : []
    staffList.value = users.filter((u:any) => u.status === 'active')
    docList.value = users.filter((u:any) => u.status === 'active' && u.role?.name === 'doctor')
  } catch {}
}

// Dialog
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)

const form = ref({
  customer_id: null, appointment_date: today(), time_slot: '', duration: 30,
  consultant_id: null, doctor_id: null, items: '', remark: ''
})

const rules: Record<string, any> = {
  customer_id: [{ required: true, message: '请选择客户', trigger: 'change' }],
  appointment_date: [{ required: true, message: '请选择日期', trigger: 'change' }],
  time_slot: [{ required: true, message: '请选择时间', trigger: 'change' }],
}

function resetForm() {
  form.value = { customer_id: null, appointment_date: today(), time_slot: '', duration: 30, consultant_id: null, doctor_id: null, items: '', remark: '' }
}

async function openCreate() {
  isEdit.value = false; editingId.value = null
  resetForm()
  dialogVisible.value = true
  await Promise.all([loadCustomers(), loadStaff()])
}

async function openEdit(row: any) {
  isEdit.value = true; editingId.value = row.id
  form.value = {
    customer_id: row.customer_id,
    appointment_date: formatDate(row.date) || row.appointment_date || today(),
    time_slot: row.time_slot || '',
    duration: row.duration || 30,
    consultant_id: row.consultant_id || null,
    doctor_id: row.doctor_id || null,
    items: row.items || '',
    remark: row.remark || '',
  }
  dialogVisible.value = true
  await Promise.all([loadCustomers(), loadStaff()])
}

async function submit() {
  saving.value = true
  try {
    const data = {
      customer_id: form.value.customer_id,
      date: form.value.appointment_date,
      duration: form.value.duration,
      time_slot: form.value.time_slot,
      consultant_id: form.value.consultant_id || 0,
      doctor_id: form.value.doctor_id || 0,
      items: form.value.items,
      remark: form.value.remark,
    }
    if (isEdit.value && editingId.value) {
      await api.put(`/appointments/${editingId.value}`, data)
      ElMessage.success('已更新')
    } else {
      await api.post('/appointments', data)
      ElMessage.success('已创建')
    }
    dialogVisible.value = false
    // Update date range to include the new appointment
    const newDate = form.value.appointment_date
    if (newDate && newDate < startDate.value) startDate.value = newDate
    if (newDate && newDate > endDate.value) endDate.value = newDate
    await loadData()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '操作失败')
  } finally { saving.value = false }
}

async function handleCheckIn(row: any) {
  try {
    await api.put(`/appointments/${row.id}/checkin`)
    ElMessage.success('已确认到店')
    row.status = 'checked_in'
    await loadData()
  } catch { ElMessage.error('操作失败') }
}

async function handleComplete(row: any) {
  try {
    await api.put(`/appointments/${row.id}/complete`)
    ElMessage.success('已完成')
    row.status = 'completed'
    await loadData()
  } catch { ElMessage.error('操作失败') }
}

async function handleCancel(row: any) {
  try {
    await ElMessageBox.confirm('确定取消此预约？', '确认', { type: 'warning' })
    await api.put(`/appointments/${row.id}/cancel`)
    ElMessage.success('已取消')
    row.status = 'cancelled'
    await loadData()
  } catch { }
}

/** Extract YYYY-MM-DD from a date string (handles RFC3339, ISO, and YYYY-MM-DD formats) */
function formatDate(d: any): string {
  if (!d) return ''
  if (typeof d !== 'string') return String(d).slice(0,10)
  const m = d.match(/^(\d{4}-\d{2}-\d{2})/)
  return m ? m[1] : d.slice(0,10)
}

const formEndTime = computed(() => {
  if (!form.value.time_slot || !form.value.duration) return '--'
  const [h, m] = form.value.time_slot.split(':').map(Number)
  const total = h * 60 + m + form.value.duration
  const eh = Math.floor(total / 60); const em = total % 60
  return String(eh).padStart(2,'0') + ':' + String(em).padStart(2,'0')
})

function calcEndTime() { /* triggers formEndTime recompute */ }

function calcEndTimeStr(slot: string, dur: number): string {
  if (!slot || !dur) return '--'
  const [h, m] = slot.split(':').map(Number)
  const total = h * 60 + m + dur
  const eh = Math.floor(total / 60); const em = total % 60
  return String(eh).padStart(2,'0') + ':' + String(em).padStart(2,'0')
}

function goToEdit(row: any) { openEdit(row) }

async function loadAllAppts() {
  try {
    const r = await api.get('/appointments')
    allAppointments.value = Array.isArray(r.data) ? r.data : []
  } catch { }
}

const calendarDays = computed(() => {
  const days: any[] = []
  const weekdays = ['周日','周一','周二','周三','周四','周五','周六']
  const now = new Date()
  const weekStart = new Date(now)
  weekStart.setDate(weekStart.getDate() - weekStart.getDay() + 1 + calendarWeekOffset.value * 7)
  for (let i = 0; i < 7; i++) {
    const d = new Date(weekStart)
    d.setDate(d.getDate() + i)
    const dateStr = d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0')
    days.push({ date: dateStr, weekday: weekdays[d.getDay()], day: String(d.getDate()).padStart(2,'0'), isToday: dateStr === today() })
  }
  return days
})

const weekLabel = computed(() => {
  if (!calendarDays.value.length) return ''
  return calendarDays.value[0].date + ' ~ ' + calendarDays.value[6].date
})

function getApptsForSlot(date: string, slot: string): any[] {
  return allAppointments.value.filter((a: any) => a.date?.startsWith(date||'') && a.time_slot === slot)
}

const allAppointments = ref<any[]>([])

function printSchedule() {
  window.print()
}

onMounted(() => { loadData(); loadCustomers(); loadStaff(); loadAllAppts() })
</script>
