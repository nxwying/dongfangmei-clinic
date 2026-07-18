<template>
  <div style="max-width:900px;margin:0 auto">
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="8"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总规则</div></div></el-card></el-col>
      <el-col :span="8"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.active }}</div><div style="font-size:12px;color:#909399;margin-top:2px">启用中</div></div></el-card></el-col>
      <el-col :span="8"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.inactive }}</div><div style="font-size:12px;color:#909399;margin-top:2px">已停用</div></div></el-card></el-col>
    </el-row>
    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">自动标签规则</span>
        <el-input v-model="search" placeholder="搜索规则名称或标签" clearable style="width:200px" @input="doFilter"/>
        <span style="color:#909399;font-size:13px">{{ filteredRules.length }}条</span>
        <div style="flex:1"/>
        <el-button size="small" text @click="exportCSV">导出</el-button>
        <el-button type="primary" size="small" @click="openCreate">+ 新建规则</el-button>
        <el-button size="small" @click="applyAll">应用规则</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="filteredRules" v-loading="loading" stripe border size="small">
        <el-table-column label="顺序" prop="apply_order" width="50" align="center"/>
        <el-table-column label="规则名称" prop="name" min-width="120"/>
        <el-table-column label="触发条件" min-width="180"><template #default="{row}"><span style="font-size:13px;color:#606266">{{ condsText(parseConds(row.conditions)) }}</span></template></el-table-column>
        <el-table-column label="标签" width="90"><template #default="{row}"><el-tag type="warning" size="small">{{ row.tag }}</el-tag></template></el-table-column>
        <el-table-column label="状态" width="60"><template #default="{row}"><el-switch :model-value="row.is_active" @change="toggleActive(row)" size="small"/></template></el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{row}">
            <el-button size="small" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text @click="duplicateRule(row)">复制</el-button>
            <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑规则':'新建规则'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="规则名称"><el-input v-model="form.name"/></el-form-item>
        <el-form-item label="标签"><el-input v-model="form.tag" placeholder="如：高价值客户"/></el-form-item>
        <el-form-item label="条件">
          <div v-for="(c,i) in form.conditions" :key="i" style="display:flex;gap:8px;margin-bottom:8px">
            <el-select v-model="c.field" style="width:130px"><el-option label="消费次数" value="orders"/><el-option label="消费金额" value="total"/><el-option label="累计到店" value="visit_count"/><el-option label="最近到店(天)" value="days_since_last_visit"/><el-option label="平均客单价" value="avg_order_amount"/><el-option label="注册天数" value="days_since_register"/></el-select>
            <el-select v-model="c.op" style="width:80px"><el-option label="≥" value=">="/><el-option label=">" value=">"/><el-option label="=" value="=="/><el-option label="<" value="<"/><el-option label="≤" value="<="/><el-option label="≠" value="!="/></el-select>
            <el-input-number v-model="c.value" :min="0" style="width:120px"/>
            <el-button v-if="form.conditions.length>1" @click="form.conditions.splice(i,1)">−</el-button>
          </div>
          <el-button size="small" @click="form.conditions.push({field:'orders',op:'>=',value:1})">+ 添加条件</el-button>
        </el-form-item>
        <el-form-item label="执行顺序"><el-input-number v-model="form.apply_order" :min="0" style="width:100%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="submitRule">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
const loading=ref(false);const rules=ref<any[]>([]);const search=ref('')
const dialogVisible=ref(false);const isEdit=ref(false);const saving=ref(false);const editId=ref(0)
const form=ref({name:'',tag:'',conditions:[{field:'orders',op:'>=',value:1}],apply_order:0,is_active:true})
function parseConds(s:any){try{if(typeof s==="string")return JSON.parse(s);if(s&&typeof s==="object")return s;return[]}catch{return[]}}
function condsText(arr:any){
  if(!arr||!arr.length)return'暂无条件'
  return arr.map(function(c:any){const fl:{[k:string]:string}={orders:'消费次数',total:'消费金额',visit_count:'累计到店',days_since_last_visit:'最近到店(天)',avg_order_amount:'平均客单价',days_since_register:'注册天数'};const f=fl[c.field]||c.field;const o=c.op==='>='?'≥':c.op==='>'?'>':c.op==='=='?'=':c.op==='<'?'<':c.op==='<='?'≤':c.op==='!='?'≠':c.op||c.op;return f+o+c.value}).join(' 且 ')
}
const stats=computed(()=>{const r=rules.value;return{total:r.length,active:r.filter(function(x){return x.is_active}).length,inactive:r.filter(function(x){return !x.is_active}).length}})
const filteredRules=computed(()=>{if(!search.value)return rules.value;const q=search.value.toLowerCase();return rules.value.filter(function(x){return(x.name||'').toLowerCase().includes(q)||(x.tag||'').toLowerCase().includes(q)})})
function doFilter(){}
async function fetchRules(){loading.value=true;try{const r=await api.get('/tag-rules');rules.value=Array.isArray(r.data)?r.data:[]}finally{loading.value=false}}
function openCreate(){isEdit.value=false;editId.value=0;form.value={name:'',tag:'',conditions:[{field:'orders',op:'>=',value:1}],apply_order:rules.value.length,is_active:true};dialogVisible.value=true}
function openEdit(row:any){isEdit.value=true;editId.value=row.id;form.value={name:row.name,tag:row.tag,conditions:parseConds(row.conditions),apply_order:row.apply_order,is_active:row.is_active};dialogVisible.value=true}
function duplicateRule(row:any){isEdit.value=false;editId.value=0;form.value={name:row.name+'(副本)',tag:row.tag,conditions:parseConds(row.conditions),apply_order:rules.value.length,is_active:row.is_active};dialogVisible.value=true}
async function submitRule(){saving.value=true;try{const data={...form.value,conditions:JSON.stringify(form.value.conditions)};if(isEdit.value&&editId.value)await api.put('/tag-rules/'+editId.value,data);else await api.post('/tag-rules',data);ElMessage.success(isEdit.value?'更新成功':'创建成功');dialogVisible.value=false;fetchRules()}catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}finally{saving.value=false}}
async function toggleActive(row:any){await api.put('/tag-rules/'+row.id,{...row,is_active:!row.is_active});row.is_active=!row.is_active}
async function handleDelete(row:any){try{await ElMessageBox.confirm('确定删除？');await api.delete('/tag-rules/'+row.id);ElMessage.success('已删除');fetchRules()}catch{}}
async function applyAll(){try{const r=await api.post('/tag-rules/apply');ElMessage.success(r.data?.message||'标签更新完成')}catch(e:any){ElMessage.error(e?.response?.data?.error||'应用失败')}}
function exportCSV(){const a=filteredRules.value;if(!a.length)return;const nl=String.fromCharCode(10);const rows=a.map(function(x){return[x.name,x.tag||'',condsText(parseConds(x.conditions)),x.apply_order||0,x.is_active?'启用':'停用'].join('","')}).join(nl);const csv='﻿"规则名称","标签","触发条件","执行顺序","状态"'+nl+'"'+rows+'"';const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='自动标签规则_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)}
onMounted(fetchRules)
</script>