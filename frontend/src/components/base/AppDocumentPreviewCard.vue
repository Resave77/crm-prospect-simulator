<script setup lang="ts">
import { computed, ref } from "vue";

import AppButton from "@/components/base/AppButton.vue";
import { isNativeMobilePlatform } from "@/lib/mobile/platform";
import { formatFileSize } from "@/lib/utils/formatters";

const props = withDefaults(
  defineProps<{
    documentExt?: string;
    documentPreviewLoading?: boolean;
    documentPreviewMessage?: string;
    documentPreviewUnavailable?: boolean;
    documentTitle?: string;
    documentUrl?: string;
    fileSize?: number;
    loading?: boolean;
    useLightweightPdfPreview?: boolean;
  }>(),
  {
    documentExt: undefined,
    documentPreviewLoading: false,
    documentPreviewMessage: "",
    documentPreviewUnavailable: false,
    documentTitle: "Document Detail",
    documentUrl: undefined,
    fileSize: undefined,
    loading: false,
    useLightweightPdfPreview: false,
  }
);

const emit = defineEmits<{
  (e: "refresh"): void;
}>();

const isModalOpen = ref(false);
const zoomScale = ref(1);
const panX = ref(0);
const panY = ref(0);

const MAX_MOBILE_PDF_PREVIEW_BYTES = 3 * 1024 * 1024; // 3 MB

const shouldLightweightPdf = computed(() => {
  if (!props.useLightweightPdfPreview) return false;

  if (
    typeof props.fileSize === "number" &&
    !isNaN(props.fileSize) &&
    props.fileSize > 0
  ) {
    return props.fileSize > MAX_MOBILE_PDF_PREVIEW_BYTES;
  }

  return true;
});

const isDownloading = ref(false);

const handleDownloadDocument = async (e?: Event) => {
  if (e) e.preventDefault();
  if (!props.documentUrl) return;

  isDownloading.value = true;
  try {
    const response = await fetch(props.documentUrl);
    if (!response.ok) throw new Error("Download failed");
    const blob = await response.blob();

    const ext = computedExt.value ? `.${computedExt.value}` : "";
    let filename = props.documentTitle || "document";
    if (ext && !filename.toLowerCase().endsWith(ext.toLowerCase())) {
      filename += ext;
    }

    if (isNativeMobilePlatform()) {
      try {
        const { Directory, Filesystem } = await import("@capacitor/filesystem");

        try {
          const perm = await Filesystem.checkPermissions();
          if (perm.publicStorage !== "granted") {
            await Filesystem.requestPermissions();
          }
        } catch (permError) {
          console.warn("Filesystem permission request ignored", permError);
        }

        // eslint-disable-next-line no-undef
        const reader = new FileReader();

        const base64Data = await new Promise<string>((resolve, reject) => {
          reader.onloadend = () => {
            const res = reader.result as string;
            const base64 = res.split(",")[1] || res;
            resolve(base64);
          };
          reader.onerror = reject;
          reader.readAsDataURL(blob);
        });

        const targetPath = `Yummy/${filename}`;

        try {
          await Filesystem.writeFile({
            data: base64Data,
            directory: Directory.Documents,
            path: targetPath,
            recursive: true,
          });
        } catch (writeErr) {
          console.warn(
            "Primary directory write failed, trying fallback Directory.Data",
            writeErr
          );
          await Filesystem.writeFile({
            data: base64Data,
            directory: Directory.Data,
            path: targetPath,
            recursive: true,
          });
        }

        const { Toast } = await import("@capacitor/toast");
        await Toast.show({
          duration: "long",
          text: `File berhasil disimpan di Folder Dokumen/Yummy (${filename})`,
        });
        return;
      } catch (nativeErr) {
        console.warn(
          "Capacitor Filesystem write failed, fallback to browser link",
          nativeErr
        );
      }
    }

    const blobUrl = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = blobUrl;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(blobUrl);
  } catch (error) {
    console.warn("Direct blob download failed, fallback to window.open", error);
    window.open(props.documentUrl, "_blank");
  } finally {
    isDownloading.value = false;
  }
};

const handleOpenPdf = (e?: Event) => {
  if (e) e.preventDefault();
  if (!props.documentUrl) return;

  if (isNativeMobilePlatform()) {
    resetZoom();
    isModalOpen.value = true;
    return;
  }

  window.open(props.documentUrl, "_blank", "noopener,noreferrer");
};

const computedPdfViewerUrl = computed(() => {
  if (!props.documentUrl) return "";
  if (isNativeMobilePlatform()) {
    return `https://docs.google.com/viewer?url=${encodeURIComponent(
      props.documentUrl
    )}&embedded=true`;
  }
  return props.documentUrl;
});

const computedExt = computed(() => {
  if (props.documentExt) return props.documentExt;
  if (!props.documentUrl) return "";
  const urlPath = props.documentUrl.split("?")[0] || "";
  const parts = urlPath.split(".");
  return parts.length > 1 ? parts[parts.length - 1] || "" : "";
});

const resetZoom = () => {
  zoomScale.value = 1;
  panX.value = 0;
  panY.value = 0;
};

const zoomIn = () => {
  if (zoomScale.value < 3.5) {
    zoomScale.value = Number((zoomScale.value + 0.25).toFixed(2));
  }
};

const zoomOut = () => {
  if (zoomScale.value > 0.5) {
    zoomScale.value = Number((zoomScale.value - 0.25).toFixed(2));
    if (zoomScale.value <= 1) {
      panX.value = 0;
      panY.value = 0;
    }
  }
};

let initialTouchDistance = 0;
let initialScale = 1;
let isPinching = false;
let lastTapTime = 0;
let startTouchX = 0;
let startTouchY = 0;
let initialPanX = 0;
let initialPanY = 0;

// eslint-disable-next-line no-undef
const handleTouchStart = (e: TouchEvent) => {
  const t0 = e.touches[0];
  const t1 = e.touches[1];
  if (e.touches.length === 2 && t0 && t1) {
    isPinching = true;
    const dist = Math.hypot(t0.clientX - t1.clientX, t0.clientY - t1.clientY);
    initialTouchDistance = dist;
    initialScale = zoomScale.value;
  } else if (e.touches.length === 1 && t0) {
    startTouchX = t0.clientX;
    startTouchY = t0.clientY;
    initialPanX = panX.value;
    initialPanY = panY.value;
  }
};

// eslint-disable-next-line no-undef
const handleTouchMove = (e: TouchEvent) => {
  const t0 = e.touches[0];
  const t1 = e.touches[1];
  if (
    e.touches.length === 2 &&
    isPinching &&
    initialTouchDistance > 0 &&
    t0 &&
    t1
  ) {
    if (e.cancelable) e.preventDefault();
    const currentDist = Math.hypot(
      t0.clientX - t1.clientX,
      t0.clientY - t1.clientY
    );
    const scaleRatio = currentDist / initialTouchDistance;
    const newScale = Math.min(Math.max(initialScale * scaleRatio, 0.5), 3.5);
    zoomScale.value = Number(newScale.toFixed(2));
  } else if (
    e.touches.length === 1 &&
    t0 &&
    zoomScale.value > 1 &&
    !isPinching
  ) {
    if (e.cancelable) e.preventDefault();
    const deltaX = t0.clientX - startTouchX;
    const deltaY = t0.clientY - startTouchY;
    panX.value = initialPanX + deltaX;
    panY.value = initialPanY + deltaY;
  }
};

// eslint-disable-next-line no-undef
const handleTouchEnd = (e: TouchEvent) => {
  if (e.touches.length === 0) {
    initialTouchDistance = 0;
    if (isPinching) {
      isPinching = false;
      return;
    }
  }

  if (e.changedTouches.length === 1 && !isPinching && e.touches.length === 0) {
    const currentTime = new Date().getTime();
    const tapLength = currentTime - lastTapTime;
    if (tapLength < 300 && tapLength > 0) {
      if (zoomScale.value > 1) {
        resetZoom();
      } else {
        zoomScale.value = 2.0;
      }
    }
    lastTapTime = currentTime;
  }
};

const openModal = () => {
  if (shouldLightweightPdf.value) return;
  resetZoom();
  isModalOpen.value = true;
};

const closeModal = () => {
  isModalOpen.value = false;
};
</script>

<template>
  <div
    class="min-h-0 flex-1 overflow-auto rounded-[14px] border border-[#e2e8f0] bg-[#f8fafc] p-3"
  >
    <div
      v-if="props.documentPreviewLoading"
      class="flex min-h-[280px] items-center justify-center text-[13px] text-[#64748b]"
    >
      <span class="pi pi-spinner pi-spin mr-2 text-[12px]" />
      Sedang menyiapkan preview file...
    </div>

    <div
      v-else-if="props.documentPreviewUnavailable"
      class="rounded-[14px] border border-amber-200 bg-amber-50 p-4 text-[13px] text-amber-900"
    >
      <p>{{ props.documentPreviewMessage || "Preview belum tersedia." }}</p>
      <div class="mt-3">
        <AppButton
          type="button"
          size="small"
          variant="secondary"
          icon="pi-refresh"
          :disabled="props.loading"
          @click="emit('refresh')"
        >
          Refresh
        </AppButton>
      </div>
    </div>

    <div
      v-else-if="props.documentUrl"
      class="group relative min-h-[360px]"
      :class="{ 'cursor-pointer': !shouldLightweightPdf }"
      @click="openModal"
    >
      <button
        v-if="!shouldLightweightPdf"
        type="button"
        class="absolute right-3 top-3 z-10 inline-flex items-center gap-1.5 rounded-xl bg-slate-900/80 px-3 py-1.5 text-[11px] font-bold text-white shadow-md backdrop-blur transition hover:bg-slate-900 active:scale-95"
        title="Open Fullscreen Detail & Zoom"
        @click.stop="openModal"
      >
        <span class="pi pi-search-plus text-xs" />
        <span>Zoom / Full Detail</span>
      </button>

      <template v-if="computedExt.toUpperCase() === 'PDF'">
        <div
          v-if="shouldLightweightPdf"
          class="flex min-h-[320px] flex-col items-center justify-center rounded-[16px] border border-slate-200/80 bg-gradient-to-b from-white to-slate-50/80 p-6 text-center shadow-sm"
        >
          <div
            class="flex h-16 w-16 items-center justify-center rounded-2xl bg-red-50 text-red-600 shadow-sm ring-1 ring-red-100"
          >
            <span class="pi pi-file-pdf text-[28px]" />
          </div>

          <p class="mt-4 text-[14px] font-bold text-slate-800">
            {{ props.documentTitle || "Dokumen PDF" }}
          </p>

          <p class="mt-1 max-w-[320px] text-[12px] text-slate-500">
            Preview langsung dimatikan agar perangkat tetap responsif.
            <span v-if="props.fileSize" class="font-semibold text-slate-700">
              ({{ formatFileSize(props.fileSize) }})
            </span>
          </p>

          <div
            class="mt-5 flex w-full max-w-[300px] flex-col gap-2 sm:max-w-none sm:flex-row sm:justify-center"
          >
            <button
              type="button"
              class="inline-flex h-10 items-center justify-center gap-2 rounded-xl bg-red-600 px-5 text-[12.5px] font-semibold !text-white shadow-md shadow-red-600/20 transition-all hover:bg-red-700 active:scale-[0.98]"
              :style="{ color: '#ffffff' }"
              @click.stop="handleOpenPdf"
            >
              <span
                class="pi pi-external-link text-xs !text-white"
                :style="{ color: '#ffffff' }"
              />
              <span class="!text-white" :style="{ color: '#ffffff' }">
                Buka Dokumen PDF
              </span>
            </button>

            <button
              type="button"
              class="inline-flex h-10 items-center justify-center gap-2 rounded-xl border border-slate-200 bg-white px-4 text-[12.5px] font-semibold text-slate-700 shadow-sm transition-all hover:bg-slate-50 hover:text-slate-900 active:scale-[0.98] disabled:opacity-60"
              :disabled="isDownloading"
              @click.stop="handleDownloadDocument"
            >
              <span
                :class="[
                  'pi text-xs',
                  isDownloading ? 'pi-spinner pi-spin' : 'pi-download',
                ]"
              />
              <span>{{ isDownloading ? "Mengunduh..." : "Unduh File" }}</span>
            </button>
          </div>
        </div>
        <iframe
          v-else
          :src="computedPdfViewerUrl"
          class="h-[min(62vh,720px)] min-h-[420px] w-full rounded-[12px] border border-[#e2e8f0] bg-white xl:h-[min(70vh,720px)]"
        />
      </template>

      <img
        v-else
        :src="props.documentUrl"
        loading="lazy"
        decoding="async"
        class="max-h-[min(62vh,720px)] min-h-[320px] w-full rounded-[12px] object-contain xl:max-h-[min(70vh,720px)]"
      />
    </div>

    <!-- Document Zoom Detail Modal -->
    <Teleport to="body">
      <div
        v-if="isModalOpen && props.documentUrl"
        class="fixed inset-0 z-[9999] flex flex-col bg-slate-950/90 pt-[env(safe-area-inset-top)] pb-[env(safe-area-inset-bottom)] pl-[env(safe-area-inset-left)] pr-[env(safe-area-inset-right)] backdrop-blur-md transition-all"
      >
        <!-- Modal Header Toolbar -->
        <div
          class="flex shrink-0 items-center justify-between border-b border-slate-800 bg-slate-900/90 px-4 py-3 text-white"
        >
          <div class="min-w-0 flex-1 pr-3">
            <p class="truncate text-[13px] font-bold text-slate-200">
              {{ props.documentTitle }}
            </p>
          </div>

          <div class="flex items-center gap-2">
            <div class="flex items-center rounded-xl bg-slate-800 p-1">
              <button
                type="button"
                class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-300 transition hover:bg-slate-700 hover:text-white disabled:opacity-40"
                :disabled="zoomScale <= 0.5"
                aria-label="Zoom Out"
                @click="zoomOut"
              >
                <span class="pi pi-minus text-xs" />
              </button>
              <span
                class="w-12 text-center text-[11px] font-bold text-slate-300"
              >
                {{ Math.round(zoomScale * 100) }}%
              </span>
              <button
                type="button"
                class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-300 transition hover:bg-slate-700 hover:text-white disabled:opacity-40"
                :disabled="zoomScale >= 3.5"
                aria-label="Zoom In"
                @click="zoomIn"
              >
                <span class="pi pi-plus text-xs" />
              </button>
              <button
                type="button"
                class="ml-1 flex h-8 items-center justify-center rounded-lg px-2 text-[10px] font-bold text-slate-300 transition hover:bg-slate-700 hover:text-white"
                aria-label="Reset Zoom"
                @click="resetZoom"
              >
                Reset
              </button>
            </div>

            <button
              type="button"
              class="flex h-8 items-center gap-1.5 rounded-xl border border-slate-700 bg-slate-800 px-3 text-[11px] font-semibold text-white shadow-sm transition hover:bg-slate-700 disabled:opacity-60"
              :disabled="isDownloading"
              @click.stop="handleDownloadDocument"
            >
              <span
                :class="[
                  'pi text-xs',
                  isDownloading ? 'pi-spinner pi-spin' : 'pi-download',
                ]"
              />
              <span class="hidden sm:inline">
                {{ isDownloading ? "Mengunduh..." : "Download" }}
              </span>
            </button>

            <button
              type="button"
              class="flex h-9 w-9 items-center justify-center rounded-xl bg-red-600/90 text-white shadow-sm transition hover:bg-red-600 active:scale-95"
              aria-label="Close modal"
              @click="closeModal"
            >
              <span class="pi pi-times text-sm" />
            </button>
          </div>
        </div>

        <!-- Modal Viewport -->
        <div
          class="relative flex min-h-0 flex-1 items-center justify-center overflow-auto p-4"
          @touchstart="handleTouchStart"
          @touchmove="handleTouchMove"
          @touchend="handleTouchEnd"
        >
          <img
            v-if="computedExt.toUpperCase() !== 'PDF'"
            :src="props.documentUrl"
            alt="Document Detail"
            class="max-h-full max-w-full rounded-lg object-contain transition-transform duration-75 ease-out"
            :style="{
              transform: `translate(${panX}px, ${panY}px) scale(${zoomScale})`,
            }"
          />
          <iframe
            v-else
            :src="computedPdfViewerUrl"
            class="h-full w-full rounded-lg border-0 bg-white shadow-2xl transition-transform duration-75 ease-out"
            :style="{
              transform: `translate(${panX}px, ${panY}px) scale(${zoomScale})`,
            }"
          />
        </div>
      </div>
    </Teleport>
  </div>
</template>
