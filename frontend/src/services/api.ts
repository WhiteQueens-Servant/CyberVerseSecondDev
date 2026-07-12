import type { AvatarModelInfo, Character, CharacterForm, ComponentsResponse, ImageInfo, KnowledgeSource, KnowledgeUploadSkippedFile, Settings, LaunchConfig, LaunchConfigUpdate, PipelineMode } from '../types'

const API_BASE = import.meta.env.VITE_API_BASE || '/api/v1'

// ── Helpers ──

async function request<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'same-origin',
    ...opts,
  })
  if (!res.ok) {
    let message = `API error ${res.status}: ${path}`
    try {
      const data = await res.clone().json() as { error?: string }
      if (data?.error) message = data.error
    } catch {
      try {
        const text = await res.text()
        if (text) message = text
      } catch {
        // keep default message
      }
    }
    throw new Error(message)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

// ── Sessions (existing) ──

export interface CreateSessionResponse {
  session_id: string
  mode: PipelineMode
  streaming_mode: string  // "direct" or "livekit"
  avatar_enabled?: boolean
  idle_strategy?: 'cached_video' | 'silent_inference'
  livekit_url?: string
  livekit_token?: string
  idle_video_url?: string
  idle_video_urls?: string[]
  warnings?: string[]
  visual_input?: {
    enabled: boolean
    frame_interval_ms: number
    max_width: number
    max_height: number
    jpeg_quality: number
    max_frame_bytes: number
    ws_max_message_bytes: number
    max_recent_frames: number
    frame_ttl_ms: number
  }
}

export interface SessionInfo {
  id: string
  state: string
}

export interface HealthResponse {
  status: string
  sessions: number
  inference_connected: boolean
  error?: string
}

export async function createSession(characterId: string, mode: PipelineMode = 'standard'): Promise<CreateSessionResponse> {
  return request('/sessions', {
    method: 'POST',
    body: JSON.stringify({ character_id: characterId, mode }),
  })
}

export async function getComponents(): Promise<ComponentsResponse> {
  return request('/components')
}

export async function deleteSession(sessionId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/sessions/${sessionId}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 404) throw new Error(`Failed to delete session: ${res.status}`)
}

export async function sendMessage(sessionId: string, text: string): Promise<void> {
  return request(`/sessions/${sessionId}/message`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  })
}

export async function listSessions(): Promise<SessionInfo[]> {
  return request('/sessions')
}

export async function getHealth(): Promise<HealthResponse> {
  return request('/health')
}

// ── Agent Tasks ──

export interface AgentTask {
  id: string
  session_id: string
  character_id?: string
  kind: string
  title: string
  user_request: string
  status: 'queued' | 'running' | 'waiting_user' | 'completed' | 'failed' | 'cancelled'
  progress: number
  result_summary?: string
  created_at: string
  updated_at: string
  finished_at?: string
}

export interface AgentTaskEvent {
  task_id: string
  seq: number
  event_type: string
  status: AgentTask['status']
  message?: string
  progress: number
  payload?: Record<string, unknown>
  created_at: string
}

export async function listSessionTasks(sessionId: string): Promise<{ tasks: AgentTask[] }> {
  return request(`/sessions/${sessionId}/tasks`)
}

export async function getTaskEvents(taskId: string, afterSeq = 0): Promise<{ events: AgentTaskEvent[] }> {
  return request(`/tasks/${taskId}/events?after_seq=${afterSeq}`)
}

export function getTaskArtifactUrl(taskId: string, artifactId: string): string {
  return `${API_BASE}/tasks/${encodeURIComponent(taskId)}/artifacts/${encodeURIComponent(artifactId)}`
}

// ── Conversation History ──

export interface ConversationMessagesResponse {
  messages: { role: string; content: string; timestamp: string; session_id: string }[]
  next_cursor: string
  has_more: boolean
}

export async function getConversationMessages(
  characterId: string,
  limit: number = 50,
  before?: string,
): Promise<ConversationMessagesResponse> {
  const params = new URLSearchParams({ limit: String(limit) })
  if (before) params.set('before', before)
  return request(`/characters/${characterId}/conversations/messages?${params}`)
}

// ── Characters ──

export async function getCharacters(): Promise<Character[]> {
  return request('/characters')
}

export async function getCharacter(id: string): Promise<Character> {
  return request(`/characters/${id}`)
}

export async function createCharacter(data: CharacterForm): Promise<Character> {
  return request('/characters', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateCharacter(id: string, data: CharacterForm): Promise<Character> {
  return request(`/characters/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteCharacter(id: string): Promise<void> {
  return request(`/characters/${id}`, { method: 'DELETE' })
}

export async function testCharacterVoice(data: { voice_provider: string; voice_type: string }): Promise<{ status: string }> {
  return request('/characters/test-voice', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function uploadAvatar(id: string, file: File): Promise<{ path: string; filename?: string }> {
  const formData = new FormData()
  formData.append('avatar', file)
  const res = await fetch(`${API_BASE}/characters/${id}/avatar`, {
    method: 'POST',
    body: formData,
  })
  if (!res.ok) throw new Error(`Failed to upload avatar: ${res.status}`)
  return res.json()
}

// ── Character Images ──

export async function getCharacterImages(id: string): Promise<ImageInfo[]> {
  return request(`/characters/${id}/images`)
}

export async function deleteCharacterImage(id: string, filename: string): Promise<void> {
  const res = await fetch(`${API_BASE}/characters/${id}/images/${filename}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 404) throw new Error(`Failed to delete image: ${res.status}`)
}

export async function activateCharacterImage(id: string, filename: string): Promise<void> {
  const res = await fetch(`${API_BASE}/characters/${id}/images/${filename}/activate`, { method: 'PUT' })
  if (!res.ok) throw new Error(`Failed to activate image: ${res.status}`)
}

// ── Character Knowledge Sources ──

export async function getKnowledgeSources(id: string): Promise<KnowledgeSource[]> {
  const data = await request<{ sources: KnowledgeSource[] }>(`/characters/${id}/knowledge`)
  return data.sources
}

export interface UploadKnowledgeFilesResult {
  sources: KnowledgeSource[]
  skipped?: KnowledgeUploadSkippedFile[]
}

export async function uploadKnowledgeFiles(
  id: string,
  files: File[],
): Promise<UploadKnowledgeFilesResult> {
  const formData = new FormData()
  for (const file of files) {
    const uploadFile = file as File & { webkitRelativePath?: string; relativePath?: string }
    const relativePath = uploadFile.relativePath || uploadFile.webkitRelativePath || file.name
    formData.append('files', file, relativePath)
    formData.append('relative_paths', relativePath)
  }
  const res = await fetch(`${API_BASE}/characters/${id}/knowledge/files`, {
    method: 'POST',
    body: formData,
  })
  if (!res.ok) {
    let message = `Failed to upload knowledge source: ${res.status}`
    try {
      const body = await res.json() as { error?: string; skipped?: KnowledgeUploadSkippedFile[] }
      if (body.error) message = body.error
      if (body.skipped?.length) {
        message += ` (${body.skipped.length} skipped)`
      }
    } catch {
      // keep default message
    }
    throw new Error(message)
  }
  const body = await res.json()
  if (Array.isArray(body?.sources)) return body
  return { sources: [body as KnowledgeSource] }
}

export async function deleteKnowledgeSource(id: string, sourceId: string): Promise<void> {
  const res = await fetch(`${API_BASE}/characters/${id}/knowledge/${sourceId}`, { method: 'DELETE' })
  if (!res.ok && res.status !== 404) throw new Error(`Failed to delete knowledge source: ${res.status}`)
}

export async function reindexKnowledgeSource(id: string, sourceId: string): Promise<KnowledgeSource> {
  return request(`/characters/${id}/knowledge/${sourceId}/reindex`, { method: 'POST' })
}

// ── Settings ──

export async function getSettings(): Promise<Settings> {
  return request('/settings')
}

export async function updateSettings(data: Settings): Promise<void> {
  return request('/settings', {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function testConnection(): Promise<{ status: string }> {
  return request('/settings/test', { method: 'POST' })
}

// ── Launch Config ──

export async function getAvatarModelInfo(): Promise<AvatarModelInfo> {
  return request('/config/avatar-model')
}

export async function getLaunchConfig(model?: string): Promise<LaunchConfig> {
  const qs = model ? `?model=${encodeURIComponent(model)}` : ''
  return request(`/config/launch${qs}`)
}

export async function updateLaunchConfig(data: LaunchConfigUpdate): Promise<{ status: string; requires_restart: boolean }> {
  return request('/config/launch', {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

// ── Scenic: Attractions ──

export interface Attraction {
  id: string
  name: string
  description: string
  category: string
  location: string
  image_url: string
  tags: string[]
  opening_hours: string
  ticket_info: string
  created_at: string
  updated_at: string
}

export interface AttractionForm {
  name: string
  description: string
  category: string
  location?: string
  image_url?: string
  tags?: string[]
  opening_hours?: string
  ticket_info?: string
}

export async function listAttractions(): Promise<Attraction[]> {
  return request('/attractions')
}

export async function getAttraction(id: string): Promise<Attraction> {
  return request(`/attractions/${id}`)
}

export async function createAttraction(data: AttractionForm): Promise<Attraction> {
  return request('/attractions', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateAttraction(id: string, data: AttractionForm): Promise<Attraction> {
  return request(`/attractions/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteAttraction(id: string): Promise<void> {
  return request(`/attractions/${id}`, { method: 'DELETE' })
}

// ── Scenic: Routes ──

export interface RouteStep {
  attraction_id: string
  attraction_name?: string
  order: number
  duration_minutes: number
  highlight: string
}

export interface ScenicRoute {
  id: string
  name: string
  description: string
  duration: string
  difficulty: string
  tags: string[]
  steps: RouteStep[]
  created_at: string
  updated_at: string
}

export interface RouteForm {
  name: string
  description: string
  duration: string
  difficulty: string
  tags?: string[]
  steps: Array<{
    attraction_id: string
    order: number
    duration_minutes: number
    highlight?: string
  }>
}

export async function listRoutes(): Promise<ScenicRoute[]> {
  return request('/routes')
}

export async function getRoute(id: string): Promise<ScenicRoute> {
  return request(`/routes/${id}`)
}

export async function createRoute(data: RouteForm): Promise<ScenicRoute> {
  return request('/routes', {
    method: 'POST',
    body: JSON.stringify(data),
  })
}

export async function updateRoute(id: string, data: RouteForm): Promise<ScenicRoute> {
  return request(`/routes/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  })
}

export async function deleteRoute(id: string): Promise<void> {
  return request(`/routes/${id}`, { method: 'DELETE' })
}

// ── Scenic: Analytics ──

export interface DashboardData {
  today_sessions: number
  week_sessions: Array<{ date: string; count: number }>
  hot_questions: Array<{ question: string; count: number }>
  sentiment_distribution: { positive: number; neutral: number; negative: number }
  hourly_distribution: Array<{ hour: number; count: number }>
}

export interface SessionLog {
  id: string
  character_id: string
  /** Backend has this field but ListSessions never populates it; resolve client-side via getCharacterName() */
  character_name?: string
  started_at: string
  ended_at?: string
  duration_s: number
  turn_count: number
  sentiment: string
}

export interface SessionLogDetail {
  session: SessionLog
  messages: Array<{ role: string; content: string; timestamp: string }>
}

export async function getDashboard(): Promise<DashboardData> {
  return request('/analytics/dashboard')
}

export async function getAnalyticsSessions(params?: {
  character_id?: string
  date_from?: string
  date_to?: string
  sentiment?: string
}): Promise<SessionLog[]> {
  const qs = new URLSearchParams()
  if (params?.character_id) qs.set('character_id', params.character_id)
  if (params?.date_from) qs.set('date_from', params.date_from)
  if (params?.date_to) qs.set('date_to', params.date_to)
  if (params?.sentiment) qs.set('sentiment', params.sentiment)
  const query = qs.toString()
  return request(`/analytics/sessions${query ? `?${query}` : ''}`)
}

export async function getAnalyticsSessionDetail(id: string): Promise<SessionLogDetail> {
  return request(`/analytics/sessions/${id}`)
}

// ── Scenic: Reports ──

export interface Report {
  id: string
  status: string
  date_from: string
  date_to: string
  character_id?: string
  content?: Record<string, unknown>
  created_at: string
  finished_at?: string
}

export async function generateReport(data?: {
  date_from?: string
  date_to?: string
  character_id?: string
}): Promise<Report> {
  return request('/reports/generate', {
    method: 'POST',
    body: JSON.stringify(data || {}),
  })
}

export async function listReports(): Promise<Report[]> {
  return request('/reports')
}

export async function getReport(id: string): Promise<Report> {
  return request(`/reports/${id}`)
}

export async function deleteReport(id: string): Promise<void> {
  return request(`/reports/${id}`, { method: 'DELETE' })
}
