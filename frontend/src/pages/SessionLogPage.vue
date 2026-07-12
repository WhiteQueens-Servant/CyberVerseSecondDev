<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppHeader from '../components/AppHeader.vue'
import {
  getAnalyticsSessions, getAnalyticsSessionDetail, getCharacters,
  type SessionLog, type SessionLogDetail, type Character,
} from '../services/api'

const { t } = useI18n()

// ── State ──
const sessions = ref<SessionLog[]>([])
const characters = ref<Character[]>([])
const loading = ref(false)
const expandedId = ref<string | null>(null)
const detail = ref<SessionLogDetail | null>(null)
const detailLoading = ref(false)

// Filters
const filterCharacterId = ref('')
const filterSentiment = ref('')
const filterDateFrom = ref('')
const filterDateTo = ref('')

const sentimentColors: Record<string, string> = {
  positive: 'text-green-400 bg-green-400/10',
  neutral: 'text-gray-400 bg-gray-400/10',
  negative: 'text-red-400 bg-red-400/10',
}

const sentimentLabels: Record<string, string> = {
  positive: '正面',
  neutral: '中性',
  negative: '负面',
}

// ── Actions ──
async function fetchData() {
  loading.value = true
  try {
    const [s, c] = await Promise.all([
      getAnalyticsSessions({
        character_id: filterCharacterId.value || undefined,
        sentiment: filterSentiment.value || undefined,
        date_from: filterDateFrom.value || undefined,
        date_to: filterDateTo.value || undefined,
      }),
      getCharacters(),
    ])
    sessions.value = s
    characters.value = c
  } catch (e) {
    console.error('Failed to load sessions:', e)
  } finally {
    loading.value = false
  }
}

function getCharacterName(id: string): string {
  return characters.value.find(c => c.id === id)?.name || id
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}秒`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return s > 0 ? `${m}分${s}秒` : `${m}分钟`
}

function formatTime(ts: string): string {
  try {
    const d = new Date(ts)
    return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch {
    return ts
  }
}

function formatDate(ts: string): string {
  try {
    const d = new Date(ts)
    return d.toLocaleDateString('zh-CN')
  } catch {
    return ts
  }
}

async function toggleExpand(id: string) {
  if (expandedId.value === id) {
    expandedId.value = null
    detail.value = null
    return
  }
  expandedId.value = id
  detailLoading.value = true
  try {
    detail.value = await getAnalyticsSessionDetail(id)
  } catch (e) {
    console.error('Failed to load session detail:', e)
    detail.value = null
  } finally {
    detailLoading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="min-h-screen bg-cv-base">
    <AppHeader />

    <main class="max-w-[1200px] mx-auto px-12 py-12">
      <div class="flex items-start justify-between mb-8">
        <div>
          <h1 class="text-[32px] font-semibold text-cv-text tracking-[-0.5px]">
            {{ t('nav.sessions') }}
          </h1>
          <p class="mt-2 text-sm text-cv-text-secondary">
            浏览历史对话记录 / Browse conversation logs
          </p>
        </div>
        <button @click="fetchData"
                class="px-4 py-2 text-sm text-cv-text-secondary hover:text-cv-text border border-cv-border rounded-cv-md cursor-pointer transition-colors">
          {{ t('common.refresh') }}
        </button>
      </div>

      <!-- Filters -->
      <div class="flex gap-4 mb-6 flex-wrap">
        <select v-model="filterCharacterId" @change="fetchData"
                class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
          <option value="">全部角色</option>
          <option v-for="c in characters" :key="c.id" :value="c.id">{{ c.name }}</option>
        </select>
        <select v-model="filterSentiment" @change="fetchData"
                class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
          <option value="">全部情感</option>
          <option value="positive">正面</option>
          <option value="neutral">中性</option>
          <option value="negative">负面</option>
        </select>
        <input v-model="filterDateFrom" type="date" @change="fetchData"
               class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none" />
        <input v-model="filterDateTo" type="date" @change="fetchData"
               class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none" />
      </div>

      <!-- Loading -->
      <div v-if="loading" class="text-center py-20 text-cv-text-muted">
        {{ t('common.loading') }}
      </div>

      <!-- Table -->
      <div v-else-if="sessions.length > 0" class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-cv-border-subtle text-left">
              <th class="px-4 py-3 text-cv-text-secondary font-medium">时间</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">角色</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">时长</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">轮次</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">情感</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="s in sessions" :key="s.id">
              <tr class="border-b border-cv-border-subtle hover:bg-cv-hover transition-colors cursor-pointer"
                  @click="toggleExpand(s.id)">
                <td class="px-4 py-3 text-cv-text">
                  <div>{{ formatDate(s.started_at) }}</div>
                  <div class="text-xs text-cv-text-muted">{{ formatTime(s.started_at) }}</div>
                </td>
                <td class="px-4 py-3 text-cv-text-secondary">{{ s.character_name || getCharacterName(s.character_id) }}</td>
                <td class="px-4 py-3 text-cv-text-secondary">{{ formatDuration(s.duration_s) }}</td>
                <td class="px-4 py-3 text-cv-text-secondary">{{ s.turn_count }}</td>
                <td class="px-4 py-3">
                  <span class="inline-block px-2 py-0.5 text-xs rounded-full"
                        :class="sentimentColors[s.sentiment] || 'text-gray-400 bg-gray-400/10'">
                    {{ sentimentLabels[s.sentiment] || s.sentiment }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="text-cv-accent text-sm">
                    {{ expandedId === s.id ? '收起' : '详情' }}
                  </span>
                </td>
              </tr>
              <!-- Expanded detail -->
              <tr v-if="expandedId === s.id">
                <td colspan="6" class="px-4 py-4 bg-cv-elevated">
                  <div v-if="detailLoading" class="text-center text-cv-text-muted text-sm py-4">
                    加载中...
                  </div>
                  <div v-else-if="detail && detail.messages.length > 0" class="space-y-2 max-h-[400px] overflow-y-auto">
                    <div v-for="(msg, i) in detail.messages" :key="i"
                         class="flex gap-3 text-sm"
                         :class="msg.role === 'user' ? '' : 'pl-8'">
                      <span class="shrink-0 text-xs font-medium"
                            :class="msg.role === 'user' ? 'text-cv-accent' : 'text-green-400'">
                        [{{ msg.role === 'user' ? '用户' : '数字人' }}]
                      </span>
                      <span class="text-cv-text-secondary flex-1">{{ msg.content }}</span>
                      <span class="shrink-0 text-xs text-cv-text-muted">{{ formatTime(msg.timestamp) }}</span>
                    </div>
                  </div>
                  <div v-else class="text-center text-cv-text-muted text-sm py-4">
                    暂无对话记录
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>

      <!-- Empty -->
      <div v-else class="flex flex-col items-center justify-center py-24 text-cv-text-muted">
        <p class="text-sm">暂无对话日志</p>
      </div>
    </main>
  </div>
</template>
