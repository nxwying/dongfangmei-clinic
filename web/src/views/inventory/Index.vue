<template>
  <div style="max-width:1100px;margin:0 auto">
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#409eff">{{ totalItems }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总品项</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#67c23a">¥{{ totalValue.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">库存总价值</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#e6a23c">{{ lowStockCount }}</div><div style="font-size:12px;color:#909399;margin-top:2px">低库存预警</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#f56c6c">{{ expiringCount }}</div><div style="font-size:12px;color:#909399;margin-top:2px">临期/过期</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#909399">¥{{ catValue.toFixed(2) }}</div><div style="font-size:12px;color:#909399;margin-top:2px">耗材价值</div></div></el-card></el-col>
    </el-row>

    <!-- 操作栏 + 采购建议 -->
    <el-card shadow="never" style="margin-bottom:12px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">📦 库存管理</span>
        <span style="flex:1"/>
        <el-button size="small" @click="showPurchase=true" :type="lowStockCount>0?'warning':''">
          📥 采购建议 {{ lowStockCount>0?'('+lowStockCount+')':'' }}
        </el-button>
        <el-button size="small" text @click="exportCSV">📤 导出</el-button>
        <el-button type="primary" size="small" @click="openAdd">+ 新增物品</el-button>
      </div>
    </el-card>

    <!-- 采购建议 -->
    <el-card v-if="showPurchase" shadow="never" style="margin-bottom:12px;border-left:4px solid #e6a23c">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px">
        <span style="font-weight:600;color:#e6a23c">📥 采购建议清单</span>
        <el-button size="small" text @click="showPurchase=false">收起 ✕</el-button>
      </div>
      <el-table :data="lowStockItems" stripe border size="small">
        <el-table-column label="物品" prop="name" min-width="100"/>
        <el-table-column label="分类" width="80"><template #default="{row}">{{ catLabel(row.category) }}</template></el-table-column>
        <el-table-column label="当前库存" width="80" align="center"><template #default="{row}">{{ row.quantity }} {{ row.unit }}</template></el-table-column>
        <el-table-column label="最低库存" width="80" align="center" prop="min_stock"/>
        <el-table-column label="建议采购" width="80" align="center"><template #default="{row}">{{ suggestQty(row) }} {{ row.unit }}</template></el-table-column>
        <el-table-column label="供应商" prop="supplier" min-width="80"/>
      </el-table>
    </el-card>

    <!-- 物品列表 -->
    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe size="small">
        <el-table-column label="名称" prop="name" min-width="100"/>
        <el-table-column label="分类" width="80"><template #default="{row}">{{ catLabel(row.category) }}</template></el-table-column>
        <el-table-column label="库存" width="80" align="center">
          <template #default="{row}">
            <span :style="{color:row.quantity<=row.min_stock?'#f56c6c':'#303133',fontWeight:row.quantity<=row.min_stock?'700':'400'}">{{ row.quantity }} {{ row.unit }}</span>
          </template>
        </el-table-column>
        <el-table-column label="最低" width="60" align="center" prop="min_stock"/>
        <el-table-column label="单价" width="85" align="right"><template #default="{row}">¥{{ (row.price||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="价值" width="85" align="right"><template #default="{row}">¥{{ ((row.price||0)*(row.quantity||0)).toFixed(2) }}</template></el-table-column>
        <el-table-column label="供应商" prop="supplier" min-width="80"/>
        <el-table-column label="有效期" width="90">
          <template #default="{row}">
            <span v-if="row.expiry_date" :style="expiryStyle(row.expiry_date)">{{ row.expiry_date }}</span>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="stockIn(row)">入库</el-button>
            <el-button size="small" @click="stockOut(row)">出库</el-button>
            <el-button size="small" text @click="openLogs(row)">流水</el-button>
            <el-button size="small" text @click="openEdit(row)">编辑</el-button>
            <el-button size="small" text @click="viewDocs(row)">证件</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 库存流水弹窗 -->
    <el-dialog v-model="logsDialog" :title="'📋 库存变动流水 — '+ (logsItem?.name||'')" width="650px">
      <el-table :data="logs" v-loading="logsLoading" stripe border size="small" max-height="400">
        <el-table-column label="时间" width="140"><template #default="{row}">{{ fmtDate(row.created_at) }}</template></el-table-column>
        <el-table-column label="类型" width="60" align="center"><template #default="{row}"><el-tag :type="row.type==='in'?'success':'danger'" size="small">{{ row.type==='in'?'入库':'出库' }}</el-tag></template></el-table-column>
        <el-table-column label="数量" width="80" align="center"><template #default="{row}">{{ row.quantity }}</template></el-table-column>
        <el-table-column label="结存" width="80" align="center"><template #default="{row}">{{ row.balance_after }}</template></el-table-column>
        <el-table-column label="备注" min-width="100" prop="note"/>
      </el-table>
      <div v-if="!logsLoading && !logs.length" style="text-align:center;padding:20px;color:#c0c4cc">暂无变动记录</div>
    </el-dialog>

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑物品':'新增物品'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="名称"><el-input v-model="form.name"/></el-form-item>
        <el-form-item label="分类"><el-select v-model="form.category" style="width:100%"><el-option label="药品" value="drug"/><el-option label="高值耗材" value="consumable"/><el-option label="仪器设备" value="instrument"/><el-option label="其他" value="other"/></el-select></el-form-item>
        <el-form-item label="数量"><el-input-number v-model="form.quantity" :min="0" style="width:100%"/></el-form-item>
        <el-form-item label="单位"><el-input v-model="form.unit" placeholder="支/盒/瓶/片"/></el-form-item>
        <el-form-item label="最低库存"><el-input-number v-model="form.min_stock" :min="0" style="width:100%"/></el-form-item>
        <el-form-item label="单价"><el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="供应商"><el-input v-model="form.supplier"/></el-form-item>
        <el-form-item label="有效期"><el-date-picker v-model="form.expiry_date" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="dialogVisible=false">取消</el-button><el-button type="primary" :loading="saving" @click="saveItem">保存</el-button></template>
    </el-dialog>

    <!-- 出入库弹窗 -->
    <el-dialog v-model="stockDialog" :title="stockType==='in'?'入库':'出库'" width="400px">
      <div v-if="stockItem" style="margin-bottom:10px">当前库存：<b>{{ stockItem.quantity }} {{ stockItem.unit }}</b> &nbsp; 物品：<b>{{ stockItem.name }}</b></div>
      <el-form>
        <el-form-item :label="stockType==='in'?'入库数量':'出库数量'"><el-input-number v-model="stockQty" :min="0.01" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="备注"><el-input v-model="stockNote" type="textarea" :rows="2"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="stockDialog=false">取消</el-button><el-button type="primary" :loading="stocking" @click="confirmStock">{{ stockType==='in'?'确认入库':'确认出库'}}</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
const router = useRouter()
import { ElMessage } from 'element-plus'
import api from '../../api'

const loading=ref(false);const items=ref<any[]>([])
const dialogVisible=ref(false);const isEdit=ref(false);const editId=ref<number|null>(null);const saving=ref(false)
const form=ref({name:'',category:'drug',quantity:0,unit:'',min_stock:0,price:0,supplier:'',expiry_date:''})
const stockDialog=ref(false);const stockType=ref('in');const stockItem=ref<any>(null);const stockQty=ref(0);const stockNote=ref('');const stocking=ref(false)
const showPurchase=ref(false)
const catL:Record<string,string>={drug:'药品',consumable:'高值耗材',instrument:'仪器设备',other:'其他'}
function catLabel(s:string){return catL[s]||s}

// 统计
const totalItems=computed(()=>items.value.length)
const totalValue=computed(()=>items.value.reduce((s:number,i:any)=>s+(i.price||0)*(i.quantity||0),0))
const lowStockCount=computed(()=>items.value.filter((i:any)=>i.quantity<=i.min_stock).length)
const catValue=computed(()=>items.value.filter((i:any)=>i.category==='consumable').reduce((s:number,i:any)=>s+(i.price||0)*(i.quantity||0),0))
function suggestQty(row:any){return Math.max(0,Math.ceil(row.min_stock*2-row.quantity))}

// 有效期
function expiryStyle(date:string){if(!date)return{};const d=new Date(date).getTime();const n=Date.now();const diff=Math.ceil((d-n)/86400000);if(diff<0)return{color:'#f56c6c',fontWeight:'700'};if(diff<=30)return{color:'#e6a23c',fontWeight:'600'};return{}}
const expiringCount=computed(()=>{
  const n=Date.now()
  return items.value.filter((i:any)=>{
    if(!i.expiry_date)return false
    const diff=Math.ceil((new Date(i.expiry_date).getTime()-n)/86400000)
    return diff<30
  }).length
})

// 采购建议
const lowStockItems=computed(()=>items.value.filter((i:any)=>i.quantity<=i.min_stock))

// 流水日志
const logsDialog=ref(false);const logsItem=ref<any>(null);const logs=ref<any[]>([]);const logsLoading=ref(false)
function fmtDate(ts:string){if(!ts)return'—';const d=new Date(ts);return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`}
async function openLogs(row:any){
  logsItem.value=row;logs.value=[];logsDialog.value=true;logsLoading.value=true
  try{const r=await api.get('/inventory/items/'+row.id+'/logs');logs.value=Array.isArray(r.data)?r.data:r.data?.items||r.data?.data||[]}catch{}
  finally{logsLoading.value=false}
}

// 导出CSV
function today(){const d=new Date();return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`}
function exportCSV(){
  if(!items.value.length)return
  const h=['名称','分类','数量','单位','最低库存','单价','价值','供应商','有效期']
  const r=items.value.map((i:any)=>[i.name,catLabel(i.category),i.quantity,i.unit,i.min_stock,(i.price||0).toFixed(2),((i.price||0)*(i.quantity||0)).toFixed(2),i.supplier||'',i.expiry_date||''].map(v=>`"${v}"`).join(',')).join('\n')
  const blob=new Blob(['\uFEFF'+h.join(',')+'\n'+r],{type:'text/csv;charset=utf-8'})
  const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download='库存清单_'+today()+'.csv'
  a.click();URL.revokeObjectURL(a.href)
}

async function loadItems(){loading.value=true;try{const r=await api.get('/inventory/items');items.value=r.data||[]}catch{}finally{loading.value=false}}
function openAdd(){isEdit.value=false;editId.value=null;form.value={name:'',category:'drug',quantity:0,unit:'',min_stock:0,price:0,supplier:'',expiry_date:''};dialogVisible.value=true}
function openEdit(row:any){isEdit.value=true;editId.value=row.id;form.value={...row};dialogVisible.value=true}
async function saveItem(){
  saving.value=true
  try{if(isEdit.value&&editId.value)await api.put('/inventory/items/'+editId.value,form.value);else await api.post('/inventory/items',form.value)
  ElMessage.success(isEdit.value?'更新成功':'创建成功');dialogVisible.value=false;await loadItems()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}finally{saving.value=false}
}
function stockIn(row:any){stockType.value='in';stockItem.value=row;stockQty.value=0;stockNote.value='';stockDialog.value=true}
function stockOut(row:any){stockType.value='out';stockItem.value=row;stockQty.value=0;stockNote.value='';stockDialog.value=true}
async function confirmStock(){
  stocking.value=true
  try{const id=stockItem.value.id
    if(stockType.value==='in')await api.post('/inventory/items/'+id+'/stock-in',{quantity:stockQty.value,note:stockNote.value})
    else await api.post('/inventory/items/'+id+'/stock-out',{quantity:stockQty.value,note:stockNote.value})
    ElMessage.success(stockType.value==='in'?'入库成功':'出库成功');stockDialog.value=false;await loadItems()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}finally{stocking.value=false}
}
function viewDocs(row:any){router.push('/documents?product_name='+encodeURIComponent(row.name))}
onMounted(loadItems)
</script>
