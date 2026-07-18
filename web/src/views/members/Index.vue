<template>
  <div>
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总会员</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ stats.newMonth }}</div><div style="font-size:12px;color:#909399;margin-top:2px">本月新增</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">¥{{ stats.balance.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总储值余额</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">¥{{ stats.recharged.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">累计充值</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">¥{{ stats.consumed.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">累计消费</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never">
      <el-tabs v-model="tab">
        <!-- Tab 1: 会员列表 -->
        <el-tab-pane label="👥 会员列表" name="list">
          <div style="display:flex;gap:8px;margin-bottom:12px;flex-wrap:wrap">
            <el-input v-model="search" placeholder="搜索姓名/手机号" clearable style="width:200px" @input="loadMembers"/>
            <el-select v-model="levelFilter" placeholder="等级" clearable style="width:120px" @change="loadMembers">
              <el-option v-for="l in levels" :key="l.val" :label="l.label" :value="l.val"/>
            </el-select>
            <span style="flex:1"/>
            <el-button type="primary" size="small" @click="openCreate">+ 添加会员</el-button>
          </div>
          <el-table :data="members" v-loading="loading" stripe size="small">
            <el-table-column prop="name" label="姓名" width="80"/>
            <el-table-column prop="phone" label="手机" width="110"/>
            <el-table-column label="等级" width="80"><template #default="{row}"><el-tag :type="levelTag(row.membership?.level)" size="small">{{ levelLabel(row.membership?.level) }}</el-tag></template></el-table-column>
            <el-table-column label="储值" width="80" align="right"><template #default="{row}">¥{{ row.membership?.balance?.toFixed(2)||'0.00' }}</template></el-table-column>
            <el-table-column label="赠金" width="75" align="right"><template #default="{row}">¥{{ row.membership?.gift_balance?.toFixed(2)||'0.00' }}</template></el-table-column>
            <el-table-column label="累计充值" width="90" align="right"><template #default="{row}">¥{{ row.membership?.total_recharged?.toFixed(2)||'0.00' }}</template></el-table-column>
            <el-table-column label="消费次数" width="75" align="center" prop="order_count"/>
            <el-table-column label="生日" width="75" prop="birthday"/>
            <el-table-column label="操作" fixed="right" min-width="200">
              <template #default="{row}">
                <el-button type="primary" size="small" @click="openRecharge(row)">充值</el-button>
                <el-button size="small" @click="openConsume(row)">消费</el-button>
                <el-button type="success" size="small" @click="openPackages(row)">套餐</el-button>
                <el-button size="small" text @click="openEdit(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 生日关怀 -->
        <el-tab-pane label="🎂 生日关怀" name="birthday">
          <div style="font-weight:600;margin-bottom:8px">本月生日客户（{{ birthdayList.length }} 人）</div>
          <el-table :data="birthdayList" v-loading="ldBirthday" stripe border size="small">
            <el-table-column label="姓名" prop="name" min-width="80"/>
            <el-table-column label="电话" prop="phone" width="120"/>
            <el-table-column label="生日" prop="birthday" width="80"/>
            <el-table-column label="来源" prop="source" width="80"/>
          </el-table>
        </el-tab-pane>

        <!-- Tab 3: 沉睡激活 -->
        <el-tab-pane label="⚠️ 沉睡激活" name="dormant">
          <div style="font-weight:600;margin-bottom:8px">近沉默客户（45-60天）</div>
          <el-table :data="nearDormant" v-loading="ldDormant" stripe border size="small" style="margin-bottom:16px">
            <el-table-column label="姓名" prop="name" min-width="70"/>
            <el-table-column label="电话" prop="phone" width="110"/>
            <el-table-column label="未到店" width="70" align="center"><template #default="{row}">{{ row.days_since_last }}天</template></el-table-column>
            <el-table-column label="历史消费" width="100" align="right"><template #default="{row}">¥{{ (row.total_spent||0).toFixed(2) }}</template></el-table-column>
          </el-table>
          <div style="font-weight:600;margin-bottom:8px">沉默客户（>60天）</div>
          <el-table :data="dormantList" stripe border size="small">
            <el-table-column label="姓名" prop="name" min-width="70"/>
            <el-table-column label="电话" prop="phone" width="110"/>
            <el-table-column label="离开" width="60" align="center"><template #default="{row}">{{ row.days_gone }}天</template></el-table-column>
            <el-table-column label="历史消费" width="100" align="right"><template #default="{row}">¥{{ (row.revenue||0).toFixed(2) }}</template></el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 4: 会员分析 -->
        <el-tab-pane label="📊 会员分析" name="analytics">
          <el-row :gutter="12">
            <el-col :span="12">
              <div style="font-weight:600;margin-bottom:8px">等级分布</div>
              <el-table :data="levelDist" stripe border size="small" style="margin-bottom:16px">
                <el-table-column label="等级" min-width="80"><template #default="{row}"><el-tag :type="levelTag(row.level)" size="small">{{ levelLabel(row.level) }}</el-tag></template></el-table-column>
                <el-table-column label="人数" prop="count" width="60" align="center"/>
                <el-table-column label="占比" width="60" align="center"><template #default="{row}">{{ row.pct }}%</template></el-table-column>
              </el-table>
            </el-col>
            <el-col :span="12">
              <div style="font-weight:600;margin-bottom:8px">来源分布</div>
              <el-table :data="sourceDist" stripe border size="small" style="margin-bottom:16px">
                <el-table-column label="来源" min-width="80"><template #default="{row}">{{ srcLabel(row.source) }}</template></el-table-column>
                <el-table-column label="人数" prop="count" width="60" align="center"/>
                <el-table-column label="占比" width="60" align="center"><template #default="{row}">{{ row.pct }}%</template></el-table-column>
              </el-table>
            </el-col>
          </el-row>
          <div style="font-weight:600;margin-bottom:8px">新增会员趋势（近12个月）</div>
          <el-table :data="custTrend" stripe border size="small">
            <el-table-column label="月份" prop="month" min-width="80"/>
            <el-table-column label="新增" prop="new_customers" width="60" align="center"/>
            <el-table-column label="转化消费" prop="converted_customers" width="80" align="center"/>
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 充值弹窗 -->
    <RechargeDialog v-model:visible="rechargeVisible" :customer="rechargeTarget" @success="onRechargeSuccess"/>

    <!-- 消费记录弹窗 -->
    <el-dialog v-model="consumeDialog" :title="'📖 消费记录 — '+ (consumeTarget?.name||'')" width="750px">
      <div v-if="consumeOrders.length" style="margin-bottom:8px;color:#909399;font-size:13px">共 {{ consumeOrders.length }} 笔，合计 ¥{{ consumeTotal.toFixed(2) }}</div>
      <el-table :data="consumeOrders" v-loading="consumeLoading" stripe border size="small" max-height="400">
        <el-table-column label="日期" width="90"><template #default="{row}">{{ fmtDate(row.created_at) }}</template></el-table-column>
        <el-table-column label="单号" prop="order_no" width="130"/>
        <el-table-column label="项目" min-width="110"><template #default="{row}"><span v-if="row.items?.length">{{ row.items.map((i:any)=>i.item_name+(i.quantity>1?'×'+i.quantity:'')).join('、') }}</span><span v-else>—</span></template></el-table-column>
        <el-table-column label="金额" width="80" align="right"><template #default="{row}">¥{{ (row.final_amount||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="状态" width="60"><template #default="{row}"><el-tag :type="row.status==='paid'?'success':row.status==='refunded'?'info':'warning'" size="small">{{ row.status==='paid'?'已付':row.status==='refunded'?'已退款':'待付' }}</el-tag></template></el-table-column>
      </el-table>
    </el-dialog>

    <!-- 套餐弹窗 -->
    <el-dialog v-model="pkgDialog" :title="'📦 套餐管理 — '+ (pkgTarget?.name||'')" width="700px">
      <el-table :data="pkgList" v-loading="pkgLoading" stripe border size="small">
        <el-table-column label="套餐名" prop="template_name" min-width="100"/>
        <el-table-column label="总次数" prop="total_sessions" width="60" align="center"/>
        <el-table-column label="已用" prop="used_sessions" width="50" align="center"/>
        <el-table-column label="剩余" width="60" align="center"><template #default="{row}">{{ (row.total_sessions||0)-(row.used_sessions||0) }}</template></el-table-column>
        <el-table-column label="到期日" width="80"><template #default="{row}">{{ row.expires_at ? row.expires_at.slice(0,10) : '—' }}</template></el-table-column>
        <el-table-column label="状态" width="70"><template #default="{row}"><el-tag :type="pkgStatusType(row)" size="small">{{ pkgStatusLabel(row) }}</el-tag></template></el-table-column>
      </el-table>
    </el-dialog>

    
    <!-- 添加会员弹窗 -->
    <el-dialog v-model="createDialog" title="添加会员" width="450px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="姓名"><el-input v-model="createForm.name" placeholder="请输入姓名"/></el-form-item>
        <el-form-item label="手机号"><el-input v-model="createForm.phone" placeholder="请输入手机号"/></el-form-item>
        <el-form-item label="性别"><el-select v-model="createForm.gender" style="width:100%"><el-option label="男" :value="1"/><el-option label="女" :value="2"/></el-select></el-form-item>
        <el-form-item label="等级"><el-select v-model="createForm.level" style="width:100%"><el-option v-for="l in levels" :key="l.val" :label="l.label" :value="l.val"/></el-select></el-form-item>
        <el-form-item label="来源"><el-select v-model="createForm.source" style="width:100%"><el-option label="到店" value="walk_in"/><el-option label="转介绍" value="referral"/><el-option label="小红书" value="xiaohongshu"/><el-option label="微信" value="wechat"/><el-option label="抖音" value="douyin"/></el-select></el-form-item>
        <el-form-item label="生日"><el-date-picker v-model="createForm.birthday" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="createDialog=false">取消</el-button><el-button type="primary" :loading="createSaving" @click="submitCreate">确认添加</el-button></template>
    </el-dialog>

    <!-- 编辑会员弹窗 -->
    <el-dialog v-model="editDialog" title="编辑会员" width="450px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="姓名"><el-input v-model="editForm.name"/></el-form-item>
        <el-form-item label="手机号"><el-input v-model="editForm.phone"/></el-form-item>
        <el-form-item label="生日"><el-date-picker v-model="editForm.birthday" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
        <el-form-item label="等级"><el-select v-model="editForm.level" style="width:100%"><el-option v-for="l in levels" :key="l.val" :label="l.label" :value="l.val"/></el-select></el-form-item>
        <el-form-item label="来源"><el-select v-model="editForm.source" style="width:100%"><el-option label="到店" value="walk_in"/><el-option label="转介绍" value="referral"/><el-option label="小红书" value="xiaohongshu"/><el-option label="微信" value="wechat"/><el-option label="抖音" value="douyin"/><el-option label="大众点评" value="dianping"/></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="editDialog=false">取消</el-button><el-button type="primary" :loading="editSaving" @click="saveEdit">保存</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'
import RechargeDialog from './Recharge.vue'

const tab=ref('list')
const search=ref('');const levelFilter=ref('');const loading=ref(false);const members=ref<any[]>([])

const levels=[
  {val:'regular',label:'普通会员'},{val:'silver',label:'白银会员'},
  {val:'gold',label:'黄金会员'},{val:'platinum',label:'铂金会员'},{val:'diamond',label:'钻石会员'},
]

const levelMap:Record<string,string>={regular:'普通',silver:'白银',gold:'黄金',platinum:'铂金',diamond:'钻石'}
const tagMap:Record<string,string>={regular:'info',silver:'',gold:'warning',platinum:'',diamond:'danger'}
function levelLabel(l?:string){return l?levelMap[l]||l:'未知'}
function levelTag(l?:string){return l?tagMap[l]||'info':'info'}


// 添加会员
const createDialog=ref(false);const createSaving=ref(false)
const createForm=ref({name:'',phone:'',gender:0,level:'regular',source:'',birthday:''})
function openCreate(){createForm.value={name:'',phone:'',gender:0,level:'regular',source:'',birthday:''};createDialog.value=true}
async function submitCreate(){
  if(!createForm.value.name||!createForm.value.phone){ElMessage.warning('请填写姓名和手机号');return}
  createSaving.value=true
  try{
    const cr=await api.post('/customers',{name:createForm.value.name,phone:createForm.value.phone,gender:createForm.value.gender,source:createForm.value.source,birthday:createForm.value.birthday||null})
    await api.post('/customers/'+cr.data.id+'/membership',{level:createForm.value.level})
    ElMessage.success('添加成功');createDialog.value=false;await loadMembers()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
  finally{createSaving.value=false}
}

async function loadMembers(){
  loading.value=true
  try{
    const params:any={page_size:200}
    if(search.value)params.keyword=search.value
    const r=await api.get('/customers',{params});const d=r.data?.data||[]
    members.value=d.filter((c:any)=>c.membership!=null)
      .filter((c:any)=>!levelFilter.value||c.membership?.level===levelFilter.value)
      .map((c:any)=>({...c,order_count:0}))
    // Get order counts for each member
    const o=await api.get('/orders',{params:{page_size:1000}});const ol=Array.isArray(o.data)?o.data:o.data?.data||[]
    const oc:Record<number,number>={};ol.forEach((od:any)=>{if(od.customer_id)oc[od.customer_id]=(oc[od.customer_id]||0)+1})
    members.value=members.value.map((m:any)=>({...m,order_count:oc[m.id]||0}))
  }catch{members.value=[]}
  finally{loading.value=false}
}

// 统计
const stats=computed(()=>{
  const m=members.value;const now=new Date();const ym=now.getFullYear()+'-'+String(now.getMonth()+1).padStart(2,'0')
  return{
    total:m.length,
    newMonth:m.filter((c:any)=>c.created_at?.startsWith(ym)).length,
    balance:m.reduce((s:number,c:any)=>s+(c.membership?.balance||0),0),
    recharged:m.reduce((s:number,c:any)=>s+(c.membership?.total_recharged||0),0),
    consumed:m.reduce((s:number,c:any)=>s+(c.membership?.total_consumed||0),0),
  }
})

// 等级分布
const levelDist=computed(()=>{
  const m=members.value;const map:Record<string,number>={};m.forEach((c:any)=>{const l=c.membership?.level||'regular';map[l]=(map[l]||0)+1})
  const total=m.length;return levels.map(({val,label})=>({level:val,label,count:map[val]||0,pct:total>0?Math.round(((map[val]||0)/total)*100):0}))
})

// 来源分布
const srcLabels:Record<string,string>={walk_in:'到店',referral:'转介绍',xiaohongshu:'小红书',wechat:'微信',douyin:'抖音',dianping:'大众点评',other:'其他'}
function srcLabel(s:string){return srcLabels[s]||s||'其他'}
const sourceDist=computed(()=>{
  const m=members.value;const map:Record<string,number>={};m.forEach((c:any)=>{const s=c.source||'other';map[s]=(map[s]||0)+1})
  const total=m.length;return Object.entries(map).map(([source,count])=>({source,count,pct:total>0?Math.round((count/total)*100):0})).sort((a,b)=>b.count-a.count)
})

// 充值
const rechargeVisible=ref(false);const rechargeTarget=ref<any>(null)
function openRecharge(row:any){rechargeTarget.value=row;rechargeVisible.value=true}
function onRechargeSuccess(){rechargeVisible.value=false;loadMembers()}

// 消费
const consumeDialog=ref(false);const consumeTarget=ref<any>(null);const consumeOrders=ref<any[]>([]);const consumeLoading=ref(false)
const consumeTotal=computed(()=>consumeOrders.value.reduce((s:number,r:any)=>s+(r.final_amount||0),0))
function fmtDate(ts:string){if(!ts)return'—';const d=new Date(ts);return d.getFullYear()+'-'+String(d.getMonth()+1).padStart(2,'0')+'-'+String(d.getDate()).padStart(2,'0')}
async function openConsume(row:any){
  consumeTarget.value=row;consumeOrders.value=[];consumeDialog.value=true;consumeLoading.value=true
  try{const r=await api.get('/orders',{params:{customer_id:row.id}});const d=r.data;consumeOrders.value=Array.isArray(d)?d:d?.data||d?.list||[]}catch{}
  finally{consumeLoading.value=false}
}

// 套餐
const pkgDialog=ref(false);const pkgTarget=ref<any>(null);const pkgList=ref<any[]>([]);const pkgLoading=ref(false)
function pkgStatusType(row:any){
  if(row.status==='expired')return'info'
  if(row.expires_at&&new Date(row.expires_at).getTime()<Date.now())return'info'
  if(row.expires_at&&(new Date(row.expires_at).getTime()-Date.now())<30*86400000)return'warning'
  if((row.total_sessions||0)-(row.used_sessions||0)<=0)return'info'
  return'success'
}
function pkgStatusLabel(row:any){
  if((row.total_sessions||0)-(row.used_sessions||0)<=0)return'已用完'
  if(row.expires_at&&new Date(row.expires_at).getTime()<Date.now())return'已过期'
  if(row.expires_at&&(new Date(row.expires_at).getTime()-Date.now())<30*86400000)return'即将到期'
  return'使用中'
}
async function openPackages(row:any){
  pkgTarget.value=row;pkgList.value=[];pkgDialog.value=true;pkgLoading.value=true
  try{const r=await api.get('/customers/'+row.id+'/packages');pkgList.value=Array.isArray(r.data)?r.data:r.data?.data||[]}catch{}
  finally{pkgLoading.value=false}
}

// 编辑
const editDialog=ref(false);const editForm=ref<any>({});const editSaving=ref(false)
function openEdit(row:any){editForm.value={id:row.id,name:row.name,phone:row.phone,birthday:row.birthday,level:row.membership?.level||'regular',source:row.source};editDialog.value=true}
async function saveEdit(){
  editSaving.value=true
  try{
    await api.put('/customers/'+editForm.value.id,{name:editForm.value.name,phone:editForm.value.phone,birthday:editForm.value.birthday,source:editForm.value.source})
    if(editForm.value.level)await api.post('/customers/'+editForm.value.id+'/membership',{level:editForm.value.level})
    ElMessage.success('保存成功');editDialog.value=false;await loadMembers()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'保存失败')}
  finally{editSaving.value=false}
}

// 生日
const ldBirthday=ref(false);const birthdayList=ref<any[]>([])
async function loadBirthday(){
  ldBirthday.value=true
  try{const r=await api.get('/marketing/birthday');birthdayList.value=Array.isArray(r.data)?r.data:r.data?.items||[]}catch{}
  finally{ldBirthday.value=false}
}

// 沉睡
const ldDormant=ref(false);const dormantList=ref<any[]>([]);const nearDormant=ref<any[]>([])
async function loadDormant(){
  ldDormant.value=true
  try{
    const [d,nd]=await Promise.all([api.get('/marketing/dormant'),api.get('/marketing/near-dormant')])
    dormantList.value=Array.isArray(d.data)?d.data:d.data?.items||[]
    nearDormant.value=Array.isArray(nd.data)?nd.data:nd.data?.items||[]
  }catch{}
  finally{ldDormant.value=false}
}

// 趋势
const custTrend=ref<any[]>([])
async function loadTrend(){try{const r=await api.get('/reports/customer-trend');custTrend.value=Array.isArray(r.data)?r.data:[]}catch{}}

onMounted(()=>{loadMembers();loadBirthday();loadDormant();loadTrend()})
</script>