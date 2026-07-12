<script setup lang="ts">
import { ref, computed } from 'vue'

interface ScenicStep {
  attraction_name: string
  duration_minutes: number
  highlight: string
}

interface ScenicCardStructured {
  type: 'route' | 'attraction'
  id: string
  name: string
  duration_minutes?: number
  difficulty?: string
  description?: string
  category?: string
  location?: string
  tags?: string[]
  steps?: ScenicStep[]
}

interface ScenicCardData {
  type: 'route' | 'attraction' | 'mixed'
  title: string
  summary: string
  details: string
  structured?: ScenicCardStructured
}

const props = defineProps<{ data: ScenicCardData }>()
const expanded = ref(false)

function toggle() {
  expanded.value = !expanded.value
}

function formatDuration(minutes: number): string {
  if (!minutes) return ''
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  if (h > 0 && m > 0) return `${h}h${m}m`
  if (h > 0) return `${h}h`
  return `${m}m`
}

const steps = computed(() => props.data.structured?.steps || [])
const hasStructured = computed(() => !!props.data.structured && steps.value.length > 0)
</script>

<template>
  <div class="scenic-card" :class="[`scenic-card--${data.type}`]" @click="toggle">
    <!-- Header -->
    <div class="scenic-card__header">
      <div class="scenic-card__icon">
        <!-- Route icon -->
        <svg v-if="data.type === 'route'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M9 6.75V15m6-6v8.25m.503 3.498l4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 00-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0z" />
        </svg>
        <!-- Attraction icon -->
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
          <path d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
        </svg>
      </div>
      <div class="scenic-card__title-area">
        <span class="scenic-card__type-label">{{ data.type === 'route' ? '推荐路线' : '景点' }}</span>
        <h4 class="scenic-card__title">{{ data.title }}</h4>
      </div>
      <svg class="scenic-card__chevron" :class="{ 'scenic-card__chevron--open': expanded }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M19 9l-7 7-7-7" />
      </svg>
    </div>

    <!-- Summary -->
    <div class="scenic-card__summary">{{ data.summary }}</div>

    <!-- Structured steps (when available) -->
    <div v-if="expanded && hasStructured" class="scenic-card__steps" @click.stop>
      <!-- Route metadata -->
      <div v-if="data.structured?.duration_minutes || data.structured?.difficulty" class="scenic-card__meta">
        <span v-if="data.structured?.duration_minutes" class="meta-tag">
          <svg class="meta-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="8" cy="8" r="6.5" /><path d="M8 4.5V8l2.5 1.5" stroke-linecap="round" />
          </svg>
          {{ formatDuration(data.structured!.duration_minutes!) }}
        </span>
        <span v-if="data.structured?.difficulty" class="meta-tag meta-tag--difficulty">
          {{ data.structured!.difficulty }}
        </span>
      </div>

      <!-- Step timeline -->
      <div class="step-timeline">
        <div v-for="(step, i) in steps" :key="i" class="step-item">
          <!-- Timeline connector -->
          <div class="step-line">
            <div class="step-dot" />
            <div v-if="i < steps.length - 1" class="step-connector" />
          </div>
          <!-- Step content -->
          <div class="step-content">
            <div class="step-header">
              <span class="step-name">{{ step.attraction_name }}</span>
              <span v-if="step.duration_minutes" class="step-duration">{{ formatDuration(step.duration_minutes) }}</span>
            </div>
            <p v-if="step.highlight" class="step-highlight">{{ step.highlight }}</p>
          </div>
        </div>
      </div>

      <!-- Description -->
      <div v-if="data.structured?.description" class="scenic-card__description">
        {{ data.structured!.description }}
      </div>
    </div>

    <!-- Fallback: full text details (no structured data) -->
    <div v-else-if="expanded && !hasStructured" class="scenic-card__details" @click.stop>
      {{ data.details }}
    </div>

    <!-- Attraction tags -->
    <div v-if="expanded && data.structured?.tags && data.structured!.tags!.length > 0" class="scenic-card__tags" @click.stop>
      <span v-for="tag in data.structured!.tags!" :key="tag" class="tag-chip">{{ tag }}</span>
    </div>

    <!-- Hint -->
    <div class="scenic-card__hint">{{ expanded ? '点击收起' : '点击查看详情' }}</div>
  </div>
</template>

<style scoped>
.scenic-card {
  margin-top: 8px;
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
  transition: box-shadow 200ms;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.scenic-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.2);
}

.scenic-card--route {
  border-left: 3px solid #22c55e;
}

.scenic-card--attraction {
  border-left: 3px solid #3b82f6;
}

.scenic-card--mixed {
  border-left: 3px solid #f59e0b;
}

/* Header */
.scenic-card__header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px 0;
}

.scenic-card__icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.08);
}

.scenic-card__icon svg {
  width: 18px;
  height: 18px;
}

.scenic-card--route .scenic-card__icon {
  color: #22c55e;
}

.scenic-card--attraction .scenic-card__icon {
  color: #3b82f6;
}

.scenic-card__title-area {
  flex: 1;
  min-width: 0;
}

.scenic-card__type-label {
  display: inline-block;
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 2px;
}

.scenic-card--route .scenic-card__type-label {
  background: rgba(34, 197, 94, 0.15);
  color: #4ade80;
}

.scenic-card--attraction .scenic-card__type-label {
  background: rgba(59, 130, 246, 0.15);
  color: #60a5fa;
}

.scenic-card__title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #eee;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.scenic-card__chevron {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  color: #666;
  transition: transform 200ms;
}

.scenic-card__chevron--open {
  transform: rotate(180deg);
}

/* Summary */
.scenic-card__summary {
  padding: 6px 12px 0;
  font-size: 13px;
  color: #aaa;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

/* Structured steps area */
.scenic-card__steps {
  padding: 8px 12px 0;
}

.scenic-card__meta {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.meta-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  background: rgba(34, 197, 94, 0.1);
  color: #4ade80;
}

.meta-tag--difficulty {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
}

.meta-icon {
  width: 14px;
  height: 14px;
}

/* Step timeline */
.step-timeline {
  display: flex;
  flex-direction: column;
}

.step-item {
  display: flex;
  gap: 12px;
  min-height: 48px;
}

.step-line {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex-shrink: 0;
  width: 20px;
}

.step-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #22c55e;
  border: 2px solid rgba(34, 197, 94, 0.3);
  flex-shrink: 0;
  margin-top: 4px;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.3);
}

.step-item:last-child .step-dot {
  background: #f59e0b;
  border-color: rgba(245, 158, 11, 0.3);
  box-shadow: 0 0 6px rgba(245, 158, 11, 0.3);
}

.step-connector {
  width: 2px;
  flex: 1;
  background: linear-gradient(180deg, rgba(34, 197, 94, 0.4), rgba(34, 197, 94, 0.1));
  margin: 2px 0;
  min-height: 16px;
}

.step-content {
  flex: 1;
  min-width: 0;
  padding-bottom: 8px;
}

.step-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.step-name {
  font-size: 13px;
  font-weight: 600;
  color: #ddd;
}

.step-duration {
  font-size: 11px;
  color: #888;
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 4px;
  flex-shrink: 0;
}

.step-highlight {
  margin: 2px 0 0;
  font-size: 12px;
  color: #999;
  line-height: 1.4;
}

/* Description */
.scenic-card__description {
  padding: 8px 12px;
  font-size: 12px;
  color: #999;
  line-height: 1.5;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  margin-top: 4px;
}

/* Tags */
.scenic-card__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 6px 12px 0;
}

.tag-chip {
  padding: 1px 8px;
  border-radius: 999px;
  font-size: 11px;
  background: rgba(255, 255, 255, 0.06);
  color: #aaa;
}

/* Fallback details */
.scenic-card__details {
  padding: 8px 12px 0;
  font-size: 13px;
  color: #bbb;
  line-height: 1.6;
  white-space: pre-wrap;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  margin-top: 8px;
}

/* Hint */
.scenic-card__hint {
  padding: 6px 12px 8px;
  font-size: 11px;
  color: #555;
  text-align: right;
}
</style>
