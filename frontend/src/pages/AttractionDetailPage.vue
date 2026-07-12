<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getAttraction } from '../services/api'
import type { Attraction } from '../services/api'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const attraction = ref<Attraction | null>(null)
const loading = ref(true)
const error = ref('')

onMounted(async () => {
  const id = route.params.id as string
  try {
    attraction.value = await getAttraction(id)
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('scenic.detail.loadError')
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
    <div v-else-if="attraction" class="detail-content">
      <!-- Full-page background image -->
      <div v-if="attraction.image_url" class="bg-backdrop">
        <img :src="attraction.image_url" :alt="attraction.name">
        <div class="bg-overlay" />
      </div>

      <!-- Back button -->
      <button class="back-floating" :title="t('common.back')" @click="goBack">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15.75 19.5L8.25 12l7.5-7.5" />
        </svg>
      </button>

      <!-- Info card -->
      <div class="detail-card">
        <div class="detail-header">
          <div class="detail-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
              <path d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
            </svg>
          </div>
          <div class="detail-title-group">
            <h1>{{ attraction.name }}</h1>
            <span v-if="attraction.category" class="detail-category">{{ attraction.category }}</span>
          </div>
        </div>

        <!-- Location -->
        <div v-if="attraction.location" class="detail-row">
          <svg class="row-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
            <path d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
          </svg>
          <span>{{ attraction.location }}</span>
        </div>

        <!-- Opening hours -->
        <div v-if="attraction.opening_hours" class="detail-row">
          <svg class="row-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ t('scenic.detail.openingHours') }}: {{ attraction.opening_hours }}</span>
        </div>

        <!-- Ticket info -->
        <div v-if="attraction.ticket_info" class="detail-row">
          <svg class="row-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M16.5 6v.75m0 3v.75m0 3v.75m0 3V18m-9-5.25h5.25M7.5 15h3M3.375 5.25c-.621 0-1.125.504-1.125 1.125v3.026a2.999 2.999 0 010 5.198v3.026c0 .621.504 1.125 1.125 1.125h17.25c.621 0 1.125-.504 1.125-1.125v-3.026a2.999 2.999 0 010-5.198V6.375c0-.621-.504-1.125-1.125-1.125H3.375z" />
          </svg>
          <span>{{ t('scenic.detail.ticketInfo') }}: {{ attraction.ticket_info }}</span>
        </div>

        <!-- Tags -->
        <div v-if="attraction.tags && attraction.tags.length > 0" class="detail-tags">
          <span v-for="tag in attraction.tags" :key="tag" class="tag">{{ tag }}</span>
        </div>

        <!-- Description -->
        <div v-if="attraction.description" class="detail-desc">
          <p>{{ attraction.description }}</p>
        </div>

        <!-- Action: start chat about this attraction -->
        <div class="detail-actions">
          <button class="chat-btn" @click="router.push('/scenic')">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M20.25 8.511c.884.284 1.5 1.128 1.5 2.097v4.286c0 1.136-.847 2.1-1.98 2.193-.34.027-.68.052-1.02.072v3.091l-3-3c-1.354 0-2.694-.055-4.02-.163a2.115 2.115 0 01-.825-.242m9.345-8.334a2.126 2.126 0 00-.476-.095 48.64 48.64 0 00-8.048 0c-1.131.094-1.976 1.057-1.976 2.192v4.286c0 .837.46 1.58 1.155 1.951m9.345-8.334V6.637c0-1.621-1.152-3.026-2.76-3.235A48.455 48.455 0 0011.25 3c-2.115 0-4.198.137-6.24.402-1.608.209-2.76 1.614-2.76 3.235v6.226c0 1.621 1.152 3.026 2.76 3.235.577.075 1.157.14 1.74.194V21l4.155-4.155" />
            </svg>
            {{ t('scenic.detail.startChatAbout') }}
          </button>
        </div>
      </div>
    </div>
  </main>
</template>

<style scoped>
.detail-page {
  min-height: 100vh;
  background: #f9fafb;
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

/* Full-page background backdrop */
.bg-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
}

.bg-backdrop img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.bg-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, rgba(255,255,255,0) 0%, rgba(255,255,255,0.6) 70%, rgba(249,250,251,1) 100%);
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

/* Detail content layout */
.detail-content {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80px 24px;
}

/* Detail card */
.detail-card {
  width: 100%;
  max-width: 720px;
  padding: 36px 28px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 20px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  position: relative;
  z-index: 1;
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
  background: rgba(219, 234, 254, 0.8);
  color: #2563eb;
  border-radius: 14px;
}

.detail-icon svg {
  width: 24px;
  height: 24px;
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

.detail-category {
  display: inline-block;
  margin-top: 6px;
  padding: 3px 12px;
  border-radius: 999px;
  background: rgba(240, 253, 244, 0.8);
  color: #16a34a;
  font-size: 13px;
  font-weight: 600;
}

/* Location row */
.detail-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 20px;
  font-size: 14px;
  color: #6b7280;
}

.row-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
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
  background: rgba(243, 244, 246, 0.8);
  color: #6b7280;
  font-size: 13px;
  font-weight: 500;
}

/* Description */
.detail-desc {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid rgba(243, 244, 246, 0.8);
}

.detail-desc p {
  margin: 0;
  font-size: 15px;
  color: #374151;
  line-height: 26px;
  white-space: pre-wrap;
}

/* Actions */
.detail-actions {
  margin-top: 28px;
  display: flex;
  gap: 12px;
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
    margin: 0 16px;
    padding: 24px 20px;
  }

  .detail-title-group h1 {
    font-size: 22px;
  }
}
</style>
