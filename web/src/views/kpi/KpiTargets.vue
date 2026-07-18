<template>
  <div style="max-width:800px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px">
        <span style="font-weight:600;font-size:15px">🎯 KPI 目标设定</span>
        <div style="flex:1"/>
        <el-date-picker v-model="ym" type="month" value-format="YYYY-MM" style="width:160px" @change="fetchData"/>
        <el-button @click="go('/kpi/leaderboard')">排行榜 →</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="users" v-loading="loading" stripe border>
        <el-table-column label="姓名" prop="real_name" min-width="80"/>
        <el-table-column label="业绩目标"><template #default="{row}"><el-input-number v-model="row.revenue_target" :min="0" size="small" style="width:120px"/></template></el-table-column>
        <el-table-column label="订单目标"><template #default="{row}"><el-input-number v-model="row.order_target" :min="0" size="small" style="width:80px"/></template></el-table-column>
        <el-table-column label="回访目标"><template #default="{row}"><el-input-number v-model="row.followup_target" :min="0" size="small" style="width:80px"/></template></el-table-column>
        <el-table-column label="新客目标"><template #default="{row}"><el-input-number v-model="row.new_customer_target" :min="0" size="small" style="width:80px"/></template></el-table-column>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button size="small" @click="saveRow(row)">保存</el-button></template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import api from '../../api'
const router=useRouter();const loading=ref(false);const ym=ref(new Date().toISOString().slice(0,7))
const users=ref<any[]>([])
async function fetchData(){
  loading.value=true
  try{
    const r=await api.get('/kpi/targets',{params:{year_month:ym.value}})
    const targets=Array.isArray(r.data)?r.data:[]
    const u=await api.get('/users');const allUsers=Array.isArray(u.data)?u.data:[]
    users.value=allUsers.filter((u:any)=>u.role_id!==1).map((u:any)=>{
      const t=targets.find((x:any)=>x.user_id===u.id)
      return{user_id:u.id,real_name:u.real_name,year_month:ym.value,
        revenue_target:t?.revenue_target||0,order_target:t?.order_target||0,
        followup_target:t?.followup_target||0,new_customer_target:t?.new_customer_target||0}
    })
  }finally{loading.value=false}
}
async function saveRow(row:any){
  await api.post('/kpi/targets',row);ElMessage.success('保存成功')
}
function go(p:string){router.push(p)}
onMounted(fetchData)
</script>
