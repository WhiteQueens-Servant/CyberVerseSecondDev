<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { createSession, getCharacter, getCharacters, listAttractions, listRoutes } from '../services/api'
import type { Attraction, ScenicRoute } from '../services/api'
import type { Character, PipelineMode } from '../types'
import { buildSessionLaunchState, saveSessionLaunchState } from '../utils/sessionLaunchState'

const router = useRouter()
const { t } = useI18n()

const SCENIC_RETURN_PATH = '/scenic'

const connecting = ref(false)
const launchError = ref('')
const guideCharacters = ref<Character[]>([])
const selectedCharacter = ref<Character | null>(null)
const attractions = ref<Attraction[]>([])
const routes = ref<ScenicRoute[]>([])
const loading = ref(true)

const categoryLabel: Record<string, string> = {
  guide: t('scenic.categoryGuide'),
  narrator: t('scenic.categoryNarrator'),
  service: t('scenic.categoryService'),
}

onMounted(async () => {
  try {
    const [chars, atts, rtes] = await Promise.allSettled([
      getCharacters(),
      listAttractions(),
      listRoutes(),
    ])

    if (chars.status === 'fulfilled') {
      const guides = chars.value.filter(c => c.scenic_category === 'guide')
      guideCharacters.value = guides.length > 0 ? guides : chars.value
      if (guideCharacters.value.length > 0) {
        selectedCharacter.value = guideCharacters.value[0]
      }
    }
    if (atts.status === 'fulfilled') {
      attractions.value = atts.value
    }
    if (rtes.status === 'fulfilled') {
      routes.value = rtes.value
    }
  } finally {
    loading.value = false
  }
})

const guideCategory = computed(() => {
  if (!selectedCharacter.value) return ''
  return categoryLabel[selectedCharacter.value.scenic_category || ''] || selectedCharacter.value.scenic_category || ''
})

async function startConversation() {
  if (!selectedCharacter.value || connecting.value) return

  connecting.value = true
  launchError.value = ''

  try {
    let launchMode: PipelineMode = 'omni'
    try {
      const char = await getCharacter(selectedCharacter.value.id)
      launchMode = char.mode || launchMode
    } catch {
      // Fallback to omni if character detail unavailable
    }

    const response = await createSession(selectedCharacter.value.id, launchMode)
    response.warnings?.forEach(w => console.warn('[ScenicLanding]', w))
    saveSessionLaunchState(buildSessionLaunchState(response, selectedCharacter.value.id, launchMode, SCENIC_RETURN_PATH))
    router.push(`/session/${response.session_id}`)
  } catch (error) {
    launchError.value = error instanceof Error ? error.message : t('scenic.errorConnect')
  } finally {
    connecting.value = false
  }
}
</script>

<template>
  <main class="scenic-page">
    <!-- Ink wash background decorations -->
    <div class="ink-bg">
      <div class="ink-circle ink-circle-1"></div>
      <div class="ink-circle ink-circle-2"></div>
      <div class="ink-circle ink-circle-3"></div>
      <div class="circuit-line circuit-h"></div>
      <div class="circuit-line circuit-v"></div>
    </div>

    <!-- Hero section -->
    <section class="hero">
      <div class="hero-content">
        <!-- Decorative top border -->
        <div class="deco-border-top">
          <span class="deco-dot"></span>
          <span class="deco-line"></span>
          <span class="deco-diamond">◆</span>
          <span class="deco-line"></span>
          <span class="deco-dot"></span>
        </div>

        <span class="hero-badge">{{ t('scenic.subtitle') }}</span>
        <h1 class="hero-title">
          <span class="title-char" v-for="(ch, i) in '泉城智游'" :key="i" :style="{ animationDelay: `${i * 0.15}s` }">{{ ch }}</span>
        </h1>
        <p class="hero-desc">{{ t('scenic.description') }}</p>

        <!-- Guide character cards -->
        <div v-if="guideCharacters.length > 0" class="guide-grid">
          <div
            v-for="char in guideCharacters"
            :key="char.id"
            class="guide-card"
            :class="{ 'guide-card-selected': selectedCharacter?.id === char.id }"
            @click="selectedCharacter = char"
          >
            <div class="guide-avatar">
              <img v-if="char.active_image" :src="`/api/v1/characters/${char.id}/images/${encodeURIComponent(char.active_image)}`" :alt="char.name">
              <img v-else-if="char.avatar_image" :src="char.avatar_image" :alt="char.name">
              <div v-else class="guide-avatar-placeholder">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                </svg>
              </div>
            </div>
            <div class="guide-info">
              <span v-if="categoryLabel[char.scenic_category || '']" class="guide-category">{{ categoryLabel[char.scenic_category || ''] }}</span>
              <h3>{{ char.name }}</h3>
              <p v-if="char.description">{{ char.description }}</p>
            </div>
          </div>
        </div>
        <div v-else-if="!loading" class="no-guide">
          {{ t('scenic.noGuide') }}
        </div>

        <button class="start-btn" type="button" :disabled="connecting || !selectedCharacter" @click="startConversation">
          <span class="btn-icon">⚔</span>
          {{ connecting ? t('scenic.connecting') : t('scenic.startChat') }}
        </button>
        <p v-if="launchError" class="launch-error" role="alert">{{ launchError }}</p>

        <!-- Decorative bottom border -->
        <div class="deco-border-bottom">
          <span class="deco-dot"></span>
          <span class="deco-line"></span>
          <span class="deco-diamond">◆</span>
          <span class="deco-line"></span>
          <span class="deco-dot"></span>
        </div>
      </div>
    </section>

    <!-- Attractions section -->
    <section v-if="attractions.length > 0" class="content-section">
      <div class="section-header">
        <span class="section-deco">〔</span>
        <h2>{{ t('scenic.attractionsTitle') }}</h2>
        <span class="section-deco">〕</span>
      </div>
      <div class="card-grid">
        <article v-for="att in attractions" :key="att.id" class="info-card" @click="router.push(`/scenic/attraction/${att.id}`)">
          <div v-if="att.image_url" class="card-image">
            <img :src="att.image_url" :alt="att.name">
            <div class="card-image-overlay"></div>
          </div>
          <div v-else class="card-image card-image-placeholder">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
              <path d="M2.25 15.75l5.159-5.159a2.25 2.25 0 013.182 0l5.159 5.159m-1.5-1.5l1.409-1.409a2.25 2.25 0 013.182 0l2.909 2.909M3.75 21h16.5A2.25 2.25 0 0022.5 18.75V5.25A2.25 2.25 0 0020.25 3H3.75A2.25 2.25 0 001.5 5.25v13.5A2.25 2.25 0 003.75 21z" />
            </svg>
          </div>
          <div class="card-body">
            <div class="card-icon attraction-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M15 10.5a3 3 0 11-6 0 3 3 0 016 0z" />
                <path d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1115 0z" />
              </svg>
            </div>
            <h3>{{ att.name }}</h3>
            <p v-if="att.description" class="card-desc">{{ att.description }}</p>
            <div class="card-tags">
              <span v-if="att.category" class="tag">{{ att.category }}</span>
              <span v-for="tag in (att.tags || []).slice(0, 2)" :key="tag" class="tag">{{ tag }}</span>
            </div>
          </div>
          <div class="card-glow"></div>
        </article>
      </div>
    </section>
    <section v-else-if="!loading" class="content-section">
      <div class="section-header">
        <span class="section-deco">〔</span>
        <h2>{{ t('scenic.attractionsTitle') }}</h2>
        <span class="section-deco">〕</span>
      </div>
      <p class="empty-state">{{ t('scenic.noAttractions') }}</p>
    </section>

    <!-- Routes section -->
    <section v-if="routes.length > 0" class="content-section">
      <div class="section-header">
        <span class="section-deco">〔</span>
        <h2>{{ t('scenic.routesTitle') }}</h2>
        <span class="section-deco">〕</span>
      </div>
      <div class="card-grid">
        <article v-for="route in routes" :key="route.id" class="info-card route-card" @click="router.push(`/scenic/route/${route.id}`)">
          <div class="card-icon route-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M9 6.75V15m6-6v8.25m.503 3.498l4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 00-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0z" />
            </svg>
          </div>
          <h3>{{ route.name }}</h3>
          <p v-if="route.description" class="card-desc">{{ route.description }}</p>
          <div class="route-meta">
            <span v-if="route.duration">{{ t('scenic.duration') }}: {{ route.duration }}</span>
            <span v-if="route.difficulty">{{ t('scenic.difficulty') }}: {{ route.difficulty }}</span>
          </div>
          <div v-if="route.steps && route.steps.length > 0" class="route-stops">
            <span class="stops-label">{{ t('scenic.stops') }}:</span>
            <span class="stops-list">{{ route.steps.map(s => s.attraction_name || s.attraction_id).join(' → ') }}</span>
          </div>
          <div class="card-glow"></div>
        </article>
      </div>
    </section>
    <section v-else-if="!loading" class="content-section">
      <div class="section-header">
        <span class="section-deco">〔</span>
        <h2>{{ t('scenic.routesTitle') }}</h2>
        <span class="section-deco">〕</span>
      </div>
      <p class="empty-state">{{ t('scenic.noRoutes') }}</p>
    </section>
  </main>
</template>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
   古风 + 数字 (Ancient + Digital) Theme
   Colors: 朱红 #C23B22 · 墨黑 #1C1C1C · 宣纸白 #F5F0E8 · 青瓷绿 #6BAF8D
   ═══════════════════════════════════════════════════════════════════════ */
.scenic-page {
  --zhu-red: #C23B22;
  --zhu-red-light: #E85D4A;
  --ink-black: #1C1C1C;
  --ink-gray: #3A3A3A;
  --xuan-white: #F5F0E8;
  --xuan-warm: #EDE6D6;
  --celadon: #6BAF8D;
  --celadon-light: #8FCDB0;
  --gold: #C9A84C;
  --gold-light: #E0C878;
  --muted: #7A7264;
  --border: #D4CBBA;
  --surface: rgba(245, 240, 232, 0.85);
  --radius: 12px;
  min-height: 100vh;
  background: var(--xuan-white);
  color: var(--ink-black);
  position: relative;
  overflow: hidden;
  font-family: "Noto Serif SC", "Source Han Serif SC", "SimSun", serif;
}

/* ── Ink wash background ── */
.ink-bg {
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.ink-circle {
  position: absolute;
  border-radius: 50%;
  opacity: 0.04;
  background: radial-gradient(circle, var(--ink-black) 0%, transparent 70%);
}

.ink-circle-1 { width: 600px; height: 600px; top: -200px; right: -100px; }
.ink-circle-2 { width: 400px; height: 400px; bottom: 10%; left: -80px; opacity: 0.03; }
.ink-circle-3 { width: 300px; height: 300px; top: 40%; right: 20%; opacity: 0.02; }

.circuit-line {
  position: absolute;
  background: linear-gradient(90deg, transparent, var(--celadon), transparent);
  opacity: 0.08;
}

.circuit-h {
  width: 100%;
  height: 1px;
  top: 30%;
  animation: circuit-flow 8s linear infinite;
}

.circuit-v {
  width: 1px;
  height: 100%;
  left: 70%;
  background: linear-gradient(180deg, transparent, var(--celadon), transparent);
  animation: circuit-flow-v 10s linear infinite;
}

@keyframes circuit-flow {
  0% { transform: translateX(-100%); }
  100% { transform: translateX(100%); }
}

@keyframes circuit-flow-v {
  0% { transform: translateY(-100%); }
  100% { transform: translateY(100%); }
}

/* ── Hero ── */
.hero {
  position: relative;
  z-index: 1;
  padding: 60px 24px 48px;
  display: flex;
  justify-content: center;
}

.hero-content {
  width: min(720px, 100%);
  text-align: center;
}

/* Decorative borders */
.deco-border-top,
.deco-border-bottom {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 24px;
  color: var(--gold);
}

.deco-border-bottom {
  margin-top: 32px;
  margin-bottom: 0;
}

.deco-dot {
  width: 6px;
  height: 6px;
  background: var(--gold);
  border-radius: 50%;
}

.deco-line {
  width: 60px;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--gold), transparent);
}

.deco-diamond {
  font-size: 10px;
  color: var(--gold);
  animation: pulse-gold 3s ease-in-out infinite;
}

@keyframes pulse-gold {
  0%, 100% { opacity: 0.6; }
  50% { opacity: 1; }
}

.hero-badge {
  display: inline-block;
  padding: 6px 20px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(194, 59, 34, 0.1), rgba(201, 168, 76, 0.1));
  color: var(--zhu-red);
  font-size: 14px;
  font-weight: 600;
  border: 1px solid rgba(194, 59, 34, 0.2);
  letter-spacing: 2px;
}

.hero-title {
  margin: 20px 0 0;
  font-size: clamp(40px, 7vw, 64px);
  font-weight: 900;
  color: var(--ink-black);
  line-height: 1.15;
  letter-spacing: 8px;
}

.title-char {
  display: inline-block;
  animation: char-reveal 0.6s ease-out both;
  background: linear-gradient(180deg, var(--ink-black) 40%, var(--zhu-red) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

@keyframes char-reveal {
  0% { opacity: 0; transform: translateY(20px) scale(0.9); }
  100% { opacity: 1; transform: translateY(0) scale(1); }
}

.hero-desc {
  margin: 16px auto 0;
  max-width: 520px;
  color: var(--muted);
  font-size: 15px;
  line-height: 28px;
  letter-spacing: 1px;
}

/* ── Guide cards ── */
.guide-grid {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 16px;
  margin: 32px auto 0;
  max-width: 720px;
}

.guide-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px 20px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  backdrop-filter: blur(8px);
  box-shadow: 0 2px 8px rgba(28, 28, 28, 0.06);
  text-align: left;
  cursor: pointer;
  transition: all 200ms ease;
  flex: 1 1 280px;
  max-width: 360px;
  position: relative;
  overflow: hidden;
}

.guide-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--celadon), transparent);
  opacity: 0;
  transition: opacity 200ms;
}

.guide-card:hover {
  border-color: var(--celadon);
  box-shadow: 0 4px 16px rgba(107, 175, 141, 0.15);
  transform: translateY(-2px);
}

.guide-card:hover::before { opacity: 1; }

.guide-card-selected {
  border-color: var(--zhu-red) !important;
  box-shadow: 0 0 0 2px rgba(194, 59, 34, 0.15), 0 4px 16px rgba(194, 59, 34, 0.1) !important;
}

.guide-card-selected::before {
  opacity: 1;
  background: linear-gradient(90deg, transparent, var(--zhu-red), transparent);
}

.guide-avatar {
  width: 60px;
  height: 60px;
  flex-shrink: 0;
  border-radius: 50%;
  overflow: hidden;
  background: var(--xuan-warm);
  border: 2px solid var(--border);
}

.guide-card-selected .guide-avatar {
  border-color: var(--zhu-red);
}

.guide-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.guide-avatar-placeholder {
  width: 100%;
  height: 100%;
  display: grid;
  place-items: center;
  color: var(--muted);
}

.guide-avatar-placeholder svg { width: 32px; height: 32px; }

.guide-info { flex: 1; min-width: 0; }

.guide-category {
  display: inline-block;
  padding: 2px 10px;
  border-radius: 999px;
  background: rgba(107, 175, 141, 0.12);
  color: var(--celadon);
  font-size: 11px;
  font-weight: 600;
  margin-bottom: 4px;
  letter-spacing: 1px;
}

.guide-info h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--ink-black);
}

.guide-info p {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 18px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.no-guide {
  margin: 32px auto 0;
  padding: 16px;
  max-width: 480px;
  background: rgba(194, 59, 34, 0.06);
  border: 1px solid rgba(194, 59, 34, 0.15);
  border-radius: var(--radius);
  color: var(--zhu-red);
  font-size: 14px;
}

/* ── Start button ── */
.start-btn {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-top: 32px;
  padding: 14px 40px;
  background: linear-gradient(135deg, var(--zhu-red), #A82F1A);
  color: var(--xuan-white);
  border: none;
  border-radius: var(--radius);
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  letter-spacing: 2px;
  box-shadow: 0 4px 16px rgba(194, 59, 34, 0.3);
  transition: all 200ms ease;
  position: relative;
  overflow: hidden;
}

.start-btn::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.15), transparent);
  transition: left 500ms;
}

.start-btn:hover:not(:disabled)::after { left: 100%; }

.start-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 24px rgba(194, 59, 34, 0.4);
}

.start-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-icon { font-size: 18px; }

.launch-error {
  margin: 16px auto 0;
  max-width: 480px;
  padding: 10px 16px;
  background: rgba(194, 59, 34, 0.06);
  border: 1px solid rgba(194, 59, 34, 0.15);
  border-radius: var(--radius);
  color: var(--zhu-red);
  font-size: 13px;
}

/* ── Content sections ── */
.content-section {
  position: relative;
  z-index: 1;
  padding: 40px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 32px;
}

.section-deco {
  font-size: 28px;
  color: var(--gold);
  font-weight: 300;
}

.section-header h2 {
  margin: 0;
  font-size: 26px;
  font-weight: 700;
  color: var(--ink-black);
  letter-spacing: 4px;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
  width: min(1100px, 100%);
}

/* ── Info cards ── */
.info-card {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  backdrop-filter: blur(8px);
  box-shadow: 0 2px 8px rgba(28, 28, 28, 0.04);
  transition: all 250ms ease;
  cursor: pointer;
  position: relative;
  overflow: hidden;
}

.info-card:hover {
  box-shadow: 0 8px 32px rgba(28, 28, 28, 0.1);
  transform: translateY(-3px);
  border-color: var(--celadon);
}

.card-glow {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at center, rgba(107, 175, 141, 0.06) 0%, transparent 60%);
  opacity: 0;
  transition: opacity 300ms;
  pointer-events: none;
}

.info-card:hover .card-glow { opacity: 1; }

.card-icon {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  margin-bottom: 12px;
}

.card-image {
  height: 180px;
  overflow: hidden;
  position: relative;
}

.card-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
  transition: transform 400ms ease;
}

.info-card:hover .card-image img { transform: scale(1.05); }

.card-image-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent 50%, rgba(28, 28, 28, 0.4) 100%);
  pointer-events: none;
}

.card-image-placeholder {
  display: grid;
  place-items: center;
  background: var(--xuan-warm);
  border-bottom: 1px solid var(--border);
}

.card-image-placeholder svg { width: 40px; height: 40px; color: var(--border); }

.card-body { padding: 20px; }

.card-icon svg { width: 22px; height: 22px; }

.attraction-icon {
  background: rgba(107, 175, 141, 0.12);
  color: var(--celadon);
}

.route-icon {
  background: rgba(194, 59, 34, 0.08);
  color: var(--zhu-red);
}

.info-card h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: var(--ink-black);
}

.card-desc {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--muted);
  line-height: 20px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 12px;
}

.tag {
  padding: 3px 10px;
  border-radius: 999px;
  background: rgba(28, 28, 28, 0.04);
  color: var(--muted);
  font-size: 11px;
  font-weight: 500;
  letter-spacing: 0.5px;
}

/* ── Route specific ── */
.route-meta {
  display: flex;
  gap: 16px;
  margin-top: 10px;
  font-size: 12px;
  color: var(--muted);
  letter-spacing: 0.5px;
}

.route-stops {
  margin-top: 10px;
  font-size: 12px;
  color: var(--muted);
  line-height: 18px;
}

.stops-label { font-weight: 600; color: var(--ink-gray); }
.stops-list { margin-left: 4px; }

.empty-state {
  color: var(--muted);
  font-size: 15px;
  letter-spacing: 1px;
}

/* ── Responsive ── */
@media (max-width: 640px) {
  .hero { padding: 48px 16px 36px; }
  .guide-card { flex: 1 1 100%; max-width: 100%; }
  .card-grid { grid-template-columns: 1fr; }
  .hero-title { letter-spacing: 4px; }
}
</style>
