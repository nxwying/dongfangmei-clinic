<template>
  <div style="max-width:1100px;margin:0 auto">
    <!-- Search bar -->
    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">↩️ 退款管理</span>
        <el-date-picker v-model="startDate" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" style="width:140px" size="small" @change="loadRefunds"/>
        <span style="color:#909399">~</span>
        <el-date-picker v-model="endDate" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" style="width:140px" size="small" @change="loadRefunds"/>
        <el-input v-model="searchQ" placeholder="搜索客户姓名或单号" clearable style="width:180px" size="small" @keyup.enter="loadRefunds"/>
        <el-button size="small" @click="loadRefunds" :loading="loading">搜索</el-button>
      </div>
    </el-card>

    <!-- Stats cards -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:8px"><div style="font-size:24px;font-weight:700;color:#409eff">{{ pendingCount }}</div><div style="font-size:12px;color:#909399;margin-top:2px">可退款订单</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:8px"><div style="font-size:20px;font-weight:700;color:#409eff">¥{{ pendingAmount.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">可退款金额</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:8px"><div style="font-size:24px;font-weight:700;color:#f56c6c">{{ refundedCount }}</div><div style="font-size:12px;color:#909399;margin-top:2px">已退款订单</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:8px"><div style="font-size:20px;font-weight:700;color:#f56c6c">¥{{ refundedAmount.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">退款总金额</div></div></el-card></el-col>
    </el-row>

    <!-- Tabs: refundable / refunded -->
    <el-card shadow="never">
      <el-tabs v-model="tab">
        <el-tab-pane label="✅ 可退款订单" name="pending">
          <el-table :data="paidOrders" v-loading="loading" stripe border size="small" empty-text="暂无已付款的订单">
            <el-table-column label="单号" prop="order_no" width="150"/>
            <el-table-column label="客户" min-width="80"><template #default="{row:r}">{{ r.customer?.name||'--' }}</template></el-table-column>
            <el-table-column label="电话" width="110"><template #default="{row:r}">{{ r.customer?.phone||'--' }}</template></el-table-column>
            <el-table-column label="项目" min-width="100"><template #default="{row:r}"><span v-if="r.items?.length">{{ r.items.map((i:any)=>i.item_name).join('、') }}</span><span v-else>--</span></template></el-table-column>
            <el-table-column label="金额" width="80" align="right"><template #default="{row:r}">¥{{ (r.final_amount||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="时间" width="140"><template #default="{row:r}">{{ r.created_at }}</template></el-table-column>
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="{row:r}">
                <el-button type="danger" size="small" :loading="refundingId===r.id" @click="doRefund(r)">退款</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="ℹ️ 已退款记录" name="refunded">
          <el-table :data="refundedOrders" v-loading="loading" stripe border size="small" empty-text="暂无退款记录">
            <el-table-column label="单号" prop="order_no" width="150"/>
            <el-table-column label="客户" min-width="80"><template #default="{row:r}">{{ r.customer?.name||'--' }}</template></el-table-column>
            <el-table-column label="电话" width="110"><template #default="{row:r}">{{ r.customer?.phone||'--' }}</template></el-table-column>
            <el-table-column label="项目" min-width="100"><template #default="{row:r}"><span v-if="r.items?.length">{{ r.items.map((i:any)=>i.item_name).join('、') }}</span><span v-else>--</span></template></el-table-column>
            <el-table-column label="金额" width="80" align="right"><template #default="{row:r}">¥{{ (r.final_amount||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="退款时间" width="140"><template #default="{row:r}">{{ r.created_at }}</template></el-table-column>
            <el-table-column label="状态" width="70"><template #default="{row:r}"><el-tag type="info" size="small">已退款</el-tag></template></el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const loading = ref(false)
const orders = ref<any[]>([])
const tab = ref('pending')
const refundingId = ref<number | null>(null)

const startDate = ref('')
const endDate = ref('')
const searchQ = ref('')

onMounted(() => {
  const d = new Date()
  startDate.value = d.getFullYear() + '-' + String(d.getMonth()+1).padStart(2,'0') + '-01'
  endDate.value = d.getFullYear() + '-' + String(d.getMonth()+1).padStart(2,'0') + '-' + String(d.getDate()).padStart(2,'0')
  loadRefunds()
})

const paidOrders = computed(() => orders.value.filter((o:any) => o.status === 'paid'))
const refundedOrders = computed(() => orders.value.filter((o:any) => o.status === 'refunded'))
const pendingCount = computed(() => paidOrders.value.length)
const pendingAmount = computed(() => paidOrders.value.reduce((s:number,o:any) => s+(o.final_amount||0), 0))
const refundedCount = computed(() => refundedOrders.value.length)
const refundedAmount = computed(() => refundedOrders.value.reduce((s:number,o:any) => s+(o.final_amount||0), 0))

async function loadRefunds() {
  loading.value = true
  try {
    const params: any = {
      start_date: startDate.value,
      end_date: endDate.value,
      page_size: 200
    }
    if (searchQ.value) params.q = searchQ.value
    const r = await api.get('/orders', { params })
    const data = r.data
    orders.value = Array.isArray(data) ? data : data?.data || data?.list || []
  } catch { orders.value = [] }
  finally { loading.value = false }
}

async function doRefund(order: any) {
  if (order.status !== 'paid') { ElMessage.warning('只有已付款订单才能退款'); return }
  try {
    await ElMessageBox.confirm(
      '确定退款订单 ' + order.order_no + '?\n客户: ' + (order.customer?.name||'') + '\n金额: ¥' + (order.final_amount||0).toFixed(2),
      '退款确认', { confirmButtonText: '确认退款', cancelButtonText: '取消', type: 'warning' }
    )
    refundingId.value = order.id
    await api.post('/orders/' + order.id + '/refund')
    ElMessage.success('退款成功')
    order.status = 'refunded'
    await loadRefunds()
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e?.response?.data?.error || '退款失败')
  } finally { refundingId.value = null }
}
</script>
