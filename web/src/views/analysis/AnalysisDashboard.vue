<template>
  <div style="max-width:1100px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:16px">
      <span style="font-weight:600;font-size:15px">📊 深度运营分析</span>
    </el-card>
    <el-tabs v-model="tab">
      <!-- LTV -->
      <el-tab-pane label="客户LTV分析" name="ltv">
        <el-card shadow="never">
          <el-table :data="ltvData" v-loading="loadingLtv" stripe border>
            <el-table-column label="获客渠道" prop="source" min-width="120"/>
            <el-table-column label="客户数" prop="customer_count" width="80" align="center"/>
            <el-table-column label="总收入" width="120" align="right"><template #default="{r}">¥{{ (r.total_revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="人均消费" width="120" align="right"><template #default="{r}">¥{{ (r.avg_revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="人均订单" prop="avg_orders" width="80" align="center"/>
          </el-table>
        </el-card>
      </el-tab-pane>
      <!-- Cross-sell -->
      <el-tab-pane label="交叉销售分析" name="cross">
        <el-card shadow="never">
          <el-table :data="crossData" v-loading="loadingCross" stripe border>
            <el-table-column label="项目A" prop="item_a" min-width="150"/>
            <el-table-column label="项目B" prop="item_b" min-width="150"/>
            <el-table-column label="同时购买次数" prop="count" width="120" align="center"/>
          </el-table>
        </el-card>
      </el-tab-pane>
      <!-- No-show -->
      <el-tab-pane label="预约不到分析" name="noshow">
        <el-card shadow="never">
          <el-table :data="noshowData" v-loading="loadingNoShow" stripe border>
            <el-table-column label="时间段" prop="time_slot" min-width="120"/>
            <el-table-column label="总预约" prop="total" width="80" align="center"/>
            <el-table-column label="失约数" prop="no_show" width="80" align="center"/>
            <el-table-column label="失约率" width="100" align="center"><template #default="{r}">{{ (r.rate||0).toFixed(1) }}%</template></el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
      <!-- Churn -->
      <el-tab-pane label="客户流失预警" name="churn">
        <el-card shadow="never">
          <el-table :data="churnData" v-loading="loadingChurn" stripe border>
            <el-table-column label="客户" prop="name" min-width="100"/>
            <el-table-column label="电话" prop="phone" width="120"/>
            <el-table-column label="最后到店" prop="last_visit" width="100"/>
            <el-table-column label="已离开" width="80" align="center"><template #default="{r}">{{ r.days_gone }} 天</template></el-table-column>
            <el-table-column label="历史消费" width="120" align="right"><template #default="{r}">¥{{ (r.revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="操作" width="100"><template #default="{r}"><el-button size="small" @click="go('/followup')">去回访</el-button></template></el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
      <!-- Profit -->
      <el-tab-pane label="项目毛利率" name="profit">
        <el-card shadow="never">
          <el-table :data="profitData" v-loading="loadingProfit" stripe border :summary-method="profitSummary" show-summary>
            <el-table-column label="项目" prop="procedure" min-width="150"/>
            <el-table-column label="收入" width="120" align="right"><template #default="{r}">¥{{ (r.revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="订单数" prop="order_count" width="70" align="center"/>
            <el-table-column label="预估成本" width="120" align="right"><template #default="{r}">¥{{ (r.est_cost||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="毛利" width="120" align="right"><template #default="{r}"><span :style="{color:(r.margin||0)>=0?'#67c23a':'#f56c6c',fontWeight:600}">¥{{ (r.margin||0).toFixed(2) }}</span></template></el-table-column>
            <el-table-column label="毛利率" width="80" align="center"><template #default="{r}"><el-tag :type="(r.margin_pct||0)>=50?'success':(r.margin_pct||0)>=20?'warning':'danger'" size="small">{{ (r.margin_pct||0).toFixed(1) }}%</el-tag></template></el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../../api'
const router=useRouter()
const tab=ref('ltv')
const loadingLtv=ref(false);const ltvData=ref<any[]>([])
const loadingCross=ref(false);const crossData=ref<any[]>([])
const loadingNoShow=ref(false);const noshowData=ref<any[]>([])
const loadingChurn=ref(false);const churnData=ref<any[]>([])
const loadingProfit=ref(false);const profitData=ref<any[]>([])
function go(p:string){router.push(p)}

function profitSummary({columns,data}:any){
  const s:string[]=[];columns.forEach((c:any,i:number)=>{
    if(i===0)s[i]='合计';
    else if(c.property==='revenue'||c.property==='est_cost'||c.property==='margin')
      s[i]='¥'+data.reduce((a:number,r:any)=>a+Number(r[c.property]||0),0).toFixed(2);
    else s[i]=''
  });return s
}
onMounted(async ()=>{
  loadingLtv.value=true;try{const r=await api.get('/analysis/ltv');ltvData.value=Array.isArray(r.data)?r.data:[]}finally{loadingLtv.value=false}
  loadingCross.value=true;try{const r=await api.get('/analysis/cross-sell');crossData.value=Array.isArray(r.data)?r.data:[]}finally{loadingCross.value=false}
  loadingNoShow.value=true;try{const r=await api.get('/analysis/no-show');noshowData.value=Array.isArray(r.data)?r.data:[]}finally{loadingNoShow.value=false}
  loadingChurn.value=true;try{const r=await api.get('/analysis/churn');churnData.value=Array.isArray(r.data)?r.data:[]}finally{loadingChurn.value=false}
  loadingProfit.value=true;try{const r=await api.get('/analysis/procedure-profit');profitData.value=Array.isArray(r.data)?r.data:[]}finally{loadingProfit.value=false}
})
</script>
