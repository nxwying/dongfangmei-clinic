<template>
  <div style="max-width: 1200px; margin: 0 auto; padding: 20px">
    <!-- Header -->
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px">
      <div>
        <h2 style="margin: 0 0 4px 0">🏥 检查模式</h2>
        <span style="color: #909399; font-size: 13px">按产品汇总所有关联证件，方便上级部门查阅</span>
      </div>
      <el-button @click="goBack">← 返回列表</el-button>
    </div>

    <!-- Summary stats -->
    <el-row :gutter="16" style="margin-bottom: 16px">
      <el-col :span="4"><el-statistic title="全部证件" :value="allDocs.length" /></el-col>
      <el-col :span="4"><el-statistic title="发票" :value="countByType('invoice')" /></el-col>
      <el-col :span="4"><el-statistic title="出库单" :value="countByType('delivery')" /></el-col>
      <el-col :span="4"><el-statistic title="厂家资质" :value="countByType('mfr_qual')" /></el-col>
      <el-col :span="4"><el-statistic title="经销商资质" :value="countByType('dist_qual')" /></el-col>
      <el-col :span="4"><el-statistic title="质检报告" :value="countByType('inspection')" /></el-col>
    </el-row>

    <!-- Search + Filters -->
    <el-card shadow="never" style="margin-bottom: 16px">
      <div style="display: flex; align-items: center; gap: 10px; flex-wrap: wrap">
        <el-input v-model="searchKeyword" placeholder="搜索产品名/供应商/标题" clearable style="width: 300px" @keyup.enter="applyFilter" />
        <el-select v-model="filterType" placeholder="全部类型" clearable style="width: 120px" @change="applyFilter">
          <el-option label="全部" value="" />
          <el-option label="发票" value="invoice" />
          <el-option label="出库单" value="delivery" />
          <el-option label="厂家资质" value="mfr_qual" />
          <el-option label="经销商资质" value="dist_qual" />
          <el-option label="质检报告" value="inspection" />
        </el-select>
        <el-checkbox v-model="showOnlyExpiring" label="仅显示即将到期/已过期" @change="applyFilter" />
        <div style="flex: 1" />
        <el-button type="primary" @click="selectAll">全选</el-button>
        <el-button @click="clearSelection">清空</el-button>
      </div>
    </el-card>

    <!-- Loading -->
    <div v-if="loading" style="text-align: center; padding: 60px; color: #909399">加载中...</div>

    <!-- Grouped list -->
    <div v-else-if="filteredGroups.length === 0" style="text-align: center; padding: 60px; color: #909399">
      暂无匹配的证件记录
    </div>

    <div v-else>
      <div v-for="group in filteredGroups" :key="group.product" style="margin-bottom: 20px">
        <el-card shadow="never">
          <template #header>
            <div style="display: flex; align-items: center; justify-content: space-between">
              <div>
                <span style="font-weight: 600; font-size: 15px">{{ group.product || '(未关联产品)' }}</span>
                <span v-if="group.supplier" style="margin-left: 10px; color: #909399; font-size: 13px">{{ group.supplier }}</span>
              </div>
              <el-tag type="info">{{ group.items.length }} 份</el-tag>
            </div>
          </template>

          <div v-for="doc in group.items" :key="doc.id" style="display: flex; align-items: center; gap: 12px; padding: 8px 0; border-bottom: 1px solid #f0f0f0">
            <el-checkbox :model-value="selectedIds.has(doc.id)" @change="toggleSelect(doc.id)" />
            <span style="color: #909399; min-width: 85px; font-size: 13px">{{ doc.issue_date || '-' }}</span>
            <el-tag :type="typeTag(doc.doc_type)" size="small" style="min-width: 60px; text-align: center">
              {{ typeLabel(doc.doc_type) }}
            </el-tag>
            <span style="flex: 1">
              <a :href="downloadUrl(doc.id)" target="_blank"
                style="color: #409EFF; text-decoration: none; cursor: pointer"
                @click.prevent="handlePreview(doc)">
                {{ doc.title }}
              </a>
            </span>
            <span style="color: #909399; font-size: 12px; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">
              {{ doc.file_name }}
            </span>
            <span v-if="doc.expiry_date" :style="expiryStyle(doc.expiry_date)" style="font-size: 12px; white-space: nowrap">
              有效至 {{ doc.expiry_date }}
            </span>
          </div>
        </el-card>
      </div>
    </div>

    <!-- Bottom action bar -->
    <div v-if="selectedIds.size > 0"
      style="position: sticky; bottom: 0; background: #fff; border-top: 1px solid #e6e6e6; padding: 12px 20px; display: flex; align-items: center; gap: 12px; box-shadow: 0 -2px 8px rgba(0,0,0,0.06)">
      <span>已选 <b>{{ selectedIds.size }}</b> 项</span>
      <el-button size="small" type="primary" @click="previewSelected">预览选中</el-button>
      <el-button size="small" @click="downloadSelected">打包下载 ZIP</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { getDocuments, getDocumentDownloadUrl } from '../../api/documents'
import type { Document } from '../../api/documents'

const router = useRouter()
const loading = ref(false)
const allDocs = ref<Document[]>([])

const searchKeyword = ref('')
const filterType = ref('')
const showOnlyExpiring = ref(false)
const selectedIds = ref(new Set<number>())

const today = new Date().toISOString().slice(0, 10)
const thirtyDaysLater = new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10)

const typeLabels: Record<string, string> = {
  invoice: '发票', delivery: '出库单', mfr_qual: '厂家资质',
  dist_qual: '经销商资质', inspection: '质检报告',
}
const typeTags: Record<string, string> = {
  invoice: '', delivery: '', mfr_qual: 'warning',
  dist_qual: 'warning', inspection: 'success',
}
function typeLabel(t: string) { return typeLabels[t] || t }
function typeTag(t: string) { return typeTags[t] || '' }

function countByType(t: string) {
  return allDocs.value.filter(d => d.doc_type === t).length
}

const groupedItems = computed(() => {
  const groups: Record<string, { product: string; supplier: string; items: Document[] }> = {}
  for (const doc of allDocs.value) {
    const key = doc.product_name || '__unlinked__'
    if (!groups[key]) groups[key] = { product: doc.product_name, supplier: '', items: [] }
    if (doc.supplier && !groups[key].supplier) groups[key].supplier = doc.supplier
    groups[key].items.push(doc)
  }
  // Sort within groups by issue_date desc
  for (const key of Object.keys(groups)) {
    groups[key].items.sort((a, b) => (b.issue_date || '').localeCompare(a.issue_date || ''))
  }
  // Sort groups by product name
  return Object.values(groups).sort((a, b) => a.product.localeCompare(b.product))
})

const filteredGroups = computed(() => {
  let groups = groupedItems.value
  if (searchKeyword.value) {
    const kw = searchKeyword.value.toLowerCase()
    groups = groups.map(g => ({
      ...g,
      items: g.items.filter(d =>
        d.product_name?.toLowerCase().includes(kw) ||
        d.supplier?.toLowerCase().includes(kw) ||
        d.title?.toLowerCase().includes(kw)
      )
    })).filter(g => g.items.length > 0)
  }
  if (filterType.value) {
    groups = groups.map(g => ({
      ...g,
      items: g.items.filter(d => d.doc_type === filterType.value)
    })).filter(g => g.items.length > 0)
  }
  if (showOnlyExpiring.value) {
    groups = groups.map(g => ({
      ...g,
      items: g.items.filter(d =>
        d.expiry_date !== '' &&
        (d.expiry_date < today || d.expiry_date <= thirtyDaysLater)
      )
    })).filter(g => g.items.length > 0)
  }
  return groups
})

function applyFilter() {
  // reactivity handles it
}

function downloadUrl(id: number) { return getDocumentDownloadUrl(id) }

function handlePreview(doc: Document) { window.open(getDocumentDownloadUrl(doc.id), '_blank') }

function toggleSelect(id: number) {
  const s = new Set(selectedIds.value)
  if (s.has(id)) s.delete(id); else s.add(id)
  selectedIds.value = s
}

function selectAll() {
  const ids = new Set<number>()
  for (const g of filteredGroups.value) {
    for (const d of g.items) ids.add(d.id)
  }
  selectedIds.value = ids
}

function clearSelection() { selectedIds.value = new Set() }

function previewSelected() {
  for (const id of selectedIds.value) window.open(getDocumentDownloadUrl(id), '_blank')
}

function downloadSelected() {
  ElMessage.info(`打包下载功能待实现：已选 ${selectedIds.value.size} 项`)
}

function expiryStyle(date: string) {
  if (date < today) return { color: '#f56c6c' }
  if (date <= thirtyDaysLater) return { color: '#e6a23c' }
  return { color: '#909399' }
}

function goBack() { router.push('/documents') }

async function fetchAll() {
  loading.value = true
  try {
    const res = await getDocuments({ page: 1, page_size: 500 })
    allDocs.value = res.items
  } catch (e: any) { ElMessage.error(e?.response?.data?.error || '加载失败')
  } finally { loading.value = false }
}

onMounted(fetchAll)
</script>
