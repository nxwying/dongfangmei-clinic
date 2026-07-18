<template>
  <div style="max-width:1100px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总员工</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.active }}</div><div style="font-size:12px;color:#909399;margin-top:2px">启用中</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">{{ stats.inactive }}</div><div style="font-size:12px;color:#909399;margin-top:2px">已停用</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.roles }}</div><div style="font-size:12px;color:#909399;margin-top:2px">角色种类</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ stats.admins }}</div><div style="font-size:12px;color:#909399;margin-top:2px">管理员</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never">
      <template #header>
        <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
          <span style="font-weight:600;font-size:14px">员工管理</span>
          <el-input v-model="search" placeholder="搜索姓名/用户名/手机" clearable style="width:200px" @input="doFilter"/>
          <el-select v-model="roleFilter" placeholder="角色" clearable style="width:110px" @change="doFilter">
            <el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id"/>
          </el-select>
          <span style="color:#909399;font-size:13px">{{ filteredUsers.length }}人</span>
          <div style="flex:1"/>
          <el-button size="small" text @click="exportCSV">导出</el-button>
          <el-button type="primary" size="small" @click="openCreate">+ 新建员工</el-button>
        </div>
      </template>
      <el-table :data="pagedUsers" v-loading="loading" stripe border size="small">
        <el-table-column prop="username" label="用户名" width="100"/>
        <el-table-column prop="real_name" label="姓名" width="80"/>
        <el-table-column prop="phone" label="手机号" width="130"/>
        <el-table-column label="角色" width="80"><template #default="{row}">{{ row.role?.name||'-' }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="65"><template #default="{row}"><el-tag :type="row.status==='active'?'success':'danger'" size="small">{{ row.status==='active'?'启用':'停用' }}</el-tag></template></el-table-column>
        <el-table-column label="最后登录" width="140"><template #default="{row}">{{ row.last_login_at?new Date(row.last_login_at*1000).toLocaleString():'-' }}</template></el-table-column>
        <el-table-column label="操作" width="170" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="openEdit(row)">编辑</el-button>
            <el-button size="small" :type="row.status==='active'?'warning':'success'" @click="toggleStatus(row)">{{ row.status==='active'?'停用':'启用' }}</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="totalPages>1" style="display:flex;justify-content:center;margin-top:16px">
        <el-pagination v-model:current-page="page" :page-size="pageSize" :total="filteredUsers.length" layout="prev,pager,next,total" @current-change="doFilter"/>
      </div>
    </el-card>

    <el-dialog v-model="dialogVisible" :title="isEditing?'编辑员工':'新建员工'" width="500px">
      <el-form :model="form" label-width="110px">
        <el-form-item label="用户名" v-if="!isEditing"><el-input v-model="form.username" placeholder="登录用户名"/></el-form-item>
        <el-form-item :label="isEditing?'新密码(选填)':'密码'"><el-input v-model="form.password" type="password" :placeholder="isEditing?'留空则不修改':'请输入密码'" show-password/></el-form-item>
        <el-form-item label="姓名"><el-input v-model="form.real_name" placeholder="真实姓名"/></el-form-item>
        <el-form-item label="手机号"><el-input v-model="form.phone" placeholder="手机号码"/></el-form-item>
        <el-form-item label="角色">
          <div style="display:flex;gap:8px;align-items:center">
            <el-select v-model="form.role_id" placeholder="选择角色" style="flex:1"><el-option v-for="r in roles" :key="r.id" :label="r.name" :value="r.id"/></el-select>
            <el-button size="small" text @click="goRoles">管理角色 →</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" @click="handleSubmit" :loading="submitting">确认</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getUsers, createUser, updateUser, getRoles, updateUserStatus } from '../../api/users'
import type { User, Role } from '../../api/users'

const router=useRouter()
const loading=ref(false);const users=ref<User[]>([]);const roles=ref<Role[]>([]);const search=ref('');const roleFilter=ref<number|null>(null)
const dialogVisible=ref(false);const isEditing=ref(false);const submitting=ref(false);const editingId=ref<number|null>(null)
const page=ref(1);const pageSize=ref(20)
const form=ref({username:'',password:'',real_name:'',phone:'',role_id:null as number|null})

const stats=computed(()=>{
  const u=users.value;const roleSet=new Set(u.map(function(x){return x.role_id}));const admins=u.filter(function(x){return x.role?.name?.includes('管理')||x.role?.name==='admin'})
  return{total:u.length,active:u.filter(function(x){return x.status==='active'}).length,inactive:u.filter(function(x){return x.status!=='active'}).length,roles:roleSet.size,admins:admins.length}
})

const filteredUsers=computed(()=>{
  let a=users.value
  if(search.value){const q=search.value.toLowerCase();a=a.filter(function(x){return(x.username?.toLowerCase().includes(q)||x.real_name?.toLowerCase().includes(q)||x.phone?.includes(q))})}
  if(roleFilter.value)a=a.filter(function(x){return x.role_id===roleFilter.value})
  return a
})
const totalPages=computed(()=>Math.ceil(filteredUsers.value.length/pageSize.value))
const pagedUsers=computed(()=>{const s=(page.value-1)*pageSize.value;return filteredUsers.value.slice(s,s+pageSize.value)})

function doFilter(){page.value=1}

async function loadData(){
  loading.value=true
  try{const[u,r]=await Promise.all([getUsers(),getRoles()]);users.value=u;roles.value=r}finally{loading.value=false}
}
function resetForm(){form.value={username:'',password:'',real_name:'',phone:'',role_id:null}}
function openCreate(){isEditing.value=false;editingId.value=null;resetForm();dialogVisible.value=true}
function openEdit(row:User){isEditing.value=true;editingId.value=row.id;form.value={username:row.username,password:'',real_name:row.real_name,phone:row.phone,role_id:row.role_id};dialogVisible.value=true}
function goRoles(){router.push('/settings/roles')}
async function toggleStatus(row:User){
  try{await ElMessageBox.confirm('确定'+(row.status==='active'?'停用':'启用')+'员工「'+row.real_name+'」？');const ns=row.status==='active'?'disabled':'active';await updateUserStatus(row.id,ns);ElMessage.success('状态已更新');await loadData()}catch{}
}
async function handleSubmit(){
  submitting.value=true
  try{if(isEditing.value&&editingId.value!==null){await updateUser(editingId.value,form.value);ElMessage.success('更新成功')}else{await createUser(form.value as any);ElMessage.success('创建成功')}dialogVisible.value=false;await loadData()}
  catch(e:any){ElMessage.error(e.response?.data?.error||'操作失败')}
  finally{submitting.value=false}
}
function exportCSV(){
  const a=filteredUsers.value;if(!a.length)return
  const nl=String.fromCharCode(10)
  const rows=a.map(function(x){return[x.username,x.real_name||'',x.phone||'',x.role?.name||'',x.status==='active'?'启用':'停用',x.last_login_at?new Date(x.last_login_at*1000).toLocaleString():''].join('","')}).join(nl)
  const csv='﻿"用户名","姓名","手机号","角色","状态","最后登录"'+nl+'"'+rows+'"'
  const blob=new Blob([csv],{type:'text/csv;charset=utf-8'});const el=document.createElement('a');el.href=URL.createObjectURL(blob);el.download='员工名单_'+new Date().toISOString().slice(0,10)+'.csv';el.click();URL.revokeObjectURL(el.href)
}
onMounted(loadData)
</script>