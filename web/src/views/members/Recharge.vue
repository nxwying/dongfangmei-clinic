<template>
  <el-dialog :model-value="visible" @update:model-value="$emit('update:visible', $event)" title="充值" width="450px">
    <div v-if="customer" style="margin-bottom:12px;color:#606266">
      会员：<b>{{ customer.name }}</b> &nbsp; 当前余额：<b>¥{{ (customer.membership?.balance||0).toFixed(2) }}</b>
    </div>
    <el-form label-width="90px">
      <el-form-item label="充值金额">
        <el-input-number v-model="amount" :min="1" :precision="2" style="width:200px"/>
        <div style="display:flex;gap:4px;margin-left:8px">
          <el-button v-for="v in [500,1000,2000,5000]" :key="v" size="small" @click="amount=v">¥{{ v }}</el-button>
        </div>
      </el-form-item>
      <el-form-item label="赠送金额"><el-input-number v-model="giftAmount" :min="0" :precision="2" style="width:200px"/></el-form-item>
      <el-form-item label="支付方式">
        <el-select v-model="payMethod" style="width:200px">
          <el-option label="微信" value="wechat"/><el-option label="支付宝" value="alipay"/>
          <el-option label="银行卡" value="card"/><el-option label="现金" value="cash"/>
        </el-select>
      </el-form-item>
      <el-form-item label="备注"><el-input v-model="note" type="textarea" :rows="2"/></el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:visible',false)">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">确认充值</el-button>
    </template>
  </el-dialog>
</template>
<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'
const props = defineProps<{visible:boolean; customer:any}>()
const emit = defineEmits<{'update:visible':[boolean]; 'success':[]}>()
const amount = ref(0); const giftAmount = ref(0); const payMethod = ref('wechat'); const note = ref(''); const submitting = ref(false)
watch(()=>props.visible,(v)=>{if(v){amount.value=0;giftAmount.value=0;payMethod.value='wechat';note.value=''}})
async function submit(){
  if(!amount.value||amount.value<=0){ElMessage.warning('请输入充值金额');return}
  submitting.value=true
  try{
    await api.post('/customers/'+props.customer.id+'/recharge',{amount:amount.value,gift_amount:giftAmount.value,pay_method:payMethod.value,note:note.value})
    ElMessage.success('充值成功');emit('success')
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'充值失败')}
  finally{submitting.value=false}
}
</script>