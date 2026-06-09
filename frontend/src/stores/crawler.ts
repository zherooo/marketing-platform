import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { OAuthToken, CrawlTask, CrawlStatistics } from '@/types'

export const useCrawlerStore = defineStore('crawler', () => {
  // OAuth 账号列表
  const accounts = ref<OAuthToken[]>([])
  
  // 当前选中的账号
  const currentAccount = ref<string>('')
  
  // 抓取任务列表
  const tasks = ref<CrawlTask[]>([])
  
  // 抓取统计
  const statistics = ref<CrawlStatistics | null>(null)
  
  // 正在运行的任务
  const runningTasks = ref<CrawlTask[]>([])

  const setAccounts = (list: OAuthToken[]) => {
    accounts.value = list
    // 默认选中第一个在线账号
    if (list.length > 0 && !currentAccount.value) {
      const onlineAccount = list.find(a => a.is_online)
      currentAccount.value = onlineAccount?.account_id || list[0].account_id
    }
  }

  const setCurrentAccount = (accountId: string) => {
    currentAccount.value = accountId
  }

  const setTasks = (list: CrawlTask[]) => {
    tasks.value = list
  }

  const setStatistics = (stats: CrawlStatistics) => {
    statistics.value = stats
  }

  const setRunningTasks = (list: CrawlTask[]) => {
    runningTasks.value = list
  }

  return {
    accounts,
    currentAccount,
    tasks,
    statistics,
    runningTasks,
    setAccounts,
    setCurrentAccount,
    setTasks,
    setStatistics,
    setRunningTasks
  }
})
