<template>
  <div style="max-width:1000px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">本期支出</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#f56c6c">¥{{ stats.highValue.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">高值耗材</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">¥{{ stats.commission.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">分成</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">¥{{ stats.general.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">一般支出</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.count }}笔</div><div style="font-size:12px;color:#909399;margin-top:2px">支出笔数</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <el-date-picker v-model="queryStart" type="date" placeholder="开始日期" value-format="YYYY-MM-DD" clearable style="width:160px" @change="fetchData"/>
        <span style="color:#909399">至</span>
        <el-date-picker v-model="queryEnd" type="date" placeholder="结束日期" value-format="YYYY-MM-DD" clearable style="width:160px" @change="fetchData"/>
        <el-radio-group v-model="typeFilter" @change="fetchData">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="high_value">高值耗材</el-radio-button>
          <el-radio-button value="commission">分成</el-radio-button>
          <el-radio-button value="general">一般支出</el-radio-button>
        </el-radio-group>
        <div style="flex:1"/>
        <el-button size="small" text @click="exportCSV">导出</el-button>
        <el-button size="small" @click="fetchData">刷新</el-button>
        <el-button type="primary" size="small" @click="openCreate">+ 新增支出</el-button>
      </div>
    </el-card>

    <el-card shadow="never" style="margin-bottom:12px">
      <el-table :data="filteredData" v-loading="loading" stripe size="small" :summary-method="summaryMethod" show-summary>
        <el-table-column label="日期" width="100" prop="date"/>
        <el-table-column label="类型" width="90"><template #default="{row}"><el-tag :type="typeTag(row.type)" size="small">{{ typeLabel(row.type) }}</el-tag></template></el-table-column>
        <el-table-column label="分类" width="90" prop="category"/>
        <el-table-column label="金额" width="100" align="right"><template #default="{row}">¥{{ (row.amount||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="备注" min-width="150" prop="note"/>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{row}">
            <el-button size="small" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 月度趋势 -->
    <el-card v-if="monthlyTrend.length" shadow="never">
      <div style="font-weight:600;margin-bottom:8px">月度支出趋势</div>
      <el-table :data="monthlyTrend" stripe border size="small">
        <el-table-column label="月份" width="80" prop="month"/>
        <el-table-column label="高值耗材" width="100" align="right"><template #default="{row}">¥{{ row.hv.toFixed(2) }}</template></el-table-column>
        <el-table-column label="分成" width="100" align="right"><template #default="{row}">¥{{ row.com.toFixed(2) }}</template></el-table-column>
        <el-table-column label="一般支出" width="100" align="right"><template #default="{row}">¥{{ row.gen.toFixed(2) }}</template></el-table-column>
        <el-table-column label="合计" width="100" align="right"><template #default="{row}">¥{{ (row.hv+row.com+row.gen).toFixed(2) }}</template></el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑支出':'新增支出'" width="450px" @close="resetForm">
      <el-form :model="form" label-width="80px" :rules="rules" ref="formRef">
        <el-form-item label="日期" prop="date"><el-date-picker v-model="form.date" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
        <el-form-item label="类型" prop="type"><el-select v-model="form.type" style="width:100%"><el-option label="高值耗材" value="high_value"/><el-option label="分成" value="commission"/><el-option label="一般支出" value="general"/></el-select></el-form-item>
        <el-form-item label="分类" prop="category"><el-input v-model="form.category" placeholder="如药品/人工/房租"/></el-form-item>
        <el-form-item label="金额" prop="amount"><el-input-number v-model="form.amount" :min="0.01" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.note" type="textarea" :rows="2"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="submitting" @click="submitForm">{{ isEdit?'保存':'创建' }}</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const loading=ref(false);const expenses=ref<any[]>([]);const queryStart=ref('');const queryEnd=ref('');const typeFilter=ref('')
const dialogVisible=ref(false);const isEdit=ref(false);const editId=ref<number|null>(null);const submitting=ref(false)
const formRef=ref<any>(null);const form=ref({date:'',type:'high_value',category:'',amount:0,note:''})
const rules={date:[{required:true,message:'请选择日期',trigger:'change'}],type:[{required:true,message:'请选择类型',trigger:'change'}],amount:[{required:true,message:'请输入金额',trigger:'change'}]}

const typeL:Record<string,string>={high_value:'高值耗材',commission:'分成',general:'一般支出'}
const typeT:Record<string,string>={high_value:'danger',commission:'warning',general:'info'}
function typeLabel(s:string){return typeL[s]||s}
function typeTag(s:string){return typeT[s]||'info'}

const filteredData=computed(()=>{
  let a=expenses.value
  if(typeFilter.value)a=a.filter(function(x){return x.type===typeFilter.value})
  return a
})

const stats=computed(()=>{
  let a=filteredData.value
  return{
    total:a.reduce(function(s:number,x:any){return s+(x.amount||0)},0),
    highValue:a.filter(x=>x.type==='high_value').reduce(function(s:number,x:any){return s+(x.amount||0)},0),
    commission:a.filter(x=>x.type==='commission').reduce(function(s:number,x:any){return s+(x.amount||0)},0),
    general:a.filter(x=>x.type==='general').reduce(function(s:number,x:any){return s+(x.amount||0)},0),
    count:a.length
  }
})

const monthlyTrend=computed(()=>{
  const map:Record<string,{hv:number,com:number,gen:number}>={}
  expenses.value.forEach(function(x:any){
    const m=(x.date||'').slice(0,7)
    if(!m)return
    if(!map[m])map[m]={hv:0,com:0,gen:0}
    if(x.type==='high_value')map[m].hv+=x.amount||0
    else if(x.type==='commission')map[m].com+=x.amount||0
    else map[m].gen+=x.amount||0
  })
  return Object.entries(map).sort(([a],[b])=>a.localeCompare(b)).map(function(e){return{month:e[0],hv:e[1].hv,com:e[1].com,gen:e[1].gen}})
})

function summaryMethod({columns,data}:any){
  const s:string[]=[]
  columns.forEach(function(c:any,i:number){
    if(i===0)s[i]='合计'
    else if(c.property==='amount')s[i]='¥'+data.reduce(function(a:number,r:any){return a+Number(r.amount||0)},0).toFixed(2)
    else s[i]=''
  })
  return s
}

async function fetchData(){
  loading.value=true
  try{
    const params:any={}
    if(queryStart.value)params.start_date=queryStart.value
    if(queryEnd.value)params.end_date=queryEnd.value
    const r=await api.get('/expenses',{params})
    const d=r.data??[];expenses.value=Array.isArray(d)?d:d?.items||d?.data||[]
  }catch{expenses.value=[]}
  finally{loading.value=false}
}

function openCreate(){isEdit.value=false;editId.value=null;form.value={date:'',type:'high_value',category:'',amount:0,note:''};dialogVisible.value=true}
function openEdit(row:any){isEdit.value=true;editId.value=row.id;form.value={...row};dialogVisible.value=true}
function resetForm(){formRef.value?.resetFields()}
async function submitForm(){
  const valid=await formRef.value?.validate().catch(()=>false);if(!valid)return
  submitting.value=true
  try{
    if(isEdit.value&&editId.value)await api.put('/expenses/'+editId.value,form.value)
    else await api.post('/expenses',form.value)
    ElMessage.success(isEdit.value?'保存成功':'创建成功');dialogVisible.value=false;await fetchData()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{submitting.value=false}
}
async function handleDelete(row:any){
  try{await ElMessageBox.confirm('确定删除该支出记录吗？','确认',{type:'warning'});await api.delete('/expenses/'+row.id);ElMessage.success('已删除');await fetchData()}catch{}
}
function exportCSV(){
  const a=filteredData.value;if(!a.length)return
  const nl=String.fromCharCode(10)
  const rows=a.map(function(x){return[x.date,typeLabel(x.type),x.category||'',(x.amount||0).toFixed(2),(x.note||'').replace(/"/g,'""')].join('","')}).join(nl)
  const csv='﻿"日期","类型","分类","金额","备注"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='支出明细_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)
}

onMounted(fetchData)
</script>