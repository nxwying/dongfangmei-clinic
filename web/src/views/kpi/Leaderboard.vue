<template>
  <div style="max-width:800px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px">
        <span style="font-weight:600;font-size:15px">🏆 业绩排行榜</span>
        <div style="flex:1"/>
        <el-date-picker v-model="ym" type="month" value-format="YYYY-MM" style="width:160px" @change="fetchData"/>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border>
        <el-table-column type="index" label="#" width="50" align="center"/>
        <el-table-column label="姓名" prop="real_name" min-width="80"/>
        <el-table-column label="总业绩" width="130" align="right"><template #default="{row}">¥{{ (row.revenue||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="订单数" prop="orders" width="70" align="center"/>
        <el-table-column label="回访完成" prop="followups" width="80" align="center"/>
        <el-table-column label="新增客户" prop="new_customers" width="80" align="center"/>
      </el-table>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
const loading=ref(false);const items=ref<any[]>([]);const ym=ref(new Date().toISOString().slice(0,7))
async function fetchData(){
  loading.value=true
  try{const r=await api.get('/kpi/leaderboard',{params:{year_month:ym.value}});items.value=Array.isArray(r.data)?r.data:[]}
  finally{loading.value=false}
}
onMounted(fetchData)
</script>
