<template>
  <div class="pos-page">

    <el-row :gutter="16">
      <!-- Left: order composition -->
      <el-col :span="14">
        <!-- Customer -->
        <el-card shadow="never" class="sc">
          <template #header><span style="font-weight:600">客户信息</span></template>
          <el-select v-model="custId" filterable remote reserve-keyword placeholder="搜索姓名或手机号/后4位" :remote-method="searchCust" :loading="custLoading" style="width:100%" @change="onCustChange">
            <el-option v-for="c in custList" :key="c.id" :label="c.name+' ('+(c.phone||'无')+')'" :value="c.id"/>
          </el-select>
          <div v-if="custInfo.id" style="margin-top:8px;display:flex;gap:8px;font-size:12px;color:#909399;flex-wrap:wrap">
            <el-tag size="small" type="success">{{ custInfo.level||'普通' }}</el-tag>
            <span>消费 <b>¥{{ fmt(custInfo.total_spent) }}</b></span>
            <span>余额 <b>¥{{ fmt(custInfo.balance) }}</b></span>
            <span>赠金 <b>¥{{ fmt(custInfo.gift_balance) }}</b></span>
            <span>套餐 <b>{{ custInfo.pkgCount||0 }}</b></span>
            <el-button size="small" text type="primary" @click="showRecharge=true">充值</el-button>
            <el-button size="small" text type="success" @click="showBuyPkg=true">购套餐</el-button>
          </div>
        </el-card>

        <!-- Items -->
        <el-card shadow="never" class="sc">
          <template #header><div style="display:flex;align-items:center;justify-content:space-between"><span style="font-weight:600">治疗项目</span><el-button type="primary" size="small" @click="showPicker=true">+ 添加</el-button></div></template>
          <el-table :data="items" stripe size="small" empty-text="尚未添加项目">
            <el-table-column label="项目" min-width="120"><template #default="{row}">{{ row.item_name }}</template></el-table-column>
            <el-table-column label="单价" width="110"><template #default="{row}"><el-input-number v-model="row.unit_price" :min="0" :precision="2" size="small" controls-position="right" style="width:105px"/></template></el-table-column>
            <el-table-column label="数量" width="90"><template #default="{row}"><el-input-number v-model="row.quantity" :min="1" :max="99" size="small" controls-position="right" style="width:75px"/></template></el-table-column>
            <el-table-column label="小计" width="80"><template #default="{row}">¥{{ fmt(row.unit_price*row.quantity) }}</template></el-table-column>
            <el-table-column label="操作" width="55"><template #default="{$index: i}"><el-button text type="danger" size="small" @click="items.splice(i,1)">删除</el-button></template></el-table-column>
          </el-table>
        </el-card>

        <!-- Discount + Rounding + Notes -->
        <el-card shadow="never" class="sc">
          <template #header><span style="font-weight:600">优惠 & 备注</span></template>
          <div style="display:flex;align-items:center;gap:16px;margin-bottom:8px;flex-wrap:wrap">
            <div><span style="font-size:13px;color:#909399;margin-right:4px">优惠</span>
              <el-input-number v-model="discount" :min="0" :precision="2" :max="subtotal" size="small" controls-position="right" style="width:130px"/>
            </div>
            <div><span style="font-size:13px;color:#909399;margin-right:4px">抹零</span>
              <el-switch v-model="rounding" size="small"/>
              <span v-if="rounding" style="font-size:12px;color:#e6a23c;margin-left:4px">去分角</span>
            </div>
            <div><span style="font-size:13px;color:#f56c6c;margin-right:4px">分成</span>
              <el-input-number v-model="commission" :min="0" :precision="2" size="small" controls-position="right" style="width:110px"/>
            </div>
            <div><span style="font-size:13px;color:#f56c6c;margin-right:4px">耗材</span>
              <el-input-number v-model="cost" :min="0" :precision="2" size="small" controls-position="right" style="width:110px"/>
            </div>
          </div>
          <el-input v-model="notes" placeholder="订单备注（选填）" size="small" maxlength="200" show-word-limit/>
        </el-card>
      </el-col>

      <!-- Right: settlement -->
      <el-col :span="10">
        <el-card shadow="never" class="sc">
          <template #header><span style="font-weight:600">结算</span></template>

          <div class="sr"><span>项目合计</span><span>¥{{ fmt(subtotal) }}</span></div>
          <div v-if="discount>0" class="sr" style="color:#e6a23c"><span>优惠</span><span>-¥{{ fmt(discount) }}</span></div>
          <div v-if="rounding && roundOff>0" class="sr" style="color:#909399"><span>抹零</span><span>-¥{{ fmt(roundOff) }}</span></div>
          <div class="sr st"><span>应付金额</span><span>¥{{ fmt(payable) }}</span></div>

          <el-divider style="margin:8px 0"/>

          <!-- Package deduction -->
          <div style="margin-bottom:8px">
            <div style="font-weight:600;font-size:13px;margin-bottom:4px">划扣套餐</div>
            <el-select v-model="dedPkg" placeholder="选择套餐" style="width:100%;margin-bottom:4px" :disabled="!custId" @change="onPkgSel">
              <el-option v-for="p in pkgList" :key="p.id" :label="p.package_name" :value="p.id"/>
            </el-select>
            <el-select v-if="dedPkg" v-model="dedItem" placeholder="选择项目" style="width:100%;margin-bottom:4px">
              <el-option v-for="pi in dedItems" :key="pi.id" :label="pi.item_name+' (余'+pi.remaining_count+'次)'" :value="pi.id"/>
            </el-select>
            <el-button size="small" :disabled="!dedItem" @click="applyDed">划扣</el-button>
            <div v-if="deds.length" style="margin-top:4px">
              <el-tag v-for="(d,i) in deds" :key="i" closable size="small" style="margin-right:4px;margin-bottom:4px" @close="deds.splice(i,1)">{{ d.pkg }} - {{ d.item }}</el-tag>
            </div>
          </div>

          <el-divider style="margin:8px 0"/>

          <!-- Quick amounts -->
          <div style="margin-bottom:6px">
            <div style="font-weight:600;font-size:13px;margin-bottom:4px">快捷金额</div>
            <div style="display:flex;gap:4px;flex-wrap:wrap">
              <el-button v-for="a in [500,1000,2000,5000]" :key="a" size="small" @click="fillAmount(a)">{{ a>=1000?(a/1000)+'k':'¥'+a }}</el-button>
              <el-button size="small" @click="fillRemain">填剩余</el-button>
            </div>
          </div>

          <!-- Payments -->
          <div>
            <div style="font-weight:600;font-size:13px;margin-bottom:4px">支付方式</div>
            <div v-for="(pm,i) in pays" :key="i" style="display:flex;gap:4px;margin-bottom:4px;align-items:center">
              <el-select v-model="pm.method" style="width:110px" size="small">
                <el-option label="余额" value="balance"/><el-option label="赠金" value="gift_balance"/>
                <el-option label="微信" value="wechat"/><el-option label="支付宝" value="alipay"/>
                <el-option label="现金" value="cash"/><el-option label="银行卡" value="bank_card"/>
              </el-select>
              <el-input-number v-model="pm.amount" :min="0" :precision="2" size="small" controls-position="right" style="width:120px"/>
              <el-button text type="danger" size="small" :disabled="pays.length===1" @click="pays.splice(i,1)">删除</el-button>
            </div>
            <el-button size="small" @click="pays.push({method:'wechat',amount:0})">+ 添加</el-button>
            <div :style="{marginTop:'4px',fontWeight:600,fontSize:'13px',color:remain>0?'#e6a23c':'#67c23a'}">剩余: ¥{{ fmt(remain) }}</div>
          </div>

          <el-button type="primary" size="large" style="width:100%;margin-top:10px" :loading="checking" :disabled="!custId||items.length===0||remain>0" @click="doCheckout">结 账</el-button>

          <div style="display:flex;gap:8px;margin-top:8px">
            <el-button size="small" style="flex:1" :disabled="!custId||!items.length" @click="hold">挂单</el-button>
            <el-button size="small" style="flex:1" @click="showHeld=!showHeld">{{ showHeld?'隐藏':'挂单('+helds.length+')' }}</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Today's orders -->
    <el-card shadow="never" class="sc" style="margin-top:16px">
      <template #header><div style="display:flex;align-items:center;justify-content:space-between"><span style="font-weight:600">今日订单 ({{ orders.length }})</span><el-button size="small" text @click="exportOrders">导出</el-button></div></template>
      <el-table :data="orders" stripe size="small" empty-text="暂无今日订单" @row-click="goDetail">
        <el-table-column prop="order_no" label="单号" width="165"/>
        <el-table-column prop="customer_name" label="客户" width="90"/>
        <el-table-column label="金额" width="80"><template #default="{row: r}">¥{{ fmt(r.total_amount) }}</template></el-table-column>
        <el-table-column label="已付" width="80"><template #default="{row: r}">¥{{ fmt(r.paid_amount) }}</template></el-table-column>
        <el-table-column label="状态" width="70"><template #default="{row: r}"><el-tag :type="stTag(r.status)" size="small">{{ stLabel(r.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="created_at" label="时间" width="145"/>
        <el-table-column label="操作" width="95" fixed="right">
          <template #default="{row: r}">
            <el-button text size="small" type="primary" @click.stop="goDetail(r)">详情</el-button>
            <el-button v-if="r.status==='paid'" text size="small" type="danger" @click.stop="refund(r)">退款</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Held orders -->
    <el-card v-if="showHeld&&helds.length" shadow="never" class="sc" style="margin-top:12px">
      <template #header><span style="font-weight:600">挂单 ({{ helds.length }})</span></template>
      <el-table :data="helds" stripe size="small">
        <el-table-column label="客户" width="90"><template #default="{row: r}">{{ r.cust||'-' }}</template></el-table-column>
        <el-table-column label="项目" width="60"><template #default="{row: r}">{{ r.items.length }}</template></el-table-column>
        <el-table-column label="合计" width="80"><template #default="{row: r}">¥{{ r.subtotal }}</template></el-table-column>
        <el-table-column label="时间" width="150"><template #default="{row: r}">{{ r.at }}</template></el-table-column>
        <el-table-column label="操作" width="110">
          <template #default="{$index: i}">
            <el-button size="small" text type="primary" @click="pickup(i)">取单</el-button>
            <el-button size="small" text type="danger" @click="helds.splice(i,1);saveHeld()">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Item picker -->
    <el-dialog v-model="showPicker" title="选择治疗项目" width="550px">
      <el-table ref="pickerRef" :data="treatItems" stripe size="small">
        <el-table-column type="selection" width="45"/>
        <el-table-column prop="name" label="项目" min-width="130"/>
        <el-table-column label="价格" width="100"><template #default="{row: r}">¥{{ fmt(r.default_price) }}</template></el-table-column>
      </el-table>
      <template #footer><el-button @click="showPicker=false">取消</el-button><el-button type="primary" @click="confirmPick">确定</el-button></template>
    </el-dialog>

    <!-- Recharge -->
    <el-dialog v-model="showRecharge" title="会员充值" width="400px">
      <div style="margin-bottom:10px;font-size:13px;color:#909399">客户: {{ custInfo.name }} | 余额: ¥{{ fmt(custInfo.balance) }} | 赠金: ¥{{ fmt(custInfo.gift_balance) }}</div>
      <el-form label-width="80px" size="small">
        <el-form-item label="充值金额"><el-input-number v-model="rechAmt" :min="1" :precision="2" style="width:200px"/></el-form-item>
        <el-form-item label="赠送金额"><el-input-number v-model="rechGift" :min="0" :precision="2" style="width:200px"/></el-form-item>
        <el-form-item label="支付方式"><el-select v-model="rechPay" style="width:200px"><el-option label="微信" value="wechat"/><el-option label="支付宝" value="alipay"/><el-option label="现金" value="cash"/><el-option label="银行卡" value="bank_card"/></el-select></el-form-item>
      </el-form>
      <template #footer><el-button @click="showRecharge=false">取消</el-button><el-button type="primary" :loading="rechLoading" :disabled="!rechAmt" @click="doRecharge">确认充值</el-button></template>
    </el-dialog>

    <!-- Buy Package -->
    <el-dialog v-model="showBuyPkg" title="购买套餐" width="480px">
      <div style="margin-bottom:10px;font-size:13px;color:#909399">客户: {{ custInfo.name }}</div>
      <el-form label-width="80px" size="small">
        <el-form-item label="套餐"><el-select v-model="buyTmpl" filterable placeholder="选择套餐模板" style="width:100%"><el-option v-for="t in tmpls" :key="t.id" :label="t.name+(t.price?' (¥'+t.price+')':'')" :value="t"/></el-select></el-form-item>
        <el-form-item v-if="buyTmpl" label="价格"><el-input-number v-model="buyPrice" :min="0" :precision="2" style="width:200px"/></el-form-item>
        <el-form-item v-if="buyTmpl" label="支付">
          <el-select v-model="buyPay" style="width:200px">
            <el-option label="微信" value="wechat"/><el-option label="支付宝" value="alipay"/>
            <el-option label="余额" value="balance"/><el-option label="现金" value="cash"/><el-option label="银行卡" value="bank_card"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="showBuyPkg=false">取消</el-button><el-button type="primary" :loading="buyLoading" :disabled="!buyTmpl" @click="doBuyPkg">确认购买</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const router = useRouter()
function fmt(v:any){return (v||0).toFixed(2)}
function stTag(s:string){return s==='paid'?'success':s==='partial'?'warning':s==='refunded'?'info':'danger'}
function stLabel(s:string){return s==='paid'?'已付':s==='partial'?'部分':s==='refunded'?'已退':'待付'}

// ===== Stats =====
const profitData=ref<any>({})
const dayStats=ref({orders:0,avgTicket:'0.00',refunds:0})
async function loadStats(){
  try{const r=await api.get('/reports/profit',{params:{date:todayStr()}});profitData.value=r.data||{}}catch{}
  try{
    const r=await api.get('/reports/daily-sales');const d=r.data||{}
    dayStats.value={orders:d.order_count||0,avgTicket:(d.avg_ticket||0).toFixed(2),refunds:d.refund_count||0}
  }catch{}
}

// ===== Customer =====
const custId=ref<number|null>(null)
const custList=ref<any[]>([])
const custLoading=ref(false)
const custInfo=ref<any>({})

async function searchCust(q:string){
  if(!q){custList.value=[];return}
  custLoading.value=true
  try{const r=await api.get('/customers',{params:{q,size:20}});custList.value=r.data?.data||r.data||[]}catch{}
  finally{custLoading.value=false}
}

async function loadCustInfo(id:number){
  try{
    const cr=await api.get('/customers/'+id);const c=cr.data||{}
    let pkgCount=0,totalSpent=0,balance=0,gift=0,level='普通会员'
    try{const mr=await api.get('/customers/'+id+'/membership');const m=mr.data||{};level=m.level||'普通';balance=m.balance||0;gift=m.gift_balance||0}catch{}
    try{const pr=await api.get('/customers/'+id+'/packages');const pkgs=pr.data?.data||pr.data||[];pkgCount=pkgs.length}catch{}
    custInfo.value={id:c.id,name:c.name,phone:c.phone,level,balance,gift,total_spent:totalSpent,pkgCount}
  }catch{custInfo.value={}}
}

async function onCustChange(){
  dedPkg.value=null;dedItem.value=null;deds.value=[];pkgList.value=[]
  if(custId.value){await Promise.all([loadCustInfo(custId.value),loadPkgList()])}else{custInfo.value={}}
}

// ===== Items =====
const treatItems=ref<any[]>([])
const items=ref<any[]>([])
const showPicker=ref(false)
const pickerRef=ref()

async function loadItems(){
  try{const r=await api.get('/settings/items');treatItems.value=r.data?.data||r.data||[]}catch{treatItems.value=[]}
}
function confirmPick(){
  const ref=pickerRef.value;if(!ref)return
  for(const item of ref.getSelectionRows()){
    const ex=items.value.find(i=>i.treatment_item_id===item.id)
    if(ex){ex.quantity++}else{items.value.push({treatment_item_id:item.id,item_name:item.name,unit_price:item.default_price||0,quantity:1})}
  }
  showPicker.value=false
}

// ===== Discount / Rounding / Notes =====
const discount=ref(0)
const commission=ref(0)
const cost=ref(0)
const rounding=ref(false)
const notes=ref('')
const subtotal=computed(()=>items.value.reduce((s,i)=>s+i.unit_price*i.quantity,0))
const dedTotal=computed(()=>deds.value.reduce((s,d)=>s+d.amt,0))
const rawPayable=computed(()=>Math.max(0,subtotal.value-discount.value-dedTotal.value))
const roundOff=computed(()=>rounding.value?rawPayable.value-Math.floor(rawPayable.value):0)
const payable=computed(()=>rawPayable.value-roundOff.value)

// ===== Payments =====
const pays=ref<{method:string;amount:number}[]>([{method:'cash',amount:0}])
const payTotal=computed(()=>pays.value.reduce((s,p)=>s+(p.amount||0),0))
const remain=computed(()=>Math.max(0,payable.value-payTotal.value))

function fillAmount(a:number){if(pays.value.length)pays.value[0].amount=Math.min(a,payable.value)}
function fillRemain(){if(pays.value.length)pays.value[0].amount=payable.value-payTotal.value+pays.value[0].amount}

// ===== Package =====
const pkgList=ref<any[]>([])
const dedPkg=ref<number|null>(null)
const dedItem=ref<number|null>(null)
const dedItems=ref<any[]>([])
const deds=ref<{pkg:string;item:string;amt:number}[]>([])

async function loadPkgList(){
  if(!custId.value)return
  try{const r=await api.get('/customers/'+custId.value+'/packages');pkgList.value=r.data?.data||r.data||[]}catch{pkgList.value=[]}
}
function onPkgSel(id:number){
  const p=pkgList.value.find(x=>x.id===id)
  dedItems.value=p?.items||[];dedItem.value=null
}
function applyDed(){
  if(!dedPkg.value||!dedItem.value)return
  const p=pkgList.value.find(x=>x.id===dedPkg.value)
  const it=dedItems.value.find((x:any)=>x.id===dedItem.value)
  if(!p||!it)return
  const oi=items.value.find(x=>x.treatment_item_id===it.item_id)
  deds.value.push({pkg_id:dedPkg.value,item_id:dedItem.value,pkg:p.package_name,item:it.item_name,amt:oi?oi.unit_price:0})
  dedPkg.value=null;dedItem.value=null;dedItems.value=[]
}

// ===== Orders =====
const orders=ref<any[]>([])
async function loadOrders(){
  try{const r=await api.get('/orders',{params:{date:todayStr()}});orders.value=r.data?.data||r.data||[]}catch{orders.value=[]}
}
function goDetail(r:any){router.push('/pos/'+r.id)}

async function refund(r:any){
  if(r.status!=='paid'){ElMessage.warning('只能退已付款订单');return}
  try{
    await ElMessageBox.confirm('确定退单 '+r.order_no+' (¥'+fmt(r.total_amount)+')?','退款确认',{type:'warning'})
    await api.post('/orders/'+r.id+'/refund')
    ElMessage.success('已退款')
    await Promise.all([loadOrders(),loadStats()])
  }catch(e:any){if(e!=='cancel')ElMessage.error(e?.response?.data?.error||'退款失败')}
}

function exportOrders(){
  if(!orders.value.length)return
  const headers=['订单号','客户','金额','已付','状态','时间']
  const rows=orders.value.map((r:any)=>[r.order_no||'',r.customer_name||'',fmt(r.total_amount),fmt(r.paid_amount),stLabel(r.status),r.created_at||''].map(v=>'"'+v+'"').join(',')).join('\n')
  const blob=new Blob(['\uFEFF'+headers.map(h=>'"'+h+'"').join(',')+'\n'+rows],{type:'text/csv;charset=utf-8'})
  const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download='今日订单_'+new Date().toISOString().slice(0,10)+'.csv'
  a.click();URL.revokeObjectURL(a.href)
}

// ===== Held Orders =====
const HELD_KEY='pos_held_orders'
const showHeld=ref(false)
const helds=ref<any[]>([])

function loadHeld(){try{const raw=localStorage.getItem(HELD_KEY);helds.value=raw?JSON.parse(raw):[]}catch{helds.value=[]}}
function saveHeld(){localStorage.setItem(HELD_KEY,JSON.stringify(helds.value))}

function hold(){
  if(!items.value.length)return
  helds.value.push({id:Date.now(),cust:custInfo.value.name||'',items:JSON.parse(JSON.stringify(items.value)),discount:discount.value,notes:notes.value,subtotal:subtotal.value.toFixed(2),at:new Date().toLocaleString()})
  saveHeld();reset();ElMessage.success('已挂单')
}
function pickup(i:number){
  const h=helds.value[i];if(!h)return
  if(items.value.length){ElMessage.warning('请先清空当前订单');return}
  if(h.cust)custId.value=null;items.value=h.items||[];discount.value=h.discount||0;notes.value=h.notes||''
  helds.value.splice(i,1);saveHeld();ElMessage.success('已取单')
}

// ===== Checkout =====
const checking=ref(false)

async function doCheckout(){
  if(!custId.value||!items.value.length)return
  if(remain.value>0){ElMessage.warning('尚有 ¥'+fmt(remain.value)+' 未分配');return}
  checking.value=true
  try{
    const payload:any={customer_id:custId.value,items:items.value.map(i=>({item_id:i.treatment_item_id,item_name:i.item_name,unit_price:i.unit_price,quantity:i.quantity})),discount_amount:discount.value,commission_amount:commission.value,cost_amount:cost.value,total_amount:payable.value+payTotal.value,remark:notes.value||''}
    const order=await api.post('/orders',payload)
    const oid=order.data?.id||order.id
    for(const d of deds.value){await api.post('/orders/'+oid+'/pay/package',{package_id:d.pkg_id,package_item_id:d.item_id})}
    const vp=pays.value.filter(p=>p.method&&p.amount>0)
    if(vp.length)await api.post('/orders/'+oid+'/pay',{payments:vp.map(p=>({pay_method:p.method,amount:p.amount}))})
    ElMessage.success('结账成功');reset();await Promise.all([loadOrders(),loadStats()])
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'结账失败')}
  finally{checking.value=false}
}

function reset(){
  custId.value=null;items.value=[];discount.value=0;commission.value=0;cost.value=0;notes.value='';rounding.value=false
  pays.value=[{method:'cash',amount:0}];deds.value=[];pkgList.value=[];dedItems.value=[]
  dedPkg.value=null;dedItem.value=null;custInfo.value={}
}

// ===== Recharge =====
const showRecharge=ref(false)
const rechAmt=ref(0)
const rechGift=ref(0)
const rechPay=ref('wechat')
const rechLoading=ref(false)

async function doRecharge(){
  if(!custId.value||!rechAmt.value)return
  rechLoading.value=true
  try{
    await api.post('/customers/'+custId.value+'/recharge',{amount:rechAmt.value,gift_amount:rechGift.value,pay_method:rechPay.value})
    ElMessage.success('充值成功');showRecharge.value=false;rechAmt.value=0;rechGift.value=0
    await loadCustInfo(custId.value)
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'充值失败')}
  finally{rechLoading.value=false}
}

// ===== Buy Package =====
const showBuyPkg=ref(false)
const tmpls=ref<any[]>([])
const buyTmpl=ref<any>(null)
const buyPrice=ref(0)
const buyPay=ref('wechat')
const buyLoading=ref(false)

async function loadTmpls(){
  try{const r=await api.get('/settings/package-templates');tmpls.value=r.data?.data||r.data||[]}catch{tmpls.value=[]}
}
async function doBuyPkg(){
  if(!custId.value||!buyTmpl.value)return
  buyLoading.value=true
  try{
    await api.post('/customers/'+custId.value+'/packages',{package_template_id:buyTmpl.value.id,price:buyPrice.value||buyTmpl.value.price||0,pay_method:buyPay.value})
    ElMessage.success('套餐购买成功');showBuyPkg.value=false;buyTmpl.value=null;buyPrice.value=0
    await Promise.all([loadPkgList(),loadCustInfo(custId.value)])
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'购买失败')}
  finally{buyLoading.value=false}
}

// ===== Init =====
onMounted(async()=>{
  await Promise.all([loadItems(),loadOrders(),loadStats(),loadTmpls()])
  loadHeld()
})
</script>

<style scoped>
.pos-page{max-width:1400px;margin:0 auto}
.sc{margin-bottom:16px}
.sr{display:flex;justify-content:space-between;padding:3px 0;font-size:14px}
.st{font-size:17px;font-weight:700;color:#303133;border-top:1px solid #ebeef5;padding-top:8px;margin-top:2px}
</style>