<template>
  <div style="max-width:1000px;margin:0 auto">
    <el-card shadow="never" style="margin-bottom:14px">
      <div style="display:flex;align-items:center;gap:10px;flex-wrap:wrap">
        <span style="font-weight:600;font-size:15px">病历模版管理</span>
        <div style="flex:1"/>
        <el-select v-model="filterCategory" placeholder="筛选分类" clearable style="width:140px" @change="loadTemplates">
          <el-option v-for="t in recordTypes" :key="t.key" :label="t.label" :value="t.key"/>
        </el-select>
        <el-button type="primary" size="small" @click="openCreate">+ 新建模版</el-button>
      </div>
    </el-card>

    <el-card shadow="never">
      <el-table :data="templates" v-loading="loading" stripe size="small" empty-text="暂无模版，点击上方按钮创建第一个模版">
        <el-table-column label="模版名称" min-width="180" prop="name"/>
        <el-table-column label="分类" width="110">
          <template #default="{row}">{{ categoryLabel(row.category) }}</template>
        </el-table-column>
        <el-table-column label="关联项目" width="150" prop="procedure_name"/>
        <el-table-column label="说明" min-width="200" prop="description" show-overflow-tooltip/>
        <el-table-column label="状态" width="70">
          <template #default="{row}">
            <el-tag :type="row.is_active?'success':'info'" size="small">{{ row.is_active?'启用':'停用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{row}">
            <el-button size="small" @click="editTemplate(row)">编辑</el-button>
            <el-button size="small" text type="danger" @click="deleteTemplate(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- ============== Create/Edit Dialog ============== -->
    <el-dialog v-model="dialogVisible" :title="isEdit?'编辑模版':'新建模版'" width="740px" top="2vh" destroy-on-close>
      <el-form label-width="100px" size="small">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="模版名称" required>
              <el-input v-model="form.name" placeholder="例如：门诊病历标准模版"/>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="分类" required>
              <el-select v-model="form.category" style="width:100%" @change="onCategoryChange">
                <el-option v-for="t in recordTypes" :key="t.key" :label="t.label" :value="t.key"/>
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="关联项目">
          <el-input v-model="form.procedure_name" placeholder="填写后新建病历时自动选中该项目"/>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="模版用途说明"/>
        </el-form-item>

        <el-divider>模版内容（填写各字段的默认值）</el-divider>
        <div v-if="currentFields.length > 0" style="max-height:400px;overflow-y:auto;padding-right:8px">
          <div v-for="(f, fi) in currentFields" :key="fi" style="margin-bottom:12px">
            <div v-if="f.type==='section'" style="font-weight:700;font-size:14px;color:#303133;border-bottom:2px solid #409eff;padding-bottom:4px;margin-bottom:8px">{{ f.label }}</div>
            <template v-else>
              <div style="font-weight:600;font-size:13px;margin-bottom:4px">
                {{ f.label }}<span v-if="f.required" style="color:#f56c6c">*</span>
                <span v-if="f.unit" style="font-weight:400;color:#909399;font-size:11px">（{{ f.unit }}）</span>
              </div>
              <el-input v-if="f.type==='textarea'" v-model="formContent[f.key]" type="textarea" :rows="3" placeholder="输入默认内容..."/>
              <el-input v-else-if="f.type==='text'" v-model="formContent[f.key]" placeholder="输入默认值"/>
              <el-input v-else-if="f.type==='number'" v-model.number="formContent[f.key]" type="number" style="width:200px"/>
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
              <el-input v-else v-model="formContent[f.key]"/>
            </template>
          </div>
        </div>
        <div v-else style="text-align:center;padding:24px;color:#909399;font-size:13px">请先选择分类</div>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible=false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveTemplate">保存模版</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
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

function categoryLabel(key: string): string {
  const t = recordTypes.find(r => r.key === key)
  return t ? t.label : key
}

// ======================== Field Definitions ========================
interface FieldDef {
  key: string; label: string; type: string; required?: boolean
  options?: string; default?: string; unit?: string
}

const FIELD_DEFS: Record<string, FieldDef[]> = {
  '门诊病历': [
    { key: 'chief_complaint', label: '主诉', type: 'textarea', required: true },
    { key: 'present_illness', label: '现病史', type: 'textarea', required: true },
    { key: 'past_history', label: '既往史', type: 'textarea' },
    { key: 'allergy', label: '过敏史', type: 'textarea' },
    { key: 'physical_exam', label: '体格检查', type: 'textarea' },
    { key: 'aux_exam', label: '辅助检查', type: 'textarea' },
    { key: 'diagnosis', label: '初步诊断', type: 'textarea', required: true },
    { key: 'treatment_plan', label: '治疗意见', type: 'textarea', required: true },
  ],
  '知情同意书': [
    { key: 'diagnosis', label: '诊断', type: 'textarea', required: true },
    { key: 'procedure_name', label: '手术/治疗名称', type: 'text', required: true },
    { key: 'purpose', label: '治疗目的', type: 'textarea', required: true },
    { key: 'expected_result', label: '预期效果', type: 'textarea' },
    { key: 'risks', label: '风险说明', type: 'textarea', required: true },
    { key: 'alternatives', label: '替代方案', type: 'textarea' },
    { key: 'postop_notes', label: '术后注意事项', type: 'textarea' },
    { key: 'patient_stmt', label: '患者声明', type: 'textarea' },
  ],
  '手术记录': [
    { key: 'preop_diag', label: '术前诊断', type: 'textarea', required: true },
    { key: 'postop_diag', label: '术后诊断', type: 'textarea', required: true },
    { key: 'procedure', label: '手术名称', type: 'text', required: true },
    { key: 'surgeon', label: '术者', type: 'text' },
    { key: 'assistant', label: '助手', type: 'text' },
    { key: 'anesthesia', label: '麻醉方式', type: 'select', options: '局部浸润,表面麻醉,静脉镇静,全麻,椎管内麻醉,无,其他' },
    { key: 'procedure_findings', label: '术中所见', type: 'textarea' },
    { key: 'procedure_steps', label: '手术步骤', type: 'textarea' },
    { key: 'blood_loss', label: '出血量', type: 'text', unit: 'ml' },
    { key: 'specimen', label: '送检标本', type: 'text' },
    { key: 'postop_orders', label: '术后医嘱', type: 'textarea' },
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
    { key: 'institution', label: '医疗机构', type: 'text' },
  ],
}

function splitOpts(s?: string) { return s ? s.split(',').map(x=>x.trim()).filter(Boolean) : [] }

// ======================== State ========================
const loading = ref(false)
const templates = ref<any[]>([])
const filterCategory = ref('')
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<number|null>(null)
const saving = ref(false)

const form = ref({ name: '', category: '', procedure_name: '', description: '' })
const formContent = ref<any>({})

// ======================== Computed ========================
const currentFields = computed(() => FIELD_DEFS[form.value.category as string] || [])

// ======================== Data Loading ========================
async function loadTemplates() {
  loading.value = true
  try {
    const params: any = {}
    if (filterCategory.value) params.category = filterCategory.value
    const r = await api.get('/medical/templates', { params })
    templates.value = Array.isArray(r.data) ? r.data : []
  } catch { templates.value = [] }
  finally { loading.value = false }
}

// ======================== Form ========================
function resetForm() {
  form.value = { name: '', category: '', procedure_name: '', description: '' }
  formContent.value = {}
}

function onCategoryChange() {
  const fields = FIELD_DEFS[form.value.category as string] || []
  formContent.value = {}
  fields.forEach((f: FieldDef) => {
    if (f.type === 'checkbox') formContent.value[f.key] = []
    else formContent.value[f.key] = f.default || ''
  })
}

function openCreate() {
  isEdit.value = false; editingId.value = null
  resetForm()
  dialogVisible.value = true
}

async function editTemplate(row: any) {
  isEdit.value = true; editingId.value = row.id
  form.value = {
    name: row.name || '',
    category: row.category || '',
    procedure_name: row.procedure_name || '',
    description: row.description || '',
  }
  try {
    formContent.value = JSON.parse(row.fields || '{}')
  } catch {
    formContent.value = {}
  }
  // Ensure all fields are initialized
  const fields = FIELD_DEFS[form.value.category as string] || []
  fields.forEach((f: FieldDef) => {
    if (formContent.value[f.key] === undefined) {
      if (f.type === 'checkbox') formContent.value[f.key] = []
      else formContent.value[f.key] = f.default || ''
    }
  })
  dialogVisible.value = true
}

async function saveTemplate() {
  if (!form.value.name) { ElMessage.warning('请输入模版名称'); return }
  if (!form.value.category) { ElMessage.warning('请选择分类'); return }
  saving.value = true
  try {
    const payload = {
      name: form.value.name,
      category: form.value.category,
      procedure_name: form.value.procedure_name,
      description: form.value.description,
      fields: JSON.stringify(formContent.value),
    }
    if (editingId.value) {
      await api.put('/medical/templates/' + editingId.value, payload)
      ElMessage.success('模版已更新')
    } else {
      await api.post('/medical/templates', payload)
      ElMessage.success('模版已创建')
    }
    dialogVisible.value = false
    await loadTemplates()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || '保存失败')
  } finally { saving.value = false }
}

async function deleteTemplate(row: any) {
  try {
    await ElMessageBox.confirm(`确定删除模版「${row.name}」？`, '确认删除', { type: 'warning' })
    await api.delete('/medical/templates/' + row.id)
    ElMessage.success('已删除')
    await loadTemplates()
  } catch { }
}

// ======================== Init ========================
onMounted(loadTemplates)
</script>
