<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, shallowRef } from "vue";

import AppButton from "@/components/base/AppButton.vue";
import { useUserPreferenceStore } from "@/stores/userPreference";

type FilterDisplayModeOverride = "drawer" | "inline" | "preference";
type InlineFilterBehavior = "always" | "header-toggle";

interface Props {
  active?: boolean;
  activeLabel?: string;
  displayMode?: FilterDisplayModeOverride;
  icon?: string;
  inlineBehavior?: InlineFilterBehavior;
  label?: string;
  prefix?: string;
  size?: "small" | "medium" | "large";
}

const props = withDefaults(defineProps<Props>(), {
  active: false,
  activeLabel: "",
  displayMode: "preference",
  icon: "pi-filter",
  inlineBehavior: "always",
  label: "",
  prefix: "",
});

const emit = defineEmits<{
  (e: "click"): void;
}>();

const userPreferences = useUserPreferenceStore();
const filterDisplayMode = computed(() => userPreferences.filterDisplayMode);
const isMobile = computed(() => userPreferences.isMobile);
const resolvedLabel = computed(() => props.label || "Filter");
const resolvedActiveLabel = computed(() => props.activeLabel || "Close Filter");

const FILTERS_DB_NAME = "app-filters-db";
const FILTERS_DB_VERSION = 1;
const FILTERS_STORE_NAME = "filters";

// eslint-disable-next-line no-undef
let filtersDbPromise: Promise<IDBDatabase | null> | null = null;
const storedFilterData = shallowRef<Record<string, unknown>>({});
let pollingInterval: ReturnType<typeof setInterval> | null = null;
let customListener: (() => void) | null = null;
const lastDataHash = ref("");
let refreshEpoch = 0;

const storagePrefix = computed(() => {
  const prefix = props.prefix?.trim();
  return "app:filters" + (prefix ? `:${prefix}` : "");
});

// eslint-disable-next-line no-undef
const getFiltersDb = async (): Promise<IDBDatabase | null> => {
  if (typeof window === "undefined" || !window.indexedDB) return null;
  if (filtersDbPromise) return filtersDbPromise;

  filtersDbPromise = new Promise((resolve) => {
    const request = window.indexedDB.open(FILTERS_DB_NAME, FILTERS_DB_VERSION);

    request.onupgradeneeded = () => {
      const db = request.result;
      if (!db.objectStoreNames.contains(FILTERS_STORE_NAME)) {
        db.createObjectStore(FILTERS_STORE_NAME);
      }
    };

    request.onsuccess = () => {
      const db = request.result;
      db.onversionchange = () => db.close();
      resolve(db);
    };

    request.onerror = () => {
      console.warn("Failed to open IndexedDB for filter button", request.error);
      resolve(null);
    };

    request.onblocked = () => {
      console.warn("IndexedDB open blocked for filter button");
    };
  });

  return filtersDbPromise;
};

const readFiltersFromStorage = async () => {
  const prefix = `${storagePrefix.value}:`;
  if (typeof window === "undefined" || !window.IDBKeyRange) return {};

  const db = await getFiltersDb();
  if (!db) return {};

  const data: Record<string, unknown> = {};

  try {
    await new Promise<void>((resolve) => {
      const tx = db.transaction(FILTERS_STORE_NAME, "readonly");
      const store = tx.objectStore(FILTERS_STORE_NAME);
      const range = window.IDBKeyRange.bound(prefix, `${prefix}\uffff`);
      const request = store.openCursor(range);

      request.onsuccess = () => {
        const cursor = request.result;
        if (!cursor) return;

        const key = String(cursor.key);
        if (!key.endsWith(":labels")) {
          data[key.slice(prefix.length)] = cursor.value;
        }
        cursor.continue();
      };

      tx.oncomplete = () => resolve();
      tx.onabort = () => resolve();
      tx.onerror = () => {
        console.warn(
          "Failed to read persisted filters for filter button",
          tx.error
        );
        resolve();
      };
    });
  } catch (error) {
    console.warn("Failed to read persisted filters for filter button", error);
  }

  return data;
};

const createDataHash = (data: Record<string, unknown>) => {
  return JSON.stringify(data);
};

const updateLocalData = async () => {
  const requestEpoch = ++refreshEpoch;
  const newData = await readFiltersFromStorage();
  if (requestEpoch !== refreshEpoch) return;

  const newHash = createDataHash(newData);

  if (newHash !== lastDataHash.value) {
    lastDataHash.value = newHash;
    storedFilterData.value = { ...newData };
  }
};

const activeCount = computed(() => {
  let count = 0;

  Object.values(storedFilterData.value).forEach((value) => {
    if (Array.isArray(value)) {
      count += value.length;
    } else if (typeof value === "string" && value.trim() !== "") {
      count++;
    } else if (typeof value === "boolean" && value === true) {
      count++;
    }
  });

  return count;
});

const hasFilters = computed(() => activeCount.value > 0);

const resolvedDisplayMode = computed(() => {
  if (isMobile.value) return "drawer";
  if (props.displayMode !== "preference") return props.displayMode;
  return filterDisplayMode.value;
});

const shouldShow = computed(() => {
  if (resolvedDisplayMode.value === "drawer") return true;
  return props.inlineBehavior === "header-toggle";
});

const buttonVariant = computed(() => {
  if (props.active) return "dangerSoft";
  return hasFilters.value ? "danger" : "secondary";
});

onMounted(() => {
  void updateLocalData();

  pollingInterval = setInterval(() => {
    void updateLocalData();
  }, 1000);

  customListener = () => {
    void updateLocalData();
  };
  window.addEventListener("filters-updated", customListener);
});

onUnmounted(() => {
  if (customListener) {
    window.removeEventListener("filters-updated", customListener);
    customListener = null;
  }

  if (pollingInterval) {
    clearInterval(pollingInterval);
    pollingInterval = null;
  }
});
</script>

<template>
  <div v-if="shouldShow" class="filter-button-wrapper">
    <AppButton
      type="button"
      :variant="buttonVariant"
      :size="size"
      :icon="icon"
      icon-pos="left"
      @click="emit('click')"
    >
      <span class="filter-button-text">{{
        active ? resolvedActiveLabel : resolvedLabel
      }}</span>
      <span
        class="filter-button-badge"
        :class="{ 'filter-button-badge--active': hasFilters }"
      >
        {{ activeCount }}
      </span>
    </AppButton>
  </div>
</template>

<style scoped>
.filter-button-wrapper {
  position: relative;
  display: inline-flex;
}

.filter-button-text {
  display: inline-flex;
  align-items: center;
}

.filter-button-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.125rem;
  height: 1.125rem;
  padding: 0 0.3rem;
  margin-left: 0.375rem;
  border-radius: 9999px;
  background: rgba(255, 255, 255, 0.75);
  color: #475569;
  font-size: 0.65rem;
  font-weight: 700;
}

.filter-button-badge--active {
  background: white;
  color: #dc2626;
}

@media (max-width: 640px) {
  .filter-button-badge {
    min-width: 1rem;
    height: 1rem;
    font-size: 0.6rem;
  }
}
</style>
