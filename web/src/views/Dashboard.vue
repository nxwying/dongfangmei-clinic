<template>
  <div style="max-width:1200px;margin:0 auto">

    <!-- Search bar -->
    <el-card shadow="never" style="margin-bottom:12px">
      <el-select v-model="searchQuery" filterable remote reserve-keyword placeholder="🔍 搜索客户姓名或手机号..." :remote-method="searchCustomer" :loading="searchLoading" style="width:100%" @change="goToCustomer">
        <el-option v-for="c in searchResults" :key="c.id" :label="c.name+' ('+c.phone+')'" :value="c.id"/>
      </el-select>
    </el-card>

    <!-- Row 1: KPI cards -->
    <el-row :gutter="12" style="margin-bottom:12px">

      <el-col :span="4">
        <el-card>
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:26px;font-weight:700;color:#67c23a">{{ newCustomers }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">今日新增客户</div>
            <div v-if="trend.customers !== 0" style="font-size:12px;margin-top:2px" :style="{color:trend.customers>=0?'#67c23a':'#f56c6c'}">
              {{ trend.customers >= 0 ? '↑' : '↓' }} {{ Math.abs(trend.customers) }} 比昨天
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card>
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:26px;font-weight:700;color:#e6a23c">{{ appts.length }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">今日预约（已到店 {{ checkedInCount }}）</div>
            <div style="font-size:12px;margin-top:2px;color:#909399">
              到店率 {{ appts.length > 0 ? Math.round(checkedInCount / appts.length * 100) : 0 }}%
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card>
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:26px;font-weight:700;color:#f56c6c">{{ dayStats.refunds }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">今日退款</div>
            <div style="font-size:12px;margin-top:2px;color:#f56c6c">¥{{ dayStats.refund_amount.toFixed(2) }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="4">
        <el-card>
          <div style="text-align:center;padding:10px 0">
            <div style="font-size:26px;font-weight:700;color:#b37feb">¥{{ (dayStats.recharge_total||0).toFixed(2) }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">今日充值</div>
            <div style="font-size:12px;margin-top:2px;color:#909399">{{ dayStats.recharge_count||0 }} 笔</div>
          </div>
        </el-card>
      </el-col>


    </el-row>

    <!-- Row 2: Alerts (资质 + 库存 + 沉默客户) -->
    <el-card v-if="hasAlerts" shadow="never" style="margin-bottom:12px;border-left:4px solid #e6a23c">
      <div v-if="expiringDocs.length > 0" style="display:flex;align-items:center;gap:12px;padding:4px 0">
        <span style="font-size:18px">⚠️</span>
        <span style="color:#e6a23c;font-weight:600">资质到期</span>
        <span style="color:#606266;font-size:13px"><b>{{ expiringDocs.length }}</b> 份证件<template v-if="expiredCount">（<b style="color:#f56c6c">{{ expiredCount }}</b> 份已过期）</template></span>
        <div style="flex:1" />
        <el-button size="small" @click="go('/documents?expiring_soon=true')">查看</el-button>
        <el-button size="small" type="primary" @click="go('/documents/inspection')">检查模式</el-button>
      </div>
      <div v-if="lowStockItems.length > 0" style="display:flex;align-items:center;gap:12px;padding:4px 0;margin-top:4px;border-top:1px solid #f0f0f0;padding-top:8px">
        <span style="font-size:18px">📦</span>
        <span style="color:#e6a23c;font-weight:600">库存不足</span>
        <span style="color:#606266;font-size:13px">{{ lowStockItems.slice(0,3).map((i:any)=>i.name+'（'+i.quantity+(i.unit||'')+'）').join('、') }}<template v-if="lowStockItems.length>3"> 等 {{ lowStockItems.length }} 种</template></span>
        <div style="flex:1" />
        <el-button size="small" @click="go('/inventory')">去入库</el-button>
      </div>
      <div v-if="silentCustomers.length > 0" style="display:flex;align-items:center;gap:12px;padding:4px 0;margin-top:4px;border-top:1px solid #f0f0f0;padding-top:8px">
        <span style="font-size:18px">⚠️</span>
        <span style="color:#f56c6c;font-weight:600">沉默客户</span>
        <span style="color:#606266;font-size:13px"><b>{{ silentCustomers.length }}</b> 位客户超过 60 天未到店</span>
        <div style="flex:1" />
        <el-button size="small" @click="go('/marketing')">去激活</el-button>
      </div>
    </el-card>

    <!-- Birthday alert -->
    <el-card v-if="birthdayCustomers.length > 0" shadow="never" style="margin-bottom:12px;border-left:4px solid #b37feb">
      <div style="display:flex;align-items:center;gap:12px;padding:6px 0">
        <span style="font-size:18px">🎂</span>
        <span style="color:#b37feb;font-weight:600">今日生日</span>
        <span style="color:#606266;font-size:13px">{{ birthdayCustomers.map((c:any)=>c.name).join("、") }} <b>{{ birthdayCustomers.length }}</b> 位客户</span>
        <div style="flex:1"/>
        <el-button size="small" @click="go('/marketing')">发送祝福</el-button>
      </div>
    </el-card>

    <!-- Row 3: 今日预约 + 快捷操作 + 今日提醒 -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="14">
        <el-card>
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between">
              <b>今日预约</b>
              <span style="font-size:13px;color:#909399">到店 {{ checkedInCount }} / {{ appts.length }}<el-button size="small" text style="margin-left:8px" @click="go('/appointments')">全部 →</el-button></span>
            </div>
          </template>
          <el-table :data="appts" size="small" v-loading="loadingAppts" stripe empty-text="今天暂无预约">
            <el-table-column label="时间" width="80">
              <template #default="{row}">{{ row.time_slot||'—' }}</template>
            </el-table-column>
            <el-table-column label="客户" min-width="85">
              <template #default="{row}">{{ row.customer?.name||'未知' }}</template>
            </el-table-column>
            <el-table-column label="项目" min-width="100">
              <template #default="{row}">{{ row.items||'—' }}</template>
            </el-table-column>
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{row}">
                <el-button v-if="row.status==='booked'" size="small" type="warning" @click="checkIn(row)">确认到店</el-button>
                <el-button v-else-if="row.status==='checked_in'" size="small" type="success" @click="complete(row)">完成</el-button>
                <el-tag v-else-if="row.status==='completed'" type="success" size="small">✓ 已完成</el-tag>
                <el-tag v-else-if="row.status==='cancelled'" type="danger" size="small">已取消</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="tomorrowAppts.length > 0" style="margin-top:8px;font-size:13px;color:#909399">
            明天预约 <b>{{ tomorrowAppts.length }}</b> 人：{{ tomorrowAppts.map((a:any)=>a.customer?.name||"未知").join("、") }}
            <el-button size="small" text @click="go('/appointments')">查看</el-button>
          </div>
        </el-card>
        <!-- Staff on duty -->
        <el-card>
          <template #header><b>今日值班</b></template>
          <div v-if="staffList.length > 0" style="display:flex;flex-wrap:wrap;gap:8px">
            <div v-for="s in staffList" :key="s.id"
              style="display:flex;align-items:center;gap:8px;padding:8px 10px;border-radius:6px;min-width:115px"
              :style="{background:staffBg(s.role?.name)}">
              <div style="width:32px;height:32px;border-radius:50%;display:flex;align-items:center;justify-content:center;color:#fff;font-weight:600;font-size:14px;flex-shrink:0"
                :style="{background:staffColor(s.role?.name)}">{{ staffInitial(s.real_name) }}</div>
              <div>
                <div style="font-size:13px;font-weight:600;color:#303133">{{ s.real_name }}</div>
                <div style="font-size:11px;color:#909399">{{ roleName(s.role?.name) }}</div>
              </div>
            </div>
          </div>
          <div v-else style="text-align:center;padding:12px;color:#c0c4cc;font-size:13px">暂无人员</div>
        </el-card>
        <!-- 今日概况 -->
        <el-card>
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between">
              <b>今日概况</b>
              <el-button size="small" text @click="go('/data')">数据中心 →</el-button>
            </div>
          </template>
          <!-- Metrics grid -->
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:6px;margin-bottom:8px">
            <div style="background:#f5f7fa;border-radius:6px;padding:10px 12px;text-align:center">
              <div style="font-size:20px;font-weight:700;color:#409eff">¥{{ avgTicket.toFixed(0) }}</div>
              <div style="font-size:11px;color:#909399;margin-top:2px">客单价</div>
            </div>
            <div style="background:#f5f7fa;border-radius:6px;padding:10px 12px;text-align:center">
              <div style="font-size:20px;font-weight:700;color:#67c23a">{{ todayOrders.length }}</div>
              <div style="font-size:11px;color:#909399;margin-top:2px">成交单数</div>
            </div>
            <div style="background:#f5f7fa;border-radius:6px;padding:10px 12px;text-align:center">
              <div style="font-size:20px;font-weight:700;color:#e6a23c">¥{{ today.gross.toFixed(0) }}</div>
              <div style="font-size:11px;color:#909399;margin-top:2px">毛收入</div>
            </div>
            <div style="background:#f5f7fa;border-radius:6px;padding:10px 12px;text-align:center">
              <div style="font-size:20px;font-weight:700;color:#b37feb">{{ checkedInCount }}</div>
              <div style="font-size:11px;color:#909399;margin-top:2px">到店人数</div>
            </div>
          </div>
          <!-- 今日新增客户 -->
          <div style="border-top:1px solid #f0f0f0;padding:8px 0;margin-top:4px">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px">
              <span style="font-size:13px;font-weight:600;color:#303133">今日新增客户</span>
              <span style="font-size:12px;color:#909399">{{ newCustomers }} 人</span>
            </div>
            <div v-if="todayCustomers.length > 0">
              <div v-for="(c,i) in todayCustomers.slice(0,4)" :key="i" style="display:flex;align-items:center;justify-content:space-between;padding:3px 0;font-size:12px;cursor:pointer" @click="go('/customers/'+c.id)">
                <span style="color:#303133">{{ c.name }}</span>
                <span style="color:#909399">{{ c.phone||'—' }}</span>
              </div>
              <div v-if="todayCustomers.length > 4" style="font-size:11px;color:#909399;text-align:center;padding:4px 0">还有 {{ todayCustomers.length-4 }} 位...</div>
            </div>
            <div v-else style="text-align:center;padding:8px;color:#c0c4cc;font-size:12px">今日暂无新增客户</div>
          </div>
          <!-- 待处理订单 -->
          <div style="border-top:1px solid #f0f0f0;padding:8px 0">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px">
              <span style="font-size:13px;font-weight:600;color:#303133">待处理订单</span>
              <span style="font-size:12px;color:#e6a23c">{{ pendingOrders.length }} 单</span>
            </div>
            <div v-if="pendingOrders.length > 0">
              <div v-for="(o,i) in pendingOrders.slice(0,4)" :key="i" style="display:flex;align-items:center;justify-content:space-between;padding:3px 0;font-size:12px;cursor:pointer" @click="go('/pos/'+o.id)">
                <span style="color:#303133">{{ o.customer?.name||'未知' }}</span>
                <span style="color:#f56c6c;font-weight:600">¥{{ (o.final_amount||0).toFixed(0) }}</span>
              </div>
              <el-button v-if="pendingOrders.length > 4" size="small" text style="width:100%;margin-top:4px" @click="go('/pos')">查看全部 →</el-button>
            </div>
            <div v-else-if="pendingOrders.length === 0" style="text-align:center;padding:8px;color:#c0c4cc;font-size:12px">暂无待处理订单</div>
          </div>
          <!-- 今日到店客户 -->
          <div style="border-top:1px solid #f0f0f0;padding:8px 0">
            <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px">
              <span style="font-size:13px;font-weight:600;color:#303133">今日到店客户</span>
              <span style="font-size:12px;color:#909399">{{ checkedInCount }} / {{ appts.length }} 人</span>
            </div>
            <div v-if="checkedInCustomers.length > 0">
              <div v-for="(a,i) in checkedInCustomers.slice(0,5)" :key="i" style="display:flex;align-items:center;justify-content:space-between;padding:3px 0;font-size:12px;cursor:pointer" @click="go('/customers/'+a.customer?.id)">
                <span style="color:#303133">{{ a.customer?.name||'未知' }}</span>
                <span style="color:#909399">{{ a.time_slot||'—' }} {{ a.items||'' }}</span>
              </div>
            </div>
            <div v-else style="text-align:center;padding:8px;color:#c0c4cc;font-size:12px">暂无到店客户</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="10">
        <!-- Quick actions -->
        <el-card style="margin-bottom:12px">
          <template #header><b>快捷操作</b></template>
          <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px">
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#ecf5ff" @click="go('/customers')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#409eff;color:#fff;font-size:20px;font-weight:700;flex-shrink:0">+</div>
              <span style="font-size:14px;font-weight:600;color:#303133">新建客户</span>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#fdf6ec" @click="go('/appointments')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#e6a23c;color:#fff;font-size:18px;flex-shrink:0">📋</div>
              <span style="font-size:14px;font-weight:600;color:#303133">新建预约</span>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#fef0f0" @click="go('/pos')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#f56c6c;color:#fff;font-size:18px;flex-shrink:0">💰</div>
              <span style="font-size:14px;font-weight:600;color:#303133">去收银</span>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#f0f9eb" @click="go('/finance/expenses')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#67c23a;color:#fff;font-size:18px;flex-shrink:0">📝</div>
              <span style="font-size:14px;font-weight:600;color:#303133">记支出</span>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#f0f0f0" @click="go('/inventory')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#909399;color:#fff;font-size:18px;flex-shrink:0">📦</div>
              <span style="font-size:14px;font-weight:600;color:#303133">入库</span>
            </div>
            <div style="display:flex;align-items:center;gap:8px;padding:14px 12px;border-radius:8px;cursor:pointer;background:#fdf6ec" @click="go('/documents')">
              <div style="width:36px;height:36px;border-radius:8px;display:flex;align-items:center;justify-content:center;background:#e6a23c;color:#fff;font-size:18px;flex-shrink:0">📄</div>
              <span style="font-size:14px;font-weight:600;color:#303133">上传证件</span>
            </div>
          </div>
        </el-card>
        <!-- Today's reminders -->
        <el-card>
          <template #header><b>今日提醒</b></template>
          <!-- Follow-up tasks table -->
          <div v-if="todayFollowups.length > 0" style="margin-bottom:8px">
            <div style="font-weight:600;font-size:13px;margin-bottom:6px">今日需回访</div>
            <el-table :data="todayFollowups.slice(0,5)" size="small" stripe>
              <el-table-column label="客户" min-width="65"><template #default="{row:r}">{{ r.customer?.name||"--" }}</template></el-table-column>
              <el-table-column label="原因" min-width="55"><template #default="{row:r}">{{ r.task_type||"回访" }}</template></el-table-column>
              <el-table-column label="操作" width="45"><template #default="{row:r}"><el-button size="small" text type="primary" @click="go('/followup')">去</el-button></template></el-table-column>
            </el-table>
          </div>
          <div v-if="silentCustomers.length > 0" style="display:flex;align-items:center;gap:8px;padding:8px 0;border-bottom:1px solid #f0f0f0">
            <span style="font-size:16px">⚠️</span>
            <span style="flex:1;font-size:13px">沉默客户 <b>{{ silentCustomers.length }}</b> 人</span>
            <el-button size="small" text @click="go('/marketing')">激活</el-button>
          </div>
          <div v-if="pendingOrders.length > 0" style="display:flex;align-items:center;gap:8px;padding:8px 0;border-bottom:1px solid #f0f0f0">
            <span style="font-size:16px">⏳</span>
            <span style="flex:1;font-size:13px">待付款 <b>{{ pendingOrders.length }}</b> 单</span>
            <el-button size="small" text @click="go('/pos')">去收款</el-button>
          </div>
          <!-- KPI progress -->
          <div v-if="kpiTargets.length > 0" style="margin-top:8px">
            <div style="font-weight:600;font-size:13px;margin-bottom:6px">本月KPI进度</div>
            <div v-for="k in kpiTargets" :key="k.id" style="margin-bottom:6px">
              <div style="display:flex;justify-content:space-between;font-size:12px;color:#909399">
                <span>{{ k.indicator }}</span>
                <span>{{ k.progress||0 }}/{{ k.target||0 }}</span>
              </div>
              <div style="width:100%;height:6px;background:#f0f2f5;border-radius:3px;overflow:hidden">
                <div :style="{width:k.target?Math.min((k.progress||0)/k.target*100,100)+'%':'0%',height:'100%',background:'#409eff',borderRadius:'3px'}"></div>
              </div>
            </div>
          </div>
          <div v-if="followupTotal > 0" style="margin-top:6px;padding-top:6px;border-top:1px solid #f0f0f0">
            <div style="display:flex;justify-content:space-between;font-size:12px;color:#909399;margin-bottom:2px">
              <span>回访完成率</span>
              <span>{{ followupDone }}/{{ followupTotal }} ({{ followupRate }}%)</span>
            </div>
            <div style="width:100%;height:6px;background:#f0f2f5;border-radius:3px;overflow:hidden">
              <div :style="{width:followupRate+'%',height:'100%',background:followupRate>=50?'#409eff':'#e6a23c',borderRadius:'3px'}"></div>
            </div>
          </div>
          <div v-if="todayFollowups.length === 0 && silentCustomers.length === 0 && pendingOrders.length === 0 && kpiTargets.length === 0 && followupTotal === 0" style="text-align:center;padding:16px;color:#c0c4cc;font-size:13px">
            暂无待办事项
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Today's orders -->
    <el-card style="margin-bottom:12px">
      <template #header><b>今日成交订单</b></template>
      <el-table :data="todayOrders.slice(0,8)" size="small" stripe empty-text="暂无今日成交">
        <el-table-column label="客户" min-width="70"><template #default="{row:r}">{{ r.customer?.name||"--" }}</template></el-table-column>
        <el-table-column label="项目" min-width="100"><template #default="{row:r}"><span v-if="r.items?.length">{{ r.items.map((i:any)=>i.item_name).join("、") }}</span><span v-else>--</span></template></el-table-column>
        <el-table-column label="金额" width="80" align="right"><template #default="{row:r}">¥{{ (r.final_amount||0).toFixed(2) }}</template></el-table-column>
        <el-table-column label="时间" width="80"><template #default="{row:r}">{{ r.created_at?.slice(11,16)||"--" }}</template></el-table-column>
        <el-table-column label="操作" width="55"><template #default="{row:r}"><el-button size="small" text type="primary" @click="go('/pos/'+r.id)">详情</el-button></template></el-table-column>
      </el-table>
    </el-card>

    <!-- Row 4: 本周收入趋势 + 项目销售排行（含客单价） -->
    <el-row :gutter="12" style="margin-bottom:12px">
      <el-col :span="14">
        <el-card>
          <template #header><b>本周毛收入趋势</b></template>
          <div v-if="week.length" style="display:flex;align-items:flex-end;gap:6px;height:160px;padding:8px 0">
            <div v-for="d in week" :key="d.date" style="flex:1;display:flex;flex-direction:column;align-items:center;height:100%">
              <div style="flex:1;width:100%;display:flex;flex-direction:column-reverse;align-items:center">
                <div :style="{width:'55%',minHeight:'2px',borderRadius:'3px 3px 0 0',transition:'height .3s',background:d.gross>0?'#e6a23c':'#e4e7ed',height:barH(d.gross)+'%'}"></div>
              </div>
              <div style="font-size:11px;color:#909399;margin-top:4px">{{ d.date.slice(5) }}</div>
              <div style="font-size:11px;color:#303133;font-weight:600;margin-top:1px">¥{{ (d.gross||0).toFixed(0) }}</div>
            </div>
          </div>
          <div v-else style="text-align:center;padding:20px;color:#909399">暂无本周数据</div>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card>
          <template #header><b>本月项目销售排行</b></template>
          <div v-if="staffRanking.length > 0" style="margin-bottom:12px">
            <div style="font-weight:600;font-size:13px;margin-bottom:6px">🏆 员工业绩排行</div>
            <div v-for="(s,i) in staffRanking.slice(0,5)" :key="i" style="display:flex;align-items:center;gap:6px;margin-bottom:4px;font-size:12px">
              <span style="width:14px;font-weight:700;color:#909399;text-align:center">{{ i+1 }}</span>
              <span style="width:50px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ s.real_name }}</span>
              <div style="flex:1;height:12px;background:#f0f2f5;border-radius:3px;overflow:hidden">
                <div :style="{width:staffBarW(s.gross)+'%',height:'100%',background:rankColor(i),borderRadius:'3px'}"></div>
              </div>
              <span style="width:60px;text-align:right;font-weight:600">¥{{ (s.gross||0).toFixed(0) }}</span>
            </div>
          </div>
          <div v-if="salesRanking.length > 0">
            <div v-for="(item, i) in salesRanking" :key="i" style="margin-bottom:14px">
              <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:2px">
                <div>
                  <span style="font-weight:600;font-size:14px;color:#303133">{{ i + 1 }}. {{ item.name }}</span>
                </div>
                <div>
                  <span style="color:#f56c6c;font-weight:700;font-size:14px">¥{{ item.revenue.toFixed(0) }}</span>
                  <span style="color:#909399;font-size:12px;margin-left:4px">{{ item.percent }}%</span>
                </div>
              </div>
              <div style="width:100%;height:8px;background:#f0f2f5;border-radius:4px;overflow:hidden">
                <div :style="{width:item.percent+'%',height:'100%',background:rankColor(i),borderRadius:'4px',transition:'width .5s'}"></div>
              </div>
              <div v-if="item.avgPrice" style="font-size:11px;color:#909399;margin-top:2px">客单价 ¥{{ item.avgPrice.toFixed(0) }}</div>
            </div>
          </div>
          <div v-else style="text-align:center;padding:20px;color:#c0c4cc;font-size:13px">暂无本月销售数据</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Row 5: 本周客户趋势 -->
    <el-card>
      <template #header><b>本周客户趋势</b></template>
      <el-row :gutter="12">
        <el-col :span="8">
          <div style="text-align:center;padding:12px 0">
            <div style="font-size:28px;font-weight:700;color:#409eff">{{ weekNewCustomers }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">本周新增客户</div>
            <div style="margin-top:8px;width:100%;height:8px;background:#f0f2f5;border-radius:4px;overflow:hidden">
              <div :style="{width:customerBarPct(weekNewCustomers)+'%',height:'100%',background:'#409eff',borderRadius:'4px'}"></div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div style="text-align:center;padding:12px 0">
            <div style="font-size:28px;font-weight:700;color:#67c23a">{{ weekVisits }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">本周到店</div>
            <div style="margin-top:8px;width:100%;height:8px;background:#f0f2f5;border-radius:4px;overflow:hidden">
              <div :style="{width:customerBarPct(weekVisits)+'%',height:'100%',background:'#67c23a',borderRadius:'4px'}"></div>
            </div>
          </div>
        </el-col>
        <el-col :span="8">
          <div style="text-align:center;padding:12px 0">
            <div style="font-size:28px;font-weight:700;color:#e6a23c">{{ weekFollowups }}</div>
            <div style="font-size:13px;color:#909399;margin-top:4px">本周回访</div>
            <div style="margin-top:8px;width:100%;height:8px;background:#f0f2f5;border-radius:4px;overflow:hidden">
              <div :style="{width:customerBarPct(weekFollowups)+'%',height:'100%',background:'#e6a23c',borderRadius:'4px'}"></div>
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../api'
import { getExpiringDocuments } from '../api/documents'

const router = useRouter()

// Data
const today = reactive({ revenue:0, commissions:0, gross:0, highValue:0, general:0, profit:0 })
const yesterday = reactive({ revenue:0 })
const trend = reactive({ revenue:0, customers:0 })
const monthProfit = ref(0)
const monthRevenue = ref(0)
const monthExpense = ref(0)
const followupDone = ref(0)
const followupTotal = ref(0)
const dayStats=reactive({refunds:0,refund_amount:0,recharge_total:0,recharge_count:0})
const searchQuery=ref('');const searchResults=ref<any[]>([]);const searchLoading=ref(false)
const todayOrders=ref<any[]>([])
const todayCustomers=ref<any[]>([])
const birthdayCustomers=ref<any[]>([])
const tomorrowAppts=ref<any[]>([])
const staffRanking=ref<any[]>([])
const pendingOrders=ref<any[]>([])
const kpiTargets=ref<any[]>([])
const todayFollowups=ref<any[]>([])
const staffList=ref<any[]>([])
const appts = ref<any[]>([])
const loadingAppts = ref(false)
const week = ref<any[]>([])
const expiringDocs = ref<any[]>([])
const newCustomers = ref(0)
const yesterdayCustomers = ref(0)
const followupTasks = ref<any[]>([])
const lowStockItems = ref<any[]>([])
const silentCustomers = ref<any[]>([])
const salesRanking = ref<any[]>([])
const weekNewCustomers = ref(0)
const weekVisits = ref(0)
const weekFollowups = ref(0)

// Computed
const expiredCount = computed(() => expiringDocs.value.filter((d: any) => d.status === 'expired').length)
const checkedInCount = computed(() => appts.value.filter((a: any) => a.status === 'checked_in' || a.status === 'completed').length)
const visitRate = computed(() => appts.value.length > 0 ? Math.round(checkedInCount.value / appts.value.length * 100) : 0)
const followupRate = computed(() => followupTotal.value > 0 ? Math.round(followupDone.value / followupTotal.value * 100) : 0)
const avgTicket = computed(() => todayOrders.value.length > 0 ? today.revenue / todayOrders.value.length : 0)
const checkedInCustomers = computed(() => appts.value.filter((a:any) => a.status==="checked_in"||a.status==="completed"))
const hasAlerts = computed(() => expiringDocs.value.length > 0 || lowStockItems.value.length > 0 || silentCustomers.value.length > 0)

const d = new Date()
const dateStr = `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,"0")}-${String(d.getDate()).padStart(2,"0")}`
const yesterdayStr = `${new Date(d.getTime()-86400000).getFullYear()}-${String(new Date(d.getTime()-86400000).getMonth()+1).padStart(2,"0")}-${String(new Date(d.getTime()-86400000).getDate()).padStart(2,"0")}`
const monthStart = dateStr.slice(0,7)+'-01'

function daysAgo(date:Date,n:number):string{
  const m = new Date(date)
  m.setDate(m.getDate() - n)
  return `${m.getFullYear()}-${String(m.getMonth()+1).padStart(2,"0")}-${String(m.getDate()).padStart(2,"0")}`
}
const weekStart = daysAgo(new Date(), 6)
const weekEnd = dateStr

const maxRev = computed(()=>Math.max(...week.value.map((w:any)=>w.gross||0),1))
function barH(v:number){return Math.max((v/maxRev.value)*100,0)}
function rankColor(i:number){const c=['#f56c6c','#e6a23c','#409eff','#67c23a','#909399'];return c[i]||'#909399'}
function customerBarPct(v:number){const m=Math.max(weekNewCustomers.value,weekVisits.value,weekFollowups.value,1);return v/m*100}

function go(path:string){router.push(path)}
async function searchCustomer(q:string){if(!q){searchResults.value=[];return}
  searchLoading.value=true
  try{const r=await api.get('/customers',{params:{q,size:20}});searchResults.value=r.data?.data||r.data||[]}catch{}
  finally{searchLoading.value=false}
}
function goToCustomer(id:number){router.push('/customers/'+id);searchQuery.value=''}
function staffBarW(v:number){const m=Math.max(...staffRanking.value.map((x:any)=>x.gross||0),1);return v/m*100}
function staffColor(role:string){const m:{[k:string]:string}={'doctor':'#409eff','nurse':'#67c23a','consultant':'#e6a23c','admin':'#909399'};return m[role]||'#b37feb'}
function staffBg(role:string){const m:{[k:string]:string}={'doctor':'#ecf5ff','nurse':'#f0f9eb','consultant':'#fdf6ec','admin':'#f4f4f5'};return m[role]||'#f9f0ff'}
function staffInitial(name:string){return name?name.charAt(0):'?'}

function roleName(role:string){const m:{[k:string]:string}={'doctor':'医生','nurse':'护士','consultant':'咨询师','admin':'管理员','receptionist':'前台','manager':'主管'};return m[role]||role||'其他'}
async function checkIn(row:any){
  try{
    await api.put('/appointments/'+row.id+'/checkin')
    ElMessage.success('已确认到店')
    row.status='checked_in'
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
}
async function complete(row:any){
  try{
    await api.put('/appointments/'+row.id+'/complete')
    ElMessage.success('已完成')
    row.status='completed'
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'操作失败')}
}

onMounted(async ()=>{
  // Today's revenue/profit
  try{
    const p=await api.get('/reports/profit',{params:{date:dateStr}})
    if(p.data){
      today.revenue=p.data.total_revenue||0
      today.commissions=p.data.commissions||0
      today.gross=p.data.gross_revenue||0
      today.highValue=p.data.high_value||0
      today.general=p.data.general||0
      today.profit=p.data.net_profit||0
    }
  }catch(e){console.log(e)}

  // Yesterday for trend
  try{
    const y=await api.get('/reports/profit',{params:{date:yesterdayStr}})
    if(y.data)yesterday.revenue=y.data.total_revenue||0
  }catch(e){console.log(e)}

  // Appointments
  try{
    loadingAppts.value=true
    const a=await api.get('/appointments',{params:{date:dateStr}})
    appts.value=Array.isArray(a.data)?a.data:[]
  }catch(e){console.log(e)}
  finally{loadingAppts.value=false}

  // Week trend
  try{
    const w=await api.get('/reports/period',{params:{start_date:weekStart,end_date:weekEnd}})
    week.value=w.data?.daily_breakdown||[]
  }catch(e){console.log(e)}

  // Expiring docs
  try{expiringDocs.value=await getExpiringDocuments()}catch(e){console.log(e)}

  // New customers today
  try{
    const c=await api.get('/customers',{params:{start_date:dateStr,end_date:dateStr}})
const custData=c.data;const custList=Array.isArray(custData)?custData:(custData?.data||[]);newCustomers.value=custList.length;todayCustomers.value=custList
  }catch(e){console.log(e)}

  // Yesterday customers for trend
  try{
    const cy=await api.get('/customers',{params:{start_date:yesterdayStr,end_date:yesterdayStr}})
    const dy=cy.data;yesterdayCustomers.value=Array.isArray(dy)?dy.length:(dy?.data?.length||dy?.total||0)
  }catch(e){console.log(e)}


  // Daily stats (refunds, recharge)
  try{const ds=await api.get('/reports/daily-sales');if(ds.data){dayStats.refunds=ds.data.refund_count||0;dayStats.refund_amount=ds.data.refund_amount||0;dayStats.recharge_total=ds.data.recharge_total||0}}catch{}

  // Today's orders
  try{const o=await api.get('/orders',{params:{date:dateStr}});const od=o.data;todayOrders.value=Array.isArray(od)?od:od?.data||od?.list||[]}catch{}


  // Staff list
  try{const u=await api.get('/users');const us=u.data;staffList.value=Array.isArray(us)?us.filter((x:any)=>x.status==='active'):[]}catch{}

  // Birthday customers
  try{const b=await api.get('/marketing/birthday');birthdayCustomers.value=Array.isArray(b.data)?b.data:(b.data?.items||[])}catch{}

  // Tomorrow's appointments
  try{const tomorrow=new Date(Date.now()+86400000);const ts=tomorrow.getFullYear()+'-'+String(tomorrow.getMonth()+1).padStart(2,'0')+'-'+String(tomorrow.getDate()).padStart(2,'0');const a=await api.get('/appointments',{params:{date:ts}});tomorrowAppts.value=Array.isArray(a.data)?a.data:[]}catch{}

  // Staff ranking (monthly)
  try{const sr=await api.get('/kpi/leaderboard',{params:{year_month:dateStr.slice(0,7)}});staffRanking.value=Array.isArray(sr.data)?sr.data:[]}catch{}

  // Pending orders
  try{const po=await api.get('/orders',{params:{status:'pending',size:10}});const pod=po.data;pendingOrders.value=Array.isArray(pod)?pod:pod?.data||pod?.list||[]}catch{}

  // Today's follow-up tasks
  try{const ft=await api.get('/followup/tasks',{params:{status:'pending',size:10}});todayFollowups.value=Array.isArray(ft.data)?ft.data:(ft.data?.items||[])}catch{}

  // KPI targets
  try{const kt=await api.get('/kpi/targets',{params:{year_month:dateStr.slice(0,7)}});const kd=kt.data;kpiTargets.value=Array.isArray(kd)?kd:[]}catch{}

  // Recharge count from orders
  try{const rc=await api.get('/orders',{params:{date:dateStr,size:200}});const rcd=rc.data;const rOrders=Array.isArray(rcd)?rcd:rcd?.data||rcd?.list||[];let rcCount=0;rOrders.forEach((o:any)=>{if(o.payments){o.payments.forEach((p:any)=>{if(p.pay_method==='balance'||p.pay_method==='gift_balance'){rcCount+=p.amount>0?1:0}})}});dayStats.recharge_count=rcCount}catch{}
  // Compute trends
  if(yesterday.revenue>0)trend.revenue=Math.round((today.revenue-yesterday.revenue)/yesterday.revenue*100)
  if(yesterdayCustomers.value>0)trend.customers=newCustomers.value-yesterdayCustomers.value

  // Followup tasks
  try{
    const f=await api.get('/followup/tasks',{params:{status:'pending',size:5}})
    followupTasks.value=Array.isArray(f.data)?f.data:(f.data?.items||[])
  }catch(e){console.log(e)}

  // Low stock
  try{
    const inv=await api.get('/inventory/items')
    const items=Array.isArray(inv.data)?inv.data:(inv.data?.items||[])
    lowStockItems.value=items.filter((i:any)=>i.quantity<=i.min_stock)
  }catch(e){console.log(e)}

  // Silent customers
  try{
    const s=await api.get('/marketing/dormant')
    silentCustomers.value=Array.isArray(s.data)?s.data:(s.data?.items||[])
  }catch(e){console.log(e)}

  // Sales ranking - calculate from orders
  async function calcSalesRanking(){
    try{
      // Try ProcedureProfit first
      const pp=await api.get('/analysis/procedure-profit')
      if(pp.data&&pp.data.length){
        const total=pp.data.reduce((s:number,x:any)=>s+(x.gross||0),0)
        salesRanking.value=pp.data.map((x:any)=>({name:x.procedure,revenue:x.gross||0,percent:total?Math.round((x.gross||0)/total*100):0,avgPrice:x.margin||0})).slice(0,5)
        return
      }
    }catch{}
    try{
      const o=await api.get('/orders',{params:{page_size:200}})
      const orders=Array.isArray(o.data)?o.data:(o.data?.data||o.data?.list||[])
      const map:any={};let total=0
      orders.forEach((order:any)=>{
        const comm=order.commission_amount||0;
        const fa=order.final_amount||0;
        if(order.items)order.items.forEach((i:any)=>{
          const n=i.item_name;
          const sub=i.unit_price*i.quantity;
          const grossItem=fa>0?sub-(comm*sub/fa):sub;
          if(!map[n])map[n]={name:n,gross:0,count:0};
          map[n].gross+=grossItem;
          map[n].count+=i.quantity;
          total+=grossItem
        })
      })
      salesRanking.value=Object.values(map).sort((a:any,b:any)=>b.gross-a.gross).slice(0,5).map((i:any)=>({name:i.name,revenue:i.gross,percent:total?Math.round(i.gross/total*100):0,avgPrice:i.count?i.gross/i.count:0}))
    }catch{salesRanking.value=[]}
  }
  calcSalesRanking()

  // Monthly revenue
  try{
    const md=await api.get('/reports/period',{params:{start_date:monthStart,end_date:dateStr}})
    if(md.data){
      monthRevenue.value=md.data.gross_revenue||0
    }
  }catch{}
  // Monthly expense (from expenses table only, to match 支出管理)
  try{
    const ex=await api.get('/expenses',{params:{start_date:monthStart,end_date:dateStr}})
    const exd=ex.data;const exList=Array.isArray(exd)?exd:exd?.data||exd?.items||[]
    monthExpense.value=exList.reduce((s:number,x:any)=>s+(x.amount||0),0)
  }catch{}

  // Followup completion rate
  try{
    const fDone=await api.get('/followup/tasks',{params:{status:'completed',start_date:monthStart,end_date:dateStr,size:200}})
    const fd=fDone.data;followupDone.value=Array.isArray(fd)?fd.length:(fd?.items?.length||fd?.total||0)
    const fTotal=await api.get('/followup/tasks',{params:{start_date:monthStart,end_date:dateStr,size:200}})
    const ft=fTotal.data;followupTotal.value=Array.isArray(ft)?ft.length:(ft?.items?.length||ft?.total||0)
  }catch{}

  // Weekly customer trend
  try{
    // New customers this week
    const cw=await api.get('/customers',{params:{start_date:weekStart,end_date:weekEnd}})
    const cwData=cw.data;weekNewCustomers.value=Array.isArray(cwData)?cwData.length:(cwData?.data?.length||cwData?.total||0)

    // Weekly visits (appointments with status checked_in/completed)
    const aw=await api.get('/appointments',{params:{start_date:weekStart,end_date:weekEnd}})
    const awData=Array.isArray(aw.data)?aw.data:(aw.data?.items||[])
    weekVisits.value=awData.filter((a:any)=>a.status==='checked_in'||a.status==='completed').length

    // Weekly followups
    const fw=await api.get('/followup/tasks',{params:{start_date:weekStart,end_date:weekEnd}})
    weekFollowups.value=Array.isArray(fw.data)?fw.data.length:(fw.data?.total||0)
  }catch(e){console.log(e)}
})
</script>

<style scoped>
.el-card { border-radius: 8px; transition: box-shadow .2s; }
.el-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,.08); }
.el-table { border-radius: 6px; overflow: hidden; }
.el-row { margin-bottom: 16px; }
</style>
