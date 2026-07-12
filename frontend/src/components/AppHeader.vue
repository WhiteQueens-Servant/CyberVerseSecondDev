<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { getHealth } from '../services/api'
import LanguageSwitcher from './LanguageSwitcher.vue'

const router = useRouter()
const { t } = useI18n()
const search = ref('')
const serviceConnected = ref(false)

onMounted(async () => {
  try {
    const h = await getHealth()
    serviceConnected.value = h.inference_connected
  } catch {
    serviceConnected.value = false
  }
})

withDefaults(defineProps<{
  showBack?: boolean
  breadcrumb?: string[]
  logoTo?: string
}>(), {
  logoTo: '/admin',
})
</script>

<template>
  <header class="app-header">
    <!-- Left -->
    <div class="flex items-center gap-3">
      <button v-if="showBack" @click="router.back()"
              class="text-cv-text-secondary hover:text-cv-text text-sm cursor-pointer transition-colors">
        {{ t('common.back') }}
      </button>
      <span v-if="showBack" class="text-cv-border">|</span>
      <span class="header-brand" @click="router.push(logoTo)">
        <span class="brand-icon-sm">泉</span>
        {{ t('scenic.title') }}
      </span>
      <LanguageSwitcher />
    </div>

    <!-- Center: Search -->
    <div class="flex-1 flex justify-center" v-if="!breadcrumb">
      <div class="relative w-[280px]">
        <input
          v-model="search"
          type="text"
          :placeholder="t('common.searchCharacters')"
          class="search-input"
        />
      </div>
    </div>

    <!-- Center: Breadcrumb -->
    <div class="flex-1 flex justify-center" v-if="breadcrumb">
      <div class="flex items-center gap-2 text-sm">
        <template v-for="(item, i) in breadcrumb" :key="i">
          <span v-if="i < breadcrumb.length - 1"
                class="text-[#6BAF8D] cursor-pointer hover:text-[#8BCBA8] transition-colors"
                @click="router.push(i === 0 ? '/admin/characters' : '')">
            {{ item }}
          </span>
          <span v-if="i < breadcrumb.length - 1" class="text-cv-text-muted">/</span>
          <span v-if="i === breadcrumb.length - 1" class="text-cv-text-secondary">{{ item }}</span>
        </template>
      </div>
    </div>

    <!-- Right: Nav Links + Status + Settings -->
    <div class="flex items-center gap-4">
      <nav class="flex items-center gap-3 text-[13px]">
        <router-link to="/admin/attractions" class="nav-link-item">
          {{ t('nav.attractions') }}
        </router-link>
        <router-link to="/admin/routes" class="nav-link-item">
          {{ t('nav.routes') }}
        </router-link>
        <router-link to="/admin/dashboard" class="nav-link-item">
          {{ t('nav.dashboard') }}
        </router-link>
        <router-link to="/admin/sessions" class="nav-link-item">
          {{ t('nav.sessions') }}
        </router-link>
        <router-link to="/admin/reports" class="nav-link-item">
          {{ t('nav.reports') }}
        </router-link>
      </nav>
      <div class="flex items-center gap-2 text-[13px]">
        <span class="w-2 h-2 rounded-full" :class="serviceConnected ? 'bg-[#6BAF8D]' : 'bg-cv-danger'" />
        <span class="text-cv-text-secondary">{{ serviceConnected ? t('common.serviceConnected') : t('common.serviceDisconnected') }}</span>
      </div>
      <button type="button"
              :aria-label="t('appHeader.statusSettingsLabel')"
              @click="router.push('/admin/settings')"
              class="w-8 h-8 flex items-center justify-center rounded-cv-md text-cv-text-secondary hover:text-cv-text hover:bg-cv-hover transition-all cursor-pointer">
        <svg class="w-[18px] h-[18px] shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor"
             stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M12 20a8 8 0 1 0 0-16 8 8 0 0 0 0 16Z" />
          <path d="M12 14a2 2 0 1 0 0-4 2 2 0 0 0 0 4Z" />
          <path d="M12 2v2" />
          <path d="M12 22v-2" />
          <path d="m17 20.66-1-1.73" />
          <path d="M11 10.27 7 3.34" />
          <path d="m20.66 17-1.73-1" />
          <path d="m3.34 7 1.73 1" />
          <path d="M14 12h8" />
          <path d="M2 12h2" />
          <path d="m20.66 7-1.73 1" />
          <path d="m3.34 17 1.73-1" />
          <path d="m17 3.34-1 1.73" />
          <path d="m11 13.73-4 6.93" />
        </svg>
      </button>
    </div>
  </header>
</template>

<style scoped>
.app-header {
  height: 56px;
  background: linear-gradient(180deg, #202020, #1C1C1C);
  border-bottom: 1px solid #2A2A2A;
  display: flex;
  align-items: center;
  padding: 0 32px;
  flex-shrink: 0;
}

.header-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 16px;
  font-weight: 700;
  color: #F5F0E8;
  letter-spacing: 2px;
  cursor: pointer;
  transition: opacity 160ms;
}

.header-brand:hover {
  opacity: 0.85;
}

.brand-icon-sm {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #C9A84C, #A8893A);
  color: #1C1C1C;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 12px;
  font-weight: 700;
  border-radius: 4px;
}

.search-input {
  width: 100%;
  height: 36px;
  background: #252525;
  border: 1px solid #333333;
  border-radius: 8px;
  padding: 0 16px;
  font-size: 13px;
  color: #F5F0E8;
  outline: none;
  transition: border-color 200ms, box-shadow 200ms;
}

.search-input::placeholder {
  color: #666666;
}

.search-input:focus {
  border-color: #6BAF8D;
  box-shadow: 0 0 0 2px rgba(107, 175, 141, 0.15);
}

.nav-link-item {
  color: #9E9E9E;
  text-decoration: none;
  transition: color 160ms;
}

.nav-link-item:hover {
  color: #F5F0E8;
}
</style>
