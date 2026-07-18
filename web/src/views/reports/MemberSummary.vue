<template>
  <div>
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-card">
            <div class="stat-value">{{ data.total_members ?? '-' }}</div>
            <div class="stat-label">会员总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-card">
            <div class="stat-value">{{ data.month_new ?? '-' }}</div>
            <div class="stat-label">本月新增</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-card">
            <div class="stat-value">¥{{ (data.total_topup ?? 0).toLocaleString() }}</div>
            <div class="stat-label">累计充值</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="never">
          <div class="stat-card">
            <div class="stat-value">¥{{ (data.total_spent ?? 0).toLocaleString() }}</div>
            <div class="stat-label">累计消费</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" style="margin-top: 20px;" v-loading="loading">
      <template #header>
        <span>会员概览</span>
      </template>
      <el-table :data="memberRows" border stripe>
        <el-table-column prop="level_name" label="等级" min-width="120" />
        <el-table-column prop="count" label="人数" min-width="100" />
        <el-table-column prop="total_topup" label="累计充值" min-width="140">
          <template #default="{ row }">
            ¥{{ (row.total_topup ?? 0).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="total_spent" label="累计消费" min-width="140">
          <template #default="{ row }">
            ¥{{ (row.total_spent ?? 0).toLocaleString() }}
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!loading && memberRows.length === 0" description="暂无数据" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getMemberSummary } from '../../api/report'

const loading = ref(false)
const data = ref<any>({})
const memberRows = ref<any[]>([])

onMounted(async () => {
  loading.value = true
  try {
    const res = await getMemberSummary()
    data.value = res
    memberRows.value = res.levels ?? []
  } catch {
    // keep defaults
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.stat-card {
  text-align: center;
  padding: 10px 0;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}
.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
</style>
