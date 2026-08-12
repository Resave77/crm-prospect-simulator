<script setup lang="ts">
import { ref } from 'vue'

const input = ref('')
const chips = ['Ringkas prospek', 'Cari peluang yoghurt', 'Buat pitch visit']

function useChip(text: string) {
  input.value = text
}

function blockSubmit(event?: Event) {
  event?.preventDefault()
}
</script>

<template>
  <article class="tanya-card">
    <div class="tanya-head">
      <p class="ai-eyebrow"><i class="pi pi-comments" /> Tanya AI</p>
      <span>Preview mode</span>
    </div>

    <div class="tanya-messages">
      <div class="tanya-empty">
        <i class="pi pi-sparkles" />
        <strong>Tanya AI belum aktif.</strong>
        <span>Pertanyaan tidak akan dikirim sampai backend generation diaktifkan.</span>
      </div>
    </div>

    <div class="tanya-chips" aria-label="AI prompt suggestions">
      <button v-for="chip in chips" :key="chip" type="button" @click="useChip(chip)">{{ chip }}</button>
    </div>

    <form class="tanya-input-row" @submit="blockSubmit">
      <textarea
        v-model="input"
        rows="2"
        placeholder="Tulis pertanyaan untuk AI..."
        @keydown.enter.exact.prevent
      />
      <button type="submit" disabled><i class="pi pi-send" /> Ask AI</button>
    </form>
  </article>
</template>

<style scoped>
.tanya-card {
  display: grid;
  gap: 0.85rem;
  min-width: 0;
  padding: 1rem;
  border: 1px solid #eadde0;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 8px 24px rgba(73, 34, 41, 0.06);
}

.tanya-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.ai-eyebrow {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  color: #d62839;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.tanya-head span {
  flex-shrink: 0;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  background: #f1e8ea;
  color: #8d7d81;
  font-size: 0.62rem;
  font-weight: 800;
}

.tanya-messages {
  min-height: 150px;
  display: grid;
  place-items: center;
  padding: 0.85rem;
  border: 1px dashed #e6dadd;
  border-radius: 12px;
  background: linear-gradient(180deg, #fff 0, #fcf9f9 100%);
}

.tanya-empty {
  display: grid;
  justify-items: center;
  gap: 0.35rem;
  text-align: center;
}

.tanya-empty i {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 13px;
  background: #fff0f1;
  color: #e63946;
}

.tanya-empty strong {
  color: var(--text-primary);
  font-size: 0.82rem;
}

.tanya-empty span {
  max-width: 260px;
  color: var(--text-muted);
  font-size: 0.74rem;
  line-height: 1.45;
}

.tanya-chips {
  display: flex;
  gap: 0.45rem;
  overflow-x: auto;
  padding-bottom: 0.2rem;
  -webkit-overflow-scrolling: touch;
}

.tanya-chips button {
  min-height: 34px;
  flex: 0 0 auto;
  padding: 0 0.65rem;
  border: 1px solid #eadde0;
  border-radius: 999px;
  background: #fff;
  color: #6b5d61;
  font-size: 0.72rem;
  font-weight: 700;
  cursor: pointer;
}

.tanya-chips button:hover {
  border-color: #f4b3ba;
  background: #fff0f1;
  color: #d62839;
}

.tanya-input-row {
  display: flex;
  align-items: flex-end;
  gap: 0.5rem;
  min-width: 0;
}

.tanya-input-row textarea {
  flex: 1;
  min-width: 0;
  min-height: 42px;
  max-height: 110px;
  resize: vertical;
  padding: 0.6rem 0.7rem;
  border: 1px solid #e6dadd;
  border-radius: 12px;
  background: #fcf9f9;
  color: var(--text-primary);
  font: inherit;
  font-size: 0.78rem;
  line-height: 1.4;
}

.tanya-input-row button {
  min-height: 42px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  flex-shrink: 0;
  padding: 0 0.75rem;
  border: 0;
  border-radius: 12px;
  background: #f1e8ea;
  color: #9b8b8f;
  font-size: 0.72rem;
  font-weight: 800;
  cursor: not-allowed;
}

@media (max-width: 480px) {
  .tanya-input-row {
    display: grid;
  }

  .tanya-input-row button {
    width: 100%;
  }
}
</style>
