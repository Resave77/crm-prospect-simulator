<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { getProspectPhotoTags, setProspectPhotoTag } from '../api/crm'
import type { PhotoCategory, PlacePhoto } from '../types/crm'
import type { UserRole } from '../types/auth'

const props = defineProps<{
  photos: PlacePhoto[]
  prospectId?: string | null
  role: UserRole
}>()

const tags = ref<Record<string, PhotoCategory>>({})
const loading = ref(true)
const savingName = ref<string | null>(null)
const lightbox = ref<PlacePhoto | null>(null)
const tagError = ref('')

const taggable = computed(() => !!props.prospectId)

const menuPhotos = computed(() => props.photos.filter((p) => tags.value[p.name] === 'MENU'))
const regularPhotos = computed(() => props.photos.filter((p) => tags.value[p.name] !== 'MENU'))

function categoryOf(photo: PlacePhoto): PhotoCategory {
  return tags.value[photo.name] ?? 'PLACE'
}

async function loadTags() {
  if (!taggable.value) {
    loading.value = false
    return
  }
  loading.value = true
  tagError.value = ''
  try {
    const items = await getProspectPhotoTags(props.prospectId!, props.role)
    const map: Record<string, PhotoCategory> = {}
    items.forEach((t) => { map[t.photoName] = t.category })
    tags.value = map
  } catch {
    tags.value = {}
  } finally {
    loading.value = false
  }
}

async function setCategory(photo: PlacePhoto, category: PhotoCategory) {
  if (savingName.value || !taggable.value) return
  savingName.value = photo.name
  tagError.value = ''
  try {
    const item = await setProspectPhotoTag(props.prospectId!, photo.name, category, props.role)
    tags.value = { ...tags.value, [item.photoName]: item.category }
  } catch (caught) {
    tagError.value = (caught as { response?: { data?: { error?: { message?: string } } } }).response?.data?.error?.message
      ?? 'Failed to save photo tag. Please try again.'
  } finally {
    savingName.value = null
  }
}

function onLightboxKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') lightbox.value = null
}

onMounted(loadTags)
watch(() => props.prospectId, () => {
  tags.value = {}
  lightbox.value = null
  loadTags()
})
onBeforeUnmount(() => { lightbox.value = null })
</script>

<template>
  <div class="ppg">
    <div v-if="tagError" class="ppg-error">{{ tagError }}</div>

    <section v-if="photos.length" class="ppg-section">
      <h2><i class="pi pi-book" /> Menu</h2>
      <p v-if="menuPhotos.length" class="ppg-count">{{ menuPhotos.length }} photo{{ menuPhotos.length === 1 ? '' : 's' }} tagged as menu</p>
      <div v-if="menuPhotos.length" class="ppg-scroll">
        <div v-for="photo in menuPhotos" :key="photo.name" class="ppg-item">
          <div class="ppg-thumb" @click="lightbox = photo">
            <img :src="photo.photoUrl" :alt="'Menu photo'" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
            <span class="ppg-badge ppg-badge-menu"><i class="pi pi-book" /> Menu</span>
          </div>
          <button
            v-if="taggable"
            class="ppg-tag-btn"
            :disabled="savingName === photo.name"
            @click="setCategory(photo, 'PLACE')"
          >
            <i v-if="savingName === photo.name" class="pi pi-spin pi-spinner" />
            <i v-else class="pi pi-image" />
            Move to Photos
          </button>
        </div>
      </div>
      <div v-else class="ppg-empty">
        <i class="pi pi-image" />
        <span>No menu photos yet</span>
        <small>Tag a photo as “Menu” below to move it here.</small>
      </div>
    </section>

    <section v-if="photos.length" class="ppg-section">
      <h2><i class="pi pi-images" /> Photos</h2>
      <p v-if="regularPhotos.length" class="ppg-count">{{ regularPhotos.length }} photo{{ regularPhotos.length === 1 ? '' : 's' }}</p>
      <div v-if="regularPhotos.length" class="ppg-scroll">
        <div v-for="photo in regularPhotos" :key="photo.name" class="ppg-item">
          <div class="ppg-thumb" @click="lightbox = photo">
            <img :src="photo.photoUrl" :alt="'Place photo'" loading="lazy" @error="($event.target as HTMLImageElement).style.display='none'" />
            <span v-if="categoryOf(photo) === 'PLACE'" class="ppg-badge ppg-badge-place"><i class="pi pi-images" /> Photo</span>
          </div>
          <button
            v-if="taggable"
            class="ppg-tag-btn"
            :disabled="savingName === photo.name"
            @click="setCategory(photo, 'MENU')"
          >
            <i v-if="savingName === photo.name" class="pi pi-spin pi-spinner" />
            <i v-else class="pi pi-book" />
            Tag as Menu
          </button>
        </div>
      </div>
      <div v-else class="ppg-empty">
        <i class="pi pi-images" />
        <span>All photos tagged as menu</span>
      </div>
    </section>

    <div v-if="lightbox" class="ppg-lightbox" @click.self="lightbox = null">
      <button class="ppg-lightbox-close" aria-label="Close" @click="lightbox = null"><i class="pi pi-times" /></button>
      <img :src="lightbox.photoUrl" :alt="'Photo preview'" @keydown="onLightboxKeydown" />
      <small v-if="lightbox.attribution" class="ppg-lightbox-attr">Photo: {{ lightbox.attribution }}</small>
    </div>
  </div>
</template>

<style scoped>
.ppg { display: grid; gap: 1.1rem; width: 100%; }
.ppg-error {
  padding: 0.5rem 0.75rem; border-radius: 10px; background: #fef2f2;
  color: #991b1b; font-size: 0.72rem;
}

.ppg-section { display: grid; gap: 0.6rem; }
.ppg-section h2 {
  margin: 0; display: flex; align-items: center; gap: 0.4rem;
  font-size: 0.68rem; font-weight: 700; text-transform: uppercase;
  letter-spacing: 0.06em; color: var(--text-muted);
}
.ppg-section h2 i { color: var(--brand-blue); font-size: 0.75rem; }
.ppg-count { margin: 0; color: var(--text-muted); font-size: 0.68rem; }

.ppg-scroll {
  display: flex; gap: 0.6rem; overflow-x: auto; scroll-snap-type: x mandatory;
  -webkit-overflow-scrolling: touch; padding-bottom: 0.3rem;
}
.ppg-scroll::-webkit-scrollbar { height: 4px; }
.ppg-scroll::-webkit-scrollbar-thumb { background: #cbd5e1; border-radius: 4px; }

.ppg-item {
  flex: 0 0 200px; display: flex; flex-direction: column; gap: 0.35rem;
  scroll-snap-align: start; min-width: 0;
}
.ppg-thumb {
  position: relative; height: 150px; border-radius: 12px; overflow: hidden;
  cursor: zoom-in; border: 1px solid var(--border-light); background: #f1f5f9;
}
.ppg-thumb img { width: 100%; height: 100%; object-fit: cover; display: block; }
.ppg-badge {
  position: absolute; left: 0.4rem; bottom: 0.4rem;
  display: inline-flex; align-items: center; gap: 0.25rem;
  padding: 0.15rem 0.5rem; border-radius: 9999px;
  font-size: 0.58rem; font-weight: 700; color: #fff; background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(2px);
}
.ppg-badge-menu i { font-size: 0.5rem; }
.ppg-badge-place i { font-size: 0.5rem; }

.ppg-tag-btn {
  display: inline-flex; align-items: center; justify-content: center; gap: 0.3rem;
  padding: 0.35rem 0.55rem; border: 1px solid var(--border-light); border-radius: 9px;
  background: #fff; color: var(--brand-blue); font-size: 0.66rem; font-weight: 600;
  cursor: pointer; transition: all 0.15s ease;
}
.ppg-tag-btn:hover:not(:disabled) { background: #eff6ff; border-color: #bfdbfe; }
.ppg-tag-btn:disabled { opacity: 0.5; cursor: not-allowed; }

.ppg-empty {
  display: flex; flex-direction: column; align-items: center; gap: 0.35rem;
  padding: 1.6rem 1rem; color: var(--text-muted); text-align: center;
  background: var(--surface-subtle); border-radius: 12px;
}
.ppg-empty i { font-size: 1.4rem; color: #cbd5e1; }
.ppg-empty span { font-size: 0.78rem; font-weight: 600; }
.ppg-empty small { font-size: 0.66rem; }

.ppg-lightbox {
  position: fixed; inset: 0; z-index: 2000;
  display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 0.6rem;
  background: rgba(15, 23, 42, 0.92); padding: 1.5rem;
}
.ppg-lightbox img {
  max-width: 100%; max-height: calc(100dvh - 5rem); border-radius: 12px;
  object-fit: contain; box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}
.ppg-lightbox-close {
  position: absolute; top: 1rem; right: 1rem; width: 40px; height: 40px;
  display: grid; place-items: center; border: 0; border-radius: 50%;
  background: rgba(255, 255, 255, 0.12); color: #fff; font-size: 1rem; cursor: pointer;
}
.ppg-lightbox-close:hover { background: rgba(255, 255, 255, 0.25); }
.ppg-lightbox-attr { color: rgba(255, 255, 255, 0.7); font-size: 0.7rem; font-style: italic; }

@media (max-width: 767px) {
  .ppg-item { flex: 0 0 160px; }
  .ppg-thumb { height: 120px; }
}
</style>
