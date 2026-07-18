<template>
  <div style="max-width:900px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px">
        <span style="font-weight:600;font-size:15px">📚 培训与认证</span>
        <div style="flex:1"/><el-button type="primary" @click="dialogVisible=true">+ 新增培训</el-button>
      </div>
    </el-card>
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col v-for="s in stats" :key="s.user_id" :span="8">
        <el-card shadow="never"><div style="text-align:center;padding:8px 0">
          <div style="font-weight:600">{{ s.real_name }}</div>
          <div style="font-size:13px;color:#909399">培训 {{ s.total_sessions }} 次 / {{ s.total_hours }} 小时</div>
        </div></el-card>
      </el-col>
    </el-row>
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border>
        <el-table-column label="培训内容" prop="title" min-width="150"/>
        <el-table-column label="讲师" prop="trainer" width="100"/>
        <el-table-column label="日期" prop="date" width="100"/>
        <el-table-column label="时长" prop="hours" width="60" align="center"/>
        <el-table-column label="证书到期" prop="cert_expiry" width="100"/>
        <el-table-column label="操作" width="80"><template #default="{row}"><el-button size="small" text type="danger" @click="handleDelete(row)">删除</el-button></template></el-table-column>
      </el-table>
    </el-card>
    <el-dialog v-model="dialogVisible" title="新增培训" width="450px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="员工"><el-select v-model="form.user_id" filterable style="width:100%"><el-option v-for="u in allUsers" :key="u.id" :label="u.real_name" :value="u.id"/></el-select></el-form-item>
        <el-form-item label="培训内容"><el-input v-model="form.title"/></el-form-item>
        <el-form-item label="讲师"><el-input v-model="form.trainer"/></el-form-item>
        <el-form-item label="日期"><el-date-picker v-model="form.date" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
        <el-form-item label="时长(小时)"><el-input-number v-model="form.hours" :min="1" style="width:100%"/></el-form-item>
        <el-form-item label="证书到期"><el-date-picker v-model="form.cert_expiry" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.notes" type="textarea"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="submit">保存</el-button></template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
const loading=ref(false);const items=ref<any[]>([]);const stats=ref<any[]>([]);const allUsers=ref<any[]>([])
const dialogVisible=ref(false);const saving=ref(false)
const form=ref({user_id:0,title:'',trainer:'',date:'',hours:1,cert_expiry:'',notes:''})
async function fetchData(){
  loading.value=true
  try{
    const r=await api.get('/training');items.value=Array.isArray(r.data)?r.data:[]
    const s=await api.get('/training/stats');stats.value=Array.isArray(s.data)?s.data:[]
    const u=await api.get('/users');allUsers.value=Array.isArray(u.data)?u.data:[]
  }finally{loading.value=false}
}
async function submit(){
  saving.value=true
  try{await api.post('/training',form.value);ElMessage.success('保存成功');dialogVisible.value=false;fetchData()}
  catch(e:any){ElMessage.error(e?.response?.data?.error||'失败')}
  finally{saving.value=false}
}
async function handleDelete(row:any){
  try{await ElMessageBox.confirm('确定删除？');await api.delete('/training/'+row.id);ElMessage.success('已删除');fetchData()}catch{}
}
onMounted(fetchData)
</script>
