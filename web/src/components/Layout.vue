<template>
  <el-container style="height: 100vh">
    <el-aside :width="isCollapse ? '64px' : '230px'" style="background: #304156">
      <div class="logo">{{ isCollapse ? '医' : '东芳美诊所管理系统' }}</div>
      <el-menu
        :default-active="route.path"
        :collapse="isCollapse"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
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
          <el-menu-item index="/settings/training">培训认证</el-menu-item>
          <el-menu-item index="/settings/tags">自动标签</el-menu-item>
          <el-menu-item index="/settings/backup">数据备份</el-menu-item>
          <el-menu-item index="/settings/audit-logs">操作日志</el-menu-item>
          <el-menu-item index="/settings/license">授权管理</el-menu-item>
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
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const isCollapse = ref(false)

const pageTitle = computed(() => {
  const title = route.meta?.title as string
  return title ? '当前：' + title : ''
})

onMounted(async () => {
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
