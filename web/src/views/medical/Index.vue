<template>
  <div style="max-width:1200px;margin:0 auto">
    <!-- Tabs -->
    <el-card shadow="never" style="margin-bottom:14px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">病历管理</span>
        <div style="flex:1"/>
        <el-input v-model="searchText" placeholder="搜索客户姓名/手机号" clearable style="width:200px" @input="loadList"/>
        <el-button type="primary" size="small" @click="openCreate">+ 新建病历</el-button>
      </div>
    </el-card>

    <el-card shadow="never" style="margin-bottom:14px">
      <el-tabs v-model="activeTab" @tab-change="loadList">
        <el-tab-pane v-for="t in recordTypes" :key="t.key" :label="t.label" :name="t.key"/>
      </el-tabs>
    </el-card>

    <!-- Stats row -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#409eff">{{ stats.total }}</div><div style="font-size:12px;color:#909399">总记录</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#67c23a">{{ stats.signed }}</div><div style="font-size:12px;color:#909399">已签字</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#e6a23c">{{ stats.draft }}</div><div style="font-size:12px;color:#909399">草稿</div></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#f56c6c">{{ stats.archived }}</div><div style="font-size:12px;color:#909399">已归档</div></div></el-card></el-col>
    </el-row>

    <!-- Record list -->
    <el-card shadow="never">
      <el-table :data="records" v-loading="loading" stripe size="small" empty-text="暂无病历记录">
        <el-table-column label="客户" min-width="120"><template #default="{row}">{{ row.customer?.name||'未知' }}</template></el-table-column>
        <el-table-column label="类型" width="90"><template #default="{row}">{{ typeLabel(row.record_type) }}</template></el-table-column>
        <el-table-column label="日期" width="90" prop="record_date"/>
        <el-table-column label="医生" width="80" prop="doctor_name"/>
        <el-table-column label="状态" width="70">
          <template #default="{row}">
            <el-tag :type="row.status==='signed'?'success':row.status==='archived'?'info':'warning'" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="260" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="viewRecord(row)">查看</el-button>
            <el-button size="small" @click="editRecord(row)">编辑</el-button>
            <el-button size="small" @click="openPhotos(row)">照片</el-button>
            <el-button v-if="row.status==='draft'" size="small" type="success" @click="signRecord(row)">签字</el-button>
            <el-button size="small" text type="danger" @click="deleteRecord(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- ============== Create/Edit Dialog ============== -->
    <el-dialog v-model="formVisible" :title="isEdit?'编辑病历':'新建病历'" width="780px" top="2vh" destroy-on-close>
      <el-form label-width="100px" size="small">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="客户" required>
              <el-select v-model="formData.customer_id" filterable remote :remote-method="searchCustomers" :loading="custLoading" placeholder="搜索客户" style="width:100%">
                <el-option v-for="c in custOptions" :key="c.id" :label="`${c.name}(${c.phone||''})`" :value="c.id"/>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="病历类型" required>
              <el-select v-model="formData.record_type" style="width:100%" @change="onTypeChange">
                <el-option v-for="t in recordTypes" :key="t.key" :label="t.label" :value="t.key"/>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="日期">
              <el-date-picker v-model="formData.record_date" type="date" value-format="YYYY-MM-DD" style="width:100%"/>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="医生">
              <el-input v-model="formData.doctor_name" placeholder="经治医生"/>
            </el-form-item>
          </el-col>
        </el-row>

        <el-divider>病历内容</el-divider>

        <div v-if="currentFields.length > 0">
          <div v-for="(f, fi) in currentFields" :key="fi" style="margin-bottom:12px">
            <div v-if="f.type==='section'" style="font-weight:700;font-size:14px;color:#303133;border-bottom:2px solid #409eff;padding-bottom:4px;margin-bottom:8px">{{ f.label }}</div>
            <template v-else>
              <div style="font-weight:600;font-size:13px;margin-bottom:4px;color:#303133">
                {{ f.label }}<span v-if="f.required" style="color:#f56c6c">*</span>
                <span v-if="f.unit" style="font-weight:400;color:#909399;font-size:11px">（{{ f.unit }}）</span>
              </div>
              <el-input v-if="f.type==='textarea'" v-model="formContent[f.key]" type="textarea" :rows="3" :placeholder="f.default||''"/>
              <el-input v-else-if="f.type==='text'" v-model="formContent[f.key]" :placeholder="f.default||''"/>
              <el-input v-else-if="f.type==='number'" v-model.number="formContent[f.key]" type="number" style="width:200px" :placeholder="f.default||''"/>
              <el-select v-else-if="f.type==='select'" v-model="formContent[f.key]" style="width:100%">
                <el-option v-for="o in splitOpts(f.options)" :key="o" :label="o" :value="o"/>
              </el-select>
              <el-checkbox-group v-else-if="f.type==='checkbox'" v-model="formContent[f.key]">
                <el-checkbox v-for="o in splitOpts(f.options)" :key="o" :label="o" :value="o"/>
              </el-checkbox-group>
              <el-date-picker v-else-if="f.type==='date'" v-model="formContent[f.key]" type="date" value-format="YYYY-MM-DD" style="width:100%"/>
              <el-radio-group v-else-if="f.type==='radio'" v-model="formContent[f.key]">
                <el-radio v-for="o in splitOpts(f.options)" :key="o" :label="o" :value="o"/>
              </el-radio-group>
              <!-- Signature canvas for consent form -->
              <div v-else-if="f.type==='signature'" style="border:1px solid #dcdfe6;border-radius:4px;padding:8px">
                <canvas :ref="(el)=>{if(el)signCanvasRefs[f.key]=el as HTMLCanvasElement}" :data-key="f.key" width="360" height="80" style="border:1px dashed #ccc;display:block;cursor:crosshair;width:100%"
                  @mousedown="startSign(f.key,$event)" @mousemove="moveSign(f.key,$event)" @mouseup="endSign" @mouseleave="endSign"/>
                <el-button size="small" @click="clearSign(f.key)" style="margin-top:4px">清除</el-button>
              </div>
              <el-input v-else v-model="formContent[f.key]" :placeholder="f.default||''"/>
            </template>
          </div>
        </div>
        <div v-else style="text-align:center;padding:24px;color:#909399;font-size:13px">请先选择病历类型</div>
      </el-form>
      <template #footer>
        <el-button @click="formVisible=false">取消</el-button>
        <el-button @click="saveDraft">存草稿</el-button>
        <el-button type="primary" :loading="saving" @click="saveAndSign">保存并签字</el-button>
      </template>
    </el-dialog>

    <!-- ============== View Dialog ============== -->
    <el-dialog v-model="viewVisible" title="病历详情" width="780px" top="2vh">
      <div v-if="viewData" style="background:#fff;padding:16px" id="print-area">
        <!-- Header -->
        <div style="text-align:center;font-size:18px;font-weight:700;margin-bottom:8px;border-bottom:2px solid #303133;padding-bottom:8px">
          {{ typeLabel(viewData.record_type) }}
        </div>
        <div style="display:flex;gap:16px;margin-bottom:8px;font-size:13px;color:#606266;flex-wrap:wrap">
          <span><b>患者：</b>{{ viewData.customer?.name||'未知' }}</span>
          <span><b>日期：</b>{{ viewData.record_date }}</span>
          <span><b>医生：</b>{{ viewData.doctor_name||'-' }}</span>
          <span><b>状态：</b>{{ statusLabel(viewData.status) }}</span>
        </div>
        <el-divider style="margin:8px 0"/>
        <!-- Content -->
        <div v-if="viewFields.length > 0" style="font-size:14px;line-height:1.8">
          <div v-for="(item, fi) in viewFields" :key="fi" style="margin-bottom:4px">
            <div v-if="item.type==='section'" style="font-weight:700;font-size:15px;color:#303133;border-bottom:1px solid #eee;padding-bottom:4px;margin:10px 0 6px">{{ item.label }}</div>
            <div v-else-if="item.type==='signature'" style="padding:4px 0">
              <span style="color:#606266">{{ item.label }}：</span>
              <img v-if="viewContentParsed[item.key]" :src="viewContentParsed[item.key]" style="max-height:50px;border:1px solid #eee;border-radius:2px;vertical-align:middle"/>
              <span v-else style="color:#c0c4cc">未签署</span>
            </div>
            <div v-else style="display:flex;padding:2px 0">
              <span style="min-width:90px;color:#606266;flex-shrink:0">{{ item.label }}：</span>
              <span style="color:#303133;white-space:pre-wrap">{{ getDisplayValue(item) }}</span>
            </div>
          </div>
        </div>
        <div v-else style="text-align:center;padding:24px;color:#909399">暂无内容</div>
      </div>
      <template #footer>
        <el-button @click="printRecord">🖨️ 打印</el-button>
        <el-button @click="viewVisible=false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- ============== Photos Dialog ============== -->
    <el-dialog v-model="photosVisible" title="病历照片" width="600px">
      <div v-if="photosRecord">
        <div style="margin-bottom:12px;display:flex;align-items:center;gap:10px">
          <el-select v-model="photoType" placeholder="选择照片类型" style="width:160px">
            <el-option label="术前" value="before"/>
            <el-option label="术中" value="during"/>
            <el-option label="术后" value="after"/>
            <el-option label="回访" value="followup"/>
          </el-select>
          <el-upload :http-request="uploadPhotoViaApi"
            :show-file-list="false" style="display:inline-block">
            <el-button size="small" type="primary">+ 上传照片</el-button>
          </el-upload>
        </div>
        <div style="display:flex;flex-wrap:wrap;gap:10px">
          <div v-for="(p,pi) in photosList" :key="pi" style="width:120px;height:120px;border:1px solid #ddd;border-radius:4px;overflow:hidden;position:relative">
            <img :src="'/api/v1/photos/'+p.id+'/download'" style="width:100%;height:100%;object-fit:cover" @click="viewPhoto('/api/v1/photos/'+p.id+'/download')"/>
            <div style="position:absolute;top:0;right:0;background:rgba(0,0,0,0.5);color:#fff;padding:0 6px;cursor:pointer;font-size:12px" @click="deletePhoto(p.id)">✕</div>
          </div>
        </div>
      </div>
      <div v-else style="text-align:center;padding:24px;color:#909399">请先保存病历后再上传照片</div>
      <template #footer><el-button @click="photosVisible=false">关闭</el-button></template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../../api'

// ======================== Record Type Definitions ========================
const recordTypes = [
  { key: '门诊病历', label: '门诊病历' },
  { key: '知情同意书', label: '知情同意书' },
  { key: '手术记录', label: '手术记录' },
  { key: '治疗记录', label: '治疗记录' },
  { key: '病程记录', label: '病程记录' },
  { key: '诊断证明', label: '诊断证明' },
]

function typeLabel(key: string): string {
  const t = recordTypes.find(r => r.key === key)
  return t ? t.label : key
}

function statusLabel(s: string): string {
  return { draft: '草稿', signed: '已签', archived: '已归档' }[s] || s
}

// ======================== Field Definitions per Type ========================
interface FieldDef {
  key: string; label: string; type: string; required?: boolean
  options?: string; default?: string; unit?: string
}

const FIELD_DEFS: Record<string, FieldDef[]> = {
  '门诊病历': [
    { key: 'chief_complaint', label: '主诉', type: 'textarea', required: true, default: '【患者主诉症状及持续时间】' },
    { key: 'present_illness', label: '现病史', type: 'textarea', required: true, default: '患者自述【症状描述】' },
    { key: 'past_history', label: '既往史', type: 'textarea', default: '【既往病史、手术史、用药史】' },
    { key: 'allergy', label: '过敏史', type: 'textarea', default: '【药物/食物过敏史】' },
    { key: 'physical_exam', label: '体格检查', type: 'textarea', default: '生命体征：【】专科检查：【】' },
    { key: 'aux_exam', label: '辅助检查', type: 'textarea', default: '【辅助检查结果】' },
    { key: 'diagnosis', label: '初步诊断', type: 'textarea', required: true, default: '【诊断名称】' },
    { key: 'treatment_plan', label: '治疗意见', type: 'textarea', required: true, default: '建议【治疗方案】' },
  ],
  '知情同意书': [
    { key: 'diagnosis', label: '诊断', type: 'textarea', required: true, default: '【诊断名称】' },
    { key: 'procedure_name', label: '手术/治疗名称', type: 'text', required: true, default: '【手术/治疗项目】' },
    { key: 'purpose', label: '治疗目的', type: 'textarea', required: true, default: '为了改善【具体问题】' },
    { key: 'expected_result', label: '预期效果', type: 'textarea', default: '【预期效果】' },
    { key: 'risks', label: '风险说明', type: 'textarea', required: true, default: '【风险及并发症说明】' },
    { key: 'alternatives', label: '替代方案', type: 'textarea', default: '【替代方案】' },
    { key: 'postop_notes', label: '术后注意事项', type: 'textarea', default: '【术后注意事项】' },
    { key: 'patient_stmt', label: '患者声明', type: 'textarea', default: '本人已认真阅读并充分理解以上内容，同意接受治疗。' },
    { key: 'patient_sign', label: '患者签名', type: 'signature' },
    { key: 'doctor_sign', label: '经治医师签名', type: 'signature' },
  ],
  '手术记录': [
    { key: 'preop_diag', label: '术前诊断', type: 'textarea', required: true, default: '【术前诊断】' },
    { key: 'postop_diag', label: '术后诊断', type: 'textarea', required: true, default: '【术后诊断】' },
    { key: 'procedure', label: '手术名称', type: 'text', required: true, default: '【手术名称】' },
    { key: 'surgeon', label: '术者', type: 'text', default: '【术者姓名】' },
    { key: 'assistant', label: '助手', type: 'text', default: '【助手姓名】' },
    { key: 'anesthesia', label: '麻醉方式', type: 'select', options: '局部浸润,表面麻醉,静脉镇静,全麻,椎管内麻醉,无,其他' },
    { key: 'procedure_findings', label: '术中所见', type: 'textarea', default: '术区情况：【】' },
    { key: 'procedure_steps', label: '手术步骤', type: 'textarea', default: '1.【手术操作步骤】' },
    { key: 'blood_loss', label: '出血量', type: 'text', unit: 'ml' },
    { key: 'specimen', label: '送检标本', type: 'text', default: '【送检标本】' },
    { key: 'postop_orders', label: '术后医嘱', type: 'textarea', default: '【术后医嘱】' },
  ],
  '治疗记录': [
    { key: 'treatment_item', label: '治疗项目', type: 'text', required: true },
    { key: 'doctor', label: '操作医师', type: 'text' },
    { key: 'findings', label: '治疗所见', type: 'textarea' },
    { key: 'treatment_desc', label: '治疗经过', type: 'textarea' },
    { key: 'reaction', label: '治疗反应', type: 'radio', options: '良好,一般,较差' },
    { key: 'notes', label: '注意事项', type: 'textarea' },
    { key: 'next_visit', label: '复诊建议', type: 'text' },
  ],
  '病程记录': [
    { key: 'record_time', label: '记录时间', type: 'date' },
    { key: 'subjective', label: '患者主诉', type: 'textarea', required: true },
    { key: 'physical_exam', label: '查体', type: 'textarea' },
    { key: 'aux_results', label: '辅助检查结果', type: 'textarea' },
    { key: 'diagnosis_change', label: '诊断变更', type: 'textarea' },
    { key: 'treatment_adjust', label: '治疗方案调整', type: 'textarea' },
    { key: 'assessment', label: '病情评估', type: 'textarea', required: true },
    { key: 'plan', label: '后续计划', type: 'textarea' },
  ],
  '诊断证明': [
    { key: 'conclusion', label: '诊断结论', type: 'textarea', required: true },
    { key: 'icd_code', label: 'ICD编码', type: 'text' },
    { key: 'doctor', label: '诊断医师', type: 'text' },
    { key: 'diagnosis_date', label: '诊断日期', type: 'date' },
    { key: 'notes', label: '注意事项', type: 'textarea' },
    { key: 'rest_days', label: '建议休养时间', type: 'text', unit: '天' },
    { key: 'institution', label: '医疗机构', type: 'text', default: '东芳美医疗美容诊所' },
  ],
}

function splitOpts(s?: string) { return s ? s.split(',').map(x=>x.trim()).filter(Boolean) : [] }
function parseContent(s: string) { try { return JSON.parse(s||'{}') } catch { return {} } }

// ======================== State ========================
const activeTab = ref('门诊病历')
const searchText = ref('')
const records = ref<any[]>([])
const loading = ref(false)
const formVisible = ref(false)
const viewVisible = ref(false)
const photosVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number|null>(null)
const saving = ref(false)
const custOptions = ref<any[]>([])
const custLoading = ref(false)
const signCanvasRefs = ref<Record<string,HTMLCanvasElement>>({})
const signDrawing = ref<{key:string;ctx:CanvasRenderingContext2D|null;drawing:boolean}>({key:'',ctx:null,drawing:false})

const formData = ref({ customer_id: null, record_type: '', record_date: '', doctor_name: '' })
const formContent = ref<any>({})
const viewData = ref<any>(null)

const photosRecord = ref<any>(null)
const photoType = ref('before')
const photosList = ref<any[]>([])

// ======================== Computed ========================
const stats = computed(() => {
  const all = records.value
  return {
    total: all.length,
    signed: all.filter((r:any)=>r.status==='signed').length,
    draft: all.filter((r:any)=>r.status==='draft').length,
    archived: all.filter((r:any)=>r.status==='archived').length,
  }
})

const currentFields = computed(() => FIELD_DEFS[formData.value.record_type as string] || [])

const viewContentParsed = computed(() => viewData.value ? parseContent(viewData.value.content) : {})

const viewFields = computed(() => {
  if (!viewData.value) return []
  const fields = FIELD_DEFS[viewData.value.record_type as string]
  if (!fields) return []
  const c = parseContent(viewData.value.content)
  return fields.map((f: FieldDef) => ({
    ...f,
    value: f.type === 'signature' ? (c[f.key]||'') : (c[f.key]||''),
    displayValue: f.type === 'checkbox' ? ((c[f.key]||[])||[]).join('、') : (c[f.key]||'-')
  }))
})

function getDisplayValue(item: any): string {
  if (item.type === 'checkbox') return ((viewContentParsed.value[item.key]||[])||[]).join('、')
  if (item.type === 'signature') return viewContentParsed.value[item.key] ? '[已签署]' : '-'
  return viewContentParsed.value[item.key] || '-'
}

// ======================== Data Loading ========================
async function loadList() {
  loading.value = true
  try {
    const params: any = { record_type: activeTab.value }
    if (searchText.value) params.q = searchText.value
    const r = await api.get('/medical/records', { params })
    records.value = Array.isArray(r.data) ? r.data : []
  } catch { records.value = [] }
  finally { loading.value = false }
}

async function searchCustomers(q: string) {
  if (!q) { custOptions.value = []; return }
  custLoading.value = true
  try {
    const r = await api.get('/customers', { params: { q, page_size: 20 } })
    custOptions.value = Array.isArray(r.data) ? r.data : (r.data?.data||[])
  } catch { custOptions.value = [] }
  finally { custLoading.value = false }
}

// ======================== Form ========================
function resetForm() {
  formData.value = { customer_id: null, record_type: activeTab.value, record_date: new Date().toISOString().slice(0,10), doctor_name: '' }
  formContent.value = {}
  const fields = FIELD_DEFS[formData.value.record_type as string] || []
  fields.forEach((f: FieldDef) => {
    if (f.type === 'checkbox') formContent.value[f.key] = []
    else if (f.type === 'signature') formContent.value[f.key] = ''
    else formContent.value[f.key] = f.default || ''
  })
}

function onTypeChange() {
  const fields = FIELD_DEFS[formData.value.record_type as string] || []
  formContent.value = {}
  fields.forEach((f: FieldDef) => {
    if (f.type === 'checkbox') formContent.value[f.key] = []
    else if (f.type === 'signature') formContent.value[f.key] = ''
    else formContent.value[f.key] = f.default || ''
  })
}

function openCreate() {
  isEdit.value = false; editingId.value = null
  resetForm()
  formVisible.value = true
}

async function editRecord(row: any) {
  isEdit.value = true; editingId.value = row.id
  formData.value = {
    customer_id: row.customer_id,
    record_type: row.record_type || '',
    record_date: row.record_date || new Date().toISOString().slice(0,10),
    doctor_name: row.doctor_name || '',
  }
  formContent.value = parseContent(row.content)
  formVisible.value = true
}

async function saveDraft() { await doSave('draft') }
async function saveAndSign() { await doSave('signed') }

async function doSave(status: string) {
  if (!formData.value.customer_id) { ElMessage.warning('请选择客户'); return }
  if (!formData.value.record_type) { ElMessage.warning('请选择病历类型'); return }
  saving.value = true
  try {
    const payload = {
      customer_id: formData.value.customer_id,
      record_type: formData.value.record_type,
      record_date: formData.value.record_date,
      doctor_name: formData.value.doctor_name,
      content: JSON.stringify(formContent.value),
      status,
    }
    if (editingId.value) {
      await api.put('/medical/records/' + editingId.value, payload)
    } else {
      await api.post('/medical/records', payload)
    }
    ElMessage.success(status === 'signed' ? '已保存并签字' : '已保存草稿')
    formVisible.value = false
    await loadList()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  } finally { saving.value = false }
}

// ======================== View / Print ========================
function viewRecord(row: any) {
  viewData.value = row
  viewVisible.value = true
}

function printRecord() {
  window.print()
}

// ======================== Signature ========================
function startSign(key: string, e: MouseEvent) {
  const canvas = signCanvasRefs.value[key]
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const rect = canvas.getBoundingClientRect()
  const scaleX = canvas.width / rect.width
  const scaleY = canvas.height / rect.height
  ctx.strokeStyle = '#000'; ctx.lineWidth = 2; ctx.lineCap = 'round'
  ctx.beginPath()
  ctx.moveTo((e.clientX - rect.left) * scaleX, (e.clientY - rect.top) * scaleY)
  signDrawing.value = { key, ctx, drawing: true }
}

function moveSign(key: string, e: MouseEvent) {
  if (!signDrawing.value.drawing || signDrawing.value.key !== key) return
  const canvas = signCanvasRefs.value[key]
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  const scaleX = canvas.width / rect.width
  const scaleY = canvas.height / rect.height
  signDrawing.value.ctx?.lineTo((e.clientX - rect.left) * scaleX, (e.clientY - rect.top) * scaleY)
  signDrawing.value.ctx?.stroke()
}

function endSign() {
  if (!signDrawing.value.drawing) return
  signDrawing.value.drawing = false
  const canvas = signCanvasRefs.value[signDrawing.value.key]
  if (canvas) formContent.value[signDrawing.value.key] = canvas.toDataURL()
}

function clearSign(key: string) {
  const canvas = signCanvasRefs.value[key]
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  ctx?.clearRect(0, 0, canvas.width, canvas.height)
  formContent.value[key] = ''
}

// ======================== Sign Record ========================
async function signRecord(row: any) {
  try {
    await ElMessageBox.confirm('签字后将锁定病历内容不可修改，确定继续？', '电子签字', { type: 'warning' })
    const r = await api.get('/medical/records/' + row.id)
    const rec = r.data
    const content = parseContent(rec.content)
    // Check if signature fields are present
    if (rec.record_type === '知情同意书' && (!content.patient_sign || !content.doctor_sign)) {
      ElMessage.warning('知情同意书需要患者和医生签名，请先编辑补充签名')
      return
    }
    await api.post('/medical/records/' + row.id + '/sign')
    ElMessage.success('签字成功')
    row.status = 'signed'
    await loadList()
  } catch { }
}

// ======================== Delete ========================
async function deleteRecord(row: any) {
  try {
    await ElMessageBox.confirm('确定删除此病历？删除后不可恢复', '确认', { type: 'warning' })
    await api.delete('/medical/records/' + row.id)
    ElMessage.success('已删除')
    await loadList()
  } catch { }
}

// ======================== Photos ========================
async function openPhotos(row: any) {
  photosRecord.value = row
  photosVisible.value = true
  await loadPhotos()
}

async function loadPhotos() {
  if (!photosRecord.value?.customer_id) return
  try {
    const r = await api.get('/photos', { params: { customer_id: photosRecord.value.customer_id } })
    photosList.value = Array.isArray(r.data) ? r.data : []
  } catch { photosList.value = [] }
}

async function deletePhoto(id: number) {
  try {
    await api.delete('/photos/' + id)
    await loadPhotos()
  } catch { ElMessage.error('删除失败') }
}

function viewPhoto(url: string) {
  window.open(url, '_blank')
}

// ======================== Custom Upload via Axios ========================
async function uploadPhotoViaApi(options: any) {
  const formData = new FormData()
  formData.append('file', options.file)
  formData.append('customer_id', String(photosRecord.value?.customer_id || ''))
  formData.append('photo_type', photoType.value)
  if (photosRecord.value?.id) formData.append('medical_record_id', String(photosRecord.value.id))
  try {
    await api.post('/photos', formData, { headers: { 'Content-Type': 'multipart/form-data' } })
    ElMessage.success('上传成功')
    loadPhotos()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '上传失败')
  }
}

// ======================== Init ========================
onMounted(() => {
  loadList()
})
</script>
