<template>
  <div>
    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:14px">操作日志</span>
        <el-date-picker v-model="fStart" type="date" placeholder="开始日期" value-format="YYYY-MM-DD" clearable style="width:150px" @change="loadLogs"/>
        <span style="color:#909399">至</span>
        <el-date-picker v-model="fEnd" type="date" placeholder="结束日期" value-format="YYYY-MM-DD" clearable style="width:150px" @change="loadLogs"/>
        <el-select v-model="fAction" placeholder="操作类型" clearable style="width:110px" @change="loadLogs">
          <el-option label="创建" value="create"/><el-option label="修改" value="update"/><el-option label="删除" value="delete"/>
          <el-option label="充值" value="recharge"/><el-option label="退款" value="refund"/><el-option label="登录" value="login"/>
        </el-select>
        <el-select v-model="fUser" placeholder="操作人" clearable style="width:110px" @change="loadLogs">
          <el-option v-for="(n,id) in userMap" :key="id" :label="n" :value="String(id)"/>
        </el-select>
        <span style="flex:1"/>
        <el-button size="small" text @click="exportCSV">导出</el-button>
        <el-button size="small" @click="loadLogs">刷新</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="logs" v-loading="loading" stripe border size="small">
        <el-table-column label="时间" width="150" prop="created_at"/>
        <el-table-column label="操作人" width="100"><template #default="{row}">{{ userName(row.user_id) }}</template></el-table-column>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-tag :type="actionTagType(row.action)" size="small">{{ actionLabel(row.action) }}</el-tag></template></el-table-column>
        <el-table-column label="资源类型" width="90"><template #default="{row}">{{ targetLabel(row.target) }}</template></el-table-column>
        <el-table-column label="资源ID" width="70" prop="target_id"/>
        <el-table-column label="详情" min-width="260" prop="detail"/>
      </el-table>
      <div style="margin-top:16px;display:flex;justify-content:center">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total" layout="prev,pager,next,total" @current-change="loadLogs"/>
      </div>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '../../api'
const actionLabels:Record<string,string>={create:'创建',update:'修改',delete:'删除',recharge:'充值',refund:'退款',login:'登录'}
const actionTagTypes:Record<string,string>={create:'success',update:'warning',delete:'danger',recharge:'',refund:'danger',login:'info'}
const targetLabels:Record<string,string>={customer:'客户',order:'订单',payment:'支付',appointment:'预约',membership:'会员',user:'员工',role:'角色',treatment_item:'项目',package_template:'套餐',consent:'同意书',follow_up:'回访',photo:'照片',document:'证件',expense:'支出',commission:'提成',kpi:'KPI',tag_rule:'标签',training:'培训',inventory:'库存',backup:'备份'}
function actionLabel(a:string){return actionLabels[a]||a}
function actionTagType(a:string){return actionTagTypes[a]||'info'}
function targetLabel(t:string){return targetLabels[t]||t}
const loading=ref(false);const logs=ref<any[]>([]);const page=ref(1);const pageSize=ref(20);const total=ref(0)
const userMap=ref<Record<number,string>>({});const fStart=ref('');const fEnd=ref('');const fAction=ref('');const fUser=ref('')
async function loadLogs(){
  loading.value=true;page.value=1
  try{
    const params:any={page:page.value,page_size:pageSize.value}
    if(fStart.value)params.start_date=fStart.value
    if(fEnd.value)params.end_date=fEnd.value
    if(fAction.value)params.action=fAction.value
    if(fUser.value)params.user_id=fUser.value
    const res=await api.get('/audit-logs',{params})
    logs.value=res.data?.data??[];total.value=res.data?.total??0
  }finally{loading.value=false}
}
async function loadUsers(){
  try{const res=await api.get('/users');const users:any[]=res.data?.data??res.data?.list??res.data??[];const map:Record<number,string>={};for(const u of users)map[u.id]=u.real_name||u.username;userMap.value=map}catch{userMap.value={}}
}
function userName(id:number){return userMap.value[id]||'用户#'+id}
function exportCSV(){
  const a=logs.value;if(!a.length)return
  const nl=String.fromCharCode(10)
  const rows=a.map(function(x){return[x.created_at,userName(x.user_id),actionLabel(x.action),targetLabel(x.target),x.target_id||'',(x.detail||'').replace(/"/g,'""')].join('","')}).join(nl)
  const csv='﻿"时间","操作人","操作","资源类型","资源ID","详情"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='操作日志_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)
}
onMounted(async()=>{await loadUsers();await loadLogs()})
</script>