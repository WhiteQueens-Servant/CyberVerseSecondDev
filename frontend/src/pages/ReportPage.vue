<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppHeader from '../components/AppHeader.vue'
import {
  generateReport, listReports, getReport, deleteReport,
  type Report,
} from '../services/api'

const { t } = useI18n()

const reports = ref<Report[]>([])
const loading = ref(false)
const generating = ref(false)
const selectedReport = ref<Report | null>(null)
const deleteConfirmId = ref<string | null>(null)

const statusLabels: Record<string, string> = {
  running: '生成中',
  completed: '已完成',
  failed: '失败',
  queued: '排队中',
}

const statusColors: Record<string, string> = {
  running: 'text-blue-400 bg-blue-400/10',
  completed: 'text-green-400 bg-green-400/10',
  failed: 'text-red-400 bg-red-400/10',
  queued: 'text-yellow-400 bg-yellow-400/10',
}

const sectionLabels: Record<string, string> = {
  tourist_focus: '游客关注点',
  sentiment_trend: '情感趋势',
  knowledge_gap: '知识盲区',
  improvement: '改进建议',
}

async function fetchReports() {
  loading.value = true
  try {
    reports.value = await listReports()
  } catch (e) {
    console.error('Failed to load reports:', e)
  } finally {
    loading.value = false
  }
}

async function handleGenerate() {
  generating.value = true
  try {
    await generateReport()
    // Poll for completion (simple approach for demo)
    setTimeout(async () => {
      await fetchReports()
      generating.value = false
    }, 3000)
  } catch (e) {
    console.error('Failed to generate report:', e)
    generating.value = false
  }
}

async function handleViewDetail(id: string) {
  try {
    selectedReport.value = await getReport(id)
  } catch (e) {
    console.error('Failed to load report:', e)
  }
}

async function handleDelete(id: string) {
  try {
    await deleteReport(id)
    deleteConfirmId.value = null
    if (selectedReport.value?.id === id) selectedReport.value = null
    await fetchReports()
  } catch (e) {
    console.error('Failed to delete report:', e)
  }
}

function formatDate(ts: string): string {
  try {
    return new Date(ts).toLocaleString('zh-CN')
  } catch {
    return ts
  }
}

function renderContent(content: Record<string, unknown> | undefined): Array<{ key: string; label: string; text: string }> {
  if (!content) return []
  return Object.entries(content)
    .filter(([, v]) => typeof v === 'string')
    .map(([k, v]) => ({
      key: k,
      label: sectionLabels[k] || k,
      text: v as string,
    }))
}

onMounted(fetchReports)
</script>

<template>
  <div class="min-h-screen bg-cv-base">
    <AppHeader />

    <main class="max-w-[1200px] mx-auto px-12 py-12">
      <div class="flex items-start justify-between mb-8">
        <div>
          <h1 class="text-[32px] font-semibold text-cv-text tracking-[-0.5px]">
            {{ t('nav.reports') }}
          </h1>
          <p class="mt-2 text-sm text-cv-text-secondary">
            生成和查看游客感受度分析报告 / Generate and view experience analysis reports
          </p>
        </div>
        <button @click="handleGenerate" :disabled="generating"
                class="px-5 py-2.5 bg-cv-accent text-white text-sm font-medium rounded-cv-md hover:bg-cv-accent-hover transition-colors cursor-pointer disabled:opacity-50">
          {{ generating ? '生成中...' : '生成报告' }}
        </button>
      </div>

      <div class="grid grid-cols-[1fr_1.5fr] gap-6">
        <!-- Report list -->
        <div>
          <div v-if="loading" class="text-center py-20 text-cv-text-muted text-sm">
            {{ t('common.loading') }}
          </div>
          <div v-else-if="reports.length > 0" class="space-y-3">
            <div v-for="r in reports" :key="r.id"
                 class="bg-cv-surface border rounded-cv-lg p-4 cursor-pointer transition-all"
                 :class="selectedReport?.id === r.id ? 'border-cv-accent' : 'border-cv-border-subtle hover:border-cv-border'"
                 @click="handleViewDetail(r.id)">
              <div class="flex items-start justify-between mb-2">
                <div class="text-sm text-cv-text">
                  {{ r.date_from }} ~ {{ r.date_to }}
                </div>
                <span class="inline-block px-2 py-0.5 text-xs rounded-full"
                      :class="statusColors[r.status] || 'text-gray-400 bg-gray-400/10'">
                  {{ statusLabels[r.status] || r.status }}
                </span>
              </div>
              <div class="text-xs text-cv-text-muted">
                {{ formatDate(r.created_at) }}
              </div>
              <div class="flex justify-end mt-2">
                <button @click.stop="deleteConfirmId = r.id"
                        class="text-xs text-cv-danger hover:text-red-400 cursor-pointer">
                  删除
                </button>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-20 text-cv-text-muted text-sm">
            暂无报告，点击"生成报告"开始
          </div>
        </div>

        <!-- Report detail -->
        <div>
          <div v-if="!selectedReport" class="flex items-center justify-center h-full min-h-[300px] text-cv-text-muted text-sm">
            ← 选择一份报告查看详情
          </div>
          <div v-else class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-6">
            <h2 class="text-lg font-semibold text-cv-text mb-1">
              感受度分析报告
            </h2>
            <p class="text-xs text-cv-text-muted mb-6">
              分析范围：{{ selectedReport.date_from }} ~ {{ selectedReport.date_to }}
            </p>

            <div v-if="selectedReport.status === 'running'" class="text-center py-12 text-cv-text-muted">
              报告生成中，请稍候...
            </div>
            <div v-else-if="selectedReport.status === 'failed'" class="text-center py-12 text-cv-danger">
              报告生成失败
            </div>
            <div v-else-if="renderContent(selectedReport.content).length > 0" class="space-y-6">
              <div v-for="section in renderContent(selectedReport.content)" :key="section.key">
                <h3 class="text-sm font-medium text-cv-text mb-2">{{ section.label }}</h3>
                <div class="text-sm text-cv-text-secondary whitespace-pre-wrap leading-relaxed bg-cv-elevated rounded-cv-md p-4">
                  {{ section.text }}
                </div>
              </div>
            </div>
            <div v-else class="text-center py-12 text-cv-text-muted text-sm">
              报告内容为空
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Delete Confirm -->
    <div v-if="deleteConfirmId" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="deleteConfirmId = null">
      <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg w-[400px] p-6">
        <h2 class="text-lg font-semibold text-cv-text mb-2">确认删除</h2>
        <p class="text-sm text-cv-text-secondary mb-6">删除报告后不可恢复。</p>
        <div class="flex justify-end gap-3">
          <button @click="deleteConfirmId = null"
                  class="px-4 py-2 text-sm text-cv-text-secondary hover:text-cv-text cursor-pointer">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleDelete(deleteConfirmId!)"
                  class="px-5 py-2 bg-cv-danger text-white text-sm font-medium rounded-cv-md hover:bg-red-600 transition-colors cursor-pointer">
            {{ t('common.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
