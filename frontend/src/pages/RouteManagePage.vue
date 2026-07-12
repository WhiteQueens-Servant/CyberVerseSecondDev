<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppHeader from '../components/AppHeader.vue'
import {
  listRoutes, createRoute, updateRoute, deleteRoute, listAttractions,
  type ScenicRoute, type RouteForm, type Attraction,
} from '../services/api'

const { t } = useI18n()

// ── State ──
const routes = ref<ScenicRoute[]>([])
const allAttractions = ref<Attraction[]>([])
const loading = ref(false)
const search = ref('')
const tagFilter = ref('')
const showDialog = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const deleteConfirmId = ref<string | null>(null)

interface StepForm {
  attraction_id: string
  order: number
  duration_minutes: number
  highlight: string
}

const form = ref<{
  name: string
  description: string
  duration: string
  difficulty: string
  tags: string[]
  steps: StepForm[]
}>({
  name: '',
  description: '',
  duration: '',
  difficulty: 'easy',
  tags: [],
  steps: [],
})
const tagInput = ref('')

const difficulties = [
  { value: 'easy', label: '简单' },
  { value: 'moderate', label: '适中' },
]

// ── Computed ──
const allTags = computed(() => {
  const set = new Set<string>()
  routes.value.forEach(r => r.tags.forEach(t => set.add(t)))
  return Array.from(set)
})

const filtered = computed(() => {
  let list = routes.value
  if (tagFilter.value) {
    list = list.filter(r => r.tags.includes(tagFilter.value))
  }
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(r =>
      r.name.toLowerCase().includes(q) || r.description.toLowerCase().includes(q)
    )
  }
  return list
})

// ── Actions ──
async function fetchData() {
  loading.value = true
  try {
    const [r, a] = await Promise.all([listRoutes(), listAttractions()])
    routes.value = r
    allAttractions.value = a
  } catch (e) {
    console.error('Failed to load data:', e)
  } finally {
    loading.value = false
  }
}

function getAttractionName(id: string): string {
  return allAttractions.value.find(a => a.id === id)?.name || id
}

function getDifficultyLabel(value: string): string {
  return difficulties.find(d => d.value === value)?.label || value
}

function openCreate() {
  editingId.value = null
  form.value = { name: '', description: '', duration: '', difficulty: 'easy', tags: [], steps: [] }
  tagInput.value = ''
  showDialog.value = true
}

function openEdit(route: ScenicRoute) {
  editingId.value = route.id
  form.value = {
    name: route.name,
    description: route.description,
    duration: route.duration,
    difficulty: route.difficulty,
    tags: [...route.tags],
    steps: route.steps.map(s => ({
      attraction_id: s.attraction_id,
      order: s.order,
      duration_minutes: s.duration_minutes,
      highlight: s.highlight || '',
    })),
  }
  tagInput.value = route.tags.join(', ')
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editingId.value = null
}

function addStep() {
  form.value.steps.push({
    attraction_id: '',
    order: form.value.steps.length,
    duration_minutes: 30,
    highlight: '',
  })
}

function removeStep(index: number) {
  form.value.steps.splice(index, 1)
  form.value.steps.forEach((s, i) => { s.order = i })
}

function moveStepUp(index: number) {
  if (index <= 0) return
  const steps = form.value.steps
  ;[steps[index - 1], steps[index]] = [steps[index], steps[index - 1]]
  steps.forEach((s, i) => { s.order = i })
}

function moveStepDown(index: number) {
  const steps = form.value.steps
  if (index >= steps.length - 1) return
  ;[steps[index], steps[index + 1]] = [steps[index + 1], steps[index]]
  steps.forEach((s, i) => { s.order = i })
}

function parseTags() {
  form.value.tags = tagInput.value
    .split(/[,，]/)
    .map(s => s.trim())
    .filter(Boolean)
}

async function handleSave() {
  parseTags()
  if (!form.value.name || !form.value.description || !form.value.duration || form.value.steps.length === 0) return
  saving.value = true
  try {
    const payload: RouteForm = {
      name: form.value.name,
      description: form.value.description,
      duration: form.value.duration,
      difficulty: form.value.difficulty,
      tags: form.value.tags,
      steps: form.value.steps.map(s => ({
        attraction_id: s.attraction_id,
        order: s.order,
        duration_minutes: s.duration_minutes,
        highlight: s.highlight,
      })),
    }
    if (editingId.value) {
      await updateRoute(editingId.value, payload)
    } else {
      await createRoute(payload)
    }
    closeDialog()
    await fetchData()
  } catch (e) {
    console.error('Failed to save route:', e)
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await deleteRoute(id)
    deleteConfirmId.value = null
    await fetchData()
  } catch (e) {
    console.error('Failed to delete route:', e)
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="min-h-screen bg-cv-base">
    <AppHeader />

    <main class="max-w-[1200px] mx-auto px-12 py-12">
      <!-- Title -->
      <div class="flex items-start justify-between mb-8">
        <div>
          <h1 class="text-[32px] font-semibold text-cv-text tracking-[-0.5px]">
            {{ t('nav.routes') }}
          </h1>
          <p class="mt-2 text-sm text-cv-text-secondary">
            管理景区路线及途经景点 / Manage scenic routes and waypoints
          </p>
        </div>
        <button @click="openCreate"
                class="px-5 py-2.5 bg-cv-accent text-white text-sm font-medium rounded-cv-md hover:bg-cv-accent-hover transition-colors cursor-pointer">
          + 新增路线
        </button>
      </div>

      <!-- Filters -->
      <div class="flex gap-4 mb-6">
        <input v-model="search" type="text" placeholder="搜索路线名称..."
               class="w-[280px] h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-4 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none transition-all" />
        <select v-model="tagFilter"
                class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
          <option value="">全部标签</option>
          <option v-for="tag in allTags" :key="tag" :value="tag">{{ tag }}</option>
        </select>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="text-center py-20 text-cv-text-muted">
        {{ t('common.loading') }}
      </div>

      <!-- Table -->
      <div v-else-if="filtered.length > 0" class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg overflow-hidden">
        <table class="w-full text-sm">
          <thead>
            <tr class="border-b border-cv-border-subtle text-left">
              <th class="px-4 py-3 text-cv-text-secondary font-medium">名称</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">时长</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">难度</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">途经景点</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">标签</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in filtered" :key="r.id" class="border-b border-cv-border-subtle last:border-b-0 hover:bg-cv-hover transition-colors">
              <td class="px-4 py-3 text-cv-text font-medium">{{ r.name }}</td>
              <td class="px-4 py-3 text-cv-text-secondary">{{ r.duration }}</td>
              <td class="px-4 py-3 text-cv-text-secondary">{{ getDifficultyLabel(r.difficulty) }}</td>
              <td class="px-4 py-3 text-cv-text-secondary">
                <div class="flex flex-wrap gap-1">
                  <span v-for="(step, i) in r.steps" :key="i"
                        class="inline-block px-2 py-0.5 text-xs bg-cv-elevated text-cv-text-secondary rounded-full">
                    {{ step.attraction_name || getAttractionName(step.attraction_id) }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-1">
                  <span v-for="tag in r.tags" :key="tag"
                        class="inline-block px-2 py-0.5 text-xs bg-cv-elevated text-cv-text-secondary rounded-full">
                    {{ tag }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3 text-right">
                <button @click="openEdit(r)" class="text-cv-accent hover:text-cv-accent-hover text-sm cursor-pointer mr-3">编辑</button>
                <button @click="deleteConfirmId = r.id" class="text-cv-danger hover:text-red-400 text-sm cursor-pointer">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty -->
      <div v-else class="flex flex-col items-center justify-center py-24 text-cv-text-muted">
        <p class="text-sm">暂无路线数据</p>
        <button @click="openCreate" class="mt-4 text-cv-accent hover:text-cv-accent-hover text-sm cursor-pointer">
          + 创建第一条路线
        </button>
      </div>
    </main>

    <!-- Create/Edit Dialog -->
    <div v-if="showDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeDialog">
      <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg w-[600px] max-h-[85vh] overflow-y-auto p-6">
        <h2 class="text-lg font-semibold text-cv-text mb-4">
          {{ editingId ? '编辑路线' : '新增路线' }}
        </h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">名称 *</label>
            <input v-model="form.name" type="text" class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">描述 *</label>
            <textarea v-model="form.description" rows="3" class="w-full bg-cv-elevated border border-cv-border rounded-cv-md px-3 py-2 text-sm text-cv-text focus:border-cv-accent focus:outline-none resize-none" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm text-cv-text-secondary mb-1">时长 *</label>
              <input v-model="form.duration" type="text" placeholder="约2小时"
                     class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
            </div>
            <div>
              <label class="block text-sm text-cv-text-secondary mb-1">难度 *</label>
              <select v-model="form.difficulty"
                      class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
                <option v-for="d in difficulties" :key="d.value" :value="d.value">{{ d.label }}</option>
              </select>
            </div>
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">标签（逗号分隔）</label>
            <input v-model="tagInput" type="text" placeholder="亲子, 休闲, 拍照"
                   class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
          </div>

          <!-- Route Steps Editor -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <label class="text-sm text-cv-text-secondary">途经景点 *</label>
              <button @click="addStep" class="text-cv-accent hover:text-cv-accent-hover text-sm cursor-pointer">
                + 添加景点
              </button>
            </div>
            <div v-if="form.steps.length === 0" class="text-sm text-cv-text-muted py-4 text-center border border-dashed border-cv-border rounded-cv-md">
              点击"添加景点"开始规划路线
            </div>
            <div v-else class="space-y-3">
              <div v-for="(step, i) in form.steps" :key="i"
                   class="bg-cv-elevated border border-cv-border rounded-cv-md p-3">
                <div class="flex items-center gap-2 mb-2">
                  <span class="text-xs text-cv-text-muted w-5 text-center">{{ i + 1 }}</span>
                  <select v-model="step.attraction_id"
                          class="flex-1 h-8 bg-cv-base border border-cv-border rounded-cv-md px-2 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
                    <option value="" disabled>选择景点</option>
                    <option v-for="a in allAttractions" :key="a.id" :value="a.id">{{ a.name }}</option>
                  </select>
                  <input v-model.number="step.duration_minutes" type="number" min="1" placeholder="分钟"
                         class="w-20 h-8 bg-cv-base border border-cv-border rounded-cv-md px-2 text-sm text-cv-text text-center focus:border-cv-accent focus:outline-none" />
                  <span class="text-xs text-cv-text-muted">分钟</span>
                  <button @click="moveStepUp(i)" :disabled="i === 0"
                          class="w-6 h-6 flex items-center justify-center text-cv-text-secondary hover:text-cv-text cursor-pointer disabled:opacity-30">↑</button>
                  <button @click="moveStepDown(i)" :disabled="i === form.steps.length - 1"
                          class="w-6 h-6 flex items-center justify-center text-cv-text-secondary hover:text-cv-text cursor-pointer disabled:opacity-30">↓</button>
                  <button @click="removeStep(i)"
                          class="w-6 h-6 flex items-center justify-center text-cv-danger hover:text-red-400 cursor-pointer">×</button>
                </div>
                <input v-model="step.highlight" type="text" placeholder="重点讲解提示（可选）"
                       class="w-full h-7 bg-cv-base border border-cv-border rounded-cv-md px-2 text-xs text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-3 mt-6">
          <button @click="closeDialog"
                  class="px-4 py-2 text-sm text-cv-text-secondary hover:text-cv-text cursor-pointer">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleSave"
                  :disabled="saving || !form.name || !form.description || !form.duration || form.steps.length === 0"
                  class="px-5 py-2 bg-cv-accent text-white text-sm font-medium rounded-cv-md hover:bg-cv-accent-hover transition-colors cursor-pointer disabled:opacity-50">
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm Dialog -->
    <div v-if="deleteConfirmId" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="deleteConfirmId = null">
      <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg w-[400px] p-6">
        <h2 class="text-lg font-semibold text-cv-text mb-2">确认删除</h2>
        <p class="text-sm text-cv-text-secondary mb-6">
          删除路线后不可恢复。如果某角色的推荐路线引用了该路线，请先从角色中移除。
        </p>
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
