<template>
  <div>
    <!-- 统计卡片 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="4"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#409eff">{{ docStats.total }}</div><div style="font-size:12px;color:#909399;margin-top:2px">总证件</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#67c23a">{{ docStats.valid }}</div><div style="font-size:12px;color:#909399;margin-top:2px">正常</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#e6a23c">{{ docStats.expiringSoon }}</div><div style="font-size:12px;color:#909399;margin-top:2px">即将到期</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#f56c6c">{{ docStats.expired }}</div><div style="font-size:12px;color:#909399;margin-top:2px">已过期</div></div></el-card></el-col>
      <el-col :span="5"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:22px;font-weight:700;color:#909399">{{ docStats.validRate }}%</div><div style="font-size:12px;color:#909399;margin-top:2px">合格率</div></div></el-card></el-col>
    </el-row>

    <el-card shadow="never" style="margin-bottom: 16px">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane label="发票" name="invoice" />
        <el-tab-pane label="出库单" name="delivery" />
        <el-tab-pane label="厂家资质" name="mfr_qual" />
        <el-tab-pane label="经销商资质" name="dist_qual" />
        <el-tab-pane label="质检报告" name="inspection" />
      </el-tabs>

      <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap; margin-top: 12px">
        <el-input v-model="search.product_name" placeholder="产品名" clearable style="width: 140px" @keyup.enter="handleSearch" />
        <el-input v-model="search.supplier" placeholder="供应商" clearable style="width: 140px" @keyup.enter="handleSearch" />
        <el-input v-model="search.keyword" placeholder="搜索标题/编号" clearable style="width: 160px" @keyup.enter="handleSearch" />
        <el-date-picker v-model="search.start_date" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" clearable style="width: 140px" />
        <span style="color: #909399">—</span>
        <el-date-picker v-model="search.end_date" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" clearable style="width: 140px" />
        <el-checkbox v-model="search.expiring_soon" label="即将到期" border size="small" />
        <el-checkbox v-model="search.expired" label="已过期" border size="small" />
        <el-button type="primary" @click="handleSearch">查询</el-button>
        <el-button @click="handleReset">重置</el-button>
        <div style="flex: 1" />
        <el-button type="primary" @click="openUpload">+ 上传文件</el-button>
        <el-button @click="goInspection">📋 检查模式</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="items" v-loading="loading" stripe border @selection-change="handleSelectionChange"
        :row-class-name="tableRowClassName">
                  <el-table-column type="selection" width="40" @selection-change="onSelectionChange"/>
          <el-table-column type="selection" width="40" />
        <el-table-column label="文件" min-width="220">
          <template #default="{ row }">
            <div style="display: flex; align-items: center; gap: 6px">
              <el-icon><Document /></el-icon>
              <a :href="downloadUrl(row.id)" target="_blank" style="color: #409EFF; text-decoration: none; cursor: pointer"
                @click.prevent="handlePreview(row)">
                {{ row.title }}
              </a>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag :type="docTypeTag(row.doc_type)" size="small">{{ docTypeLabel(row.doc_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="产品名" prop="product_name" min-width="120" />
        <el-table-column label="供应商" prop="supplier" min-width="120" />
        <el-table-column label="编号" prop="serial_no" width="130" />
        <el-table-column label="日期" prop="issue_date" width="100" />
        <el-table-column label="有效期" width="140">
          <template #default="{ row }">
            <span v-if="row.expiry_date" :style="expiryDateStyle(row.expiry_date)">
              {{ row.expiry_date }}
              <el-tag v-if="row.expiry_date < today" type="danger" size="small" effect="dark" style="margin-left:4px">过期</el-tag>
              <el-tag v-else-if="row.expiry_date <= thirtyDaysLater" type="warning" size="small" style="margin-left:4px">即将到期</el-tag>
            </span>
            <span v-else style="color: #c0c4cc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="文件大小" width="90">
          <template #default="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="handlePreview(row)">预览</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div v-if="total > 0" style="display: flex; justify-content: flex-end; margin-top: 16px">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next"
          @current-change="fetchData"
        />
      </div>

      <div v-if="selectedIds.length > 0" style="margin-top: 12px; padding: 8px 0; border-top: 1px solid #ebeef5">
        已选 <b>{{ selectedIds.length }}</b> 项
        <el-button size="small" style="margin-left: 12px" @click="batchPreview">预览选中</el-button>
        <el-button size="small" @click="batchDownload">打包下载 ZIP</el-button>
      </div>
    </el-card>

    <el-dialog v-model="uploadVisible" title="上传证件" width="550px" @close="resetUploadForm">
      <el-form ref="uploadFormRef" :model="uploadForm" label-width="100px" :rules="uploadRules">
        <el-form-item label="证件类型" prop="doc_type">
          <el-select v-model="uploadForm.doc_type" style="width: 100%">
            <el-option label="发票" value="invoice" />
            <el-option label="出库单" value="delivery" />
            <el-option label="厂家资质" value="mfr_qual" />
            <el-option label="经销商资质" value="dist_qual" />
            <el-option label="质检报告" value="inspection" />
          </el-select>
        </el-form-item>
        <el-form-item label="选择文件" prop="file">
          <el-upload ref="fileUploadRef" :auto-upload="false" :limit="1"
            :on-change="handleFileChange" accept=".pdf,.jpg,.jpeg,.png">
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <span style="font-size: 12px; color: #909399; margin-left: 8px">支持 PDF/JPG/PNG，最大 50MB</span>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="文件标题" prop="title">
          <el-input v-model="uploadForm.title" placeholder="如：华熙生物玻尿酸发票2024-01" />
        </el-form-item>
        <el-form-item label="产品名称">
          <el-input v-model="uploadForm.product_name" placeholder="关联产品/药品/耗材名称" />
        </el-form-item>
        <el-form-item label="供应商/厂家">
          <el-input v-model="uploadForm.supplier" />
        </el-form-item>
        <el-form-item label="编号">
          <el-input v-model="uploadForm.serial_no" :placeholder="serialNoPlaceholder" />
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="uploadForm.issue_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item v-if="isQualificationType" label="有效期至" prop="expiry_date">
          <el-date-picker v-model="uploadForm.expiry_date" type="date" value-format="YYYY-MM-DD" style="width: 100%" placeholder="请填写资质到期日期" />
        </el-form-item>
        <el-form-item v-if="isInvoiceOrDelivery" label="金额">
          <el-input-number v-model="uploadForm.amount" :min="0" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="uploadForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="submitUpload">上传</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDocuments, createDocument, deleteDocument, getDocumentDownloadUrl } from '../../api/documents'
import type { Document, DocumentListParams } from '../../api/documents'
import { Document as DocumentIcon } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()

const docTypeLabels: Record<string, string> = {
  all: '全部', invoice: '发票', delivery: '出库单',
  mfr_qual: '厂家资质', dist_qual: '经销商资质', inspection: '质检报告',
}
const docTypeTags: Record<string, string> = {
  invoice: '', delivery: '', mfr_qual: 'warning',
  dist_qual: 'warning', inspection: 'success',
}
function docTypeLabel(t: string) { return docTypeLabels[t] || t }
function docTypeTag(t: string) { return docTypeTags[t] || '' }

const today = new Date().toISOString().slice(0, 10)
const thirtyDaysLater = new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10)

const loading = ref(false)
const items = ref<Document[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const activeTab = ref((route.query.doc_type as string) || 'all')
const selectedIds = ref<number[]>([])

const search = reactive({
  product_name: (route.query.product_name as string) || '',
  supplier: '',
  keyword: '',
  start_date: '',
  end_date: '',
  expiring_soon: route.query.expiring_soon === 'true',
  expired: false,
})

const uploadVisible = ref(false)
const uploading = ref(false)
const selectedFile = ref<File | null>(null)
const fileUploadRef = ref<any>(null)
const uploadFormRef = ref<any>(null)

const docStats = computed(() => {
  const all = items.value || []
  const total = all.length
  let valid = 0, expiringSoon = 0, expired = 0
  all.forEach((d: any) => {
    if (!d.expiry_date) { valid++; return }
    if (d.expiry_date < today) { expired++; return }
    if (d.expiry_date <= thirtyDaysLater) { expiringSoon++; return }
    valid++
  })
  return {
    total,
    valid,
    expiringSoon,
    expired,
    validRate: total > 0 ? Math.round((valid / total) * 100) : 100
  }
})

const uploadForm = reactive({
  doc_type: 'invoice',
  title: '',
  product_name: '',
  supplier: '',
  serial_no: '',
  issue_date: '',
  expiry_date: '',
  amount: 0,
  remark: '',
})

const uploadRules = {
  doc_type: [{ required: true, message: '请选择证件类型', trigger: 'change' }],
  title: [{ required: true, message: '请输入文件标题', trigger: 'blur' }],
}

const isQualificationType = computed(() =>
  uploadForm.doc_type === 'mfr_qual' || uploadForm.doc_type === 'dist_qual'
)
const isInvoiceOrDelivery = computed(() =>
  uploadForm.doc_type === 'invoice' || uploadForm.doc_type === 'delivery'
)

const serialNoPlaceholder = computed(() => {
  const map: Record<string, string> = {
    invoice: '发票号', delivery: '出库单号', mfr_qual: '许可证号',
    dist_qual: '经营许可证号', inspection: '报告编号/批号',
  }
  return map[uploadForm.doc_type] || '编号'
})

function downloadUrl(id: number) { return getDocumentDownloadUrl(id) }

function handleTabChange() { page.value = 1; fetchData() }
function handleSearch() { page.value = 1; fetchData() }
function handleReset() {
  search.product_name = ''; search.supplier = ''; search.keyword = ''
  search.start_date = ''; search.end_date = ''
  search.expiring_soon = false; search.expired = false
  page.value = 1; fetchData()
}

async function fetchData() {
  loading.value = true
  try {
    const params: DocumentListParams = { page: page.value, page_size: pageSize.value }
    if (activeTab.value !== 'all') params.doc_type = activeTab.value
    if (search.product_name) params.product_name = search.product_name
    if (search.supplier) params.supplier = search.supplier
    if (search.keyword) params.keyword = search.keyword
    if (search.start_date) params.start_date = search.start_date
    if (search.end_date) params.end_date = search.end_date
    if (search.expiring_soon) params.expiring_soon = true
    if (search.expired) params.expired = true
    const res = await getDocuments(params)
    items.value = res.items; total.value = res.total
  } catch (e: any) { ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally { loading.value = false }
}

function openUpload() {
  uploadForm.doc_type = 'invoice'; uploadForm.title = ''
  uploadForm.product_name = ''; uploadForm.supplier = ''
  uploadForm.serial_no = ''; uploadForm.issue_date = ''
  uploadForm.expiry_date = ''; uploadForm.amount = 0; uploadForm.remark = ''
  selectedFile.value = null; fileUploadRef.value?.clearFiles()
  uploadVisible.value = true
}

function handleFileChange(_file: any) { selectedFile.value = _file.raw }

async function submitUpload() {
  if (!selectedFile.value) { ElMessage.warning('请选择要上传的文件'); return }
  if (!uploadForm.title) { ElMessage.warning('请输入文件标题'); return }
  uploading.value = true
  try {
    const fd = new FormData()
    fd.append('file', selectedFile.value)
    fd.append('doc_type', uploadForm.doc_type)
    fd.append('title', uploadForm.title)
    fd.append('product_name', uploadForm.product_name)
    fd.append('supplier', uploadForm.supplier)
    fd.append('serial_no', uploadForm.serial_no)
    fd.append('issue_date', uploadForm.issue_date)
    fd.append('expiry_date', uploadForm.expiry_date)
    fd.append('amount', String(uploadForm.amount))
    fd.append('remark', uploadForm.remark)
    await createDocument(fd)
    ElMessage.success('上传成功')
    uploadVisible.value = false; await fetchData()
  } catch (e: any) { ElMessage.error(e?.response?.data?.error || '上传失败')
  } finally { uploading.value = false }
}

function resetUploadForm() { selectedFile.value = null }

function handlePreview(row: Document) { window.open(getDocumentDownloadUrl(row.id), '_blank') }

async function handleDelete(row: Document) {
  try {
    await ElMessageBox.confirm(`确定删除「${row.title}」吗？`)
    await deleteDocument(row.id); ElMessage.success('删除成功'); await fetchData()
  } catch { /* cancelled */ }
}

function handleSelectionChange(selection: any[]) { selectedIds.value = selection.map((s: any) => s.id) }
function batchPreview() { for (const id of selectedIds.value) window.open(getDocumentDownloadUrl(id), '_blank') }
function batchDownload() { ElMessage.info(`打包下载功能待实现：已选 ${selectedIds.value.length} 项`) }

function tableRowClassName({ row }: { row: Document }) {
  if (row.expiry_date && row.expiry_date < today) return 'expired-row'
  if (row.expiry_date && row.expiry_date <= thirtyDaysLater) return 'expiring-row'
  return ''
}
function expiryDateStyle(date: string) {
  if (date < today) return { color: '#f56c6c', fontWeight: 600 }
  if (date <= thirtyDaysLater) return { color: '#e6a23c', fontWeight: 600 }
  return {}
}
function formatSize(bytes: number) {
  if (!bytes) return ''
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1048576).toFixed(1) + ' MB'
}
function goInspection() { router.push('/documents/inspection') }

onMounted(fetchData)
</script>

<style scoped>
:deep(.expired-row) { background-color: #fef0f0 !important; }
:deep(.expiring-row) { background-color: #fdf6ec !important; }
</style>
