<template>
  <div>
    <el-card shadow="never">
      <template #header>
        <div style="display: flex; align-items: center; gap: 12px;">
          <span>每日销售报表</span>
          <el-date-picker
            v-model="date"
            type="date"
            placeholder="选择日期"
            value-format="YYYY-MM-DD"
            :clearable="false"
            style="width: 160px;"
            @change="fetchData"
          />
        </div>
      </template>
      <el-table :data="rows" border stripe v-loading="loading">
        <el-table-column prop="order_count" label="订单数" min-width="120" />
        <el-table-column prop="total_revenue" label="总收入" min-width="140">
          <template #default="{ row }">
            ¥{{ row.total_revenue.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="total_discount" label="总优惠" min-width="140">
          <template #default="{ row }">
            ¥{{ row.total_discount.toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="total_topup" label="充值总额" min-width="140">
          <template #default="{ row }">
            ¥{{ row.total_topup.toLocaleString() }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && rows.length === 0" description="暂无数据" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDailySales } from '../../api/report'

const date = ref(new Date().toISOString().slice(0, 10))
const loading = ref(false)
const rows = ref<any[]>([])

async function fetchData() {
  loading.value = true
  try {
    const res = await getDailySales(date.value)
    rows.value = Array.isArray(res) ? res : [res]
  } catch {
    rows.value = []
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
