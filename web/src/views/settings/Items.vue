<template>
  <div style="max-width:1000px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总项目</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">¥{{ stats.avgPrice.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">平均价格</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">¥{{ stats.maxPrice.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">最高价</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.categories }}</div><div style="font-size:12px;color:#909399;margin-top:2px">分类数</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.active }}</div><div style="font-size:12px;color:#909399;margin-top:2px">启用中</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>
        <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
          <span style="font-weight:600;font-size:14px">项目管理</span>
          <el-input v-model="search" placeholder="搜索项目名称" clearable style="width:180px" @input="doFilter"/>
          <el-select v-model="catFilter" placeholder="分类" clearable style="width:110px" @change="doFilter">
            <el-option v-for="(l,v) in categoryMap" :key="v" :label="l" :value="v"/>
          </el-select>
          <span style="color:#909399;font-size:13px">{{ filteredItems.length }}条</span>
          <div style="flex:1"/>
          <el-button size="small" text @click="exportCSV">导出</el-button>
          <el-button type="primary" size="small" @click="openCreate">+ 新建项目</el-button>
        </div>
      </template>
      <el-table :data="filteredItems" v-loading="loading" stripe border size="small">
        <el-table-column prop="name" label="名称" min-width="120"/>
        <el-table-column label="分类" width="80"><template #default="{row}">{{ categoryLabel(row.category) }}</template></el-table-column>
        <el-table-column label="价格" width="90" align="right"><template #default="{row}">¥{{ (row.price||0).toFixed(2) }}</template></el-table-column>
        <el-table-column prop="duration" label="时长" width="70" align="center"/>
        <el-table-column label="使用次数" width="80" align="center"><template #default="{row}"><span :style="{color:itemUsage[row.name]?'#409eff':'#c0c4cc'}">{{ itemUsage[row.name]||0 }}</span></template></el-table-column>
        <el-table-column label="占比" width="60" align="center"><template #default="{row}"><span v-if="itemPct[row.name]" style="color:#909399;font-size:12px">{{ itemPct[row.name] }}%</span><span v-else style="color:#c0c4cc">-</span></template></el-table-column>
        <el-table-column prop="status" label="状态" width="65"><template #default="{row}"><el-tag :type="row.status==='active'?'success':'info'" size="small">{{ row.status==='active'?'启用':'停用' }}</el-tag></template></el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text @click="duplicateItem(row)">复制</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEditing?'编辑项目':'新建项目'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="名称"><el-input v-model="form.name" placeholder="请输入项目名称"/></el-form-item>
        <el-form-item label="分类"><el-select v-model="form.category" style="width:100%"><el-option v-for="(l,v) in categoryMap" :key="v" :label="l" :value="v"/></el-select></el-form-item>
        <el-form-item label="价格"><el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="时长(分钟)"><el-input-number v-model="form.duration" :min="0" style="width:100%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="submitting">确认</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getItems, createItem, updateItem } from '../../api/settings'
import api from '../../api'

const loading=ref(false);const items=ref<any[]>([]);const search=ref('');const catFilter=ref('')
const dialogVisible=ref(false);const isEditing=ref(false);const submitting=ref(false);const editingId=ref<number|null>(null)
const itemUsage=ref<Record<string,number>>({});const itemPct=ref<Record<string,number>>({})
const categoryMap:Record<string,string>={surgery:'外科',skin:'皮肤美容',laser:'光电',injection:'注射',scar:'疤痕',wound:'创面',other:'其他'}
function categoryLabel(val:string){return categoryMap[val]||val}
const form=ref({name:'',category:'',price:0,duration:0})

const filteredItems=computed(()=>{
  let a=items.value
  if(search.value)a=a.filter(function(x){return x.name&&x.name.includes(search.value)})
  if(catFilter.value)a=a.filter(function(x){return x.category===catFilter.value})
  return a
})

const stats=computed(()=>{
  const a=items.value;const cats=new Set(a.map(function(x){return x.category}))
  return{total:a.length,avgPrice:a.length>0?a.reduce(function(s:number,x:any){return s+(x.price||0)},0)/a.length:0,maxPrice:a.length>0?Math.max(...a.map(function(x:any){return x.price||0})):0,categories:cats.size,active:a.filter(function(x){return x.status!=='inactive'}).length}
})

async function loadItems(){
  loading.value=true
  try{
    items.value=await getItems()
    // Load usage frequency from analytics
    try{
      const a=await api.get('/reports/analytics');const top=a.data?.top_items||[]
      const t:Record<string,number>={};const pcts:Record<string,number>={};const totalRev=top.reduce(function(s:number,x:any){return s+(x.total||0)},0)
      top.forEach(function(x:any){t[x.item_name]=x.count||0;pcts[x.item_name]=totalRev>0?Math.round((x.total||0)/totalRev*100):0})
      itemUsage.value=t;itemPct.value=pcts
    }catch{}
  }finally{loading.value=false}
}
function doFilter(){}
function resetForm(){form.value={name:'',category:'',price:0,duration:0}}
function openCreate(){isEditing.value=false;editingId.value=null;resetForm();dialogVisible.value=true}
function openEdit(row:any){isEditing.value=true;editingId.value=row.id;form.value={name:row.name,category:row.category,price:row.price,duration:row.duration};dialogVisible.value=true}
function duplicateItem(row:any){isEditing.value=false;editingId.value=null;form.value={name:row.name+' - 副本',category:row.category,price:row.price,duration:row.duration};dialogVisible.value=true}
async function handleSubmit(){
  submitting.value=true
  try{if(isEditing.value&&editingId.value!==null){await updateItem(editingId.value,form.value)}else{await createItem(form.value)}ElMessage.success(isEditing.value?'更新成功':'创建成功');dialogVisible.value=false;await loadItems()}
  catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{submitting.value=false}
}
function exportCSV(){
  const a=filteredItems.value;if(!a.length)return
  const nl=String.fromCharCode(10)
  const rows=a.map(function(x){return[x.name,categoryLabel(x.category),(x.price||0).toFixed(2),x.duration||0,x.status==='active'?'启用':'停用',itemUsage[x.name]||0].join('","')}).join(nl)
  const csv='﻿"名称","分类","价格","时长","状态","使用次数"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='项目清单_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)
}

onMounted(loadItems)
</script>