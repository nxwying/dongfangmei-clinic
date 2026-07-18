<template>
  <div>
    <!-- 1. Monthly Trend -->
    <el-card shadow="never" style="margin-bottom:16px">
      <template #header><span style="font-weight:600">月度经营趋势</span></template>
      <div v-if="data.monthly_trend?.length" class="chart-container">
        <div v-for="m in data.monthly_trend" :key="m.month" class="bar-col">
          <div class="bar-stack">
            <div class="bar bar-revenue" :style="{height: barHeight(m.revenue)+'%', background:'#409eff'}" :title="'收入 ¥'+m.revenue.toFixed(0)"></div>
            <div class="bar bar-cost" :style="{height: barHeight(m.high_value+m.general+m.commission)+'%', background:'#f56c6c'}" :title="'支出 ¥'+(m.high_value+m.general+m.commission).toFixed(0)"></div>
          </div>
          <div class="bar-label">{{ m.month.slice(5) }}</div>
        </div>
      </div>
      <div v-else style="color:#909399;padding:20px;text-align:center">暂无数据</div>
    </el-card>

    <el-row :gutter="16">
      <!-- 2. Top Items -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">项目销售排行</span></template>
          <el-table :data="data.top_items" size="small" stripe v-if="data.top_items?.length">
            <el-table-column prop="item_name" label="项目" min-width="120" />
            <el-table-column prop="count" label="销量" width="60" align="center" />
            <el-table-column label="金额" width="100" align="right">
              <template #default="{row}">¥{{ row.total.toFixed(0) }}</template>
            </el-table-column>
          </el-table>
          <div v-else style="color:#909399;padding:20px;text-align:center">暂无数据</div>
        </el-card>
      </el-col>

      <!-- 3. Customer Sources -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">客户来源分析</span></template>
          <div v-if="data.customer_sources?.length" class="source-chart">
            <div v-for="s in data.customer_sources" :key="s.source" class="source-row">
              <div class="source-label">{{ sourceLabel(s.source) }}</div>
              <div class="source-bar-wrap">
                <div class="source-bar" :style="{width: sourcePct(s.count)+'%', background: sourceColor(s.source)}"></div>
              </div>
              <div class="source-count">{{ s.count }}人</div>
            </div>
          </div>
          <div v-else style="color:#909399;padding:20px;text-align:center">暂无客户数据</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- 4. Staff Performance -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">员工业绩</span></template>
          <el-table :data="data.staff_perf" size="small" stripe v-if="data.staff_perf?.length">
            <el-table-column prop="real_name" label="员工" width="100" />
            <el-table-column prop="orders" label="订单数" width="70" align="center" />
            <el-table-column label="业绩" width="120" align="right">
              <template #default="{row}">¥{{ row.total.toFixed(0) }}</template>
            </el-table-column>
          </el-table>
          <div v-else style="color:#909399;padding:20px;text-align:center">暂无数据</div>
        </el-card>
      </el-col>

      <!-- 5. Package Stats -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">套餐使用统计</span></template>
          <div v-if="data.package_stats?.length" class="pkg-grid">
            <div v-for="s in data.package_stats" :key="s.status" class="pkg-card">
              <div class="pkg-count">{{ s.count }}</div>
              <div class="pkg-label">{{ pkgLabel(s.status) }}</div>
            </div>
          </div>
          <div v-else style="color:#909399;padding:20px;text-align:center">暂无数据</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <!-- 6. Average Order -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">客单价统计</span></template>
          <div style="display:flex;gap:24px;padding:12px 0">
            <div><div class="stat-val">¥{{ (data.avg_order?.avg_order || 0).toFixed(0) }}</div><div class="stat-lbl">平均客单价</div></div>
            <div><div class="stat-val">{{ data.avg_order?.total_orders || 0 }}</div><div class="stat-lbl">总订单数</div></div>
            <div><div class="stat-val">¥{{ (data.avg_order?.total_revenue || 0).toFixed(0) }}</div><div class="stat-lbl">总收入</div></div>
          </div>
        </el-card>
      </el-col>

      <!-- 7. Appointment Funnel -->
      <el-col :span="12">
        <el-card shadow="never" style="margin-bottom:16px">
          <template #header><span style="font-weight:600">预约到店率</span></template>
          <div v-if="data.appointment_funnel?.length">
            <div v-for="s in data.appointment_funnel" :key="s.status" class="funnel-row">
              <el-tag :type="apptTagType(s.status)" size="small" style="width:70px">{{ apptLabel(s.status) }}</el-tag>
              <div class="funnel-bar-wrap">
                <div class="funnel-bar" :style="{width: apptPct(s.count)+'%'}"></div>
              </div>
              <span style="margin-left:8px;font-size:13px">{{ s.count }}单</span>
            </div>
          </div>
          <div v-else style="color:#909399;padding:20px;text-align:center">暂无预约数据</div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '../../api'

const data = ref<any>({})

const sourceColors: Record<string, string> = {
  walk_in:'#409eff', referral:'#67c23a', xiaohongshu:'#f56c6c',
  wechat:'#e6a23c', douyin:'#9b59b6', dianping:'#1abc9c', other:'#95a5a6'
}
const sourceLabelsMap: Record<string, string> = {
  walk_in:'到店', referral:'转介绍', xiaohongshu:'小红书',
  wechat:'微信', douyin:'抖音', dianping:'大众点评', other:'其他'
}
function sourceLabel(s: string) { return sourceLabelsMap[s] || s }
function sourceColor(s: string) { return sourceColors[s] || '#95a5a6' }

const maxRevenue = computed(() => Math.max(...(data.value.monthly_trend || []).map((m: any) => m.revenue || 0), 1))
function barHeight(v: number) { return Math.max((v / maxRevenue.value) * 100, 0) }

function pkgLabel(s: string) {
  const m: Record<string, string> = { active:'使用中', completed:'已完成', expired:'已过期', refunded:'已退款' }
  return m[s] || s
}

const maxSource = computed(() => Math.max(...(data.value.customer_sources || []).map((s: any) => s.count || 0), 1))
function sourcePct(c: number) { return (c / maxSource.value) * 100 }

const maxFunnel = computed(() => Math.max(...(data.value.appointment_funnel || []).map((s: any) => s.count || 0), 1))
function apptPct(c: number) { return (c / maxFunnel.value) * 100 }

function apptLabel(s: string) {
  const m: Record<string, string> = { booked:'已预约', checked_in:'已到店', completed:'已完成', cancelled:'已取消' }
  return m[s] || s
}
function apptTagType(s: string) {
  const m: Record<string, string> = { booked:'info', checked_in:'warning', completed:'success', cancelled:'danger' }
  return m[s] || 'info'
}

async function loadData() {
  try { data.value = await (await api.get('/reports/analytics')).data } catch {}
}
onMounted(loadData)
</script>

<style scoped>
.chart-container { display:flex; align-items:flex-end; gap:6px; height:200px; padding:10px 0 }
.bar-col { flex:1; display:flex; flex-direction:column; align-items:center; height:100% }
.bar-stack { flex:1; width:100%; display:flex; flex-direction:column-reverse; align-items:center; gap:2px }
.bar { width:60%; min-height:2px; border-radius:3px 3px 0 0; transition:height 0.3s }
.bar-label { font-size:11px; color:#909399; margin-top:4px }
.source-chart { padding:4px 0 }
.source-row { display:flex; align-items:center; gap:8px; margin-bottom:8px }
.source-label { width:70px; font-size:13px; text-align:right }
.source-bar-wrap { flex:1; height:20px; background:#f0f2f5; border-radius:4px }
.source-bar { height:100%; border-radius:4px; transition:width 0.3s; min-width:4px }
.source-count { width:40px; font-size:13px; color:#606266 }
.pkg-grid { display:flex; gap:12px; padding:8px 0 }
.pkg-card { flex:1; text-align:center; padding:16px; background:#f0f2f5; border-radius:6px }
.pkg-count { font-size:28px; font-weight:700; color:#409eff }
.pkg-label { font-size:13px; color:#909399; margin-top:4px }
.funnel-row { display:flex; align-items:center; gap:8px; margin-bottom:10px }
.funnel-bar-wrap { flex:1; height:22px; background:#f0f2f5; border-radius:4px }
.funnel-bar { height:100%; background:#409eff; border-radius:4px; transition:width 0.3s; min-width:4px }
.stat-val { font-size:22px; font-weight:700; color:#303133 }
.stat-lbl { font-size:13px; color:#909399; margin-top:2px }
</style>
