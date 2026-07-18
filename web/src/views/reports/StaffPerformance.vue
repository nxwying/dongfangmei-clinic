<template>
  <div style="max-width:800px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">👤 咨询师/医生效能看板</span>
        <div style="flex:1" />
        <el-date-picker v-model="startDate" type="date" value-format="YYYY-MM-DD" placeholder="开始" clearable style="width:130px" />
        <span style="color:#909399">—</span>
        <el-date-picker v-model="endDate" type="date" value-format="YYYY-MM-DD" placeholder="结束" clearable style="width:130px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border>
        <el-table-column label="姓名" prop="real_name" min-width="120" />
        <el-table-column label="订单数" prop="orders" width="100" align="center" />
        <el-table-column label="业绩(毛收入)" width="150" align="right">
         <template #default="{row}">¥{{ (row.gross||0).toFixed(2) }}</template>
       </el-table-column>
       <el-table-column label="平均客单价" width="130" align="right">
          <template #default="{row}">¥{{ row.orders ? ((row.gross||0)/row.orders).toFixed(2) : '0.00' }}</template>
       </el-table-column>
      </el-table>
      <div v-if="!loading && items.length===0" style="text-align:center;padding:40px;color:#909399">暂无数据</div>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
const loading=ref(false);const items=ref<any[]>([])
const startDate=ref('');const endDate=ref('')
async function fetchData(){
  loading.value=true
  try{
    const params:any={}
    if(startDate.value)params.start_date=startDate.value
    if(endDate.value)params.end_date=endDate.value
    const res=await api.get('/reports/insights',{params})
    items.value=res.data?.staff_perf||[]
  }finally{loading.value=false}
}
onMounted(fetchData)
</script>
