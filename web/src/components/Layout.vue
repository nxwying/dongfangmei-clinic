<template>
  <el-container style="height: 100vh">
    <el-aside :width="isCollapse ? '64px' : '230px'" style="background: #304156">
      <div class="logo" style="display:flex;align-items:center;justify-content:space-between;padding:0 12px">
        <span style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1">{{ isCollapse ? (appName.slice(0,1)||'医') : appName }}</span>
        <el-button v-if="!isCollapse" text size="small" style="color:#bfcbd9;font-size:14px;padding:0;min-width:auto" @click.stop="showNameDialog=true">✎</el-button>
        <el-button v-if="!isCollapse" text size="small" style="color:#bfcbd9;font-size:14px;padding:0;min-width:auto" @click.stop="showThemeDialog=true">🎨</el-button>
      </div>
      
      <el-dialog v-model="showNameDialog" title="修改系统名称" width="400px" append-to-body>
        <el-input v-model="editName" placeholder="输入系统名称" @keyup.enter="saveName"/>
        <template #footer>
          <el-button @click="showNameDialog=false">取消</el-button>
          <el-button type="primary" :loading="savingName" @click="saveName">保存</el-button>
        </template>
      </el-dialog>
      <el-dialog v-model="showThemeDialog" title="自定义主题" width="500px" append-to-body>
        <el-form label-width="100px" size="small">
          <el-form-item label="主题色">
            <div style="display:flex;align-items:center;gap:8px">
              <el-color-picker v-model="themeForm.primary_color" show-alpha :predefine="['#409EFF','#5B8FF9','#722ED1','#EB2F96','#FA541C','#F6BD16','#52C41A','#13C2C2']"/>
              <span style="color:#909399;font-size:12px">按钮、选中项颜色</span>
            </div>
          </el-form-item>
          <el-form-item label="侧栏背景">
            <div style="display:flex;align-items:center;gap:8px">
              <el-color-picker v-model="themeForm.sidebar_bg" :predefine="['#304156','#001529','#1E1E2E','#283046','#2C3E50','#FFFFFF']"/>
              <span style="color:#909399;font-size:12px">左侧菜单栏背景色</span>
            </div>
          </el-form-item>
          <el-form-item label="侧栏文字">
            <div style="display:flex;align-items:center;gap:8px">
              <el-color-picker v-model="themeForm.sidebar_text" :predefine="['#bfcbd9','#FFFFFF','#333333','#606266']"/>
              <span style="color:#909399;font-size:12px">左侧菜单文字颜色</span>
            </div>
          </el-form-item>
          <el-form-item label="侧栏选中">
            <div style="display:flex;align-items:center;gap:8px">
              <el-color-picker v-model="themeForm.sidebar_active" :predefine="['#409EFF','#5B8FF9','#722ED1','#52C41A','#FFFFFF']"/>
              <span style="color:#909399;font-size:12px">选中菜单项文字颜色</span>
            </div>
          </el-form-item>
          <el-form-item label="字号">
            <el-select v-model="themeForm.font_size" style="width:120px">
              <el-option label="小 (13px)" value="13px"/>
              <el-option label="中 (14px)" value="14px"/>
              <el-option label="大 (16px)" value="16px"/>
            </el-select>
          </el-form-item>
        </el-form>
        <div style="padding:12px;background:#f5f7fa;border-radius:6px;margin-top:8px">
          <div style="font-size:13px;color:#909399;margin-bottom:4px">实时预览</div>
          <div style="display:flex;gap:8px;align-items:center">
            <div :style="{background:themeForm.sidebar_bg,color:themeForm.sidebar_text,padding:'6px 12px',borderRadius:'4px',fontSize:themeForm.font_size,width:'120px'}">
              侧栏
            </div>
            <el-tag :color="themeForm.primary_color" style="color:#fff;border:none">按钮</el-tag>
            <span :style="{color:themeForm.sidebar_active,fontWeight:600,fontSize:themeForm.font_size}">选中项</span>
          </div>
        </div>
        <template #footer>
          <el-button @click="showThemeDialog=false">取消</el-button>
          <el-button type="primary" :loading="savingTheme" @click="saveTheme">保存主题</el-button>
        </template>
      </el-dialog>
      <el-menu
        :default-active="route.path"
        :collapse="isCollapse"
        :background-color="themeVars.sidebarBg"
        :text-color="themeVars.sidebarText"
        :active-text-color="themeVars.sidebarActive"
        router
        style="border-right: none"
      >
        <!-- ====== 日常运营 ====== -->
        <el-sub-menu index="ops">
          <template #title>
            <el-icon><Odometer /></el-icon>
            <span>日常运营</span>
          </template>
          <el-menu-item index="/dashboard">
            <el-icon><DataAnalysis /></el-icon>
            <span>工作台</span>
          </el-menu-item>
          <el-menu-item index="/appointments">
            <el-icon><Calendar /></el-icon>
            <span>预约管理</span>
          </el-menu-item>
          <el-menu-item index="/pos">
            <el-icon><ShoppingCart /></el-icon>
            <span>收银台</span>
          </el-menu-item>
          <el-menu-item index="/pos/refund">
            <el-icon><Coin /></el-icon>
            <span>退款管理</span>
          </el-menu-item>
          <el-menu-item index="/followup">
            <el-icon><Phone /></el-icon>
            <span>回访管理</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- ====== 客户管理 ====== -->
        <el-sub-menu index="crm">
          <template #title>
            <el-icon><UserFilled /></el-icon>
            <span>客户管理</span>
          </template>
          <el-menu-item index="/customers">
            <el-icon><User /></el-icon>
            <span>客户档案</span>
          </el-menu-item>
          <el-menu-item index="/members">
            <el-icon><Wallet /></el-icon>
            <span>会员管理</span>
          </el-menu-item>
          <el-menu-item index="/marketing">
            <el-icon><Promotion /></el-icon>
            <span>营销工具</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- ====== 医疗管理 ====== -->
        <el-sub-menu index="med">
          <template #title>
            <el-icon><FirstAidKit /></el-icon>
            <span>医疗管理</span>
          </template>
          <el-menu-item index="/medical">
            <el-icon><Folder /></el-icon>
            <span>病历管理</span>
          </el-menu-item>
          <el-menu-item index="/medical/templates">
            <el-icon><Document /></el-icon>
            <span>病历模版</span>
          </el-menu-item>
          <el-menu-item index="/inventory">
            <el-icon><Box /></el-icon>
            <span>库存管理</span>
          </el-menu-item>
          <el-menu-item index="/documents">
            <el-icon><Files /></el-icon>
            <span>证件档案</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- ====== 数据分析 ====== -->
        <el-sub-menu index="data_center">
          <template #title>
            <el-icon><DataLine /></el-icon>
            <span>数据分析</span>
          </template>
          <el-menu-item index="/data">
            <el-icon><DataAnalysis /></el-icon>
            <span>数据中心</span>
          </el-menu-item>
          <el-menu-item index="/performance">
            <el-icon><TrendCharts /></el-icon>
            <span>绩效中心</span>
          </el-menu-item>
          <el-menu-item index="/reports-hub">
            <el-icon><Document /></el-icon>
            <span>报表中心</span>
          </el-menu-item>
          <el-menu-item index="/analysis">
            <el-icon><Histogram /></el-icon>
            <span>运营分析</span>
          </el-menu-item>
          <el-menu-item index="/kpi">
            <el-icon><DataBoard /></el-icon>
            <span>KPI目标</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- ====== 系统设置 ====== -->
        <el-sub-menu index="/settings">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </template>
          <el-menu-item index="/finance/expenses">支出管理</el-menu-item>
          <el-menu-item index="/settings/items">项目管理</el-menu-item>
          <el-menu-item index="/settings/package-templates">套餐模板</el-menu-item>
          <el-menu-item index="/settings/users">员工管理</el-menu-item>
          <el-menu-item index="/settings/commission">提成规则</el-menu-item>
          <el-menu-item index="/settings/roles">角色管理</el-menu-item>
          <el-menu-item index="/settings/license">授权管理</el-menu-item>
          <el-menu-item index="/settings/training">培训认证</el-menu-item>
          <el-menu-item index="/settings/tags">自动标签</el-menu-item>
          <el-menu-item index="/settings/backup">数据备份</el-menu-item>
          <el-menu-item index="/settings/audit-logs">操作日志</el-menu-item>
        </el-sub-menu>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header style="background: #fff; border-bottom: 1px solid #e6e6e6; display: flex; align-items: center; justify-content: space-between; padding: 0 20px; height: 60px;">
        <div style="display:flex;align-items:center;gap:12px">
          <el-icon style="cursor: pointer; font-size: 20px;" @click="isCollapse = !isCollapse">
            <Fold v-if="!isCollapse" />
            <Expand v-else />
          </el-icon>
          <span style="font-size:13px;color:#909399">{{ pageTitle }}</span>
        </div>
        <div>
          <el-dropdown @command="handleCommand">
            <span style="cursor: pointer; display: flex; align-items: center; gap: 6px;">
              {{ auth.user?.real_name || '用户' }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main style="background: #f0f2f5; padding: 20px;">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { getSystemConfig, updateSystemConfig } from '../api/settings'
import type { ThemeConfig } from '../api/settings'
import { ElMessage } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const appName = ref('东芳美诊所管理系统')
const isCollapse = ref(false)
const showNameDialog = ref(false)
const editName = ref('')
const showThemeDialog = ref(false)
const savingTheme = ref(false)

const defaultTheme: ThemeConfig = {
  primary_color: '#409EFF',
  sidebar_bg: '#304156',
  sidebar_text: '#bfcbd9',
  sidebar_active: '#409EFF',
  font_size: '14px',
  primaryColor: '#409EFF',
  sidebarBg: '#304156',
  sidebarText: '#bfcbd9',
  sidebarActive: '#409EFF',
}
const themeVars = reactive({...defaultTheme, sidebarBg: defaultTheme.sidebar_bg, sidebarText: defaultTheme.sidebar_text, sidebarActive: defaultTheme.sidebar_active, primaryColor: defaultTheme.primary_color})
const themeForm = reactive<ThemeConfig>({...defaultTheme})

function applyTheme(t: ThemeConfig) {
  const root = document.documentElement
  root.style.setProperty('--el-color-primary', t.primary_color)
  root.style.setProperty('--sidebar-bg', t.sidebar_bg)
  root.style.setProperty('--sidebar-text', t.sidebar_text)
  root.style.setProperty('--sidebar-active', t.sidebar_active)
  root.style.setProperty('--el-font-size-base', t.font_size)
  themeVars.primaryColor = t.primary_color
  themeVars.sidebarBg = t.sidebar_bg
  themeVars.sidebarText = t.sidebar_text
  themeVars.sidebarActive = t.sidebar_active
  themeVars.font_size = t.font_size
  themeVars.primary_color = t.primary_color
  themeVars.sidebar_bg = t.sidebar_bg
  themeVars.sidebar_text = t.sidebar_text
  themeVars.sidebar_active = t.sidebar_active
}

async function saveTheme() {
  savingTheme.value = true
  try {
    await updateSystemConfig({ theme: { ...themeForm } })
    applyTheme(themeForm)
    showThemeDialog.value = false
    ElMessage.success('主题已保存')
  } catch { ElMessage.error('保存失败') }
  finally { savingTheme.value = false }
}
const savingName = ref(false)

async function saveName() {
  if (!editName.value.trim()) { ElMessage.warning('名称不能为空'); return }
  savingName.value = true
  try {
    await updateSystemConfig({ app_name: editName.value.trim() })
    appName.value = editName.value.trim()
    document.title = appName.value
    showNameDialog.value = false
    ElMessage.success('名称已修改')
  } catch { ElMessage.error('保存失败') }
  finally { savingName.value = false }
}

const pageTitle = computed(() => {
  const title = route.meta?.title as string
  return title ? '当前：' + title : ''
})

onMounted(async () => {
  try {
    const cfg = await getSystemConfig()
    if (cfg?.app_name) {
      appName.value = cfg.app_name
      document.title = cfg.app_name
    }
    if (cfg?.theme) {
      applyTheme(cfg.theme)
    }
  } catch {}

  if (!auth.user) {
    try {
      await auth.fetchProfile()
    } catch {
      auth.logout()
      router.push('/login')
    }
  }
})

function hasPerm(key: string): boolean {
  if (!key) return true
  const perms = (auth.user as any)?.permissions || []
  return perms.includes('admin') || perms.includes(key)
}

function handleCommand(cmd: string) {
  if (cmd === 'logout') {
    auth.logout()
    router.push('/login')
  }
}
</script>

<style scoped>
.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  color: #fff;
  font-size: 15px;
  font-weight: bold;
  overflow: hidden;
  white-space: nowrap;
  border-bottom: 1px solid rgba(255,255,255,.08);
}
.el-menu {
  overflow-x: hidden;
}
.el-menu--collapse .logo {
  font-size: 18px;
}
</style>
