<template>
  <div style="max-width:1000px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总套餐</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">¥{{ stats.avgPrice.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">平均价格</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">{{ stats.avgSessions }}</div><div style="font-size:12px;color:#909399;margin-top:2px">平均次数</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#f56c6c">¥{{ stats.maxPrice.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">最高价</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.active }}</div><div style="font-size:12px;color:#909399;margin-top:2px">启用中</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>
        <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
          <span style="font-weight:600;font-size:14px">套餐模板管理</span>
          <span style="color:#909399;font-size:13px">{{ templates.length }}条记录</span>
          <div style="flex:1"/>
          <el-button size="small" text @click="exportCSV">导出</el-button>
          <el-button type="primary" size="small" @click="openCreate">+ 新建套餐</el-button>
        </div>
      </template>
      <el-table :data="templates" v-loading="loading" stripe border size="small">
        <el-table-column prop="name" label="名称" min-width="120"/>
        <el-table-column prop="total_sessions" label="总次数" width="70" align="center"/>
        <el-table-column label="价格" width="90" align="right"><template #default="{row}">¥{{ (row.price||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="次均价" width="80" align="right"><template #default="{row}"><span style="color:#909399;font-size:12px">¥{{ (row.total_sessions>0?(row.price/row.total_sessions):0).toFixed(2) }}</span></template></el-table-column>
        <el-table-column label="项目列表" min-width="140"><template #default="{row}"><div style="display:flex;gap:4px;flex-wrap:wrap"><el-tag v-for="(item,i) in parseItems(row.items)" :key="i" size="small" style="margin:1px">{{ item.name }}×{{ item.sessions }}</el-tag></div></template></el-table-column>
        <el-table-column prop="status" label="状态" width="65"><template #default="{row}"><el-tag :type="row.status==='active'?'success':'info'" size="small">{{ row.status==='active'?'启用':'停用' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text @click="duplicateTemplate(row)">复制</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEditing?'编辑套餐':'新建套餐'" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="请输入套餐名称"/></el-form-item>
        <el-form-item label="套餐项目">
          <div style="margin-bottom:8px">
            <el-select v-model="addItemId" filterable placeholder="选择项目添加" style="width:100%" @change="onAddItem">
              <el-option v-for="i in allItems" :key="i.id" :label="i.name+' (¥'+i.price+')'" :value="i.id"/>
            </el-select>
          </div>
          <div v-if="itemRows.length">
            <div v-for="(r,idx) in itemRows" :key="idx" style="display:flex;align-items:center;gap:8px;margin-bottom:6px;padding:6px 8px;background:#f5f7fa;border-radius:4px">
              <span style="flex:1;font-size:14px">{{ r.name }}</span>
              <span style="color:#909399;font-size:12px;margin-right:4px">×</span>
              <el-input-number v-model="r.sessions" :min="1" size="small" style="width:80px"/>
              <el-button text size="small" type="danger" @click="itemRows.splice(idx,1)" style="padding:0 4px">✕</el-button>
            </div>
          </div>
          <div v-else style="text-align:center;padding:12px;color:#c0c4cc;font-size:13px">请从上方选择项目</div>
        </el-form-item>
        <el-form-item label="总次数"><el-input-number v-model="form.total_sessions" :min="1" style="width:100%"/></el-form-item>
        <el-form-item label="价格"><el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="submitting">确认</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getPackageTemplates, createPackageTemplate, updatePackageTemplate } from '../../api/settings'
import api from '../../api'

const loading=ref(false);const templates=ref<any[]>([]);const allItems=ref<any[]>([])
const dialogVisible=ref(false);const isEditing=ref(false);const submitting=ref(false);const editingId=ref<number|null>(null)
const addItemId=ref<number|null>(null);const itemRows=ref<{name:string,sessions:number}[]>([])
const form=ref({name:'',total_sessions:1,price:0})

const stats=computed(()=>{
  const a=templates.value
  return{total:a.length,avgPrice:a.length>0?a.reduce(function(s:number,x:any){return s+(x.price||0)},0)/a.length:0,avgSessions:a.length>0?Math.round(a.reduce(function(s:number,x:any){return s+(x.total_sessions||0)},0)/a.length):0,maxPrice:a.length>0?Math.max(...a.map(function(x:any){return x.price||0})):0,active:a.filter(function(x){return x.status!=='inactive'}).length}
})

function parseItems(items:any):any[]{
  if(!items)return[]
  if(typeof items==='string'){try{return JSON.parse(items)}catch{return[]}}
  if(Array.isArray(items))return items
  return[]
}

function onAddItem(id:number){
  const item=allItems.value.find(function(x){return x.id===id})
  if(item){const ex=itemRows.value.find(function(x){return x.name===item.name});if(ex){ex.sessions++}else{itemRows.value.push({name:item.name,sessions:1})}}
  addItemId.value=null
}

async function loadTemplates(){
  loading.value=true
  try{
    const[t,items]=await Promise.all([getPackageTemplates(),api.get('/settings/items')])
    templates.value=t;allItems.value=items.data?.list||items.data?.data||items.data||[]
  }finally{loading.value=false}
}

function resetForm(){form.value={name:'',total_sessions:1,price:0};itemRows.value=[]}
function openCreate(){isEditing.value=false;editingId.value=null;resetForm();dialogVisible.value=true}
function openEdit(row:any){
  isEditing.value=true;editingId.value=row.id
  form.value={name:row.name,total_sessions:row.total_sessions,price:row.price}
  itemRows.value=parseItems(row.items)
  dialogVisible.value=true
}
function duplicateTemplate(row:any){
  isEditing.value=false;editingId.value=null
  form.value={name:row.name+' - 副本',total_sessions:row.total_sessions,price:row.price}
  itemRows.value=parseItems(row.items)
  dialogVisible.value=true
}
async function handleSubmit(){
  submitting.value=true
  try{
    const items=itemRows.value.length?JSON.stringify(itemRows.value):''
    const payload={name:form.value.name,total_sessions:form.value.total_sessions,price:form.value.price,items}
    if(isEditing.value&&editingId.value!==null){await updatePackageTemplate(editingId.value,payload);ElMessage.success('更新成功')}
    else{await createPackageTemplate(payload);ElMessage.success('创建成功')}
    dialogVisible.value=false;await loadTemplates()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{submitting.value=false}
}
function exportCSV(){
  const a=templates.value;if(!a.length)return
  const nl=String.fromCharCode(10)
  const rows=a.map(function(x){
    const items=parseItems(x.items).map(function(i){return i.name+'×'+i.sessions}).join('、')
    const pp=x.total_sessions>0?(x.price/x.total_sessions).toFixed(2):'0.00'
    return[x.name,x.total_sessions,(x.price||0).toFixed(2),pp,items,x.status==='active'?'启用':'停用'].join('","')
  }).join(nl)
  const csv='﻿"名称","总次数","价格","次均价","项目","状态"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='套餐模板_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)
}
onMounted(loadTemplates)
</script>