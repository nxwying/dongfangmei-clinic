<template>
  <div style="max-width:1200px;margin:0 auto">
    <el-card shadow="never">
      <el-tabs v-model="tab">

        <!-- Tab 1: KPI目标 -->
        <el-tab-pane label="🎯 KPI目标" name="kpi">
          <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px">
            <el-date-picker v-model="ym" type="month" value-format="YYYY-MM" style="width:160px" @change="loadKpi"/>
          </div>
          <el-table :data="kpiData" v-loading="kpiLoading" stripe border>
            <el-table-column label="姓名" min-width="80"><template #default="{row:r}">{{ userName(r.user_id) }}</template></el-table-column>
            <el-table-column label="业绩目标"><template #default="{row:r}"><el-input-number v-model="r.revenue_target" size="small" style="width:120px"/></template></el-table-column>
            <el-table-column label="订单目标"><template #default="{row:r}"><el-input-number v-model="r.order_target" size="small" style="width:80px"/></template></el-table-column>
            <el-table-column label="回访目标"><template #default="{row:r}"><el-input-number v-model="r.followup_target" size="small" style="width:80px"/></template></el-table-column>
            <el-table-column label="操作" width="80"><template #default="{row:r}"><el-button size="small" @click="saveKpi(r)">保存</el-button></template></el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 2: 排行榜 -->
        <el-tab-pane label="🏆 排行榜" name="rank">
          <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;flex-wrap:wrap">
            <el-date-picker v-model="rankStart" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" style="width:160px" @change="loadRank"/>
            <el-date-picker v-model="rankEnd" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" style="width:160px" @change="loadRank"/>
          </div>
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
            <span style="font-weight:600">业绩排行</span>
            <span style="flex:1"/>
            <el-button size="small" text @click="exportTable(rankData,'排行榜',['#','姓名','总业绩','订单数','回访','新增客户'],[(_,i)=>i+1,d=>d.real_name,(d.revenue||0).toFixed(2),d=>d.orders,d=>d.followups,d=>d.new_customers])">📤</el-button>
          </div>
          <el-table :data="rankData" v-loading="rankLoading" stripe border style="margin-bottom:16px">
            <el-table-column type="index" label="#" width="50" align="center"/>
            <el-table-column label="姓名" min-width="80"><template #default="{row:r}">{{ userName(r.user_id) }}</template></el-table-column>
            <el-table-column label="总业绩" width="120" align="right"><template #default="{row:r}">¥{{ (r.revenue||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="订单数" prop="orders" width="70" align="center"/>
            <el-table-column label="回访完成" prop="followups" width="80" align="center"/>
            <el-table-column label="新增客户" prop="new_customers" width="80" align="center"/>
          </el-table>

          <!-- 回访完成排行 -->
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
            <span style="font-weight:600">📊 回访完成排行</span>
            <span style="color:#909399;font-size:13px">共 {{ followupTotal }} 条 / 完成率 {{ followupTotal>0?Math.round(followupDone/followupTotal*100):0 }}%</span>
          </div>
          <el-table :data="fuRanking" stripe border size="small">
            <el-table-column type="index" label="#" width="40" align="center"/>
            <el-table-column label="姓名" min-width="70"><template #default="{row:r}">{{ userName(r.user_id) }}</template></el-table-column>
            <el-table-column label="总任务" prop="total" width="60" align="center"/>
            <el-table-column label="已完成" prop="completed" width="60" align="center"/>
            <el-table-column label="完成率" width="60" align="center"><template #default="{row:r}"><el-tag :type="r.rate>=80?'success':r.rate>=50?'warning':'danger'" size="small">{{ r.rate }}%</el-tag></template></el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 3: 提成规则 -->
        <el-tab-pane label="💰 提成规则" name="commission">
          <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;flex-wrap:wrap">
            <el-date-picker v-model="commStart" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" style="width:160px" @change="loadCommResults"/>
            <el-date-picker v-model="commEnd" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" style="width:160px" @change="loadCommResults"/>
            <el-switch v-model="autoCalcComm" @change="saveAutoCalc" style="margin-left:4px"/>
            <span style="font-size:13px;color:#909399">{{ autoCalcComm?'已开启自动计算（每月1号自动触发）':'启用自动计算' }}</span>
          </div>
          <div style="margin-bottom:12px;display:flex;gap:8px">
            <el-button type="primary" size="small" @click="commDlg=true">+ 新建规则</el-button>
            <el-button size="small" @click="calcComm">计算提成</el-button>
          </div>
          <el-table :data="commRules" v-loading="commLoading" stripe border style="margin-bottom:16px">
            <el-table-column label="规则名称" prop="name" min-width="120"/>
            <el-table-column label="角色" width="80"><template #default="{row:r}">{{ {consultant:'咨询师',doctor:'医生',nurse:'护士'}[r.role]||r.role }}</template></el-table-column>
            <el-table-column label="比例" width="80"><template #default="{row:r}">{{ r.rate }}%</template></el-table-column>
            <el-table-column label="项目" prop="procedure" min-width="100"/>
            <el-table-column label="状态" width="70"><template #default="{row:r}"><el-switch v-model="r.is_active" @change="toggleComm(r)"/></template></el-table-column>
          </el-table>

          <div v-if="commResults.length">
            <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px">
              <span style="font-weight:600">📋 提成明细（{{ commStart }} ~ {{ commEnd }}）</span>
              <el-button size="small" text @click="exportTable(commResults,'提成明细',['姓名','业绩','提成'],[d=>d.real_name,(d.total_revenue||0).toFixed(2),(d.total_commission||0).toFixed(2)])">📤</el-button>
            </div>
            <el-table :data="commResults" stripe border size="small">
              <el-table-column label="姓名" prop="real_name" min-width="80"/>
              <el-table-column label="业绩" width="120" align="right"><template #default="{row:r}">¥{{ (r.total_revenue||0).toFixed(2) }}</template></el-table-column>
              <el-table-column label="提成" width="120" align="right"><template #default="{row:r}">¥{{ (r.total_commission||0).toFixed(2) }}</template></el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- Tab 4: 收费明细 -->
        <el-tab-pane label="📋 收费明细" name="orders">
          <div style="display:flex;align-items:center;gap:12px;margin-bottom:12px;flex-wrap:wrap">
            <el-date-picker v-model="orderStart" type="date" value-format="YYYY-MM-DD" placeholder="开始日期" style="width:160px" @change="loadOrders"/>
            <el-date-picker v-model="orderEnd" type="date" value-format="YYYY-MM-DD" placeholder="结束日期" style="width:160px" @change="loadOrders"/>
            <span style="color:#909399;font-size:13px">共 {{ orders.length }} 条，合计 ¥{{ orderTotal.toFixed(2) }}</span>
            <el-button size="small" text @click="exportTable(orders,'收费明细',['日期','单号','客户','电话','项目','金额','支付方式','状态'],[d=>fmtDate(d.created_at),d=>d.order_no,d=>d.customer?.name||'—',d=>d.customer?.phone||'—',d=>d.items?.map((i:any)=>i.item_name+(i.quantity>1?'×'+i.quantity:'')).join(' ')||'—',(d.final_amount||0).toFixed(2),d=>d.payments?.map((p:any)=>pmLabel(p.pay_method)).join('/')||'—',d=>d.status==='paid'?'已付':d.status==='refunded'?'已退款':'待付'])">📤</el-button>
          </div>
          <el-table :data="orders" v-loading="ordersLoading" stripe border max-height="600">
            <el-table-column label="日期" width="90"><template #default="{row:r}">{{ fmtDate(r.created_at) }}</template></el-table-column>
            <el-table-column label="单号" prop="order_no" width="140"/>
            <el-table-column label="客户" min-width="80"><template #default="{row:r}">{{ r.customer?.name||'—' }}</template></el-table-column>
            <el-table-column label="电话" width="110"><template #default="{row:r}">{{ r.customer?.phone||'—' }}</template></el-table-column>
            <el-table-column label="项目明细" min-width="130"><template #default="{row:r}"><span v-if="r.items?.length">{{ r.items.map((i:any)=>i.item_name+(i.quantity>1?'×'+i.quantity:'')).join('、') }}</span><span v-else>—</span></template></el-table-column>
            <el-table-column label="金额" width="90" align="right"><template #default="{row:r}">¥{{ (r.final_amount||0).toFixed(2) }}</template></el-table-column>
            <el-table-column label="支付方式" width="90"><template #default="{row:r}"><span v-if="r.payments?.length">{{ r.payments.map((p:any)=>pmLabel(p.pay_method)).join('/') }}</span><span v-else>—</span></template></el-table-column>
            <el-table-column label="状态" width="70"><template #default="{row:r}"><el-tag :type="statusType(r.status)" size="small">{{ statusLabel(r.status) }}</el-tag></template></el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- Tab 5: 培训认证 -->
        <el-tab-pane label="🎓 培训认证" name="training">
          <!-- 统计卡片: 10个 -->
          <el-row :gutter="8" style="margin-bottom:8px">
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#409eff">{{ trStats.total }}</div><div style="font-size:11px;color:#909399">总培训</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#67c23a">{{ trStats.hours }}</div><div style="font-size:11px;color:#909399">总学时</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#409eff">{{ trStats.staff }}</div><div style="font-size:11px;color:#909399">参与员工</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#e6a23c">{{ trStats.expiring }}</div><div style="font-size:11px;color:#909399">即将到期</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#f56c6c">{{ trStats.expired }}</div><div style="font-size:11px;color:#909399">已过期</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#b37feb">{{ trStats.points }}</div><div style="font-size:11px;color:#909399">总积分</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#f5222d">{{ trStats.cost }}</div><div style="font-size:11px;color:#909399">总费用</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#13c2c2">{{ trStats.pass_rate }}%</div><div style="font-size:11px;color:#909399">考核通过率</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#722ed1">{{ trStats.catCount }}类</div><div style="font-size:11px;color:#909399">培训分类</div></div></el-card></el-col>
            <el-col :span="6" :lg="2"><el-card shadow="never"><div style="text-align:center;padding:6px"><div style="font-size:20px;font-weight:700;color:#fa8c16">{{ trStats.avgScore }}</div><div style="font-size:11px;color:#909399">平均分</div></div></el-card></el-col>
          </el-row>

          <!-- 筛选栏 -->
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:8px;flex-wrap:wrap">
            <el-input v-model="trSearch" placeholder="搜索课程名称" clearable style="width:150px" size="small"/>
            <el-select v-model="trCatFilter" placeholder="分类" clearable style="width:100px" size="small">
              <el-option label="内部培训" value="internal"/>
              <el-option label="厂家培训" value="manufacturer"/>
              <el-option label="外出进修" value="external"/>
              <el-option label="线上课程" value="online"/>
              <el-option label="法规培训" value="regulation"/>
              <el-option label="考核培训" value="exam"/>
            </el-select>
            <el-select v-model="trStaffFilter" placeholder="员工" clearable style="width:100px" size="small">
              <el-option v-for="u in users" :key="u.id" :label="u.real_name||u.username" :value="u.id"/>
            </el-select>
            <el-select v-model="trPassFilter" placeholder="考核状态" clearable style="width:100px" size="small">
              <el-option label="已通过" value="passed"/>
              <el-option label="未通过" value="failed"/>
              <el-option label="待考核" value="pending"/>
            </el-select>
            <span style="color:#909399;font-size:12px">{{ filteredTraining.length }}条</span>
            <div style="flex:1"/>
            <el-button size="small" text @click="exportTraining">导出</el-button>
            <el-button type="primary" size="small" @click="openTrainingCreate">+ 新增培训</el-button>
          </div>

          <!-- 主表格 -->
          <el-table :data="filteredTraining" v-loading="trLoading" stripe border size="small" style="margin-bottom:12px" max-height="420">
            <el-table-column label="课程" prop="title" min-width="100"/>
            <el-table-column label="分类" width="65"><template #default="{row:r}">{{ catLabel(r.category) }}</template></el-table-column>
            <el-table-column label="员工" width="60"><template #default="{row:r}">{{ trainerName(r.user_id) }}</template></el-table-column>
            <el-table-column label="讲师" prop="trainer" width="65"/>
            <el-table-column label="日期" prop="date" width="70"/>
            <el-table-column label="学时" prop="hours" width="40" align="center"/>
            <el-table-column label="考核" width="75"><template #default="{row:r}"><el-tag :type="r.passed==='passed'?'success':r.passed==='failed'?'danger':'info'" size="small">{{ r.passed==='passed'?'通过':r.passed==='failed'?'未过':'待考' }}</el-tag></template></el-table-column>
            <el-table-column label="费用" width="65"><template #default="{row:r}">{{ r.cost?'¥'+r.cost:'—' }}</template></el-table-column>
            <el-table-column label="积分" prop="points" width="40" align="center"/>
            <el-table-column label="证书" width="75"><template #default="{row:r}"><span :style="certStyle(r.cert_expiry)">{{ r.cert_number||'—' }}</span></template></el-table-column>
            <el-table-column label="满意度" width="50" align="center"><template #default="{row:r}">{{ r.satisfaction||'—' }}</template></el-table-column>
            <el-table-column label="操作" width="90">
              <template #default="{row:r}">
                <el-button size="small" text @click="editTraining(r)">编辑</el-button>
                <el-button size="small" text type="danger" @click="deleteTraining(r)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 培训计划/必修课 -->
          <el-collapse style="margin-bottom:12px">
            <el-collapse-item title="📋 培训计划与必修课" name="plans">
              <div style="display:flex;gap:8px;margin-bottom:8px;flex-wrap:wrap">
                <el-select v-model="planStatusFilter" placeholder="状态" clearable style="width:100px" size="small">
                  <el-option label="进行中" value="active"/>
                  <el-option label="已完成" value="completed"/>
                </el-select>
                <div style="flex:1"/>
                <el-button size="small" type="primary" @click="planDialog=true">+ 新增计划</el-button>
              </div>
              <el-table :data="filteredPlans" stripe border size="small" v-loading="plansLoading">
                <el-table-column label="计划名称" prop="title" min-width="100"/>
                <el-table-column label="分类" width="65"><template #default="{row:r}">{{ catLabel(r.category) }}</template></el-table-column>
                <el-table-column label="必修岗位" prop="role_required" width="80"/>
                <el-table-column label="目标学时" prop="target_hours" width="60" align="center"/>
                <el-table-column label="必修" width="45" align="center"><template #default="{row:r}"><el-tag :type="r.required?'danger':'info'" size="small">{{ r.required?'是':'否' }}</el-tag></template></el-table-column>
                <el-table-column label="截止日期" prop="deadline" width="80"/>
                <el-table-column label="状态" width="65"><template #default="{row:r}"><el-tag :type="r.status==='active'?'success':'info'" size="small">{{ r.status==='active'?'进行中':'已完成' }}</el-tag></template></el-table-column>
                <el-table-column label="操作" width="80">
                  <template #default="{row:r}">
                    <el-button size="small" text @click="editPlan(r)">编辑</el-button>
                    <el-button size="small" text type="danger" @click="deletePlan(r)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </el-collapse-item>
          </el-collapse>

          <!-- 培训分类分布 & 月度趋势 -->
          <el-row :gutter="12" style="margin-bottom:12px">
            <el-col :span="8">
              <el-card shadow="never"><template #header><span style="font-weight:600;font-size:13px">📊 培训分类分布</span></template>
                <div v-if="catStats.length">
                  <div v-for="c in catStats" :key="c.category" style="display:flex;align-items:center;margin-bottom:4px;gap:6px">
                    <span style="width:64px;font-size:12px;text-align:right">{{ catLabel(c.category) }}</span>
                    <div style="flex:1;height:18px;background:#f0f0f0;border-radius:3px;overflow:hidden">
                      <div :style="{height:'100%',background:catColor(c.category),width:catPct(c)}"></div>
                    </div>
                    <span style="width:40px;font-size:11px;color:#909399">{{ c.count }}</span>
                  </div>
                </div>
                <div v-else style="color:#909399;font-size:13px;text-align:center;padding:12px">暂无数据</div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card shadow="never"><template #header><span style="font-weight:600;font-size:13px">📈 月度培训趋势</span></template>
                <div v-if="monthlyStats.length">
                  <div v-for="m in monthlyStats" :key="m.mon" style="display:flex;align-items:center;margin-bottom:4px;gap:6px">
                    <span style="width:54px;font-size:12px;text-align:right">{{ m.mon }}</span>
                    <div style="flex:1;height:18px;background:#f0f0f0;border-radius:3px;overflow:hidden">
                      <div :style="{height:'100%',background:'#409eff',width:monthPct(m)}"></div>
                    </div>
                    <span style="width:40px;font-size:11px;color:#909399">{{ m.count }}次</span>
                  </div>
                </div>
                <div v-else style="color:#909399;font-size:13px;text-align:center;padding:12px">暂无数据</div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card shadow="never"><template #header><span style="font-weight:600;font-size:13px">🏆 员工培训排名</span></template>
                <div v-if="staffStats.length">
                  <div v-for="(s,i) in staffStats.slice(0,8)" :key="s.user_id" style="display:flex;align-items:center;margin-bottom:4px;gap:6px">
                    <span style="width:16px;font-size:12px;font-weight:700;color:#909399">{{ i+1 }}</span>
                    <span style="width:56px;font-size:12px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap">{{ s.real_name }}</span>
                    <div style="flex:1;height:16px;background:#f0f0f0;border-radius:3px;overflow:hidden">
                      <div :style="{height:'100%',background:rankColor(i),width:rankPct(s)}"></div>
                    </div>
                    <span style="width:36px;font-size:11px;color:#909399">{{ s.total_hours }}h</span>
                  </div>
                </div>
                <div v-else style="color:#909399;font-size:13px;text-align:center;padding:12px">暂无数据</div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 按员工详细统计 -->
          <el-collapse style="margin-bottom:8px">
            <el-collapse-item title="📝 按员工详细培训统计" name="staffDetail">
              <el-table :data="staffStats" stripe border size="small">
                <el-table-column label="姓名" width="70"><template #default="{row:r}">{{ r.real_name }}</template></el-table-column>
                <el-table-column label="培训次数" prop="total_sessions" width="70" align="center"/>
                <el-table-column label="总学时" prop="total_hours" width="60" align="center"/>
                <el-table-column label="总积分" prop="total_points" width="60" align="center"/>
                <el-table-column label="培训费用" width="70" align="center"><template #default="{row:r}">¥{{ r.total_cost||0 }}</template></el-table-column>
                <el-table-column label="通过率" width="60" align="center"><template #default="{row:r}">{{ (r.pass_rate||0).toFixed(0) }}%</template></el-table-column>
              </el-table>
            </el-collapse-item>
          </el-collapse>
        </el-tab-pane>

      </el-tabs>
    </el-card>

    <el-dialog v-model="commDlg" title="新建提成规则" width="450px">
      <el-form :model="commForm" label-width="80px">
        <el-form-item label="名称"><el-input v-model="commForm.name"/></el-form-item>
        <el-form-item label="角色"><el-select v-model="commForm.role" style="width:100%"><el-option label="咨询师" value="consultant"/><el-option label="医生" value="doctor"/><el-option label="护士" value="nurse"/></el-select></el-form-item>
        <el-form-item label="比例%"><el-input-number v-model="commForm.rate" :min="0" :precision="2" style="width:100%"/></el-form-item>
        <el-form-item label="项目"><el-input v-model="commForm.procedure" placeholder="留空=全部"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="commDlg=false">取消</el-button><el-button type="primary" @click="saveComm">保存</el-button></template>
    </el-dialog>
  
    <!-- 培训认证弹窗（新增/编辑） -->
    <el-dialog v-model="trDialog" :title="trEditId?'编辑培训记录':'新增培训记录'" width="600px">
      <el-form :model="trForm" label-width="100px" size="small">
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="课程名称"><el-input v-model="trForm.title" placeholder="培训课程名称"/></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="培训分类"><el-select v-model="trForm.category" style="width:100%">
            <el-option label="内部培训" value="internal"/><el-option label="厂家培训" value="manufacturer"/>
            <el-option label="外出进修" value="external"/><el-option label="线上课程" value="online"/>
            <el-option label="法规培训" value="regulation"/><el-option label="考核培训" value="exam"/>
          </el-select></el-form-item></el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12"><el-form-item label="参训员工"><el-select v-model="trForm.user_id" filterable style="width:100%" placeholder="选择员工"><el-option v-for="u in users" :key="u.id" :label="u.real_name||u.username" :value="u.id"/></el-select></el-form-item></el-col>
          <el-col :span="12"><el-form-item label="讲师"><el-input v-model="trForm.trainer" placeholder="培训讲师姓名"/></el-form-item></el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="8"><el-form-item label="日期"><el-date-picker v-model="trForm.date" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="学时"><el-input-number v-model="trForm.hours" :min="0" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="培训地点"><el-input v-model="trForm.location" placeholder="地点"/></el-form-item></el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="8"><el-form-item label="培训费用"><el-input-number v-model="trForm.cost" :min="0" :precision="2" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="培训积分"><el-input-number v-model="trForm.points" :min="0" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="满意度(1-5)"><el-select v-model="trForm.satisfaction" style="width:100%">
            <el-option v-for="n in 5" :key="n" :label="'⭐'.repeat(n)" :value="n"/>
          </el-select></el-form-item></el-col>
        </el-row>
        <el-divider style="margin:8px 0">考核信息</el-divider>
        <el-row :gutter="12">
          <el-col :span="8"><el-form-item label="笔试分数"><el-input-number v-model="trForm.exam_score" :min="0" :precision="1" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="实操评分"><el-input-number v-model="trForm.practical_score" :min="0" :precision="1" style="width:100%"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="考核结果"><el-select v-model="trForm.passed" style="width:100%">
            <el-option label="待考核" value="pending"/><el-option label="已通过" value="passed"/><el-option label="未通过" value="failed"/>
          </el-select></el-form-item></el-col>
        </el-row>
        <el-divider style="margin:8px 0">证书信息</el-divider>
        <el-row :gutter="12">
          <el-col :span="8"><el-form-item label="证书编号"><el-input v-model="trForm.cert_number" placeholder="证书编号"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="发证机构"><el-input v-model="trForm.cert_issuer" placeholder="发证机构名称"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="证书到期"><el-date-picker v-model="trForm.cert_expiry" type="date" value-format="YYYY-MM-DD" style="width:100%"/></el-form-item></el-col>
        </el-row>
        <el-divider style="margin:8px 0">其他信息</el-divider>
        <el-row :gutter="12">
          <el-col :span="8"><el-form-item label="是否必修"><el-switch v-model="trForm.is_mandatory"/></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="必修岗位"><el-select v-model="trForm.role_required" clearable style="width:100%">
            <el-option label="全部" value=""/><el-option label="医生" value="doctor"/><el-option label="护士" value="nurse"/>
            <el-option label="咨询师" value="consultant"/><el-option label="前台" value="receptionist"/>
          </el-select></el-form-item></el-col>
          <el-col :span="8"><el-form-item label="材料附件"><el-input v-model="trForm.material_url" placeholder="附件URL"/></el-form-item></el-col>
        </el-row>
        <el-form-item label="备注"><el-input v-model="trForm.notes" type="textarea" :rows="2" placeholder="备注"/></el-form-item>
      </el-form>
      <template #footer><el-button @click="trDialog=false">取消</el-button><el-button type="primary" :loading="trSaving" @click="submitTraining">保存</el-button></template>
    </el-dialog>
</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../../api'

function today(){const d=new Date();return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`}
function fmtDate(ts:string){if(!ts)return'—';const d=new Date(ts);return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}-${String(d.getDate()).padStart(2,'0')}`}
const now=new Date();const curMon=today().slice(0,7)
function prevMonth(m:string){const y=parseInt(m.slice(0,4)),mo=parseInt(m.slice(5))-1;return mo===0?(y-1)+'-12':y+'-'+String(mo).padStart(2,'0')}

const tab=ref('kpi');const ym=ref(curMon)

// 排行榜日期
const rankStart=ref(curMon+'-01');const rankEnd=ref(today())
const commStart=ref(rankStart.value);const commEnd=ref(rankEnd.value)
const orderStart=ref(rankStart.value);const orderEnd=ref(rankEnd.value)

// KPI
const kpiLoading=ref(false);const kpiData=ref<any[]>([]);const users=ref<any[]>([])
function userName(id:number){return users.value.find(u=>u.id===id)?.real_name||'#'+id}
function trainerName(id:number){return userName(id)}
async function loadKpi(){
  kpiLoading.value=true
  try{
    const t=await api.get('/kpi/targets',{params:{year_month:ym.value}})
    const targets=Array.isArray(t.data)?t.data:[]
    const u=await api.get('/users');users.value=Array.isArray(u.data)?u.data:[]
    kpiData.value=users.value.filter((u:any)=>u.role_id!==1).map((u:any)=>{
      const tx=targets.find((x:any)=>x.user_id===u.id)
      return{user_id:u.id,year_month:ym.value,revenue_target:tx?.revenue_target||0,order_target:tx?.order_target||0,followup_target:tx?.followup_target||0,new_customer_target:tx?.new_customer_target||0}
    })
  }finally{kpiLoading.value=false}
}
async function saveKpi(r:any){await api.post('/kpi/targets',r);ElMessage.success('已保存')}

// 排行榜
const rankLoading=ref(false);const rankData=ref<any[]>([])
async function loadRank(){
  rankLoading.value=true
  try{const r=await api.get('/kpi/leaderboard',{params:{year_month:rankEnd.value.slice(0,7)}});rankData.value=Array.isArray(r.data)?r.data:[]}finally{rankLoading.value=false}
}

// 提成
const commLoading=ref(false);const commRules=ref<any[]>([]);const commResults=ref<any[]>([]);const commDlg=ref(false)
const commForm=ref({name:'',role:'consultant',rate:0,procedure:'',rule_type:'percentage',is_active:true})
async function loadComm(){commLoading.value=true;try{const r=await api.get('/commission/rules');commRules.value=Array.isArray(r.data)?r.data:[]}finally{commLoading.value=false}}
async function calcComm(){try{const r=await api.post('/commission/calculate',{},{params:{year_month:commEnd.value.slice(0,7)}});ElMessage.success(r.data?.message||'计算完成');await loadCommResults()}catch(e:any){ElMessage.error(e?.response?.data?.error||'计算失败')}}
async function loadCommResults(){try{const r=await api.get('/commission/results',{params:{year_month:commEnd.value.slice(0,7)}});commResults.value=Array.isArray(r.data)?r.data:[]}catch{}}
async function saveComm(){await api.post('/commission/rules',commForm.value);ElMessage.success('已创建');commDlg.value=false;loadComm()}
async function toggleComm(r:any){await api.put('/commission/rules/'+r.id,{...r,is_active:!r.is_active});r.is_active=!r.is_active}

// 收费明细
const ordersLoading=ref(false);const orders=ref<any[]>([])
const orderTotal=computed(()=>orders.value.reduce((s:number,r:any)=>s+(r.final_amount||0),0))
const pmLabels:Record<string,string>={wechat:'微信',alipay:'支付宝',card:'银行卡',balance:'余额',gift_balance:'赠金',cash:'现金'}
function pmLabel(m:string){return pmLabels[m]||m}
function statusType(s:string){return s==='paid'?'success':s==='refunded'?'info':'warning'}
function statusLabel(s:string){return s==='paid'?'已付':s==='refunded'?'已退款':'待付'}
async function loadOrders(){
  if(!orderStart.value||!orderEnd.value){ElMessage.warning('请选择日期范围');return}
  ordersLoading.value=true
  try{const r=await api.get('/orders',{params:{start_date:orderStart.value,end_date:orderEnd.value}});const d=r.data;orders.value=Array.isArray(d)?d:d?.data||d?.list||[]}catch{orders.value=[]}
  finally{ordersLoading.value=false}
}

// 回访排行
const fuRanking=ref<any[]>([]);const followupTotal=ref(0);const followupDone=ref(0)
async function loadFollowupRanking(){
  try{
    const f=await api.get('/followup/tasks');const fd=f.data;const fl=Array.isArray(fd)?fd:fd?.items||[]
    followupTotal.value=fl.length
    followupDone.value=fl.filter((t:any)=>t.status==='completed').length
    const umap:Record<number,{total:number,completed:number}>={}
    fl.forEach((t:any)=>{
      const uid=t.created_by||0
      if(!umap[uid])umap[uid]={total:0,completed:0}
      umap[uid].total++
      if(t.status==='completed')umap[uid].completed++
    })
    fuRanking.value=Object.entries(umap).map(([uid,data])=>({user_id:parseInt(uid),total:data.total,completed:data.completed,rate:data.total>0?Math.round((data.completed/data.total)*100):0})).sort((a,b)=>b.rate-a.rate)
  }catch{}
}

// 业绩对比
const cmpBase=ref(prevMonth(curMon));const cmpTarget=ref(curMon)
const cmpLoading=ref(false);const compareData=ref<any[]>([])
async function loadCompare(){
  if(!cmpBase.value||!cmpTarget.value){ElMessage.warning('请选择对比月份');return}
  cmpLoading.value=true
  try{
    const [br,tr]=await Promise.all([
      api.get('/kpi/leaderboard',{params:{year_month:cmpBase.value}}),
      api.get('/kpi/leaderboard',{params:{year_month:cmpTarget.value}})
    ])
    const base=Array.isArray(br.data)?br.data:[]
    const target=Array.isArray(tr.data)?tr.data:[]
    const umap:Record<number,any>={}
    base.forEach((b:any)=>{const uid=b.user_id;umap[uid]={user_id:uid,name:userName(uid),baseRev:b.revenue||0,baseOrd:b.orders||0,targetRev:0,targetOrd:0}})
    target.forEach((t:any)=>{const uid=t.user_id;if(umap[uid]){umap[uid].targetRev=t.revenue||0;umap[uid].targetOrd=t.orders||0}else{umap[uid]={user_id:uid,name:userName(uid),baseRev:0,baseOrd:0,targetRev:t.revenue||0,targetOrd:t.orders||0}}})
    compareData.value=Object.values(umap).map((d:any)=>({...d,change:d.targetRev-d.baseRev,changePct:d.baseRev>0?Math.round((d.targetRev-d.baseRev)/d.baseRev*100):0})).sort((a:any,b:any)=>b.changePct-a.changePct)
  }catch{ElMessage.error('加载对比数据失败')}
  finally{cmpLoading.value=false}
}

// ===== 培训认证 (全功能增强版) =====

// CRUD 状态
const trSearch=ref('');const trCatFilter=ref('');const trStaffFilter=ref<number|null>(null);const trPassFilter=ref('')
const trDialog=ref(false);const trSaving=ref(false);const trEditId=ref<number|null>(null)
const emptyForm={title:'',user_id:null,trainer:'',date:'',hours:1,cert_expiry:'',notes:'',
  category:'internal',location:'',cost:0,exam_score:0,practical_score:0,passed:'pending',
  cert_number:'',cert_issuer:'',cert_image:'',material_url:'',satisfaction:0,
  is_mandatory:false,role_required:'',points:0}
const trForm=ref({...emptyForm})

// 筛选后列表
const filteredTraining=computed(()=>{
  let a=trainingList.value
  if(trSearch.value){const q=trSearch.value.toLowerCase();a=a.filter(function(x){return x.title?.toLowerCase().includes(q)})}
  if(trCatFilter.value)a=a.filter(function(x){return x.category===trCatFilter.value})
  if(trStaffFilter.value)a=a.filter(function(x){return x.user_id===trStaffFilter.value})
  if(trPassFilter.value)a=a.filter(function(x){return x.passed===trPassFilter.value})
  return a
})

// 统计指标
const trStats=computed(()=>{
  const a=trainingList.value;const now=Date.now();const d30=30*86400000
  const staff=new Set(a.map(function(x){return x.user_id}))
  let hours=0,expiring=0,expired=0,points=0,cost=0,passed=0,examined=0,score=0,scored=0,cats=new Set()
  a.forEach(function(x){
    hours+=x.hours||0;points+=x.points||0;cost+=x.cost||0
    if(x.category)cats.add(x.category)
    if(x.cert_expiry){const d=new Date(x.cert_expiry).getTime()-now;if(d<=0)expired++;else if(d<=d30)expiring++}
    if(x.passed==='passed')passed++
    if(x.passed==='passed'||x.passed==='failed')examined++
    if(x.exam_score>0||x.practical_score>0){scored++;score+=(x.exam_score||0)+(x.practical_score||0)}
  })
  return{
    total:a.length,hours,staff:staff.size,expiring,expired,
    points,cost:cost?'¥'+cost.toFixed(0):'¥0',
    pass_rate:examined?((passed/examined)*100).toFixed(1):'—',
    catCount:cats.size,
    avgScore:scored?(score/scored).toFixed(1):'—'
  }
})

function catLabel(v:string){
  const m:{[k:string]:string}={internal:'内部培训',manufacturer:'厂家培训',external:'外出进修',online:'线上课程',regulation:'法规培训',exam:'考核培训'}
  return m[v]||v||'—'
}
function certStyle(date:string){
  if(!date)return{}
  const d=new Date(date).getTime()-Date.now()
  if(d<=0)return{color:'#f56c6c',fontWeight:'700'}
  if(d<=30*86400000)return{color:'#e6a23c',fontWeight:'600'}
  return{color:'#67c23a'}
}

const catColors=['#409eff','#67c23a','#e6a23c','#f56c6c','#b37feb','#13c2c2']
function catColor(cat:string){const idx=['internal','manufacturer','external','online','regulation','exam'].indexOf(cat);return catColors[idx>=0?idx:0]}
function catPct(c:any){const total=catStats.value.reduce((s,x:any)=>s+x.count,0);return total?((c.count/total)*100).toFixed(0)+'%':'0%'}
function monthPct(m:any){const max=Math.max(...monthlyStats.value.map((x:any)=>x.count));return max?((m.count/max)*100).toFixed(0)+'%':'0%'}
function rankPct(s:any){const max=Math.max(...staffStats.value.map((x:any)=>x.total_hours));return max?((s.total_hours/max)*100).toFixed(0)+'%':'0%'}
function rankColor(i:number){return['#f5222d','#fa8c16','#fadb14','#409eff','#67c23a','#13c2c2','#b37feb','#909399'][i]||'#909399'}

// 打开新增
function openTrainingCreate(){trForm.value={...emptyForm};trEditId.value=null;trDialog.value=true}
// 打开编辑
function editTraining(r:any){
  trEditId.value=r.id
  trForm.value={
    title:r.title||'',user_id:r.user_id,trainer:r.trainer||'',date:r.date||'',hours:r.hours||1,cert_expiry:r.cert_expiry||'',notes:r.notes||'',
    category:r.category||'internal',location:r.location||'',cost:r.cost||0,exam_score:r.exam_score||0,practical_score:r.practical_score||0,passed:r.passed||'pending',
    cert_number:r.cert_number||'',cert_issuer:r.cert_issuer||'',cert_image:r.cert_image||'',material_url:r.material_url||'',satisfaction:r.satisfaction||0,
    is_mandatory:r.is_mandatory||false,role_required:r.role_required||'',points:r.points||0
  }
  trDialog.value=true
}
// 提交保存
async function submitTraining(){
  trSaving.value=true
  try{
    if(trEditId.value){
      await api.put('/training/'+trEditId.value,trForm.value);ElMessage.success('更新成功')
    }else{
      await api.post('/training',trForm.value);ElMessage.success('保存成功')
    }
    trDialog.value=false;await loadTraining()
  }
  catch(e:any){ElMessage.error(e?.response?.data?.error||'保存失败')}
  finally{trSaving.value=false}
}
// 删除
async function deleteTraining(row:any){
  try{
    await ElMessageBox.confirm('确定删除培训记录「'+row.title+'」？','确认',{type:'warning'})
    await api.delete('/training/'+row.id);ElMessage.success('已删除');await loadTraining()
  }catch{}
}
// 导出
function exportTraining(){
  var labels=["课程","分类","员工","讲师","日期","学时","笔试分","实操分","考核","费用","积分","证书编号","证书到期","满意度"];
  var getters=[
    function(d){return d.title},
    function(d){return catLabel(d.category)},
    function(d){return trainerName(d.user_id)},
    function(d){return d.trainer||""},
    function(d){return d.date||""},
    function(d){return String(d.hours||0)},
    function(d){return String(d.exam_score||"")},
    function(d){return String(d.practical_score||"")},
    function(d){return d.passed==="passed"?"通过":d.passed==="failed"?"未过":"待考"},
    function(d){return d.cost?"Y"+d.cost:""},
    function(d){return String(d.points||"")},
    function(d){return d.cert_number||""},
    function(d){return d.cert_expiry||""},
    function(d){return d.satisfaction?String(d.satisfaction)+"xing":""}
  ];
  exportTable(filteredTraining.value,"培训记录",labels,getters);
}

async function loadTraining(){
  trLoading.value=true
  try{
    const [tl,ss,cs,ms]=await Promise.all([
      api.get("/training"),
      api.get("/training/staff-stats"),
      api.get("/training/category-stats"),
      api.get("/training/monthly-stats")
    ])
    trainingList.value=Array.isArray(tl.data)?tl.data:[]
    staffStats.value=Array.isArray(ss.data)?ss.data:[]
    catStats.value=Array.isArray(cs.data)?cs.data:[]
    monthlyStats.value=Array.isArray(ms.data)?ms.data:[]
  }catch{}
  finally{trLoading.value=false}
}

// ===== 培训计划 =====
const plansLoading=ref(false);const plans=ref<any[]>([]);const planStatusFilter=ref('');const planDialog=ref(false);const planSaving=ref(false);const planEditId=ref<number|null>(null)
const emptyPlan={title:'',category:'internal',role_required:'',target_hours:0,required:false,deadline:'',description:'',status:'active'}
const planForm=ref({...emptyPlan})

const filteredPlans=computed(()=>{
  let a=plans.value
  if(planStatusFilter.value)a=a.filter(function(x){return x.status===planStatusFilter.value})
  return a
})

async function loadPlans(){
  plansLoading.value=true
  try{
    const r=await api.get('/training/plans')
    plans.value=Array.isArray(r.data)?r.data:[]
  }finally{plansLoading.value=false}
}
function openPlanCreate(){planForm.value={...emptyPlan};planEditId.value=null;planDialog.value=true}
function editPlan(r:any){
  planEditId.value=r.id
  planForm.value={title:r.title||'',category:r.category||'internal',role_required:r.role_required||'',target_hours:r.target_hours||0,required:r.required||false,deadline:r.deadline||'',description:r.description||'',status:r.status||'active'}
  planDialog.value=true
}
async function submitPlan(){
  planSaving.value=true
  try{
    if(planEditId.value){await api.put('/training/plans/'+planEditId.value,planForm.value)}else{await api.post('/training/plans',planForm.value)}
    ElMessage.success('保存成功');planDialog.value=false;await loadPlans()
  }catch(e:any){ElMessage.error(e?.response?.data?.error||'保存失败')}
  finally{planSaving.value=false}
}
async function deletePlan(row:any){
  try{
    await ElMessageBox.confirm('确定删除培训计划「'+row.title+'」？','确认',{type:'warning'})
    await api.delete('/training/plans/'+row.id);ElMessage.success('已删除');await loadPlans()
  }catch{}
}

const trLoading=ref(false);const trainingList=ref<any[]>([]);const trainingStats=ref<any[]>([])

// 导出
function exportTable(data:any[],name:string,labels:string[],getters:((d:any,i:number)=>string)[]){
  if(!data.length)return
  const header=labels.map(l=>`"${l}"`).join(',')
  const rows=data.map((d,i)=>getters.map(g=>`"${g(d,i)}"`).join(',')).join('\n')
  const blob=new Blob(['\uFEFF'+header+'\n'+rows],{type:'text/csv;charset=utf-8'})
  const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download=name+'_'+today()+'.csv'
  a.click();URL.revokeObjectURL(a.href)
}


const autoCalcComm=ref(false)
const CALC_KEY='clinic_auto_calc_comm'
function saveAutoCalc(){
  localStorage.setItem(CALC_KEY,JSON.stringify({enabled:autoCalcComm.value}))
}
function loadAutoCalc(){
  try{const raw=localStorage.getItem(CALC_KEY);if(raw){const s=JSON.parse(raw);autoCalcComm.value=s.enabled}}catch{}
}


function autoCalcTrigger(){
  if(!autoCalcComm.value)return
  const now=new Date()
  if(now.getDate()!==1)return
  const key='clinic_auto_calc_done_'+now.getFullYear()+'-'+String(now.getMonth()+1).padStart(2,'0')
  if(localStorage.getItem(key))return
  setTimeout(async function(){
    try{
      await api.post('/commission/calculate',{},{params:{year_month:now.getFullYear()+'-'+String(now.getMonth()+1).padStart(2,'0')}})
      localStorage.setItem(key,'1')
      await loadCommResults()
    }catch{}
  },2000)
}

onMounted(()=>{loadKpi();loadRank();loadComm();loadOrders();loadFollowupRanking();loadCompare();loadTraining();loadPlans()});loadAutoCalc();autoCalcTrigger()
</script>
