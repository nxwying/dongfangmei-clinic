<template>
  <div style="max-width:1000px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">📊 渠道 ROI 分析</span>
        <div style="flex:1" />
        <el-date-picker v-model="startDate" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" clearable style="width:150px" />
        <span style="color:#909399">—</span>
        <el-date-picker v-model="endDate" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" clearable style="width:150px" />
        <el-button type="primary" @click="fetchData">查询</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border :summary-method="summary" show-summary>
        <el-table-column label="渠道" prop="source" min-width="120" />
        <el-table-column label="客户数" prop="customer_count" width="100" align="center" />
        <el-table-column label="订单数" prop="order_count" width="100" align="center" />
        <el-table-column label="总收入" width="120" align="right">
          <template #default="{row}">¥{{ (row.total_revenue||0).toFixed(2) }}</template>
        </el-table-column>
        <el-table-column label="客单价" width="120" align="right">
          <template #default="{row}">¥{{ (row.avg_order_value||0).toFixed(2) }}</template>
        </el-table-column>
      </el-table>
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
    const res=await api.get('/reports/source-roi',{params})
    items.value=Array.isArray(res.data)?res.data:[]
  }finally{loading.value=false}
}
function summary({columns,data}:any){
  const sums:string[]=[]
  columns.forEach((col:any,i:number)=>{
    if(i===0){sums[i]='合计';return}
    if(col.property==='total_revenue'||col.property==='avg_order_value')
      sums[i]='¥'+data.reduce((s:number,r:any)=>s+Number(r[col.property]||0),0).toFixed(2)
    else if(col.property==='customer_count'||col.property==='order_count')
      sums[i]=data.reduce((s:number,r:any)=>s+Number(r[col.property]||0),0)
    else sums[i]=''
  })
  return sums
}
onMounted(fetchData)
</script>
