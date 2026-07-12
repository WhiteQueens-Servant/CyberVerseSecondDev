<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import AppHeader from '../components/AppHeader.vue'
import {
  listAttractions, createAttraction, updateAttraction, deleteAttraction,
  type Attraction, type AttractionForm,
} from '../services/api'

const { t } = useI18n()

// ── State ──
const attractions = ref<Attraction[]>([])
const loading = ref(false)
const search = ref('')
const categoryFilter = ref('')
const showDialog = ref(false)
const editingId = ref<string | null>(null)
const saving = ref(false)
const deleteConfirmId = ref<string | null>(null)

const form = ref<AttractionForm>({
  name: '',
  description: '',
  category: 'natural',
  location: '',
  image_url: '',
  tags: [],
})
const tagInput = ref('')

const categories = [
  { value: 'natural', label: '自然风光' },
  { value: 'historical', label: '历史古迹' },
  { value: 'family', label: '亲子游乐' },
]

// ── Computed ──
const filtered = computed(() => {
  let list = attractions.value
  if (categoryFilter.value) {
    list = list.filter(a => a.category === categoryFilter.value)
  }
  if (search.value) {
    const q = search.value.toLowerCase()
    list = list.filter(a =>
      a.name.toLowerCase().includes(q) || a.description.toLowerCase().includes(q)
    )
  }
  return list
})

// ── Actions ──
async function fetchAttractions() {
  loading.value = true
  try {
    attractions.value = await listAttractions()
  } catch (e) {
    console.error('Failed to load attractions:', e)
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  form.value = { name: '', description: '', category: 'natural', location: '', image_url: '', tags: [] }
  tagInput.value = ''
  showDialog.value = true
}

function openEdit(attraction: Attraction) {
  editingId.value = attraction.id
  form.value = {
    name: attraction.name,
    description: attraction.description,
    category: attraction.category,
    location: attraction.location || '',
    image_url: attraction.image_url || '',
    tags: [...attraction.tags],
  }
  tagInput.value = attraction.tags.join(', ')
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
  editingId.value = null
}

function parseTags() {
  form.value.tags = tagInput.value
    .split(/[,，]/)
    .map(s => s.trim())
    .filter(Boolean)
}

async function handleSave() {
  parseTags()
  if (!form.value.name || !form.value.description) return
  saving.value = true
  try {
    if (editingId.value) {
      await updateAttraction(editingId.value, form.value)
    } else {
      await createAttraction(form.value)
    }
    closeDialog()
    await fetchAttractions()
  } catch (e) {
    console.error('Failed to save attraction:', e)
  } finally {
    saving.value = false
  }
}

async function handleDelete(id: string) {
  try {
    await deleteAttraction(id)
    deleteConfirmId.value = null
    await fetchAttractions()
  } catch (e) {
    console.error('Failed to delete attraction:', e)
  }
}

function getCategoryLabel(value: string): string {
  return categories.find(c => c.value === value)?.label || value
}

onMounted(fetchAttractions)
</script>

<template>
  <div class="min-h-screen bg-cv-base">
    <AppHeader />

    <main class="max-w-[1200px] mx-auto px-12 py-12">
      <!-- Title -->
      <div class="flex items-start justify-between mb-8">
        <div>
          <h1 class="text-[32px] font-semibold text-cv-text tracking-[-0.5px]">
            {{ t('nav.attractions') }}
          </h1>
          <p class="mt-2 text-sm text-cv-text-secondary">
            管理景区景点信息 / Manage scenic attractions
          </p>
        </div>
        <button @click="openCreate"
                class="px-5 py-2.5 bg-cv-accent text-white text-sm font-medium rounded-cv-md hover:bg-cv-accent-hover transition-colors cursor-pointer">
          + 新增景点
        </button>
      </div>

      <!-- Filters -->
      <div class="flex gap-4 mb-6">
        <input v-model="search" type="text" placeholder="搜索景点名称..."
               class="w-[280px] h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-4 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none transition-all" />
        <select v-model="categoryFilter"
                class="h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
          <option value="">全部分类</option>
          <option v-for="cat in categories" :key="cat.value" :value="cat.value">{{ cat.label }}</option>
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
              <th class="px-4 py-3 text-cv-text-secondary font-medium">分类</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">位置</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">简介</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium">标签</th>
              <th class="px-4 py-3 text-cv-text-secondary font-medium text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="a in filtered" :key="a.id" class="border-b border-cv-border-subtle last:border-b-0 hover:bg-cv-hover transition-colors">
              <td class="px-4 py-3 text-cv-text font-medium">{{ a.name }}</td>
              <td class="px-4 py-3 text-cv-text-secondary">{{ getCategoryLabel(a.category) }}</td>
              <td class="px-4 py-3 text-cv-text-secondary">{{ a.location || '—' }}</td>
              <td class="px-4 py-3 text-cv-text-secondary max-w-[200px] truncate">{{ a.description }}</td>
              <td class="px-4 py-3">
                <div class="flex flex-wrap gap-1">
                  <span v-for="tag in a.tags" :key="tag"
                        class="inline-block px-2 py-0.5 text-xs bg-cv-elevated text-cv-text-secondary rounded-full">
                    {{ tag }}
                  </span>
                </div>
              </td>
              <td class="px-4 py-3 text-right">
                <button @click="openEdit(a)" class="text-cv-accent hover:text-cv-accent-hover text-sm cursor-pointer mr-3">编辑</button>
                <button @click="deleteConfirmId = a.id" class="text-cv-danger hover:text-red-400 text-sm cursor-pointer">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty -->
      <div v-else class="flex flex-col items-center justify-center py-24 text-cv-text-muted">
        <p class="text-sm">暂无景点数据</p>
        <button @click="openCreate" class="mt-4 text-cv-accent hover:text-cv-accent-hover text-sm cursor-pointer">
          + 创建第一个景点
        </button>
      </div>
    </main>

    <!-- Create/Edit Dialog -->
    <div v-if="showDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeDialog">
      <div class="bg-cv-surface border border-cv-border-subtle rounded-cv-lg w-[520px] max-h-[80vh] overflow-y-auto p-6">
        <h2 class="text-lg font-semibold text-cv-text mb-4">
          {{ editingId ? '编辑景点' : '新增景点' }}
        </h2>

        <div class="space-y-4">
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">名称 *</label>
            <input v-model="form.name" type="text" class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">简介 *</label>
            <textarea v-model="form.description" rows="3" class="w-full bg-cv-elevated border border-cv-border rounded-cv-md px-3 py-2 text-sm text-cv-text focus:border-cv-accent focus:outline-none resize-none" />
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">分类 *</label>
            <select v-model="form.category" class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text focus:border-cv-accent focus:outline-none">
              <option v-for="cat in categories" :key="cat.value" :value="cat.value">{{ cat.label }}</option>
            </select>
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">位置</label>
            <input v-model="form.location" type="text" placeholder="如：景区东部，莲花湖畔"
                   class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">图片 URL</label>
            <input v-model="form.image_url" type="text" placeholder="https://..."
                   class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
          </div>
          <div>
            <label class="block text-sm text-cv-text-secondary mb-1">标签（逗号分隔）</label>
            <input v-model="tagInput" type="text" placeholder="湖泊, 免费, 适合拍照"
                   class="w-full h-9 bg-cv-elevated border border-cv-border rounded-cv-md px-3 text-sm text-cv-text placeholder:text-cv-text-muted focus:border-cv-accent focus:outline-none" />
          </div>
        </div>

        <div class="flex justify-end gap-3 mt-6">
          <button @click="closeDialog"
                  class="px-4 py-2 text-sm text-cv-text-secondary hover:text-cv-text cursor-pointer">
            {{ t('common.cancel') }}
          </button>
          <button @click="handleSave" :disabled="saving || !form.name || !form.description"
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
          删除景点后不可恢复。如果该景点被路线引用，请先从路线中移除。
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
