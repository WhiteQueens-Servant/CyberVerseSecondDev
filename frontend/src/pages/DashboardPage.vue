<script setup lang="ts">
import { ref, onMounted, shallowRef, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts/core'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent, TitleComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import { getDashboard, type DashboardData } from '../services/api'

echarts.use([LineChart, BarChart, PieChart, GridComponent, TooltipComponent, LegendComponent, TitleComponent, CanvasRenderer])

const { t } = useI18n()

const loading = ref(false)
const data = ref<DashboardData | null>(null)
const weekChartRef = ref<HTMLDivElement | null>(null)
const sentimentChartRef = ref<HTMLDivElement | null>(null)
const hourlyChartRef = ref<HTMLDivElement | null>(null)
const hotQuestionsChartRef = ref<HTMLDivElement | null>(null)

const weekChart = shallowRef<echarts.ECharts | null>(null)
const sentimentChart = shallowRef<echarts.ECharts | null>(null)
const hourlyChart = shallowRef<echarts.ECharts | null>(null)
const hotQuestionsChart = shallowRef<echarts.ECharts | null>(null)

async function fetchData() {
  loading.value = true
  try {
    data.value = await getDashboard()
    // Let loading=false so template renders chart containers, then wait for DOM
    loading.value = false
    await nextTick()
    renderCharts()
  } catch (e) {
    console.error('Failed to load dashboard:', e)
    loading.value = false
  }
}

function renderCharts() {
  if (!data.value) return

  // Week trend line chart
  if (weekChartRef.value && weekChartRef.value.offsetHeight > 0) {
    weekChart.value = echarts.init(weekChartRef.value)
    weekChart.value.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 20, top: 30, bottom: 30 },
      xAxis: {
        type: 'category',
        data: data.value.week_sessions.map(d => d.date),
        axisLabel: { color: '#9ca3af' },
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#9ca3af' },
        splitLine: { lineStyle: { color: '#374151' } },
      },
      series: [{
        type: 'line',
        data: data.value.week_sessions.map(d => d.count),
        smooth: true,
        lineStyle: { color: '#6BAF8D', width: 2 },
        areaStyle: { color: 'rgba(107, 175, 141, 0.1)' },
        itemStyle: { color: '#6BAF8D' },
      }],
    })
  }

  // Sentiment pie chart
  if (sentimentChartRef.value && sentimentChartRef.value.offsetHeight > 0) {
    sentimentChart.value = echarts.init(sentimentChartRef.value)
    const sd = data.value.sentiment_distribution
    sentimentChart.value.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: 0, textStyle: { color: '#9ca3af' } },
      series: [{
        type: 'pie',
        radius: ['40%', '65%'],
        center: ['50%', '45%'],
        data: [
          { name: '正面', value: sd.positive, itemStyle: { color: '#22c55e' } },
          { name: '中性', value: sd.neutral, itemStyle: { color: '#9ca3af' } },
          { name: '负面', value: sd.negative, itemStyle: { color: '#ef4444' } },
        ],
        label: { show: true, formatter: '{b}: {d}%', color: '#d1d5db' },
      }],
    })
  }

  // Hourly distribution bar chart
  if (hourlyChartRef.value && hourlyChartRef.value.offsetHeight > 0) {
    hourlyChart.value = echarts.init(hourlyChartRef.value)
    const hours = data.value.hourly_distribution
    hourlyChart.value.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 40, right: 20, top: 20, bottom: 30 },
      xAxis: {
        type: 'category',
        data: hours.map(h => `${h.hour}:00`),
        axisLabel: { color: '#9ca3af', interval: 2 },
      },
      yAxis: {
        type: 'value',
        axisLabel: { color: '#9ca3af' },
        splitLine: { lineStyle: { color: '#374151' } },
      },
      series: [{
        type: 'bar',
        data: hours.map(h => h.count),
        itemStyle: { color: '#C9A84C', borderRadius: [4, 4, 0, 0] },
        barWidth: '60%',
      }],
    })
  }

  // Hot questions horizontal bar chart
  if (hotQuestionsChartRef.value && hotQuestionsChartRef.value.offsetHeight > 0 && data.value.hot_questions.length > 0) {
    hotQuestionsChart.value = echarts.init(hotQuestionsChartRef.value)
    const questions = [...data.value.hot_questions].reverse()
    hotQuestionsChart.value.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 180, right: 40, top: 10, bottom: 20 },
      xAxis: {
        type: 'value',
        axisLabel: { color: '#9ca3af' },
        splitLine: { lineStyle: { color: '#374151' } },
      },
      yAxis: {
        type: 'category',
        data: questions.map(q => q.question.length > 20 ? q.question.slice(0, 20) + '...' : q.question),
        axisLabel: { color: '#d1d5db', width: 160, overflow: 'truncate' },
      },
      series: [{
        type: 'bar',
        data: questions.map(q => q.count),
        itemStyle: { color: '#C23B22', borderRadius: [0, 4, 4, 0] },
        barWidth: '50%',
      }],
    })
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="max-w-[1200px] mx-auto px-12 py-12">
      <div class="flex items-start justify-between mb-8">
        <div>
          <h1 class="text-[32px] font-semibold text-cv-text tracking-[-0.5px]">
            {{ t('nav.dashboard') }}
          </h1>
          <p class="mt-2 text-sm text-cv-text-secondary">
            核心运营指标概览 / Operational metrics overview
          </p>
        </div>
        <button @click="fetchData"
                class="px-4 py-2 text-sm text-cv-text-secondary hover:text-cv-text border border-cv-border rounded-cv-md cursor-pointer transition-colors">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-if="loading" class="text-center py-20 text-cv-text-muted">
        {{ t('common.loading') }}
      </div>

      <template v-else-if="data">
        <!-- Today's sessions card -->
        <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-6 mb-6">
          <div class="text-sm text-cv-text-secondary mb-1">今日服务人次</div>
          <div class="text-5xl font-bold text-cv-text">{{ data.today_sessions }}</div>
        </div>

        <!-- Charts grid -->
        <div class="grid grid-cols-2 gap-6">
          <!-- Week trend -->
          <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-4">
            <h3 class="text-sm font-medium text-cv-text-secondary mb-3">本周服务趋势</h3>
            <div ref="weekChartRef" class="w-full h-[260px]" />
          </div>

          <!-- Sentiment -->
          <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-4">
            <h3 class="text-sm font-medium text-cv-text-secondary mb-3">情感分布</h3>
            <div ref="sentimentChartRef" class="w-full h-[260px]" />
          </div>

          <!-- Hourly distribution -->
          <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-4">
            <h3 class="text-sm font-medium text-cv-text-secondary mb-3">活跃时段分布</h3>
            <div ref="hourlyChartRef" class="w-full h-[260px]" />
          </div>

          <!-- Hot questions -->
          <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg p-4">
            <h3 class="text-sm font-medium text-cv-text-secondary mb-3">热门问答 TOP 10</h3>
            <div v-if="data.hot_questions.length === 0" class="flex items-center justify-center h-[260px] text-cv-text-muted text-sm">
              暂无数据
            </div>
            <div v-else ref="hotQuestionsChartRef" class="w-full h-[260px]" />
          </div>
        </div>
      </template>
  </div>
</template>
