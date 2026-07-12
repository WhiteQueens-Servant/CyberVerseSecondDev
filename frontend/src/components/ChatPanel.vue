<script setup lang="ts">
import { ref, nextTick, watch, onMounted, onUnmounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { ChatMessage, AvatarStatus } from '../composables/useChat'
import TaskProgressCard from './TaskProgressCard.vue'
import ScenicCard from './ScenicCard.vue'

const props = withDefaults(defineProps<{
  messages: ChatMessage[]
  currentTranscript: string
  currentLLMResponse: string
  avatarStatus: AvatarStatus
  historyLoading?: boolean
  historyHasMore?: boolean
  theme?: 'light' | 'dark'
}>(), {
  theme: 'dark',
})

const emit = defineEmits<{
  sendText: [text: string]
  loadMore: []
}>()

const { t } = useI18n()
const inputText = ref('')
const messagesContainer = ref<HTMLElement | null>(null)
const sentinel = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null
let prevMessageCount = 0
let isLoadingHistory = false
let scrollHeightBeforeLoad = 0
let initialLoadDone = false

function handleSend() {
  const text = inputText.value.trim()
  if (!text) return
  emit('sendText', text)
  inputText.value = ''
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

// Scenic card parsing: detect route/attraction info in LLM text responses
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

const SCENIC_JSON_PATTERN = /<!--SCENIC_CARD_JSON\s*([\s\S]*?)\s*SCENIC_CARD_JSON-->/
const ROUTE_PATTERN = /(?:推荐|介绍|有一条|这条|以下是).{0,10}路线[：:]/i
const ATTRACTION_PATTERN = /(?:推荐|介绍|以下是|有名).{0,10}(?:景点|景区|地方)[：:]/i
const STOPS_PATTERN = /途经[：:]/
const DURATION_PATTERN = /[时长耗时].{0,5}[：:]/i

function parseScenicCards(content: string): ScenicCardData[] {
  // Primary: extract structured JSON block injected by persona_agent
  const jsonMatch = content.match(SCENIC_JSON_PATTERN)
  if (jsonMatch) {
    try {
      const parsed = JSON.parse(jsonMatch[1])
      const scenicCards: ScenicCardStructured[] = parsed.scenic_cards || parsed
      if (Array.isArray(scenicCards) && scenicCards.length > 0) {
        return scenicCards.map(sc => ({
          type: sc.type === 'route' ? 'route' as const : 'attraction' as const,
          title: sc.name || '',
          summary: buildSummary(sc),
          details: content.replace(SCENIC_JSON_PATTERN, '').trim(),
          structured: sc,
        }))
      }
    } catch {
      // JSON parse failed, fall through to regex fallback
    }
  }

  // Fallback: regex-based heuristic parsing
  const cards: ScenicCardData[] = []
  const lines = content.split('\n').map(l => l.trim()).filter(Boolean)
  const fullText = content

  const routeTitleMatch = fullText.match(/(?:推荐|介绍|有一条|这条|以下是).{0,20}路线[：:，,]?\s*[《「]?(.{2,30})[》」]?/)
  if (routeTitleMatch || STOPS_PATTERN.test(fullText) || (ROUTE_PATTERN.test(fullText) && DURATION_PATTERN.test(fullText))) {
    const title = routeTitleMatch?.[1]?.trim() || extractFirstLine(content)
    const stopsMatch = fullText.match(/途经[：:]\s*(.+)/)
    const durationMatch = fullText.match(/[时长耗时].{0,3}[：:]\s*(.+?)[\n,，。]/)
    const difficultyMatch = fullText.match(/难度[：:]\s*(.+?)[\n,，。]/)

    const summaryParts: string[] = []
    if (durationMatch) summaryParts.push(durationMatch[1].trim())
    if (difficultyMatch) summaryParts.push(difficultyMatch[1].trim())
    if (stopsMatch) summaryParts.push('途经 ' + stopsMatch[1].trim())

    cards.push({
      type: 'route',
      title: title,
      summary: summaryParts.join(' · ') || lines.slice(0, 2).join(' '),
      details: fullText,
    })
  }

  if (ATTRACTION_PATTERN.test(fullText) && cards.length === 0) {
    const titleMatch = fullText.match(/(?:景点|景区)[：:，,]?\s*[《「]?(.{2,20})[》」]?/)
    const categoryMatch = fullText.match(/分类[：:]\s*(.+?)[\n,，。]/)
    const title = titleMatch?.[1]?.trim() || extractFirstLine(content)

    const summaryParts: string[] = []
    if (categoryMatch) summaryParts.push(categoryMatch[1].trim())

    cards.push({
      type: 'attraction',
      title: title,
      summary: summaryParts.join(' · ') || lines.slice(0, 2).join(' '),
      details: fullText,
    })
  }

  return cards
}

function buildSummary(sc: ScenicCardStructured): string {
  if (sc.type === 'route') {
    const parts: string[] = []
    if (sc.duration_minutes) {
      const h = Math.floor(sc.duration_minutes / 60)
      const m = sc.duration_minutes % 60
      parts.push(h > 0 ? `${h}小时${m > 0 ? m + '分钟' : ''}` : `${m}分钟`)
    }
    if (sc.difficulty) parts.push(sc.difficulty)
    if (sc.steps && sc.steps.length > 0) {
      parts.push('途经 ' + sc.steps.map(s => s.attraction_name).join(' → '))
    }
    return parts.join(' · ')
  }
  // attraction
  const parts: string[] = []
  if (sc.category) parts.push(sc.category)
  if (sc.location) parts.push(sc.location)
  return parts.join(' · ') || (sc.description || '').slice(0, 60)
}

function extractFirstLine(text: string): string {
  const line = text.split('\n')[0]?.trim() || ''
  return line.length > 30 ? line.slice(0, 30) + '...' : line
}

/**
 * Render a small subset of Markdown to HTML.
 * Covers bold, italic, headings, unordered lists, and links.
 * Intentionally lightweight — does not handle tables, code blocks, etc.
 */
function renderMarkdown(text: string): string {
  let html = text
    // Escape HTML entities
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
  // Bold: **text**
  html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
  // Italic: *text*
  html = html.replace(/\*(.+?)\*/g, '<em>$1</em>')
  // Headings: ### text
  html = html.replace(/^(#{1,6})\s+(.+)$/gm, '<strong>$2</strong>')
  // Unordered list items: - text or * text
  html = html.replace(/^[\-\*]\s+(.+)$/gm, '• $1')
  // Links: [text](url)
  html = html.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
  // Line breaks
  html = html.replace(/\n/g, '<br>')
  return html
}

// Compute session separators: indices where session_id changes between history messages
const sessionBreaks = computed(() => {
  const breaks = new Set<number>()
  for (let i = 1; i < props.messages.length; i++) {
    const prev = props.messages[i - 1]
    const curr = props.messages[i]
    if (prev.isHistory && curr.isHistory && prev.sessionId && curr.sessionId && prev.sessionId !== curr.sessionId) {
      breaks.add(i)
    }
    // Separator between history and live messages
    if (prev.isHistory && !curr.isHistory) {
      breaks.add(i)
    }
  }
  return breaks
})

const messageRenderKey = computed(() => props.messages.map((msg) => {
  if (msg.kind === 'task' && msg.task) {
    return [
      msg.id,
      msg.task.status,
      msg.task.progress,
      msg.task.eventCount,
      msg.task.artifacts.map(artifact => artifact.id).join(','),
    ].join(':')
  }
  return `${msg.id || ''}:${msg.timestamp}:${msg.content}`
}).join('|'))

watch(
  () => props.historyLoading,
  (loading) => {
    if (loading) {
      isLoadingHistory = true
      // Capture scroll height BEFORE history messages are inserted
      const container = messagesContainer.value
      if (container) {
        scrollHeightBeforeLoad = container.scrollHeight
      }
    }
  }
)

watch(
  messageRenderKey,
  async () => {
    const container = messagesContainer.value
    if (!container) return
    const newLen = props.messages.length

    if (isLoadingHistory && newLen > prevMessageCount) {
      // History was prepended: preserve scroll position using height captured before load
      await nextTick()
      const newHeight = container.scrollHeight
      container.scrollTop += newHeight - scrollHeightBeforeLoad
      isLoadingHistory = false
      initialLoadDone = true
    } else {
      // New message appended: scroll to bottom
      await nextTick()
      container.scrollTop = container.scrollHeight
    }
    prevMessageCount = newLen
  }
)

onMounted(() => {
  prevMessageCount = props.messages.length
  // Set up IntersectionObserver for infinite scroll up
  if (sentinel.value && messagesContainer.value) {
    observer = new IntersectionObserver(
      (entries) => {
        // Skip triggers before initial history load completes (sentinel is visible in empty container)
        if (!initialLoadDone) return
        if (entries[0]?.isIntersecting && props.historyHasMore && !props.historyLoading) {
          emit('loadMore')
        }
      },
      { root: messagesContainer.value, threshold: 0.1 }
    )
    observer.observe(sentinel.value)
  }
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>

<template>
  <div class="chat-panel" :class="`theme-${theme}`">
    <div ref="messagesContainer" class="messages">
      <!-- Sentinel for infinite scroll up -->
      <div ref="sentinel" class="sentinel">
        <div v-if="historyLoading" class="history-loading">
          <span class="loading-dot" /><span class="loading-dot" /><span class="loading-dot" />
        </div>
        <div v-else-if="historyHasMore" class="load-more-hint">
          {{ t('chat.loadMore') }}
        </div>
        <div v-else-if="messages.some(m => m.isHistory)" class="history-end">
          {{ t('chat.historyEnd') }}
        </div>
      </div>

      <template v-for="(msg, i) in messages" :key="msg.id || `msg-${msg.timestamp}-${i}`">
        <!-- Session separator -->
        <div v-if="sessionBreaks.has(i)" class="session-separator">
          <span class="separator-line" />
          <span v-if="msg.isHistory" class="separator-label">{{ t('chat.previousConversation') }}</span>
          <span v-else class="separator-label">{{ t('chat.currentConversation') }}</span>
          <span class="separator-line" />
        </div>

        <div
          v-if="msg.kind === 'task' && msg.task"
          class="message task"
          :class="{ history: msg.isHistory }"
        >
          <TaskProgressCard :task="msg.task" />
        </div>

        <div
          v-else
          class="message"
          :class="[msg.role, { history: msg.isHistory }]"
        >
          <template v-if="msg.role === 'assistant' && parseScenicCards(msg.content).length > 0">
            <div class="message-content" v-html="renderMarkdown(msg.content)" />
            <ScenicCard v-for="(card, ci) in parseScenicCards(msg.content)" :key="ci" :data="card" />
          </template>
          <template v-else-if="msg.role === 'assistant'">
            <div class="message-content" v-html="renderMarkdown(msg.content)" />
            <a
              v-if="msg.artifactUrl"
              class="artifact-link"
              :href="msg.artifactUrl"
              target="_blank"
              rel="noreferrer"
            >
              打开资料
            </a>
          </template>
          <div v-else class="message-content">
            {{ msg.content }}
            <a
              v-if="msg.artifactUrl"
              class="artifact-link"
              :href="msg.artifactUrl"
              target="_blank"
              rel="noreferrer"
            >
              打开资料
            </a>
          </div>
        </div>
      </template>

      <div v-if="currentTranscript" class="message user typing">
        <div class="message-content">{{ currentTranscript }}...</div>
      </div>

      <div v-if="currentLLMResponse" class="message assistant typing">
        <div class="message-content" v-html="renderMarkdown(currentLLMResponse)" />
      </div>
    </div>

    <div class="input-bar">
      <input
        v-model="inputText"
        type="text"
        :placeholder="t('chat.inputPlaceholder')"
        @keydown="handleKeydown"
      />
      <button class="send-btn" @click="handleSend" :disabled="!inputText.trim()">
        {{ t('chat.send') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.chat-panel {
  /* Dark theme (default) */
  --chat-bg: #1e1e1e;
  --chat-input-bg: #2a2a2a;
  --chat-border: #333;
  --chat-border-focus: #2563eb;
  --chat-text: #eee;
  --chat-text-muted: #666;
  --chat-text-secondary: #555;
  --chat-user-bubble: #2563eb;
  --chat-user-text: white;
  --chat-assistant-bubble: #333;
  --chat-system-bg: #24303a;
  --chat-system-text: #cbd5e1;
  --chat-link-color: #66d9ef;
  --chat-dot-color: #666;
  --chat-separator: #333;
  --chat-btn-bg: #2563eb;
  --chat-btn-text: white;

  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--chat-bg);
  overflow: hidden;
}

/* Light scenic theme */
.chat-panel.theme-light {
  --chat-bg: #f8faf8;
  --chat-input-bg: #ffffff;
  --chat-border: #d4e8d4;
  --chat-border-focus: #16a34a;
  --chat-text: #1a1a1a;
  --chat-text-muted: #6b7280;
  --chat-text-secondary: #9ca3af;
  --chat-user-bubble: #16a34a;
  --chat-user-text: white;
  --chat-assistant-bubble: #ffffff;
  --chat-system-bg: #f0fdf4;
  --chat-system-text: #374151;
  --chat-link-color: #16a34a;
  --chat-dot-color: #9ca3af;
  --chat-separator: #e5e7eb;
  --chat-btn-bg: #16a34a;
  --chat-btn-text: white;
}

.messages {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sentinel {
  min-height: 1px;
  display: flex;
  justify-content: center;
  padding: 4px 0;
}

.history-loading {
  display: flex;
  gap: 4px;
  align-items: center;
  padding: 8px;
}

.loading-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--chat-dot-color);
  animation: pulse 1.2s ease-in-out infinite;
}
.loading-dot:nth-child(2) { animation-delay: 0.2s; }
.loading-dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes pulse {
  0%, 100% { opacity: 0.3; }
  50% { opacity: 1; }
}

.load-more-hint {
  font-size: 12px;
  color: var(--chat-text-muted);
  padding: 4px 0;
}

.history-end {
  font-size: 12px;
  color: var(--chat-text-secondary);
  padding: 4px 0;
}

.session-separator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
}

.separator-line {
  flex: 1;
  height: 1px;
  background: var(--chat-separator);
}

.separator-label {
  font-size: 11px;
  color: var(--chat-text-muted);
  white-space: nowrap;
}

.message {
  box-sizing: border-box;
  width: max-content;
  max-width: 80%;
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 14px;
  line-height: 1.4;
}
.message-content {
  max-width: 100%;
  white-space: pre-wrap;
  overflow-wrap: break-word;
}
.artifact-link {
  display: inline-flex;
  margin-top: 6px;
  color: var(--chat-link-color);
  text-decoration: none;
}
.artifact-link:hover {
  text-decoration: underline;
}
.message.user {
  align-self: flex-end;
  background: var(--chat-user-bubble);
  color: var(--chat-user-text);
}
.message.assistant {
  align-self: flex-start;
  background: var(--chat-assistant-bubble);
  color: var(--chat-text);
  border: 1px solid var(--chat-border);
}
.message.system {
  align-self: center;
  max-width: 88%;
  background: var(--chat-system-bg);
  color: var(--chat-system-text);
}
.message.task {
  align-self: stretch;
  width: 100%;
  max-width: 100%;
  padding: 0;
  background: transparent;
  color: inherit;
}
.message.history {
  opacity: 0.75;
}
.message.typing {
  opacity: 0.7;
}

.input-bar {
  display: flex;
  gap: 8px;
  padding: 12px;
  border-top: 1px solid var(--chat-border);
}

.input-bar input {
  flex: 1;
  padding: 8px 12px;
  background: var(--chat-input-bg);
  border: 1px solid var(--chat-border);
  border-radius: 8px;
  color: var(--chat-text);
  outline: none;
}
.input-bar input:focus {
  border-color: var(--chat-border-focus);
}

.send-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  background: var(--chat-btn-bg);
  color: var(--chat-btn-text);
}
.send-btn:disabled {
  opacity: 0.5;
  cursor: default;
}
</style>
