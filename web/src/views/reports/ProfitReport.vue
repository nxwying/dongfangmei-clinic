<template>
  <div>
    <el-card shadow="never" style="margin-bottom: 16px">
      <div style="display: flex; align-items: center; gap: 12px">
        <el-date-picker v-model="queryDate" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 160px" @change="loadData" />
        <span style="color: #909399; font-size: 14px">日期：{{ queryDate || '今天' }}</span>
      </div>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="8">
        <el-card shadow="never">
          <div class="card-value" style="color: #409eff">¥{{ (data.total_revenue || 0).toFixed(2) }}</div>
          <div class="card-label">总收入（订单实收）</div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="never">
          <div class="card-value" style="color: #f56c6c">- ¥{{ (data.high_value || 0).toFixed(2) }}</div>
          <div class="card-label">高值药品耗材</div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="never">
          <div class="card-value" style="color: #f56c6c">- ¥{{ (data.high_value || 0).toFixed(2) }}</div>
          <div class="card-label">高值药品耗材</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="card-value" style="color: #e6a23c">- ¥{{ (data.general || 0).toFixed(2) }}</div>
          <div class="card-label">一般支出</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="card-value" :style="{ color: data.net_profit >= 0 ? '#67c23a' : '#f56c6c', fontSize: '32px' }">
            ¥{{ (data.net_profit || 0).toFixed(2) }}
          </div>
          <div class="card-label" style="font-weight: 600">净利润（毛收入 - 成本）</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 16px">
      <template #header><span>计算公式</span></template>
      <div style="font-size: 14px; line-height: 2">
        <div>总收入 <span style="color: #409eff">¥{{ (data.total_revenue || 0).toFixed(2) }}</span></div>
        <div style="margin-left: 16px">− 高值药品耗材 <span style="color: #f56c6c">¥{{ (data.high_value || 0).toFixed(2) }}</span></div>
        <div style="margin-left: 16px">− 一般支出 <span style="color: #e6a23c">¥{{ (data.general || 0).toFixed(2) }}</span></div>
        <div style="margin-left: 16px; font-weight: 600; font-size: 16px">= 净利润 ¥{{ (data.net_profit || 0).toFixed(2) }}</div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getProfitReport } from '../../api/expense'

const queryDate = ref('')
const data = ref<any>({})

async function loadData() {
  try {
    data.value = await getProfitReport({ date: queryDate.value || undefined })
  } catch {
    data.value = {}
  }
}

onMounted(loadData)
</script>

<style scoped>
.card-value { font-size: 28px; font-weight: bold; padding: 8px 0; }
.card-label { font-size: 14px; color: #909399; margin-top: 4px; }
</style>
