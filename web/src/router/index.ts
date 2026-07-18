import { createRouter, createWebHistory } from 'vue-router'
import Layout from '../components/Layout.vue'

const router = createRouter({
  history: createWebHistory("/v2/"),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '',
      component: Layout,
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('../views/Dashboard.vue'),
          meta: { requiresAuth: true, title: '工作台' },
        },
        {
          path: 'customers',
          name: 'Customers',
          component: () => import('../views/customers/Index.vue'),
          meta: { requiresAuth: true, title: '客户管理' },
        },
        {
          path: 'customers/:id',
          name: 'CustomerDetail',
          component: () => import('../views/customers/Detail.vue'),
          meta: { requiresAuth: true, title: '客户详情' },
        },
        {
          path: 'appointments',
          name: 'Appointments',
          component: () => import('../views/appointments/Index.vue'),
          meta: { requiresAuth: true, title: '预约管理' },
        },
        {
          path: 'members',
          name: 'Members',
          component: () => import('../views/members/Index.vue'),
          meta: { requiresAuth: true, title: '会员管理' },
        },

        {
          path: 'pos',
          name: 'Checkout',
          component: () => import('../views/pos/Index.vue'),
          meta: { requiresAuth: true, title: '收银台' },
        },
        {
          path: 'pos/refund',
          name: 'RefundMgmt',
          component: () => import('../views/pos/Refund.vue'),
          meta: { requiresAuth: true, title: '退款管理' },
        },
        {
          path: 'pos/:id',
          name: 'POSDetail',
          component: () => import('../views/pos/Detail.vue'),
          meta: { requiresAuth: true, title: '订单详情' },
        },
        {
          path: 'finance/expenses',
          name: 'Expenses',
          component: () => import('../views/finance/Expenses.vue'),
          meta: { requiresAuth: true, title: '支出管理' },
        },
        {
          path: 'settings/items',
          name: 'SettingsItems',
          component: () => import('../views/settings/Items.vue'),
          meta: { requiresAuth: true, title: '项目管理' },
        },
        {
          path: 'settings/package-templates',
          name: 'SettingsPackages',
          component: () => import('../views/settings/PackageTemplates.vue'),
          meta: { requiresAuth: true, title: '套餐模板' },
        },
        {
          path: 'settings/users',
          name: 'SettingsUsers',
          component: () => import('../views/settings/Users.vue'),
          meta: { requiresAuth: true, title: '员工管理' },
        },
        {
          path: 'settings/backup',
          name: 'BackupRestore',
          component: () => import('../views/settings/BackupRestore.vue'),
          meta: { requiresAuth: true, title: '数据备份' },
        },
        {
          path: 'settings/audit-logs',
          name: 'AuditLogs',
          component: () => import('../views/settings/AuditLogs.vue'),
          meta: { requiresAuth: true, title: '操作日志' },
        },
        {
          path: 'settings/license',
          name: 'License',
          component: () => import('../views/settings/License.vue'),
          meta: { requiresAuth: true, title: '授权管理' },
        },
        {
          path: 'settings/tags',
          name: 'AutoTags',
          component: () => import('../views/settings/AutoTags.vue'),
          meta: { requiresAuth: true, title: '自动标签' },
        },
        {
          path: 'settings/commission',
          name: 'CommissionRules',
          component: () => import('../views/settings/CommissionRules.vue'),
          meta: { requiresAuth: true, title: '提成规则' },
        },
        {
          path: 'settings/training',
          name: 'Training',
          component: () => import('../views/settings/Training.vue'),
          meta: { requiresAuth: true, title: '培训认证' },
        },
        {
          path: 'settings/roles',
          name: 'SettingsRoles',
          component: () => import('../views/settings/Roles.vue'),
          meta: { requiresAuth: true, title: '角色管理' },
        },
        {
          path: 'inventory',
          name: 'Inventory',
          component: () => import('../views/inventory/Index.vue'),
          meta: { requiresAuth: true, title: '库存管理' },
        },
        {
          path: 'followup',
          name: 'FollowUp',
          component: () => import('../views/followup/Index.vue'),
          meta: { requiresAuth: true, title: '回访管理' },
        },
        {
          path: 'marketing',
          name: 'Marketing',
          component: () => import('../views/marketing/Index.vue'),
          meta: { requiresAuth: true, title: '营销工具' },
        },
        {
          path: 'medical',
          name: 'MedicalRecords',
          component: () => import('../views/medical/Index.vue'),
          meta: { requiresAuth: true, title: '病历管理' },
        },
        {
          path: 'performance',
          name: 'Performance',
          component: () => import('../views/performance/Index.vue'),
          meta: { requiresAuth: true, title: '绩效中心' },
        },
        {
          path: 'documents',
          name: 'Documents',
          component: () => import('../views/documents/Index.vue'),
          meta: { requiresAuth: true, title: '证件档案' },
        },
        {
          path: 'documents/inspection',
          name: 'DocumentsInspection',
          component: () => import('../views/documents/Inspection.vue'),
          meta: { requiresAuth: true, title: '检查模式' },
        },

        {
          path: 'data',
          name: 'DataCenter',
          component: () => import('../views/data/Index.vue'),
          meta: { requiresAuth: true, title: '数据中心' },
        },
        {
          path: 'kpi',
          name: 'KpiTargets',
          component: () => import('../views/kpi/KpiTargets.vue'),
          meta: { requiresAuth: true, title: 'KPI目标' },
        },
        {
          path: 'kpi/leaderboard',
          name: 'Leaderboard',
          component: () => import('../views/kpi/Leaderboard.vue'),
          meta: { requiresAuth: true, title: '排行榜' },
        },
        {
          path: 'analysis',
          name: 'AnalysisDashboard',
          component: () => import('../views/analysis/AnalysisDashboard.vue'),
          meta: { requiresAuth: true, title: '运营分析' },
        },
        {
          path: 'reports',
          redirect: '/reports-hub',
          meta: { requiresAuth: true, title: '报表中心' },
        },
        {
          path: 'reports-hub',
          name: 'ReportsHub',
          component: () => import('../views/reports/ReportsHub.vue'),
          meta: { requiresAuth: true, title: '报表中心' },
        },
        {
          path: 'reports/profit',
          name: 'ProfitReport',
          component: () => import('../views/reports/ProfitReport.vue'),
          meta: { requiresAuth: true, title: '经营利润' },
        },
        {
          path: 'reports/daily-sales',
          name: 'DailySales',
          component: () => import('../views/reports/DailySales.vue'),
          meta: { requiresAuth: true, title: '日销售报表' },
        },
        {
          path: 'reports/source-roi',
          name: 'SourceROI',
          component: () => import('../views/reports/SourceROI.vue'),
          meta: { requiresAuth: true, title: '渠道ROI分析' },
        },
        {
          path: 'reports/staff-performance',
          name: 'StaffPerformance',
          component: () => import('../views/reports/StaffPerformance.vue'),
          meta: { requiresAuth: true, title: '效能看板' },
        },
      ],
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
