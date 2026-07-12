<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import LanguageSwitcher from '../components/LanguageSwitcher.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const sidebarCollapsed = ref(false)

interface NavItem {
  path: string
  icon: string
  labelKey: string
}

const navItems: NavItem[] = [
  { path: '/admin/dashboard', icon: 'chart', labelKey: 'nav.dashboard' },
  { path: '/admin/characters', icon: 'users', labelKey: 'nav.characters' },
  { path: '/admin/attractions', icon: 'map-pin', labelKey: 'nav.attractions' },
  { path: '/admin/routes', icon: 'route', labelKey: 'nav.routes' },
  { path: '/admin/sessions', icon: 'message', labelKey: 'nav.sessions' },
  { path: '/admin/reports', icon: 'file-text', labelKey: 'nav.reports' },
  { path: '/admin/settings', icon: 'settings', labelKey: 'nav.settings' },
]

const activePath = computed(() => {
  const p = route.path
  const match = navItems.find(n => p === n.path || p.startsWith(n.path + '/'))
  return match?.path || ''
})

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}
</script>

<template>
  <div class="admin-layout" :class="{ collapsed: sidebarCollapsed }">
    <!-- Sidebar -->
    <aside class="admin-sidebar">
      <!-- Decorative top border -->
      <div class="sidebar-deco-top" />

      <div class="sidebar-header">
        <div v-if="!sidebarCollapsed" class="sidebar-brand">
          <span class="brand-icon">泉</span>
          <div class="brand-text">
            <span class="brand-name">{{ t('scenic.title') }}</span>
            <span class="brand-sub">{{ t('nav.adminPortal') }}</span>
          </div>
        </div>
        <button class="collapse-btn" @click="toggleSidebar" :title="sidebarCollapsed ? '展开' : '收起'">
          <svg :class="{ rotated: sidebarCollapsed }" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.8">
            <path d="M10 3 5 8l5 5" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </button>
      </div>

      <!-- Circuit line decoration -->
      <div class="sidebar-circuit" />

      <nav class="sidebar-nav">
        <router-link
          v-for="item in navItems"
          :key="item.path"
          :to="item.path"
          class="nav-item"
          :class="{ active: activePath === item.path }"
        >
          <!-- Chart icon -->
          <svg v-if="item.icon === 'chart'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M3 13V7m4 6V5m4 12V9m4 4V3" stroke-linecap="round" />
          </svg>
          <!-- Users icon -->
          <svg v-else-if="item.icon === 'users'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="7" cy="6" r="3" /><path d="M1 17v-1a4 4 0 0 1 4-4h4a4 4 0 0 1 4 4v1" /><circle cx="15" cy="7" r="2.5" /><path d="M15 12a3.5 3.5 0 0 1 3.5 3.5V17" />
          </svg>
          <!-- Map pin icon -->
          <svg v-else-if="item.icon === 'map-pin'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M10 11a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5Z" /><path d="M10 18s-6-4.35-6-9a6 6 0 1 1 12 0c0 4.65-6 9-6 9Z" />
          </svg>
          <!-- Route icon -->
          <svg v-else-if="item.icon === 'route'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="5" cy="5" r="2" /><circle cx="15" cy="15" r="2" /><path d="M7 5h4a2 2 0 0 1 2 2v4a2 2 0 0 0 2 2h2" />
          </svg>
          <!-- Message icon -->
          <svg v-else-if="item.icon === 'message'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M3 4h14a1 1 0 0 1 1 1v8a1 1 0 0 1-1 1H6l-3 3V5a1 1 0 0 1 1-1Z" />
          </svg>
          <!-- File text icon -->
          <svg v-else-if="item.icon === 'file-text'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M5 3h7l4 4v10a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V4a1 1 0 0 1 1-1Z" /><path d="M7 8h6M7 11h4" stroke-linecap="round" />
          </svg>
          <!-- Settings icon -->
          <svg v-else-if="item.icon === 'settings'" class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="10" cy="10" r="3" /><path d="M10 1v2m0 14v2M1 10h2m14 0h2m-3.1-6.9-1.4 1.4M5.5 14.5l-1.4 1.4m0-12.8 1.4 1.4m9 9 1.4 1.4" stroke-linecap="round" />
          </svg>
          <span v-if="!sidebarCollapsed" class="nav-label">{{ t(item.labelKey) }}</span>
          <!-- Active indicator diamond -->
          <span v-if="activePath === item.path && !sidebarCollapsed" class="active-diamond">◆</span>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <router-link to="/scenic" class="footer-link" :title="t('nav.backToScenic')">
          <svg class="nav-icon" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M13 4H6a1 1 0 0 0-1 1v10m0 0 3-3m-3 3 3 3" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
          <span v-if="!sidebarCollapsed" class="nav-label">{{ t('nav.backToScenic') }}</span>
        </router-link>
      </div>

      <!-- Decorative bottom border -->
      <div class="sidebar-deco-bottom" />
    </aside>

    <!-- Main content -->
    <main class="admin-main">
      <header class="admin-header">
        <div class="header-left">
          <span class="header-deco-left">〔</span>
          <span class="header-title">{{ t('nav.adminPortal') }}</span>
          <span class="header-deco-right">〕</span>
          <div class="header-line" />
        </div>
        <div class="header-right">
          <LanguageSwitcher />
        </div>
      </header>
      <div class="admin-content">
        <router-view />
      </div>
    </main>
  </div>
</template>

<style scoped>
.admin-layout {
  --sidebar-width: 220px;
  --sidebar-collapsed-width: 60px;
  --header-height: 56px;
  --sidebar-bg: #1C1C1C;
  --sidebar-active: #252525;
  --sidebar-hover: #222222;
  --sidebar-border: #333333;
  --sidebar-text: #9E9E9E;
  --sidebar-text-active: #F5F0E8;
  --sidebar-accent: #6BAF8D;
  --sidebar-gold: #C9A84C;

  display: flex;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

/* Sidebar */
.admin-sidebar {
  flex: 0 0 var(--sidebar-width);
  width: var(--sidebar-width);
  display: flex;
  flex-direction: column;
  background: var(--sidebar-bg);
  border-right: 1px solid var(--sidebar-border);
  transition: width 200ms ease, flex-basis 200ms ease;
  overflow: hidden;
  position: relative;
}

/* Decorative gold borders */
.sidebar-deco-top,
.sidebar-deco-bottom {
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--sidebar-gold), transparent);
  flex-shrink: 0;
}

.admin-layout.collapsed .admin-sidebar {
  flex-basis: var(--sidebar-collapsed-width);
  width: var(--sidebar-collapsed-width);
}

.sidebar-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px 0 16px;
  flex-shrink: 0;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  overflow: hidden;
}

.brand-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--sidebar-gold), #A8893A);
  color: #1C1C1C;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 16px;
  font-weight: 700;
  border-radius: 6px;
  flex-shrink: 0;
}

.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
  overflow: hidden;
}

.brand-name {
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 14px;
  font-weight: 700;
  color: var(--sidebar-text-active);
  letter-spacing: 2px;
  white-space: nowrap;
}

.brand-sub {
  font-size: 10px;
  color: var(--sidebar-text);
  letter-spacing: 1px;
  white-space: nowrap;
}

.collapse-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  color: var(--sidebar-text);
  cursor: pointer;
  border-radius: 6px;
  flex-shrink: 0;
  transition: background 160ms, color 160ms;
}

.collapse-btn:hover {
  background: var(--sidebar-hover);
  color: var(--sidebar-text-active);
}

.collapse-btn svg {
  width: 14px;
  height: 14px;
  transition: transform 200ms ease;
}

.collapse-btn svg.rotated {
  transform: rotate(180deg);
}

/* Circuit line decoration */
.sidebar-circuit {
  height: 1px;
  margin: 0 16px;
  background: linear-gradient(90deg,
    transparent 0%,
    var(--sidebar-accent) 20%,
    var(--sidebar-accent) 40%,
    transparent 45%,
    transparent 55%,
    var(--sidebar-gold) 60%,
    var(--sidebar-gold) 80%,
    transparent 100%
  );
  opacity: 0.3;
  flex-shrink: 0;
  animation: circuit-pulse 4s ease-in-out infinite;
}

@keyframes circuit-pulse {
  0%, 100% { opacity: 0.2; }
  50% { opacity: 0.5; }
}

/* Navigation */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  color: var(--sidebar-text);
  text-decoration: none;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 13px;
  font-weight: 500;
  transition: background 160ms, color 160ms;
  white-space: nowrap;
  overflow: hidden;
  position: relative;
}

.nav-item:hover {
  background: var(--sidebar-hover);
  color: var(--sidebar-text-active);
}

.nav-item.active {
  background: var(--sidebar-active);
  color: var(--sidebar-text-active);
  border-left: 3px solid var(--sidebar-accent);
  padding-left: 9px;
}

.nav-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  color: inherit;
}

.nav-item.active .nav-icon {
  color: var(--sidebar-accent);
}

.nav-label {
  overflow: hidden;
  text-overflow: ellipsis;
}

.active-diamond {
  margin-left: auto;
  font-size: 8px;
  color: var(--sidebar-accent);
  animation: diamond-glow 2s ease-in-out infinite;
}

@keyframes diamond-glow {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}

/* Footer */
.sidebar-footer {
  padding: 8px;
  border-top: 1px solid var(--sidebar-border);
  flex-shrink: 0;
}

.footer-link {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  color: var(--sidebar-text);
  text-decoration: none;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 13px;
  font-weight: 500;
  transition: background 160ms, color 160ms;
  white-space: nowrap;
  overflow: hidden;
}

.footer-link:hover {
  background: var(--sidebar-hover);
  color: var(--sidebar-text-active);
}

/* Main content */
.admin-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: #181818;
}

.admin-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  flex-shrink: 0;
  background: linear-gradient(180deg, #202020, #1C1C1C);
  border-bottom: 1px solid #2A2A2A;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 4px;
}

.header-deco-left,
.header-deco-right {
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 18px;
  color: var(--sidebar-gold);
  font-weight: 300;
}

.header-title {
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
  font-size: 15px;
  font-weight: 600;
  color: var(--sidebar-text-active);
  letter-spacing: 2px;
}

.header-line {
  width: 60px;
  height: 1px;
  background: linear-gradient(90deg, var(--sidebar-gold), transparent);
  margin-left: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin-content {
  flex: 1;
  overflow-y: auto;
}

/* Responsive */
@media (max-width: 768px) {
  .admin-layout {
    --sidebar-width: 200px;
    --sidebar-collapsed-width: 0px;
  }

  .admin-layout.collapsed .admin-sidebar {
    display: none;
  }

  .admin-header {
    padding: 0 16px;
  }
}
</style>
