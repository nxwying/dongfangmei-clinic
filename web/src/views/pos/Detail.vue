<template>
  <div class="detail-page">
    <el-card shadow="never">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;">
          <span style="font-weight:600;">订单详情</span>
          <div>
            <el-button @click="$router.back()">返回</el-button>
            <el-button
              v-if="order && order.status !== 'refunded'"
              type="danger"
              :loading="refunding"
              @click="handleRefund"
            >
              退款
            </el-button>
          </div>
        </div>
      </template>

      <el-skeleton :loading="loading" animated :rows="6">
        <template v-if="order">
          <!-- Order header info -->
          <div style="margin-bottom:20px;">
            <el-descriptions :column="3" border size="small">
              <el-descriptions-item label="订单号">{{ order.order_no }}</el-descriptions-item>
              <el-descriptions-item label="客户">{{ order.customer_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="状态">
                <el-tag v-if="order.status === 'paid'" type="success" size="small">已支付</el-tag>
                <el-tag v-else-if="order.status === 'partial'" type="warning" size="small">部分支付</el-tag>
                <el-tag v-else-if="order.status === 'refunded'" type="info" size="small">已退款</el-tag>
                <el-tag v-else type="danger" size="small">待支付</el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="项目合计">¥{{ (order.items_total || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="优惠">¥{{ (order.discount_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="应付总额">¥{{ (order.total_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="已支付">¥{{ (order.paid_amount || 0).toFixed(2) }}</el-descriptions-item>
              <el-descriptions-item label="创建时间" :span="2">{{ order.created_at }}</el-descriptions-item>
            </el-descriptions>
          </div>

          <!-- Items table -->
          <div style="margin-bottom:20px;">
            <div style="font-weight:600;margin-bottom:8px;">项目明细</div>
            <el-table :data="order.items || []" stripe size="small" style="width:100%;">
              <el-table-column prop="item_name" label="项目名称" min-width="140" />
              <el-table-column label="单价" width="100">
                <template #default="{ row }">¥{{ (row.unit_price || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column prop="quantity" label="数量" width="70" />
              <el-table-column label="小计" width="100">
                <template #default="{ row }">
                  ¥{{ ((row.unit_price || 0) * (row.quantity || 1)).toFixed(2) }}
                </template>
              </el-table-column>
            </el-table>
          </div>

          <!-- Package deductions -->
          <div v-if="order.package_deductions && order.package_deductions.length" style="margin-bottom:20px;">
            <div style="font-weight:600;margin-bottom:8px;">套餐划扣</div>
            <el-table :data="order.package_deductions" stripe size="small" style="width:100%;">
              <el-table-column prop="package_name" label="套餐" min-width="140" />
              <el-table-column prop="item_name" label="项目" min-width="140" />
              <el-table-column label="划扣金额" width="100">
                <template #default="{ row }">¥{{ (row.deduct_amount || 0).toFixed(2) }}</template>
              </el-table-column>
            </el-table>
          </div>

          <!-- Payments table -->
          <div>
            <div style="font-weight:600;margin-bottom:8px;">支付明细</div>
            <el-table :data="order.payments || []" stripe size="small" style="width:100%;">
              <el-table-column label="支付方式" width="120">
                <template #default="{ row }">{{ payMethodLabel(row.pay_method) }}</template>
              </el-table-column>
              <el-table-column label="金额" width="100">
                <template #default="{ row }">¥{{ (row.amount || 0).toFixed(2) }}</template>
              </el-table-column>
              <el-table-column label="支付时间" width="160">
                <template #default="{ row }">{{ row.paid_at || row.created_at || '-' }}</template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </el-skeleton>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getOrder, refundOrder } from '../../api/order'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const refunding = ref(false)
const order = ref<any>(null)

const payMethodLabels: Record<string, string> = {
  balance: '余额',
  gift_balance: '赠送余额',
  wechat: '微信',
  alipay: '支付宝',
  cash: '现金',
  bank_card: '银行卡',
}

function payMethodLabel(method: string) {
  return payMethodLabels[method] || method || '-'
}

onMounted(async () => {
  const id = Number(route.params.id)
  if (!id) {
    ElMessage.error('无效订单ID')
    router.push('/pos')
    return
  }
  try {
    const res = await getOrder(id)
    order.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '加载订单失败')
  } finally {
    loading.value = false
  }
})

async function handleRefund() {
  try {
    await ElMessageBox.confirm('确定要退款吗？此操作不可撤销。', '退款确认', {
      type: 'warning',
      confirmButtonText: '确认退款',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }

  refunding.value = true
  try {
    await refundOrder(order.value.id)
    ElMessage.success('退款成功')
    // Reload
    const res = await getOrder(order.value.id)
    order.value = res?.data || res
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || '退款失败')
  } finally {
    refunding.value = false
  }
}
</script>

<style scoped>
.detail-page {
  max-width: 960px;
  margin: 0 auto;
}
</style>
