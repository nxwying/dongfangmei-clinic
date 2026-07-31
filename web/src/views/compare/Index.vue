<template>
  <div>
    <el-card shadow="never" style="margin-bottom: 16px">
      <div style="display:flex;align-items:center;gap:16px;flex-wrap:wrap">
        <span style="font-size:16px;font-weight:600">术前术后对比</span>
        <el-select v-model="selectedCustomer" filterable remote :remote-method="searchCustomer"
          placeholder="搜索客户姓名或手机号" style="width:280px" @change="loadPhotos">
          <el-option v-for="c in customerOptions" :key="c.id" :label="c.name + ' (' + c.phone + ')'" :value="c.id" />
        </el-select>
        <el-select v-model="selectedPart" placeholder="选择部位" style="width:160px" clearable>
          <el-option v-for="g in groups" :key="g.body_part" :label="g.body_part" :value="g.body_part" />
        </el-select>
      </div>
    </el-card>

    <div v-if="loading" style="text-align:center;padding:60px">
      <el-icon class="is-loading" :size="32"><Loading /></el-icon>
      <p style="color:#909399;margin-top:12px">加载中...</p>
    </div>

    <div v-else-if="!selectedCustomer" style="text-align:center;padding:60px;color:#c0c4cc">
      <el-icon :size="48"><Picture /></el-icon>
      <p style="margin-top:12px;font-size:14px">请在上方选择客户查看对比图</p>
    </div>

    <div v-else-if="filteredGroups.length === 0" style="text-align:center;padding:60px;color:#c0c4cc">
      <el-icon :size="48"><PictureFilled /></el-icon>
      <p style="margin-top:12px;font-size:14px">该客户暂无术前术后照片</p>
    </div>

    <div v-else>
      <el-card v-for="g in filteredGroups" :key="g.body_part" shadow="never" style="margin-bottom:16px">
        <template #header>
          <div style="display:flex;align-items:center;justify-content:space-between">
            <span style="font-weight:600">{{ g.body_part }}</span>
            <el-tag size="small" type="info">{{ g.photos.length }} 张</el-tag>
          </div>
        </template>

        <!-- Pair-based comparison: find before/after pairs -->
        <div v-for="(pair, idx) in buildPairs(g.photos)" :key="idx" style="margin-bottom:20px">
          <div style="display:flex;gap:12px;flex-wrap:wrap">
            <!-- Before -->
            <div style="text-align:center">
              <div style="position:relative;width:240px;height:240px;border:1px solid #ebeef5;border-radius:8px;overflow:hidden;background:#f5f7fa;display:flex;align-items:center;justify-content:center">
                <img v-if="pair.before" :src="photoUrl(pair.before.id)" style="width:100%;height:100%;object-fit:cover" />
                <div v-else style="color:#c0c4cc;font-size:13px">无术前照</div>
              </div>
              <div style="margin-top:6px;font-size:13px;color:#909399">
                {{ pair.before ? formatDate(pair.before.created_at) : '术前' }}
              </div>
            </div>
            <!-- Arrow -->
            <div style="display:flex;align-items:center">
              <el-icon :size="24" color="#409eff"><Right /></el-icon>
            </div>
            <!-- After -->
            <div style="text-align:center">
              <div style="position:relative;width:240px;height:240px;border:1px solid #ebeef5;border-radius:8px;overflow:hidden;background:#f5f7fa;display:flex;align-items:center;justify-content:center">
                <img v-if="pair.after" :src="photoUrl(pair.after.id)" style="width:100%;height:100%;object-fit:cover" />
                <div v-else style="color:#c0c4cc;font-size:13px">无术后照</div>
              </div>
              <div style="margin-top:6px;font-size:13px;color:#909399">
                {{ pair.after ? formatDate(pair.after.created_at) : '术后' }}
              </div>
            </div>
            <!-- Slide comparison (only if both before and after exist) -->
            <div v-if="pair.before && pair.after" style="text-align:center">
              <div style="position:relative;width:240px;height:240px;border:1px solid #ebeef5;border-radius:8px;overflow:hidden;cursor:ew-resize"
                   @mousemove="onSlide($event, 'slide-' + g.body_part + '-' + idx)"
                   :ref="'slide-' + g.body_part + '-' + idx">
                <img :src="photoUrl(pair.after.id)" style="width:100%;height:100%;object-fit:cover;position:absolute;top:0;left:0" />
                <div :style="{position:'absolute',top:0,left:0,width:'100%',height:'100%',overflow:'hidden',clipPath:'inset(0 ' + (100 - slidePos[g.body_part + '-' + idx] || 50) + '% 0 0)'}">
                  <img :src="photoUrl(pair.before.id)" style="width:240px;height:240px;object-fit:cover" />
                </div>
                <div :style="{position:'absolute',top:0,bottom:0,left:(slidePos[g.body_part + '-' + idx] || 50)+'%',width:'2px',background:'#409eff'}"></div>
                <div :style="{position:'absolute',top:'50%',left:(slidePos[g.body_part + '-' + idx] || 50)+'%',transform:'translate(-50%,-50%)',background:'#409eff',color:'#fff',borderRadius:'50%',width:'28px',height:'28px',display:'flex',alignItems:'center',justifyContent:'center',fontSize:'12px'}">
                  <el-icon><Rank /></el-icon>
                </div>
              </div>
              <div style="margin-top:6px;font-size:13px;color:#909399">滑动对比</div>
            </div>
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { Picture, PictureFilled, Right, Rank, Loading } from '@element-plus/icons-vue'
import api from '../../api'

const selectedCustomer = ref(null)
const selectedPart = ref('')
const customerOptions = ref([])
const groups = ref([])
const loading = ref(false)
const slidePos = reactive({})

async function searchCustomer(q) {
  try {
    const res = await api.get('/customers', { params: { q, page_size: 20 } })
    customerOptions.value = res.data?.data || res.data || []
  } catch(e) {}
}

async function loadPhotos() {
  if (!selectedCustomer.value) { groups.value = []; return }
  loading.value = true
  try {
    const res = await api.get(`/customers/${selectedCustomer.value}/photos/grouped`)
    groups.value = res.data || []
  } catch(e) {
    groups.value = []
  }
  loading.value = false
}

const filteredGroups = computed(() => {
  if (!selectedPart.value) return groups.value
  return groups.value.filter(g => g.body_part === selectedPart.value)
})

function buildPairs(photos) {
  const befores = photos.filter(p => p.photo_type === 'before')
  const afters = photos.filter(p => p.photo_type === 'after' || p.photo_type === 'followup')
  const maxLen = Math.max(befores.length, afters.length, 1)
  const pairs = []
  for (let i = 0; i < maxLen; i++) {
    pairs.push({
      before: befores[i] || null,
      after: afters[i] || null
    })
  }
  return pairs.length ? pairs : [{ before: null, after: null }]
}

function photoUrl(id) {
  return `/api/v1/photos/${id}/download?token=${localStorage.getItem('token')}`
}

function formatDate(d) {
  if (!d) return ''
  return String(d).substring(0, 10)
}

function onSlide(e, key) {
  const rect = e.currentTarget.getBoundingClientRect()
  const x = e.clientX - rect.left
  const pct = Math.max(0, Math.min(100, (x / rect.width) * 100))
  slidePos[key] = pct
}
</script>

