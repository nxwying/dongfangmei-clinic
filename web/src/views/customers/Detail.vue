<template>
  <div class="detail-page">
    <!-- Header -->
    <el-card shadow="never" style="margin-bottom: 16px">
      <div style="display: flex; align-items: center; justify-content: space-between">
        <div style="display: flex; align-items: center; gap: 12px">
          <el-button @click="$router.back()">返回</el-button>
          <span style="font-size: 18px; font-weight: 600">{{ customer?.name || '客户详情' }}</span>
        </div>
        <el-button type="primary" @click="openEdit">编辑资料</el-button>
      </div>
    </el-card>
    <el-skeleton :loading="loading" animated :rows="10">
      <template v-if="customer">
        <!-- Section 1: 基本信息 -->
        <el-card shadow="never" style="margin-bottom: 16px">
          <template #header><span style="font-weight: 600">基本信息</span></template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="姓名">{{ customer.name }}</el-descriptions-item>
            <el-descriptions-item label="手机号">{{ customer.phone }}</el-descriptions-item>
            <el-descriptions-item label="性别">{{ genderLabel(customer.gender) }}</el-descriptions-item>
            <el-descriptions-item label="身份证号">{{ customer.id_card || '-' }}</el-descriptions-item>
            <el-descriptions-item label="生日">{{ customer.birthday || '-' }}</el-descriptions-item>
            <el-descriptions-item label="来源">{{ sourceLabel(customer.source) }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="statusTagType(customer.status)" size="small">
                {{ statusLabel(customer.status) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="customer.remark" label="备注" :span="2">
              {{ customer.remark }}
            </el-descriptions-item>
          </el-descriptions>
        </el-card>
        <!-- Section 2: 会员信息 -->
        <el-card shadow="never" style="margin-bottom: 16px">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between">
              <span style="font-weight: 600">会员信息</span>
              <div v-if="customer.membership">
                <el-button type="warning" size="small" @click="rechargeVisible = true">充值</el-button>
              </div>
              <el-button v-else type="success" size="small" @click="membershipVisible = true">开卡</el-button>
            </div>
          </template>
          <template v-if="customer.membership">
            <el-descriptions :column="3" border size="small">
              <el-descriptions-item label="会员等级">
                <el-tag :type="levelTagType(customer.membership.level)" size="small">
                  {{ levelLabel(customer.membership.level) }}
                </el-tag>
              </el-descriptions-item>
              <el-descriptions-item label="储值余额">
                ¥{{ (customer.membership.balance || 0).toFixed(2) }}
              </el-descriptions-item>
              <el-descriptions-item label="赠送余额">
                ¥{{ (customer.membership.gift_balance || 0).toFixed(2) }}
              </el-descriptions-item>
              <el-descriptions-item label="累计充值">
                ¥{{ (customer.membership.total_recharged || 0).toFixed(2) }}
              </el-descriptions-item>
            </el-descriptions>
          </template>
          <div v-else style="color: #909399; font-size: 14px; padding: 8px 0">
            尚未开通会员
          </div>
        </el-card>
        <!-- Section 3: 套餐 -->
        <el-card shadow="never" style="margin-bottom: 16px">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between">
              <span style="font-weight: 600">套餐</span>
              <el-button type="primary" size="small" @click="packageVisible = true">购买套餐</el-button>
            </div>
          </template>
          <el-table v-if="packages.length" :data="packages" stripe size="small" style="width: 100%">
            <el-table-column prop="name" label="套餐名称" min-width="140" />
            <el-table-column label="总次数" width="80" align="center">
              <template #default="{ row }">{{ row.total_sessions }}</template>
            </el-table-column>
            <el-table-column label="已用次数" width="80" align="center">
              <template #default="{ row }">{{ row.used_sessions }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.used_sessions >= row.total_sessions ? 'info' : 'success'" size="small">
                  {{ row.used_sessions >= row.total_sessions ? '已用完' : '使用中' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div v-else style="color: #909399; font-size: 14px; padding: 8px 0">
            暂无套餐
          </div>
        </el-card>
        <!-- Section 4: 历史订单 -->
        <el-card shadow="never" style="margin-bottom: 16px">
          <template #header><span style="font-weight: 600">历史订单</span></template>
          <el-table v-if="orders.length" :data="orders" stripe size="small" style="width: 100%">
            <el-table-column prop="order_no" label="订单号" width="180" />
            <el-table-column label="金额" width="120" align="right">
              <template #default="{ row }">¥{{ (row.total_amount || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="orderStatusTagType(row.status)" size="small">
                  {{ orderStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="日期" width="160" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-link type="primary" size="small" @click="$router.push(`/pos/${row.id}`)">查看</el-link>
              </template>
            </el-table-column>
          </el-table>
          <div v-else style="color: #909399; font-size: 14px; padding: 8px 0">
            暂无订单
          </div>
        </el-card>
        <!-- Section 5: 回访记录 -->
        <el-card shadow="never" style="margin-bottom: 16px">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between">
              <span style="font-weight: 600">回访记录</span>
              <el-button type="primary" size="small" @click="followUpVisible = true">新增回访</el-button>
            </div>
          </template>
          <div v-if="followUps.length" style="display: flex; flex-direction: column; gap: 12px">
            <div v-for="item in followUps" :key="item.id" class="follow-up-item">
              <div style="display: flex; justify-content: space-between; align-items: center">
                <span style="font-weight: 500">{{ followUpMethodLabel(item.method) }}</span>
                <span style="color: #909399; font-size: 13px">{{ item.created_at }}</span>
              </div>
              <div style="margin-top: 6px; color: #606266">{{ item.content }}</div>
              <div v-if="item.next_at" style="margin-top: 4px; font-size: 13px; color: #909399">
                下次回访：{{ item.next_at }}
              </div>
            </div>
          </div>
          <div v-else style="color: #909399; font-size: 14px; padding: 8px 0">
            暂无回访记录
          </div>
        </el-card>
      </template>
    </el-skeleton>
    <!-- ── Dialogs ────────────────────────────────────── -->
    <!-- Edit dialog -->
    <el-dialog v-model="editVisible" title="编辑资料" width="520px" @close="resetEditForm">
      <el-form ref="editFormRef" :model="editForm" label-width="80px" :rules="editRules">
        <el-form-item label="姓名" prop="name">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="editForm.phone" />
        </el-form-item>
        <el-form-item label="性别" prop="gender">
          <el-select v-model="editForm.gender" style="width: 100%">
            <el-option label="男" :value="1" />
            <el-option label="女" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="身份证号" prop="id_card">
          <el-input v-model="editForm.id_card" placeholder="18 位身份证号" maxlength="18" @input="onEditIdCardInput" />
        </el-form-item>
        <el-form-item label="生日" prop="birthday">
          <el-date-picker v-model="editForm.birthday" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="来源" prop="source">
          <el-select v-model="editForm.source" style="width: 100%">
            <el-option label="到店" value="walk_in" />
            <el-option label="转介绍" value="referral" />
            <el-option label="小红书" value="xiaohongshu" />
            <el-option label="微信" value="wechat" />
            <el-option label="抖音" value="douyin" />
            <el-option label="大众点评" value="dianping" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="editForm.remark" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" :loading="editSubmitting" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
    <!-- Recharge dialog -->
    <el-dialog v-model="rechargeVisible" title="会员充值" width="420px" @close="resetRechargeForm">
      <el-form ref="rechargeFormRef" :model="rechargeForm" label-width="100px" :rules="rechargeRules">
        <el-form-item label="充值金额" prop="amount">
          <el-input-number v-model="rechargeForm.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="赠送金额" prop="gift_amount">
          <el-input-number v-model="rechargeForm.gift_amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeVisible = false">取消</el-button>
        <el-button type="primary" :loading="rechargeSubmitting" @click="submitRecharge">确认充值</el-button>
      </template>
    </el-dialog>
    <!-- Open membership dialog -->
    <el-dialog v-model="membershipVisible" title="开通会员" width="420px" @close="resetMembershipForm">
      <el-form ref="membershipFormRef" :model="membershipForm" label-width="100px" :rules="membershipRules">
        <el-form-item label="会员等级" prop="level">
          <el-select v-model="membershipForm.level" style="width: 100%">
            <el-option label="普通会员" value="regular" />
            <el-option label="白银会员" value="silver" />
            <el-option label="黄金会员" value="gold" />
            <el-option label="铂金会员" value="platinum" />
            <el-option label="钻石会员" value="diamond" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="membershipVisible = false">取消</el-button>
        <el-button type="primary" :loading="membershipSubmitting" @click="submitMembership">确认开卡</el-button>
      </template>
    </el-dialog>
    <!-- New follow-up dialog -->
    <el-dialog v-model="followUpVisible" title="新增回访" width="520px" @close="resetFollowUpForm">
      <el-form ref="followUpFormRef" :model="followUpForm" label-width="100px" :rules="followUpRules">
        <el-form-item label="回访内容" prop="content">
          <el-input v-model="followUpForm.content" type="textarea" :rows="4" placeholder="请填写回访内容" />
        </el-form-item>
        <el-form-item label="回访方式" prop="method">
          <el-select v-model="followUpForm.method" style="width: 100%">
            <el-option label="电话" value="phone" />
            <el-option label="微信" value="wechat" />
            <el-option label="短信" value="sms" />
            <el-option label="到店" value="visit" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="下次回访" prop="next_at">
          <el-date-picker
            v-model="followUpForm.next_at"
            type="datetime"
            placeholder="选择下次回访时间（可选）"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="followUpVisible = false">取消</el-button>
        <el-button type="primary" :loading="followUpSubmitting" @click="submitFollowUp">保存</el-button>
      </template>
    </el-dialog>
    <!-- Buy package dialog -->
    <el-dialog v-model="packageVisible" title="购买套餐" width="420px" @close="resetPackageForm">
      <el-form ref="packageFormRef" :model="packageForm" label-width="100px" :rules="packageRules">
        <el-form-item label="套餐模板" prop="template_id">
          <el-select v-model="packageForm.template_id" filterable placeholder="选择套餐" style="width: 100%">
            <el-option v-for="tpl in packageTemplates" :key="tpl.id" :label="tpl.name" :value="tpl.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="packageVisible = false">取消</el-button>
        <el-button type="primary" :loading="packageSubmitting" @click="submitPackage">确认购买</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../../api'
import {
  getCustomer,
  updateCustomer,
  getFollowUps,
  createFollowUp,
  openMembership,
  recharge,
  getCustomerPackages,
  createPackage,
} from '../../api/customer'
import { getOrders } from '../../api/order'
const route = useRoute()
const router = useRouter()
const customerId = computed(() => Number(route.params.id))

// ── State ──────────────────────────────────────────────
const loading = ref(true)
const customer = ref<any>(null)
const orders = ref<any[]>([])
const packages = ref<any[]>([])
const followUps = ref<any[]>([])
const packageTemplates = ref<any[]>([])
async function loadCustomer() {
  if (!customerId.value) {
    ElMessage.error('无效客户ID')
    router.push('/customers')
    return
  }
  loading.value = true
  try {
    const res = await getCustomer(customerId.value)
    customer.value = res?.data ?? res
  } catch (e) {
    ElMessage.error('加载客户信息失败')
  } finally {
    loading.value = false
  }
}
async function loadOrders() {
  try {
    const res = await getOrders({ customer_id: customerId.value, page_size: 50 })
    orders.value = res?.data ?? res?.list ?? []
  } catch {
    orders.value = []
  }
}
async function loadPackages() {
  try {
    const res = await getCustomerPackages(customerId.value)
    packages.value = res?.data ?? res?.list ?? []
  } catch {
    packages.value = []
  }
}
async function loadFollowUps() {
  try {
    const res = await getFollowUps(customerId.value)
    followUps.value = res?.data ?? res?.list ?? []
  } catch {
    followUps.value = []
  }
}
async function loadPackageTemplates() {
  try {
    const res = await api.get('/settings/package-templates')
    packageTemplates.value = res?.data?.data ?? res?.data?.list ?? res?.data ?? []
  } catch {
    packageTemplates.value = []
  }
}
// ── Label helpers ──────────────────────────────────────
function genderLabel(gender?: number): string {
  if (gender === 1) return '男'
  if (gender === 2) return '女'
  return '-'
}
function sourceLabel(source?: string): string {
  const map: Record<string, string> = {
    walk_in: '到店',
    referral: '转介绍',
    xiaohongshu: '小红书',
    wechat: '微信',
    douyin: '抖音',
    dianping: '大众点评',
    other: '其他',
  }
  return source ? map[source] || source : '-'
}
function statusTagType(status?: string): 'info' | 'success' | 'danger' | '' {
  const map: Record<string, 'info' | 'success' | 'danger' | ''> = {
    potential: 'info',
    active: 'success',
    lost: 'danger',
  }
  return status ? map[status] ?? 'info' : 'info'
}
function statusLabel(status?: string): string {
  const map: Record<string, string> = {
    potential: '潜在客户',
    active: '活跃客户',
    lost: '流失客户',
  }
  return status ? map[status] || status : '-'
}
function levelLabel(level?: string): string {
  const map: Record<string, string> = {
    regular: '普通会员',
    silver: '白银会员',
    gold: '黄金会员',
    platinum: '铂金会员',
    diamond: '钻石会员',
  }
  return level ? map[level] || level : '-'
}
function levelTagType(level?: string): 'info' | 'success' | 'warning' | 'danger' | '' {
  const map: Record<string, 'info' | 'success' | 'warning' | 'danger' | ''> = {
    regular: 'info',
    silver: '',
    gold: 'warning',
    platinum: '',
    diamond: 'danger',
  }
  return level ? map[level] || 'info' : 'info'
}
function orderStatusTagType(status?: string): 'info' | 'success' | 'warning' | 'danger' | '' {
  const map: Record<string, 'info' | 'success' | 'warning' | 'danger' | ''> = {
    pending: 'warning',
    paid: 'success',
    partial: '',
    refunded: 'danger',
  }
  return status ? map[status] ?? 'info' : 'info'
}
function orderStatusLabel(status?: string): string {
  const map: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    partial: '部分支付',
    refunded: '已退款',
  }
  return status ? map[status] || status : '-'
}
function followUpMethodLabel(method?: string): string {
  const map: Record<string, string> = {
    phone: '电话',
    wechat: '微信',
    sms: '短信',
    visit: '到店',
    other: '其他',
  }
  return method ? map[method] || method : '-'
}
// ── Edit dialog ────────────────────────────────────────
const editVisible = ref(false)
const editSubmitting = ref(false)
const editFormRef = ref<any>(null)
const editForm = ref({ name: '', phone: '', gender: '', id_card: '', birthday: '', source: '', remark: '' })
const editRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
}
function resetEditForm() {
  editFormRef.value?.resetFields()
}
function openEdit() {
  if (!customer.value) return
  editForm.value = {
    name: customer.value.name || '',
    phone: customer.value.phone || '',
    gender: customer.value.gender ?? 0,
    birthday: customer.value.birthday || '',
    source: customer.value.source || '',
    remark: customer.value.remark || '',
  }
  editVisible.value = true
}
function onEditIdCardInput() {
  const v = editForm.id_card
  if (v && v.length === 18) {
    editForm.birthday = v.substring(6, 10) + '-' + v.substring(10, 12) + '-' + v.substring(12, 14)
    const d = parseInt(v.charAt(16))
    editForm.gender = d % 2 === 1 ? 1 : 2
  }
}

async function submitEdit() {
  const valid = await editFormRef.value?.validate().catch(() => false)
  if (!valid) return
  editSubmitting.value = true
  try {
    await updateCustomer(customerId.value, editForm.value)
    ElMessage.success('更新成功')
    editVisible.value = false
    await loadCustomer()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '更新失败')
  } finally {
    editSubmitting.value = false
  }
}
// ── Recharge dialog ────────────────────────────────────
const rechargeVisible = ref(false)
const rechargeSubmitting = ref(false)
const rechargeFormRef = ref<any>(null)
const rechargeForm = ref({ amount: 0, gift_amount: 0 })
const rechargeRules = {
  amount: [{ required: true, message: '请输入充值金额', trigger: 'blur' }],
}
function resetRechargeForm() {
  rechargeFormRef.value?.resetFields()
  rechargeForm.value = { amount: 0, gift_amount: 0 }
}
async function submitRecharge() {
  const valid = await rechargeFormRef.value?.validate().catch(() => false)
  if (!valid) return
  rechargeSubmitting.value = true
  try {
    await recharge(customerId.value, rechargeForm.value)
    ElMessage.success('充值成功')
    rechargeVisible.value = false
    await loadCustomer()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '充值失败')
  } finally {
    rechargeSubmitting.value = false
  }
}
// ── Membership (open card) dialog ──────────────────────
const membershipVisible = ref(false)
const membershipSubmitting = ref(false)
const membershipFormRef = ref<any>(null)
const membershipForm = ref({ level: 'regular' })
const membershipRules = {
  level: [{ required: true, message: '请选择会员等级', trigger: 'change' }],
}
function resetMembershipForm() {
  membershipFormRef.value?.resetFields()
  membershipForm.value = { level: 'regular' }
}
async function submitMembership() {
  const valid = await membershipFormRef.value?.validate().catch(() => false)
  if (!valid) return
  membershipSubmitting.value = true
  try {
    await openMembership(customerId.value, membershipForm.value)
    ElMessage.success('开卡成功')
    membershipVisible.value = false
    await loadCustomer()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '开卡失败')
  } finally {
    membershipSubmitting.value = false
  }
}
// ── Follow-up dialog ───────────────────────────────────
const followUpVisible = ref(false)
const followUpSubmitting = ref(false)
const followUpFormRef = ref<any>(null)
const followUpForm = ref({ content: '', method: 'phone', next_at: '' })
const followUpRules = {
  content: [{ required: true, message: '请填写回访内容', trigger: 'blur' }],
  method: [{ required: true, message: '请选择回访方式', trigger: 'change' }],
}
function resetFollowUpForm() {
  followUpFormRef.value?.resetFields()
  followUpForm.value = { content: '', method: 'phone', next_at: '' }
}
async function submitFollowUp() {
  const valid = await followUpFormRef.value?.validate().catch(() => false)
  if (!valid) return
  followUpSubmitting.value = true
  try {
    await createFollowUp(customerId.value, followUpForm.value)
    ElMessage.success('回访记录已保存')
    followUpVisible.value = false
    await loadFollowUps()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  } finally {
    followUpSubmitting.value = false
  }
}
// ── Buy package dialog ─────────────────────────────────
const packageVisible = ref(false)
const packageSubmitting = ref(false)
const packageFormRef = ref<any>(null)
const packageForm = ref({ template_id: null as number | null })
const packageRules = {
  template_id: [{ required: true, message: '请选择套餐', trigger: 'change' }],
}
function resetPackageForm() {
  packageFormRef.value?.resetFields()
  packageForm.value = { template_id: null }
}
async function submitPackage() {
  const valid = await packageFormRef.value?.validate().catch(() => false)
  if (!valid) return
  packageSubmitting.value = true
  try {
    await createPackage(customerId.value, packageForm.value)
    ElMessage.success('套餐购买成功')
    packageVisible.value = false
    await loadPackages()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || '购买失败')
  } finally {
    packageSubmitting.value = false
  }
}
// ── Init ───────────────────────────────────────────────
onMounted(async () => {
  await Promise.all([
    loadCustomer(),
    loadOrders(),
    loadPackages(),
    loadFollowUps(),
    loadPackageTemplates(),
  ])
})
</script>
<style scoped>
.detail-page {
  max-width: 960px;
  margin: 0 auto;
}
.follow-up-item {
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
}
.follow-up-item + .follow-up-item {
  margin-top: 8px;
}
</style>
