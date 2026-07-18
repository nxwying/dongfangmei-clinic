<template>
  <div style="max-width:1000px;margin:0 auto">
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总规则</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.active }}</div><div style="font-size:12px;color:#909399;margin-top:2px">启用中</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.inactive }}</div><div style="font-size:12px;color:#909399;margin-top:2px">已停用</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">{{ stats.totalComm }}</div><div style="font-size:12px;color:#909399;margin-top:2px">本月提成</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.personCount }}</div><div style="font-size:12px;color:#909399;margin-top:2px">有提成人数</div></div></el-card></el-col>
    </el-row>
    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">提成管理</span>
        <template v-if="tab==='rules'">
          <el-input v-model="search" placeholder="搜索规则名称" clearable style="width:180px" @input="doFilter"/>
          <el-select v-model="roleFilter" placeholder="角色" clearable style="width:100px" @change="doFilter">
            <el-option label="咨询师" value="consultant"/><el-option label="医生" value="doctor"/><el-option label="护士" value="nurse"/>
          </el-select>
        </template>
        <span style="color:#909399;font-size:13px">{{ filteredRules.length }}条规则</span>
        <div style="flex:1"/>
        <el-button size="small" text @click="exportCSV">导出</el-button>
        <el-button type="primary" size="small" @click="openCreate">+ 新建规则</el-button>
        <el-button size="small" @click="calculate">计算本月</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-tabs v-model="tab">
        <el-tab-pane label="提成规则" name="rules">
          <el-table :data="filteredRules" v-loading="loading" stripe border size="small">
            <el-table-column label="规则名称" prop="name" min-width="110"/>
            <el-table-column label="角色" width="70"><template #default="{row}">{{ {consultant:'咨询师',doctor:'医生',nurse:'护士'}[row.role]||row.role }}</template></el-table-column>
            <el-table-column label="类型" width="65"><template #default="{row}">{{ {percentage:'比例',fixed:'固定',tiered:'阶梯'}[row.rule_type]||row.rule_type }}</template></el-table-column>
            <el-table-column label="比例" width="70" align="right"><template #default="{row}">{{ row.rate }}%</template></el-table-column>
            <el-table-column label="适用项目" prop="procedure" min-width="90"/>
            <el-table-column label="状态" width="60"><template #default="{row}"><el-switch v-model="row.is_active" @change="toggle(row)" size="small"/></template></el-table-column>
            <el-table-column label="操作" width="130">
              <template #default="{row}">
                <el-button size="small" text @click="openEdit(row)">编辑</el-button>
                <el-button size="small" text @click="duplicateRule(row)">复制</el-button>
                <el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="计算结果" name="results">
          <div style="margin-bottom:12px;display:flex;gap:8px;align-items:center">
            <el-date-picker v-model="calcMonth" type="month" value-format="YYYY-MM" placeholder="选择月份" style="width:160px"/>
            <el-button @click="fetchResults">查询</el-button>
            <el-button @click="calculate">重新计算</el-button>
            <span style="color:#909399;font-size:13px">{{ results.length }}人有提成</span>
            <div style="flex:1"/>
            <el-button size="small" text @click="exportResults">导出结果</el-button>
          </div>
          <el-table :data="results" v-loading="loadingResults" stripe border size="small" :summary-method="summarize" show-summary>
            <el-table-column label="姓名" prop="real_name" min-width="80"/>
            <el-table-column label="总业绩" width="100" align="right"><template #default="{row}">¥{{ (row.total_revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="总提成" width="100" align="right"><template #default="{row}">¥{{ (row.total_commission||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="提成占比" width="80" align="center"><template #default="{row}">{{ row.total_revenue?((row.total_commission/row.total_revenue)*100).toFixed(1):0 }}%</template></el-table-column>
            <el-table-column label="状态" width="70"><template #default="{row}"><el-tag :type="row.status==='paid'?'success':row.status==='confirmed'?'warning':'info'" size="small">{{ {draft:'草稿',confirmed:'已确认',paid:'已发放'}[row.status]||row.status }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="80"><template #default="{row}"><el-button v-if="row.status==='draft'" size="small" @click="confirm(row)">确认</el-button></template></el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑规则':'新建规则'" width="500px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="规则名称"><el-input v-model="form.name"/></el-form-item>
        <el-form-item label="适用角色"><el-select v-model="form.role" style="width:100%"><el-option label="咨询师" value="consultant"/><el-option label="医生" value="doctor"/><el-option label="护士" value="nurse"/></el-select></el-form-item>
        <el-form-item label="规则类型"><el-select v-model="form.rule_type" style="width:100%"><el-option label="按比例" value="percentage"/><el-option label="固定金额" value="fixed"/><el-option label="阶梯" value="tiered"/></el-select></el-form-item>
        <el-form-item label="比例(%)"><el-input-number v-model="form.rate" :min="0" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="适用项目"><el-input v-model="form.procedure" placeholder="留空=全部项目"/></el-form-item>
        <el-form-item v-if="form.rule_type==='tiered'" label="阶梯区间"><el-input-number v-model="form.tier_min" :min="0" style="width:45%"/> ~ <el-input-number v-model="form.tier_max" :min="0" style="width:45%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="submitRule">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
const loading=ref(false);const rules=ref<any[]>([]);const results=ref<any[]>([]);const loadingResults=ref(false)
const tab=ref('rules');const dialogVisible=ref(false);const isEdit=ref(false);const editId=ref(0);const saving=ref(false)
const search=ref('');const roleFilter=ref('')
const calcMonth=ref(new Date().toISOString().slice(0,7))
const form=ref({name:'',role:'consultant',rule_type:'percentage',rate:0,procedure:'',tier_min:0,tier_max:0,is_active:true})

const stats=computed(()=>{
  const r=rules.value;const res=results.value
  return{total:r.length,active:r.filter(function(x){return x.is_active}).length,inactive:r.filter(function(x){return !x.is_active}).length,totalComm:res.reduce(function(s:number,x:any){return s+(x.total_commission||0)},0),personCount:res.filter(function(x:any){return (x.total_commission||0)>0}).length}
})

const filteredRules=computed(()=>{
  let a=rules.value
  if(search.value){const q=search.value.toLowerCase();a=a.filter(function(x){return x.name?.toLowerCase().includes(q)||x.procedure?.toLowerCase().includes(q)})}
  if(roleFilter.value)a=a.filter(function(x){return x.role===roleFilter.value})
  return a
})

function doFilter(){}
async function fetchRules(){loading.value=true;try{const r=await api.get('/commission/rules');rules.value=Array.isArray(r.data)?r.data:[]}finally{loading.value=false}}
async function fetchResults(){loadingResults.value=true;try{const r=await api.get('/commission/results',{params:{year_month:calcMonth.value}});results.value=Array.isArray(r.data)?r.data:[]}finally{loadingResults.value=false}}
function openCreate(){isEdit.value=false;editId.value=0;form.value={name:'',role:'consultant',rule_type:'percentage',rate:0,procedure:'',tier_min:0,tier_max:0,is_active:true};dialogVisible.value=true}
function openEdit(row:any){isEdit.value=true;editId.value=row.id;form.value={...row};dialogVisible.value=true}
function duplicateRule(row:any){isEdit.value=false;editId.value=0;form.value={name:row.name+'(副本)',role:row.role,rule_type:row.rule_type,rate:row.rate,procedure:row.procedure,tier_min:row.tier_min||0,tier_max:row.tier_max||0,is_active:row.is_active};dialogVisible.value=true}
async function submitRule(){saving.value=true;try{if(isEdit.value)await api.put('/commission/rules/'+editId.value,form.value);else await api.post('/commission/rules',form.value);ElMessage.success('保存成功');dialogVisible.value=false;fetchRules()}catch(e:any){ElMessage.error(e?.response?.data?.error||'失败')}finally{saving.value=false}}
async function toggle(row:any){await api.put('/commission/rules/'+row.id,{...row,is_active:!row.is_active});row.is_active=!row.is_active}
async function handleDelete(row:any){try{await ElMessageBox.confirm('确定删除？');await api.delete('/commission/rules/'+row.id);ElMessage.success('已删除');fetchRules()}catch{}}
async function calculate(){try{const r=await api.post('/commission/calculate',{},{params:{year_month:calcMonth.value}});ElMessage.success(r.data?.message||'计算完成');fetchResults()}catch(e:any){ElMessage.error(e?.response?.data?.error||'计算失败')}}
async function confirm(row:any){await api.put('/commission/results/'+row.id+'/confirm');ElMessage.success('已确认');fetchResults()}
function summarize({columns,data}:any){const s:string[]=[];columns.forEach((c:any,i:number)=>{if(i===0)s[i]='合计';else if(c.property==='total_revenue'||c.property==='total_commission')s[i]='¥'+data.reduce((a:number,r:any)=>a+Number(r[c.property]||0),0).toFixed(2);else s[i]=''});return s}
function exportCSV(){const a=filteredRules.value;if(!a.length)return;const nl=String.fromCharCode(10);const rows=a.map(function(x){const r={consultant:'咨询师',doctor:'医生',nurse:'护士'}[x.role]||x.role;return[x.name,r,(x.rate||0)+'%',x.procedure||'全部',x.is_active?'启用':'停用'].join('","')}).join(nl);const csv='﻿"规则名称","角色","比例","适用项目","状态"'+nl+'"'+rows+'"';const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='提成规则_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)}
function exportResults(){const a=results.value;if(!a.length)return;const nl=String.fromCharCode(10);const rows=a.map(function(x){return[x.real_name,(x.total_revenue||0).toFixed(2),(x.total_commission||0).toFixed(2),x.status==='draft'?'草稿':x.status==='confirmed'?'已确认':'已发放'].join('","')}).join(nl);const csv='﻿"姓名","总业绩","总提成","状态"'+nl+'"'+rows+'"';const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='提成结果_'+calcMonth.value+'.csv';el.click();URL.revokeObjectURL(el.href)}
onMounted(()=>{fetchRules();fetchResults()})
</script>