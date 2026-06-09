<template>
  <div class="hosting-rule-form">
    <el-card>
      <template #header>
        <el-page-header @back="$router.back()">
          <template #content>{{ isEdit ? '编辑规则' : '创建规则' }}</template>
        </el-page-header>
      </template>

      <el-form ref="formRef" :model="form" :rules="rules" label-width="120px" style="max-width: 800px">
        <!-- 基本信息 -->
        <el-divider content-position="left">基本信息</el-divider>
        <el-form-item label="规则名称" prop="rule_name">
          <el-input v-model="form.rule_name" placeholder="请输入规则名称" maxlength="100" />
        </el-form-item>
        <el-form-item label="规则分类" prop="category">
          <el-select v-model="form.category" placeholder="选择分类">
            <el-option label="成本控制" value="cost_control" />
            <el-option label="预算管理" value="budget_manage" />
            <el-option label="效果优化" value="effect_optimize" />
            <el-option label="风险预警" value="risk_alert" />
          </el-select>
        </el-form-item>
        <el-form-item label="托管场景" prop="scene">
          <el-input v-model="form.scene" placeholder="如：转化成本超出预期" maxlength="100" />
        </el-form-item>
        <el-form-item label="规则描述">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="规则描述" maxlength="500" />
        </el-form-item>
        <el-form-item label="优先级" prop="priority">
          <el-slider v-model="form.priority" :min="1" :max="10" show-input />
        </el-form-item>

        <!-- 触发条件 -->
        <el-divider content-position="left">触发条件 (IF)</el-divider>
        <el-form-item label="监控指标" prop="trigger_condition.metric">
          <el-select v-model="form.trigger_condition.metric" placeholder="选择监控指标">
            <el-option label="转化成本(CPA)" value="conversion_cost" />
            <el-option label="单次点击成本(CPC)" value="cpc" />
            <el-option label="日消耗(Spend)" value="spend" />
            <el-option label="日预算消耗比例" value="budget_ratio" />
            <el-option label="曝光量" value="impressions" />
            <el-option label="转化量" value="conversions" />
            <el-option label="投放时长" value="delivery_hours" />
          </el-select>
        </el-form-item>
        <el-form-item label="比较运算符">
          <el-select v-model="form.trigger_condition.operator" placeholder="选择运算符">
            <el-option label="大于 (>) " value="gt" />
            <el-option label="大于等于 (>=)" value="gte" />
            <el-option label="小于 (<) " value="lt" />
            <el-option label="小于等于 (<=)" value="lte" />
          </el-select>
        </el-form-item>
        <el-form-item label="阈值">
          <el-input-number v-model="form.trigger_condition.threshold" :precision="2" :step="1" style="width: 200px" />
        </el-form-item>
        <el-form-item label="持续时长(分)">
          <el-input-number v-model="form.trigger_condition.duration" :min="1" :step="5" style="width: 200px" />
        </el-form-item>
        <el-form-item v-if="form.category === 'cost_control'" label="时间范围">
          <el-input v-model="form.trigger_condition.time_range" placeholder="如 0-6 表示凌晨 0-6 点" />
        </el-form-item>
        <el-form-item v-if="form.category === 'effect_optimize'" label="最低投放天数">
          <el-input-number v-model="form.trigger_condition.ad_min_days" :min="1" />
        </el-form-item>
        <el-form-item v-if="form.category === 'effect_optimize'" label="最低转化数">
          <el-input-number v-model="form.trigger_condition.max_conv_count" :min="1" />
        </el-form-item>

        <!-- 执行动作 -->
        <el-divider content-position="left">执行动作 (THEN)</el-divider>
        <el-form-item label="动作类型" prop="execution_action.type">
          <el-select v-model="form.execution_action.type" placeholder="选择执行动作">
            <el-option label="暂停广告" value="pause_ad" />
            <el-option label="调整出价" value="adjust_bid" />
            <el-option label="提升日限额" value="raise_budget" />
            <el-option label="发送通知" value="notify" />
            <el-option label="启用广告" value="resume_ad" />
            <el-option label="一键起量" value="quick_start" />
          </el-select>
        </el-form-item>

        <template v-if="form.execution_action.type === 'adjust_bid'">
          <el-form-item label="出价调整比例">
            <el-input-number v-model="form.execution_action.bid_adjust_ratio" :min="0.1" :max="2" :step="0.05" :precision="4" style="width: 200px" />
            <span class="form-tip">（1=不变, 0.9=降价10%, 1.1=提价10%）</span>
          </el-form-item>
        </template>

        <template v-if="form.execution_action.type === 'raise_budget'">
          <el-form-item label="预算提升比例">
            <el-input-number v-model="form.execution_action.budget_raise_ratio" :min="1" :max="5" :step="0.1" :precision="2" style="width: 200px" />
            <span class="form-tip">（如 1.2 表示提升20%）</span>
          </el-form-item>
        </template>

        <template v-if="form.execution_action.type === 'notify'">
          <el-form-item label="通知渠道">
            <el-checkbox-group v-model="form.execution_action.notify_channels">
              <el-checkbox label="email" value="email">邮件</el-checkbox>
              <el-checkbox label="sms" value="sms">短信</el-checkbox>
              <el-checkbox label="dingtalk" value="dingtalk">钉钉</el-checkbox>
              <el-checkbox label="feishu" value="feishu">飞书</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="自定义消息">
            <el-input v-model="form.execution_action.message" type="textarea" :rows="2" placeholder="可选自定义通知内容" />
          </el-form-item>
        </template>

        <!-- 作用范围 -->
        <el-divider content-position="left">作用范围</el-divider>
        <el-form-item label="账号范围">
          <el-input v-model="form.account_ids" placeholder='JSON 数组，如 ["account1","account2"]' />
        </el-form-item>
        <el-form-item label="广告系列">
          <el-input v-model="form.campaign_ids" placeholder='JSON 数组' />
        </el-form-item>
        <el-form-item label="广告组">
          <el-input v-model="form.ad_group_ids" placeholder='JSON 数组' />
        </el-form-item>
        <el-form-item label="广告">
          <el-input v-model="form.ad_ids" placeholder='JSON 数组' />
        </el-form-item>

        <!-- 执行限制 -->
        <el-divider content-position="left">执行限制</el-divider>
        <el-form-item label="冷却时间(分)">
          <el-input-number v-model="form.cooldown_minutes" :min="1" :step="10" style="width: 200px" />
          <span class="form-tip">两次执行最小间隔</span>
        </el-form-item>
        <el-form-item label="每日最大执行">
          <el-input-number v-model="form.max_executions_per_day" :min="1" :max="100" style="width: 200px" />
          <span class="form-tip">单日最多执行次数</span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">
            {{ isEdit ? '保存修改' : '创建规则' }}
          </el-button>
          <el-button @click="$router.back()">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { createHostingRule, updateHostingRule, getHostingRule } from '@/api'
import type { HostingRule } from '@/types'

const route = useRoute()
const router = useRouter()
const isEdit = computed(() => !!route.params.id)

const formRef = ref()
const submitting = ref(false)

const getDefaultForm = () => ({
  rule_name: '',
  category: '',
  scene: '',
  description: '',
  priority: 5,
  trigger_condition: { type: 'cost_control', metric: '', operator: 'gt', threshold: 0, duration: 5, time_range: '', ad_min_days: 3, max_conv_count: 5 } as any,
  execution_action: { type: '', bid_adjust_ratio: 0.9, budget_raise_ratio: 1.2, notify_channels: [] as string[], message: '' } as any,
  account_ids: '[]',
  campaign_ids: '[]',
  ad_group_ids: '[]',
  ad_ids: '[]',
  cooldown_minutes: 60,
  max_executions_per_day: 10
})

const form = reactive(getDefaultForm())

const rules = {
  rule_name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  category: [{ required: true, message: '请选择分类', trigger: 'change' }],
  scene: [{ required: true, message: '请输入场景', trigger: 'blur' }],
  'execution_action.type': [{ required: true, message: '请选择动作类型', trigger: 'change' }]
}

onMounted(async () => {
  if (isEdit.value) {
    try {
      const res = await getHostingRule(Number(route.params.id))
      const data = res.data
      Object.assign(form, {
        rule_name: data.rule_name,
        category: data.category,
        scene: data.scene,
        description: data.description,
        priority: data.priority,
        trigger_condition: { ...getDefaultForm().trigger_condition, ...data.trigger_condition },
        execution_action: { ...getDefaultForm().execution_action, ...data.execution_action },
        account_ids: data.account_ids || '[]',
        campaign_ids: data.campaign_ids || '[]',
        ad_group_ids: data.ad_group_ids || '[]',
        ad_ids: data.ad_ids || '[]',
        cooldown_minutes: data.cooldown_minutes || 60,
        max_executions_per_day: data.max_executions_per_day || 10
      })
    } catch {
      ElMessage.error('获取规则信息失败')
    }
  }
})

const handleSubmit = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (isEdit.value) {
      await updateHostingRule(Number(route.params.id), form)
      ElMessage.success('规则已更新')
    } else {
      await createHostingRule(form)
      ElMessage.success('规则已创建')
    }
    router.push('/hosting/rules')
  } catch (err: any) {
    ElMessage.error(err?.response?.data?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.form-tip { margin-left: 12px; font-size: 12px; color: #909399 }
</style>
