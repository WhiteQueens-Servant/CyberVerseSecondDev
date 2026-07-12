<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getRoute } from '../services/api'
import type { ScenicRoute } from '../services/api'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const scenicRoute = ref<ScenicRoute | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const id = route.params.id as string
  try {
    scenicRoute.value = await getRoute(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('scenic.routeDetail.loadError')
  } finally {
    loading.value = false
  }
})

function goBack() {
  router.push('/scenic')
}
</script>

<template>
  <main class="detail-page">
    <!-- Loading state -->
    <div v-if="loading" class="status-state">
      <div class="spinner" />
      <p>{{ t('common.loading') }}</p>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="status-state">
      <div class="error-icon">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
        </svg>
      </div>
      <p class="error-text">{{ error }}</p>
      <button class="back-btn" @click="goBack">{{ t('common.back') }}</button>
    </div>

    <!-- Detail content -->
    <div v-else-if="scenicRoute" class="detail-content">
      <!-- Back button -->
      <button class="back-floating" :title="t('common.back')" @click="goBack">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15.75 19.5L8.25 12l7.5-7.5" />
        </svg>
      </button>

      <!-- Header card -->
      <div class="detail-card">
        <div class="detail-header">
          <div class="detail-icon route-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M9 6.75V15m6-6v8.25m.503 3.498l4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 00-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0z" />
            </svg>
          </div>
          <div class="detail-title-group">
            <h1>{{ scenicRoute.name }}</h1>
            <div class="route-meta-inline">
              <span v-if="scenicRoute.duration" class="meta-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                {{ t('scenic.duration') }}: {{ scenicRoute.duration }}
              </span>
              <span v-if="scenicRoute.difficulty" class="meta-item">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M3 3v1.5M3 21v-6m0 0l2.77-.693a9 9 0 016.208.682l.108.054a9 9 0 006.086.71l3.114-.732a48.524 48.524 0 01-.005-10.499l-3.11.732a9 9 0 01-6.085-.711l-.108-.054a9 9 0 00-6.208-.682L3 4.5M3 15V4.5" />
                </svg>
                {{ t('scenic.difficulty') }}: {{ scenicRoute.difficulty }}
              </span>
            </div>
          </div>
        </div>

        <!-- Tags -->
        <div v-if="scenicRoute.tags && scenicRoute.tags.length > 0" class="detail-tags">
          <span v-for="tag in scenicRoute.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <!-- Description -->
        <div v-if="scenicRoute.description" class="detail-desc">
          <p>{{ scenicRoute.description }}</p>
        </div>
      </div>

      <!-- Steps timeline -->
      <div v-if="scenicRoute.steps && scenicRoute.steps.length > 0" class="steps-section">
        <h2>{{ t('scenic.routeDetail.routeSteps') }}</h2>
        <div class="timeline">
          <div v-for="(step, idx) in scenicRoute.steps" :key="idx" class="timeline-item">
            <div class="timeline-marker">
              <span class="step-number">{{ step.order || idx + 1 }}</span>
            </div>
            <div class="timeline-content">
              <div class="step-header">
                <h3 v-if="step.attraction_name">{{ step.attraction_name }}</h3>
                <h3 v-else class="step-unknown">{{ t('scenic.routeDetail.unknownAttraction') }}</h3>
                <span v-if="step.duration_minutes" class="step-duration">{{ step.duration_minutes }}{{ t('scenic.routeDetail.minutes') }}</span>
              </div>
              <p v-if="step.highlight" class="step-highlight">{{ step.highlight }}</p>
              <button
                v-if="step.attraction_id"
                class="step-link-btn"
                @click="router.push(`/scenic/attraction/${step.attraction_id}`)"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M13.5 6H5.25A2.25 2.25 0 003 8.25v10.5A2.25 2.25 0 005.25 21h10.5A2.25 2.25 0 0018 18.75V10.5m-10.5 6L21 3m0 0h-5.25M21 3v5.25" />
                </svg>
                {{ t('scenic.routeDetail.viewAttraction') }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Action -->
      <div class="detail-actions">
        <button class="chat-btn" @click="router.push('/scenic')">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155" />
          </svg>
          {{ t('scenic.detail.startChatAbout') }}
        </button>
      </div>
    </div>
  </main>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #f0fdf4 0%, #ffffff 40%);
}

/* Status states */
.status-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
  gap: 16px;
  color: #6b7280;
  font-size: 15px;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e5e7eb;
  border-top-color: #16a34a;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.error-icon {
  width: 48px;
  height: 48px;
  color: #ef4444;
}

.error-icon svg {
  width: 100%;
  height: 100%;
}

.error-text {
  color: #b91c1c;
  font-size: 14px;
}

/* Floating back button */
.back-floating {
  position: fixed;
  top: 20px;
  left: 20px;
  z-index: 10;
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(8px);
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: background 150ms, box-shadow 150ms;
}

.back-floating:hover {
  background: #ffffff;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.back-floating svg {
  width: 20px;
  height: 20px;
  color: #374151;
}

/* Detail card */
.detail-card {
  max-width: 720px;
  margin: 20px auto 0;
  padding: 28px;
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 20px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.06);
}

.detail-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
}

.detail-icon {
  width: 48px;
  height: 48px;
  flex-shrink: 0;
  display: grid;
  place-items: center;
  border-radius: 14px;
}

.detail-icon svg {
  width: 24px;
  height: 24px;
}

.route-icon {
  background: #f0fdf4;
  color: #16a34a;
}

.detail-title-group {
  flex: 1;
  min-width: 0;
}

.detail-title-group h1 {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  color: #1a1a1a;
  line-height: 1.3;
}

.route-meta-inline {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-top: 8px;
}

.meta-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #6b7280;
}

.meta-item svg {
  width: 16px;
  height: 16px;
  color: #9ca3af;
}

/* Tags */
.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 16px;
}

.tag {
  padding: 4px 12px;
  border-radius: 999px;
  background: #f3f4f6;
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
}

/* Description */
.detail-desc {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #f3f4f6;
}

.detail-desc p {
  margin: 0;
  font-size: 15px;
  color: #374151;
  line-height: 26px;
  white-space: pre-wrap;
}

/* Steps section */
.steps-section {
  max-width: 720px;
  margin: 32px auto 0;
  padding: 0 28px;
}

.steps-section h2 {
  margin: 0 0 24px;
  font-size: 22px;
  font-weight: 700;
  color: #1a1a1a;
}

.timeline {
  position: relative;
  padding-left: 32px;
}

.timeline::before {
  content: '';
  position: absolute;
  left: 15px;
  top: 8px;
  bottom: 8px;
  width: 2px;
  background: #e5e7eb;
}

.timeline-item {
  position: relative;
  padding-bottom: 28px;
}

.timeline-item:last-child {
  padding-bottom: 0;
}

.timeline-marker {
  position: absolute;
  left: -32px;
  top: 0;
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
}

.step-number {
  width: 32px;
  height: 32px;
  display: grid;
  place-items: center;
  background: #16a34a;
  color: #ffffff;
  border-radius: 50%;
  font-size: 13px;
  font-weight: 700;
}

.timeline-content {
  background: #ffffff;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 16px 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.step-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.step-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: #1a1a1a;
}

.step-unknown {
  color: #9ca3af !important;
  font-style: italic;
}

.step-duration {
  flex-shrink: 0;
  font-size: 13px;
  color: #6b7280;
  background: #f3f4f6;
  padding: 2px 10px;
  border-radius: 999px;
}

.step-highlight {
  margin: 8px 0 0;
  font-size: 14px;
  color: #6b7280;
  line-height: 22px;
}

.step-link-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-top: 10px;
  padding: 6px 14px;
  background: transparent;
  color: #2563eb;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 150ms, border-color 150ms;
}

.step-link-btn:hover {
  background: #eff6ff;
  border-color: #93c5fd;
}

.step-link-btn svg {
  width: 14px;
  height: 14px;
}

/* Actions */
.detail-actions {
  max-width: 720px;
  margin: 36px auto 60px;
  padding: 0 28px;
}

.chat-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 24px;
  background: #16a34a;
  color: #ffffff;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  box-shadow: 0 4px 12px rgba(22, 163, 74, 0.2);
  transition: background 150ms, transform 150ms, box-shadow 150ms;
}

.chat-btn:hover {
  background: #22c55e;
  transform: translateY(-1px);
  box-shadow: 0 6px 16px rgba(22, 163, 74, 0.25);
}

.chat-btn svg {
  width: 20px;
  height: 20px;
}

/* Responsive */
@media (max-width: 768px) {
  .detail-card {
    margin: 20px 16px 0;
    padding: 20px;
  }

  .detail-title-group h1 {
    font-size: 22px;
  }

  .steps-section {
    padding: 0 16px;
  }

  .detail-actions {
    padding: 0 16px;
  }
}
</style>
