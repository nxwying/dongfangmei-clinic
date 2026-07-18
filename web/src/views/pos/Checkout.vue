<template>
  <div style="max-width:1100px;margin:0 auto">

    <!-- 上半部：选客户 + 项目 -->
    <el-row :gutter="12">
      <el-col :span="14">
        <el-card shadow="never" style="margin-bottom:12px">
          <template #header><b>客户信息</b></template>
          <el-select v-model="customer" filterable remote reserve-keyword
            placeholder="搜索姓名或手机号"
            :remote-method="searchCust" :loading="loadingCust" style="width:100%"
            @change="c=>selectedCustomer=c">
            <el-option v-for="c in custList" :key="c.id"
              :label="`${c.name} (${c.phone||'无'})`" :value="c"/>
          </el-select>
        </el-card>

        <el-card shadow="never" style="margin-bottom:12px">
          <template #header><b>项目明细</b></template>
          <el-table :data="items" size="small" max-height="220" stripe style="width:100%">
            <el-table-column label="项目" min-width="130">
              <template #default="{row}">{{ row.name }}</template>
            </el-table-column>
            <el-table-column label="单价" width="80" align="right">
              <template #default="{row}">¥{{ row.price.toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="数量" width="80" align="center">
              <template #default="{row}">
                <el-input-number v-model="row.qty" :min="1" size="small" style="width:70px"/>
              </template>
            </el-table-column>
            <el-table-column label="小计" width="80" align="right">
              <template #default="{row}">¥{{ (row.price*row.qty).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="50" fixed="right">
              <template #default="{row,$index}">
                <el-button text size="small" type="danger" @click="items.splice($index,1)">✕</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div style="margin-top:8px;display:flex;gap:8px">
            <el-select v-model="addItem" filterable placeholder="+ 添加项目" style="flex:1"
              @change="(v:any)=>{if(v){const ex=items.find(i=>i.id===v.id);if(ex){ex.qty++}else{items.push({...v,qty:1})};addItem=null}}">
              <el-option v-for="i in treatItems" :key="i.id" :label="i.name+(i.price?'（¥'+i.price+'）':'')" :value="i"/>
            </el-select>
          </div>
        </el-card>

        <el-card shadow="never">
          <template #header><b>支付信息</b></template>
          <div style="margin-bottom:8px">
            <div style="display:flex;align-items:center;gap:8px;font-size:13px;color:#909399;">
              折扣 <el-input-number v-model="discount" :min="0" size="small" style="width:100px"/> 元
            </div>
            <div style="display:flex;align-items:center;gap:8px;font-size:13px;color:#909399;margin-top:6px;">
              分成 <el-input-number v-model="commission" :min="0" size="small" style="width:100px" :precision="2"/> 元
              耗材 <el-input-number v-model="cost" :min="0" size="small" style="width:100px" :precision="2"/> 元
            </div>
          </div>
          <el-divider style="margin:10px 0"/>
          <div v-for="(pm,i) in payMethods" :key="i" style="display:flex;gap:8px;margin-bottom:6px;align-items:center">
            <el-select v-model="pm.method" style="width:130px">
              <el-option label="微信" value="wechat"/>
              <el-option label="支付宝" value="alipay"/>
              <el-option label="银行卡" value="card"/>
              <el-option label="余额" value="balance"/>
              <el-option label="赠金" value="gift_balance"/>
              <el-option label="现金" value="cash"/>
            </el-select>
            <el-input-number v-model="pm.amount" :min="0" :precision="2" style="width:130px"/>
            <el-button v-if="payMethods.length>1" text size="small" type="danger" @click="payMethods.splice(i,1)">✕</el-button>
          </div>
          <el-button size="small" @click="payMethods.push({method:'wechat',amount:0})">+ 添加支付方式</el-button>
          <el-divider style="margin:10px 0"/>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <div>
              <span style="color:#909399;font-size:13px">小计：¥{{ subtotal.toFixed(2) }} | 应收：¥{{ receivable.toFixed(2) }} | 已付：¥{{ paid.toFixed(2) }}</span>
              <span v-if="paid>=receivable&&paid>0" style="color:#67c23a;font-weight:600;margin-left:8px">✅ 找零 ¥{{ (paid-receivable).toFixed(2) }}</span>
            </div>
            <client-only>
              <el-button type="primary" :loading="paying" :disabled="!customer||items.length===0||paid<receivable" @click="doCheckout">
                {{ paying?'支付中...':'确认收款' }}
              </el-button>
            </client-only>
          </div>
        </el-card>
      </el-col>

      <el-col :span="10">
        <!-- 下半部：近期订单 -->
        <el-card shadow="never">
          <template #header><b>近期订单</b></template>
          <el-table :data="orders" v-loading="loadingOrder" size="small" stripe>
            <el-table-column prop="order_no" label="单号" width="140"/>
            <el-table-column label="应收" width="80" align="right">
              <template #default="{row}">¥{{(row.final_amount||0).toFixed(2)}}</template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{row}">
                <el-tag :type="row.status==='paid'?'success':row.status==='refunded'?'info':'warning'" size="small">
                  {{row.status==='paid'?'已付':row.status==='refunded'?'已退款':'待付'}}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="客户" min-width="100">
              <template #default="{row}">{{row.customer?.name||'-'}}</template>
            </el-table-column>
            <el-table-column label="操作" width="70" fixed="right">
              <template #default="{row}">
                <el-button v-if="row.status==='paid'" type="danger" size="small" text @click="handleRefund(row)">退款</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

const customer = ref<any>(null)
const custList = ref<any[]>([])
const loadingCust = ref(false)
const treatItems = ref<any[]>([])
const addItem = ref<any>(null)
const items = ref<{id:number;name:string;price:number;qty:number}[]>([])
const discount = ref(0)
const commission = ref(0)
const cost = ref(0)
const payMethods = ref([{method:'wechat',amount:0}])
const orders = ref<any[]>([])
const loadingOrder = ref(false)
const paying = ref(false)

const subtotal = computed(()=>items.value.reduce((s,i)=>s+i.price*i.qty,0))
const receivable = computed(()=>Math.max(subtotal.value-discount.value,0))
const paid = computed(()=>payMethods.value.reduce((s,p)=>s+(p.amount||0),0))

async function searchCust(q:string){
  if(!q||q.length<1){custList.value=[];return}
  loadingCust.value=true
  try{const r=await api.get('/customers',{params:{keyword:q,page_size:10}});custList.value=r.data?.data||r.data||[]}catch{custList.value=[]}
  finally{loadingCust.value=false}
}

async function loadItems(){
  try{const r=await api.get('/settings/items');treatItems.value=r.data||[]}catch{}
}

async function loadOrders(){
  loadingOrder.value=true
  try{const r=await api.get('/orders',{params:{page_size:5}});const d=r.data;orders.value=Array.isArray(d)?d:d?.data||d?.list||[]}
  catch{orders.value=[]}
  finally{loadingOrder.value=false}
}

async function doCheckout(){
  paying.value=true
  try{
    const payload={
      customer_id:customer.value.id,
      discount_amount:discount.value,
      final_amount:receivable.value,
      total_amount:subtotal.value,
      remark:'',
      items:items.value.map(i=>({item_type:'treatment',item_id:i.id,item_name:i.name,quantity:i.qty,unit_price:i.price,subtotal:i.price*i.qty}))
    }
    const orderRes=await api.post('/orders',payload)
    const order=orderRes.data, orderId=order.id||order.ID

    const validPm=payMethods.value.filter(p=>p.amount>0)
    if(validPm.length>0){
      await api.post(`/orders/${orderId}/pay`,{payments:validPm.map(p=>({pay_method:p.method,amount:p.amount}))})
    }

    if(commission.value>0){
      const today=new Date()
      await api.post('/expenses',{type:'commission',category:'分成支出',amount:commission.value,note:`订单${order.order_no||''}分成`,date:`${today.getFullYear()}-${String(today.getMonth()+1).padStart(2,'0')}-${String(today.getDate()).padStart(2,'0')}`})
    }
    if(cost.value>0){
      const today=new Date()
      await api.post('/expenses',{type:'high_value',category:'高值耗材',amount:cost.value,note:`订单${order.order_no||''}成本`,date:`${today.getFullYear()}-${String(today.getMonth()+1).padStart(2,'0')}-${String(today.getDate()).padStart(2,'0')}`})
    }

    ElMessage.success('收款成功')
    items.value=[];discount.value=0;commission.value=0;cost.value=0
    payMethods.value=[{method:'wechat',amount:0}];customer.value=null
    await loadOrders()
  }catch(e:any){
    ElMessage.error(e?.response?.data?.error||e?.message||'操作失败')
  }finally{paying.value=false}
}

// 退款功能 - 独立顶层函数，模板可直接访问
async function handleRefund(order:any){
  if(order.status!=='paid'){
    ElMessage.warning('只有已付款订单才能退款')
    return
  }
  try{
    await ElMessageBox.confirm('确定要退单 '+order.order_no+'（¥'+(order.final_amount||0).toFixed(2)+'）吗？', '退款确认', {confirmButtonText:'确认退款',cancelButtonText:'取消',type:'warning'})
    await api.post('/orders/'+order.id+'/refund')
    ElMessage.success('退款成功')
    await loadOrders()
  }catch(e:any){
    if(e!=='cancel') ElMessage.error(e?.response?.data?.error||'退款失败')
  }
}

onMounted(()=>{loadItems();loadOrders()})
</script>
