<template>
  <div style="max-width:1000px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">{{ dormant.length }}</div><div style="font-size:12px;color:#909399;margin-top:2px">沉睡客户</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ nearDormant.length }}</div><div style="font-size:12px;color:#909399;margin-top:2px">近沉默</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ birthday.length }}</div><div style="font-size:12px;color:#909399;margin-top:2px">本月生日</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ activeRate }}%</div><div style="font-size:12px;color:#909399;margin-top:2px">可激活占比</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">¥{{ totalLost.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">潜在损失</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never" style="margin-bottom:14px">
      <el-tabs v-model="tab">
        <el-tab-pane label="沉睡客户" name="dormant"/>
        <el-tab-pane label="近沉默" name="near"/>
        <el-tab-pane label="本月生日" name="birthday"/>
        <el-tab-pane label="渠道分析" name="sources"/>
      </el-tabs>
    </el-card>

    <!-- 沉睡客户 -->
    <el-card shadow="never" v-if="tab==='dormant'">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>超过90天未到店</span>
          <div style="display:flex;gap:8px">
            <el-button size="small" text @click="exportCSV(dormant,'沉睡客户')">导出</el-button>
            <el-button size="small" @click="loadDormant">刷新</el-button>
            <el-button size="small" type="warning" @click="batchCreateFollowup(dormant,'dormant')">全部创建回访</el-button>
          </div>
        </div>
      </template>
      <el-table :data="dormant" v-loading="dLoading" stripe size="small" empty-text="暂无沉睡客户">
        <el-table-column label="客户" min-width="80" prop="name"/>
        <el-table-column label="电话" width="110" prop="phone"/>
        <el-table-column label="上次到店" width="100" prop="last_visit"/>
        <el-table-column label="离开" width="60" align="center" prop="days_since_last"/>
        <el-table-column label="消费" width="100" align="right">
          <template #default="{row}"><span :style="{color:row.total_spent>0?'#e6a23c':'#909399'}">¥{{ (row.total_spent||0).toFixed(2) }}</span></template>
        </el-table-column>
        <el-table-column label="备注" min-width="100"><template #default="{row}"><el-input v-model="dormantNotes[row.name]" size="small" placeholder="输入跟进备注"/></template></el-table-column>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button size="small" @click="createFollowup(row,'dormant')">创建回访</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <!-- 近沉默 -->
    <el-card shadow="never" v-if="tab==='near'">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>60-89天未到店</span>
          <div style="display:flex;gap:8px">
            <el-button size="small" text @click="exportCSV(nearDormant,'近沉默客户')">导出</el-button>
            <el-button size="small" @click="loadNearDormant">刷新</el-button>
            <el-button size="small" type="warning" @click="batchCreateFollowup(nearDormant,'dormant')">全部创建回访</el-button>
          </div>
        </div>
      </template>
      <el-table :data="nearDormant" v-loading="nLoading" stripe size="small" empty-text="暂无近沉默客户">
        <el-table-column label="客户" min-width="80" prop="name"/>
        <el-table-column label="电话" width="110" prop="phone"/>
        <el-table-column label="上次到店" width="100" prop="last_visit"/>
        <el-table-column label="离开" width="60" align="center" prop="days_since_last"/>
        <el-table-column label="消费" width="100" align="right">
          <template #default="{row}"><span :style="{color:row.total_spent>0?'#e6a23c':'#909399'}">¥{{ (row.total_spent||0).toFixed(2) }}</span></template>
        </el-table-column>
        <el-table-column label="备注" min-width="100"><template #default="{row}"><el-input v-model="nearNotes[row.name]" size="small" placeholder="输入跟进备注"/></template></el-table-column>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button size="small" @click="createFollowup(row,'dormant')">提前回访</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <!-- 生日客户 -->
    <el-card shadow="never" v-if="tab==='birthday'">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>本月生日客户</span>
          <div style="display:flex;gap:8px">
            <el-button size="small" text @click="exportCSV(birthday,'本月生日')">导出</el-button>
            <el-button size="small" @click="loadBirthday">刷新</el-button>
            <el-button size="small" type="success" :loading="bdayGenLoading" @click="generateBirthdayTasks">生日回访</el-button>
          </div>
        </div>
      </template>
      <el-table :data="birthday" v-loading="bLoading" stripe size="small" empty-text="本月无生日客户">
        <el-table-column label="客户" min-width="80" prop="name"/>
        <el-table-column label="电话" width="120" prop="phone"/>
        <el-table-column label="生日" width="90" prop="birthday"/>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button size="small" @click="createFollowup(row,'birthday')">生日关怀</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <!-- 渠道分析 -->
    <el-card shadow="never" v-if="tab==='sources'">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>客户来源分布</span>
          <el-button size="small" text @click="exportCSV(sourceData,'客户来源')">导出</el-button>
        </div>
      </template>
      <el-table :data="sourceData" stripe border size="small" empty-text="加载中...">
        <el-table-column label="来源" min-width="100"><template #default="{row}">{{ srcLabel(row.source) }}</template></el-table-column>
        <el-table-column label="人数" prop="count" width="80" align="center"/>
        <el-table-column label="占比" width="80" align="center"><template #default="{row}">{{ row.pct }}%</template></el-table-column>
        <el-table-column label="累计消费" width="120" align="right"><template #default="{row}">¥{{ (row.total_spent||0).toFixed(2) }}</template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'

const tab=ref('dormant')
const dormant=ref<any[]>([]); const dLoading=ref(false)
const nearDormant=ref<any[]>([]); const nLoading=ref(false)
const birthday=ref<any[]>([]); const bLoading=ref(false)
const genLoading=ref(false); const bdayGenLoading=ref(false)
const dormantNotes=ref<Record<string,string>>({}); const nearNotes=ref<Record<string,string>>({})
const allCustomers=ref<any[]>([])

// 统计
const activeRate=computed(()=>{
  const total=dormant.value.length+nearDormant.value.length
  return total>0?Math.round(nearDormant.value.length/total*100):0
})
const totalLost=computed(()=>{
  return dormant.value.reduce((s:number,r:any)=>s+(r.total_spent||0),0)+nearDormant.value.reduce((s:number,r:any)=>s+(r.total_spent||0),0)
})

// 来源分析
const srcL:Record<string,string>={walk_in:'到店',referral:'转介绍',xiaohongshu:'小红书',wechat:'微信',douyin:'抖音',dianping:'大众点评',other:'其他'}
function srcLabel(s:string){return srcL[s]||s||'其他'}
const sourceData=computed(()=>{
  const m=allCustomers.value;const map:Record<string,{count:number,total_spent:number}>={}
  m.forEach((c:any)=>{const s=c.source||'other';if(!map[s])map[s]={count:0,total_spent:0};map[s].count++;map[s].total_spent+=c.membership?.total_consumed||0})
  const total=m.length;return Object.entries(map).map(([source,v])=>({source,count:v.count,total_spent:v.total_spent,pct:total>0?Math.round(v.count/total*100):0})).sort((a,b)=>b.count-a.count)
})

async function loadDormant(){dLoading.value=true;try{const r=await api.get('/marketing/dormant');dormant.value=r.data||[]}catch{dormant.value=[]}finally{dLoading.value=false}}
async function loadNearDormant(){nLoading.value=true;try{const r=await api.get('/marketing/near-dormant');nearDormant.value=r.data||[]}catch{nearDormant.value=[]}finally{nLoading.value=false}}
async function loadBirthday(){bLoading.value=true;try{const r=await api.get('/marketing/birthday');birthday.value=r.data||[]}catch{birthday.value=[]}finally{bLoading.value=false}}

async function loadAllCustomers(){
  try{const r=await api.get('/customers',{params:{page_size:500}});const d=r.data?.data||[];allCustomers.value=d}catch{allCustomers.value=[]}
}

async function createFollowup(customer:any,type:string){
  try{
    const note=type==='birthday'?'生日关怀：祝'+customer.name+'生日快乐！':'沉睡客户唤醒回访'
    await api.post('/followup/tasks',{customer_id:customer.id,type:type==='birthday'?'birthday':'dormant',due_date:new Date().toISOString().slice(0,10),note})
    ElMessage.success('已为 '+customer.name+' 创建回访任务')
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
}

async function batchCreateFollowup(list:any[],type:string){
  genLoading.value=true;let count=0
  try{
    for(const c of list){
      await api.post('/followup/tasks',{customer_id:c.id,type,due_date:new Date().toISOString().slice(0,10),note:type==='birthday'?'生日关怀':'沉睡客户唤醒回访'})
      count++
    }
    ElMessage.success('已为'+count+'位客户创建回访任务')
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{genLoading.value=false}
}

async function generateBirthdayTasks(){
  bdayGenLoading.value=true;let count=0
  try{
    for(const c of birthday.value){
      await api.post('/followup/tasks',{customer_id:c.id,type:'birthday',due_date:new Date().toISOString().slice(0,10),note:'生日关怀：祝'+c.name+'生日快乐！'})
      count++
    }
    ElMessage.success('已为'+count+'位生日客户创建回访任务')
    await loadBirthday()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{bdayGenLoading.value=false}
}

// 导出
function exportCSV(data:any[],name:string){
  if(!data.length)return
  const nl=String.fromCharCode(10)
  const sep='","'
  const h='客户'+sep+'电话'+sep+'状态'+sep+'消费'
  const rows=data.map(function(x){
    const status=name.includes('生日')?x.birthday:(x.days_since_last||'')+'天'
    return [x.name,x.phone,status,(x.total_spent||x.total_deposited||0).toFixed(2)].join('","')
  }).join(nl)
  const csv='﻿"'+h+'"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'})
  const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download=name+'_'+new Date().toISOString().slice(0,10)+'.csv'
  el.click();URL.revokeObjectURL(el.href)
}

onMounted(()=>{loadDormant();loadNearDormant();loadBirthday();loadAllCustomers()})
</script>