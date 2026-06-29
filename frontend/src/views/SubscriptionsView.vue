<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto">
      <div class="p-[30px] space-y-7">
        <!-- Header -->
        <div class="flex items-start justify-between">
          <div class="space-y-1">
            <h1 class="text-[28px] font-bold tracking-tight text-foreground">Subscriptions</h1>
            <p class="text-sm text-muted-foreground">Manage remote proxy subscriptions</p>
          </div>
          <button
            v-if="store.subscriptions.length > 0"
            :disabled="isUpdatingAll"
            class="h-9 px-4 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2"
            @click="updateAll"
          >
            <RefreshCw
              class="w-4 h-4"
              :class="{ 'animate-spin': isUpdatingAll }"
            />
            {{ isUpdatingAll ? 'Updating...' : 'Update All' }}
          </button>
        </div>

        <!-- Subscriptions Section -->
        <SectionCard
          title="Subscriptions"
          :icon="Link"
          icon-bg-class="bg-blue-500"
          :count="store.subscriptions.length"
          :collapsed="collapsed"
          @update:collapsed="collapsed = $event"
        >
          <template #actions>
            <button
              v-if="store.subscriptions.length > 0"
              class="h-7 px-3 rounded-md border border-border bg-card text-xs font-semibold text-red-500 hover:bg-red-50 hover:text-red-600 hover:border-red-200 dark:hover:bg-red-950 dark:hover:text-red-400 dark:hover:border-red-800 transition-all duration-150 inline-flex items-center gap-1.5"
              @click.stop="showDeleteAll = true"
            >
              <Trash2 class="w-3 h-3" />
              Delete All
            </button>
            <button
              class="h-7 px-3 rounded-md bg-blue-600 text-xs font-semibold text-white hover:bg-blue-700 transition-all duration-150 inline-flex items-center gap-1.5"
              @click.stop="showAdd = true"
            >
              <Plus class="w-3 h-3" />
              Add
            </button>
          </template>

          <!-- Empty State -->
          <div v-if="store.subscriptions.length === 0" class="flex flex-col items-center py-7 gap-3.5">
            <p class="text-sm text-muted-foreground">No Subscriptions</p>
            <button
              class="h-8 px-4 rounded-lg border border-border bg-card text-sm font-medium text-card-foreground hover:bg-accent hover:text-accent-foreground transition-all duration-150"
              @click="showAdd = true"
            >
              Add Subscription
            </button>
          </div>

          <!-- Subscription List -->
          <div v-else>
            <div
              v-for="sub in store.subscriptions"
              :key="sub.id"
              class="flex items-center gap-2.5 px-[18px] py-3 cursor-pointer select-none transition-colors duration-150 hover:bg-accent/50"
              :class="{ 'bg-accent/30': store.selectedSubID === sub.id }"
              @click="selectSub(sub.id)"
            >
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-card-foreground truncate">{{ sub.name }}</p>
                <div class="flex items-center gap-1 text-xs text-muted-foreground mt-0.5">
                  <span>{{ sub.nodes.length }} nodes</span>
                  <span v-if="sub.lastUpdated">· Updated {{ getTimeAgo(sub.lastUpdated) }}</span>
                </div>
              </div>

              <div class="flex items-center gap-1.5 flex-shrink-0">
                <!-- Copy URL -->
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-accent hover:text-accent-foreground"
                  :class="copiedID === sub.id ? 'border-emerald-300 bg-emerald-50 text-emerald-600 dark:border-emerald-700 dark:bg-emerald-950 dark:text-emerald-400' : ''"
                  title="Copy subscription URL"
                  @click.stop="copyURL(sub)"
                >
                  <Check v-if="copiedID === sub.id" class="w-3.5 h-3.5" />
                  <Copy v-else class="w-3.5 h-3.5" />
                </button>

                <!-- Edit -->
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-accent hover:text-accent-foreground"
                  title="Edit subscription"
                  @click.stop="startEdit(sub)"
                >
                  <Pencil class="w-3.5 h-3.5" />
                </button>

                <!-- Refresh -->
                <button
                  :disabled="refreshingID === sub.id"
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-accent hover:text-accent-foreground disabled:opacity-50 disabled:cursor-not-allowed"
                  :title="refreshingID === sub.id ? 'Refreshing...' : 'Refresh subscription'"
                  @click.stop="refreshOne(sub.id)"
                >
                  <RefreshCw
                    class="w-3.5 h-3.5"
                    :class="{ 'animate-spin': refreshingID === sub.id }"
                  />
                </button>

                <!-- Delete -->
                <button
                  class="w-8 h-8 rounded-lg border border-border bg-card flex items-center justify-center transition-all duration-150 hover:bg-red-50 hover:text-red-600 hover:border-red-200 dark:hover:bg-red-950 dark:hover:text-red-400 dark:hover:border-red-800"
                  title="Delete subscription"
                  @click.stop="confirmDelete(sub)"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </SectionCard>
      </div>
    </div>

    <!-- Add Subscription Modal -->
    <ModalSheet
      v-model="showAdd"
      title="Add Subscription"
      :icon="Link2"
      icon-bg-class="bg-blue-500"
      action-label="Add"
      action-btn-class="bg-blue-600 hover:bg-blue-700"
      :disabled="!addFormValid"
      :loading="addLoading"
      @action="handleAdd"
    >
      <div class="space-y-5">
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Name</label>
          <input
            v-model="addName"
            type="text"
            placeholder="My Subscription"
            class="h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
            @keydown.enter="addFormValid && handleAdd()"
          />
        </div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">URL</label>
          <input
            ref="addUrlInput"
            v-model="addUrl"
            type="url"
            placeholder="https://example.com/sub"
            class="h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
            @keydown.enter="addFormValid && handleAdd()"
          />
        </div>
        <p v-if="addError" class="text-xs text-red-500 flex items-center gap-1.5">
          <AlertTriangle class="w-3.5 h-3.5 flex-shrink-0" />
          {{ addError }}
        </p>
      </div>
    </ModalSheet>

    <!-- Edit Subscription Modal -->
    <ModalSheet
      v-model="showEdit"
      title="Edit Subscription"
      :icon="Pencil"
      icon-bg-class="bg-blue-500"
      action-label="Save"
      action-btn-class="bg-blue-600 hover:bg-blue-700"
      :disabled="!editFormValid"
      :loading="editLoading"
      @action="handleEdit"
    >
      <div class="space-y-5">
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">Name</label>
          <input
            v-model="editName"
            type="text"
            placeholder="Name"
            class="h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
            @keydown.enter="editFormValid && handleEdit()"
          />
        </div>
        <div class="space-y-1.5">
          <label class="text-xs font-medium text-muted-foreground">URL</label>
          <input
            v-model="editUrl"
            type="url"
            placeholder="URL"
            class="h-9 w-full rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1 transition-shadow duration-150"
            @keydown.enter="editFormValid && handleEdit()"
          />
        </div>
      </div>
    </ModalSheet>

    <!-- Delete Confirmation -->
    <ConfirmDialog
      v-model="showDeleteConfirm"
      :title="deleteTarget ? `Delete Subscription?` : 'Delete All Subscriptions?'"
      :message="deleteTarget
        ? `This will remove '${deleteTarget.name}' and its ${deleteTarget.nodes.length} nodes. This cannot be undone.`
        : `This will remove all ${store.subscriptions.length} subscriptions. This cannot be undone.`"
      confirm-label="Delete"
      :loading="deleteLoading"
      @confirm="handleDelete"
    />

    <!-- Delete All Confirmation -->
    <ConfirmDialog
      v-model="showDeleteAll"
      title="Delete All Subscriptions?"
      :message="`This will remove all ${store.subscriptions.length} subscriptions. This cannot be undone.`"
      confirm-label="Delete All"
      :loading="deleteLoading"
      @confirm="handleDeleteAll"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useNotifications } from '@/composables/useNotifications'
import { relativeTime } from '@/lib/utils'
import type { Subscription } from '@/types'
import SectionCard from '@/components/ui/SectionCard.vue'
import ModalSheet from '@/components/ui/ModalSheet.vue'
import ConfirmDialog from '@/components/ui/ConfirmDialog.vue'
import {
  Link,
  Link2,
  Plus,
  Pencil,
  Trash2,
  RefreshCw,
  Copy,
  Check,
  AlertTriangle,
} from 'lucide-vue-next'

const store = useAppStore()
const notify = useNotifications()

// --- Section collapse ---
const collapsed = ref(false)

// --- Timestamp refresh ---
let timestampTimer: ReturnType<typeof setInterval> | null = null
const tick = ref(0)

// Update tick every 60 seconds for auto-refresh
onMounted(() => {
  timestampTimer = setInterval(() => { tick.value++ }, 60000)
})
onUnmounted(() => {
  if (timestampTimer) clearInterval(timestampTimer)
})

// Also update tick when subscriptions change (after refresh)
watch(() => store.subscriptions, () => { tick.value++ }, { deep: true })

// force recomputation of relativeTime
watch(tick, () => {})

// Reactive wrapper that depends on tick for auto-refresh
function getTimeAgo(dateStr: string): string {
  void tick.value // track dependency
  return relativeTime(dateStr)
}

// --- Copy URL ---
const copiedID = ref<string | null>(null)
let copyTimer: ReturnType<typeof setTimeout> | null = null

function copyURL(sub: Subscription) {
  navigator.clipboard.writeText(sub.url).then(() => {
    copiedID.value = sub.id
    if (copyTimer) clearTimeout(copyTimer)
    copyTimer = setTimeout(() => { copiedID.value = null }, 1500)
  }).catch(() => {
    notify.error('Failed to copy', 'Could not access clipboard')
  })
}

// --- Select ---
function selectSub(id: string) {
  store.setSelectedSubID(id)
}

// --- Add ---
const showAdd = ref(false)
const addName = ref('')
const addUrl = ref('')
const addError = ref('')
const addLoading = ref(false)
const addUrlInput = ref<HTMLInputElement>()

const addFormValid = computed(() => {
  if (!addUrl.value) return false
  try {
    const u = new URL(addUrl.value)
    return !!u.protocol && !!u.host
  } catch {
    return false
  }
})

watch(showAdd, (open) => {
  if (open) {
    addName.value = ''
    addUrl.value = ''
    addError.value = ''
    nextTick(() => addUrlInput.value?.focus())
  }
})

async function handleAdd() {
  if (!addFormValid.value) {
    addError.value = 'Please enter a valid URL'
    return
  }
  addLoading.value = true
  addError.value = ''
  const name = addName.value.trim() || new URL(addUrl.value).host || 'Subscription'
  await store.addSubscription(name, addUrl.value)
  addLoading.value = false
  showAdd.value = false
}

// --- Edit ---
const showEdit = ref(false)
const editID = ref('')
const editName = ref('')
const editUrl = ref('')
const editLoading = ref(false)

const editFormValid = computed(() => editName.value.trim().length > 0)

function startEdit(sub: Subscription) {
  editID.value = sub.id
  editName.value = sub.name
  editUrl.value = sub.url
  showEdit.value = true
}

async function handleEdit() {
  if (!editFormValid.value) return
  editLoading.value = true
  await store.updateSubscription(editID.value, editName.value.trim(), editUrl.value)
  editLoading.value = false
  showEdit.value = false
}

// --- Refresh ---
const refreshingID = computed(() => store.refreshingSubID)
const isUpdatingAll = computed(() => store.refreshingAll)

async function refreshOne(id: string) {
  await store.refreshSubscription(id)
}

async function updateAll() {
  await store.refreshAllSubscriptions()
}

// --- Delete ---
const showDeleteAll = ref(false)
const showDeleteConfirm = ref(false)
const deleteTarget = ref<Subscription | null>(null)
const deleteLoading = ref(false)

function confirmDelete(sub: Subscription) {
  deleteTarget.value = sub
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!deleteTarget.value) return
  deleteLoading.value = true
  await store.deleteSubscription(deleteTarget.value.id)
  deleteLoading.value = false
  showDeleteConfirm.value = false
  deleteTarget.value = null
}

async function handleDeleteAll() {
  deleteLoading.value = true
  const ids = store.subscriptions.map(s => s.id)
  for (const id of ids) {
    await store.deleteSubscription(id)
  }
  deleteLoading.value = false
  showDeleteAll.value = false
}
</script>
