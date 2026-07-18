<template>
  <div style="max-width:1000px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">📋 电子知情同意书</span>
        <div style="flex:1"/>
        <el-select v-model="searchCid" filterable remote :remote-method="searchCust" placeholder="搜索客户" style="width:200px" clearable @change="fetchData">
          <el-option v-for="c in custOpts" :key="c.id" :label="c.name" :value="c.id"/>
        </el-select>
        <el-button type="primary" @click="openCreate">+ 新建</el-button>
      </div>
    </el-card>
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border>
        <el-table-column label="客户" min-width="100"><template #default="{r}">{{ custName(r.customer_id) }}</template></el-table-column>
        <el-table-column label="项目" prop="procedure_name" min-width="140"/>
        <el-table-column label="医生" prop="doctor_name" width="80"/>
        <el-table-column label="签署日期" prop="sign_date" width="100"/>
        <el-table-column label="状态" width="80">
          <template #default="{r}"><el-tag :type="r.status==='signed'?'success':r.status==='archived'?'info':'warning'">{{ {draft:'草稿',signed:'已签',archived:'已归档'}[r.status]||r.status }}</el-tag></template>
        </el-table-column>
        <el-table-column label="操作" width="160">
          <template #default="{r}">
            <el-button size="small" @click="viewDetail(r)">查看</el-button>
            <el-button v-if="r.status==='draft'" size="small" @click="openSign(r)">签署</el-button>
            <el-button size="small" text type="danger" @click="handleDelete(r)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Create Dialog -->
    <el-dialog v-model="createVisible" title="新建知情同意书" width="600px">
      <el-form :model="form" label-width="100px">
        <el-form-item label="客户"><el-select v-model="form.customer_id" filterable style="width:100%"><el-option v-for="c in custOpts" :key="c.id" :label="c.name" :value="c.id"/></el-select></el-form-item>
        <el-form-item label="项目名称"><el-input v-model="form.procedure_name" placeholder="如：玻尿酸填充"/></el-form-item>
        <el-form-item label="医生"><el-input v-model="form.doctor_name"/></el-form-item>
        <el-form-item label="同意书内容">
          <el-input v-model="form.content" type="textarea" :rows="8" placeholder="输入告知内容，包括：项目说明、风险告知、术后注意事项等"/>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="createVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="submitCreate">创建</el-button></template>
    </el-dialog>

    <!-- Detail/Sign Dialog -->
    <el-dialog v-model="detailVisible" :title="'知情同意书 - '+current?.procedure_name" width="700px">
      <div v-if="current" style="padding:20px;border:1px solid #dcdfe6;border-radius:8px;min-height:400px">
        <div style="text-align:center;font-size:18px;font-weight:700;margin-bottom:20px">医疗美容知情同意书</div>
        <div style="margin-bottom:12px"><b>客户：</b>{{ custName(current.customer_id) }} &nbsp;&nbsp; <b>项目：</b>{{ current.procedure_name }} &nbsp;&nbsp; <b>医生：</b>{{ current.doctor_name }}</div>
        <div style="margin-bottom:12px"><b>日期：</b>{{ current.sign_date || '未签署' }}</div>
        <div style="border-top:1px solid #eee;padding-top:12px;white-space:pre-wrap;line-height:1.8">{{ current.content }}</div>
        <div v-if="current.status==='signed'" style="border-top:1px solid #eee;margin-top:20px;padding-top:12px">
          <div style="display:flex;gap:40px">
            <div><b>患者签名：</b><img v-if="current.patient_sign" :src="current.patient_sign" style="max-width:200px;max-height:60px;border:1px solid #ddd"/><span v-else style="color:#909399">未签名</span></div>
            <div><b>医生签名：</b><img v-if="current.doctor_sign" :src="current.doctor_sign" style="max-width:200px;max-height:60px;border:1px solid #ddd"/><span v-else style="color:#909399">未签名</span></div>
          </div>
        </div>
        <div v-if="current.status==='draft'" style="border-top:1px solid #eee;margin-top:20px;padding-top:12px">
          <div style="margin-bottom:12px"><b>患者签名：</b><canvas ref="patientCanvas" width="300" height="80" style="border:1px solid #ccc;display:block;cursor:crosshair" @mousedown="startSign('patient',$event)" @mousemove="moveSign('patient',$event)" @mouseup="endSign"/><el-button size="small" @click="clearSign('patient')">清除</el-button></div>
          <div style="margin-bottom:12px"><b>医生签名：</b><canvas ref="doctorCanvas" width="300" height="80" style="border:1px solid #ccc;display:block;cursor:crosshair" @mousedown="startSign('doctor',$event)" @mousemove="moveSign('doctor',$event)" @mouseup="endSign"/><el-button size="small" @click="clearSign('doctor')">清除</el-button></div>
        </div>
      </div>
      <template #footer>
        <el-button @click="detailVisible=false">关闭</el-button>
        <el-button v-if="current?.status==='draft'" type="primary" :loading="saving" @click="submitSign">确认签署</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'
const loading=ref(false);const items=ref<any[]>([]);const custOpts=ref<any[]>([]);const searchCid=ref(0)
const createVisible=ref(false);const detailVisible=ref(false);const saving=ref(false)
const current=ref<any>(null)
const form=ref({customer_id:0,procedure_name:'',doctor_name:'',content:'',status:'draft'})
const custNameMap=ref<Record<number,string>>({})

function custName(id:number){return custNameMap.value[id]||'#'+id}

async function loadAllCustomers(){
  try{
    const r=await api.get('/customers',{params:{page_size:500}})
    const list=Array.isArray(r.data)?r.data:(r.data?.items||[])
    list.forEach((c:any)=>{custNameMap.value[c.id]=c.name})
    custOpts.value=list
  }catch{}}


// Signature state
const patientCanvas=ref<HTMLCanvasElement|null>(null)
const doctorCanvas=ref<HTMLCanvasElement|null>(null)
const signing=ref<{type:string;ctx:CanvasRenderingContext2D|null;drawing:boolean}>({type:'',ctx:null,drawing:false})
const signData=ref<{patient:string;doctor:string}>({patient:'',doctor:''})

async function searchCust(q:string){
  if(!q){
    try{const r=await api.get('/customers',{params:{page_size:500}});custOpts.value=Array.isArray(r.data)?r.data:(r.data?.items||[])}catch(e){console.log(e)}
    return
  }
  try{const r=await api.get('/customers',{params:{keyword:q}});custOpts.value=Array.isArray(r.data)?r.data:(r.data?.items||[])}catch(e){console.log(e)}
}
async function fetchData(){
  loading.value=true
  try{const r=await api.get('/consent',{params:{customer_id:searchCid.value||undefined}});items.value=Array.isArray(r.data)?r.data:[]}finally{loading.value=false}
}
function openCreate(){form.value={customer_id:0,procedure_name:'',doctor_name:'',content:'',status:'draft'};loadAllCustomers();createVisible.value=true}
async function submitCreate(){
  saving.value=true
  try{await api.post('/consent',form.value);ElMessage.success('创建成功');createVisible.value=false;fetchData()}catch(e:any){ElMessage.error(e?.response?.data?.error||'失败')}finally{saving.value=false}
}
async function viewDetail(r:any){
  try{const d=await api.get('/consent/'+r.id);current.value=d.data;detailVisible.value=true;signData.value={patient:'',doctor:''};setTimeout(initCanvas,100)}catch{}
}
function openSign(r:any){viewDetail(r)}
function initCanvas(){
  ;['patient','doctor'].forEach(t=>{
    const canvas=t==='patient'?patientCanvas.value:doctorCanvas.value
    if(!canvas)return
    const ctx=canvas.getContext('2d')
    if(ctx){ctx.strokeStyle='#000';ctx.lineWidth=2;ctx.lineCap='round'}
  })
}
function startSign(type:string,e:MouseEvent){
  const canvas=type==='patient'?patientCanvas.value:doctorCanvas.value
  if(!canvas)return
  const ctx=canvas.getContext('2d')
  if(!ctx)return
  const rect=canvas.getBoundingClientRect()
  ctx.beginPath();ctx.moveTo(e.clientX-rect.left,e.clientY-rect.top)
  signing.value={type,ctx,drawing:true}
}
function moveSign(type:string,e:MouseEvent){
  if(!signing.value.drawing||signing.value.type!==type)return
  const canvas=type==='patient'?patientCanvas.value:doctorCanvas.value
  if(!canvas)return
  const rect=canvas.getBoundingClientRect()
  signing.value.ctx?.lineTo(e.clientX-rect.left,e.clientY-rect.top)
  signing.value.ctx?.stroke()
}
function endSign(){
  signing.value.drawing=false
  const canvas=signing.value.type==='patient'?patientCanvas.value:doctorCanvas.value
  if(canvas)signData.value[signing.value.type as 'patient'|'doctor']=canvas.toDataURL()
}
function clearSign(type:string){
  const canvas=type==='patient'?patientCanvas.value:doctorCanvas.value
  if(!canvas)return
  const ctx=canvas.getContext('2d')
  ctx?.clearRect(0,0,canvas.width,canvas.height)
  signData.value[type as 'patient'|'doctor']=''
}
async function submitSign(){
  if(!current.value?.id)return
  saving.value=true
  try{
    await api.put('/consent/'+current.value.id,{
      patient_sign:signData.value.patient,doctor_sign:signData.value.doctor,
      status:'signed',sign_date:new Date().toISOString().slice(0,10)
    })
    ElMessage.success('签署成功');detailVisible.value=false;fetchData()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'失败')}finally{saving.value=false}
}
async function handleDelete(r:any){try{await ElMessageBox.confirm('确定删除？');await api.delete('/consent/'+r.id);ElMessage.success('已删除');fetchData()}catch{}}

onMounted(()=>{fetchData();loadAllCustomers()})
</script>
