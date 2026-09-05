<!-- eslint-disable no-undef -->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";

import type { DialogMobileBottomBarAction } from "@/composables/useDialogMobileBottomBar";

import AppButton from "@/components/base/AppButton.vue";
import AppDialog from "@/components/base/AppDialog.vue";
import { isNativeMobilePlatform } from "@/lib/mobile/platform";

type CropFilter =
  | "blackWhite"
  | "cleanPaper"
  | "contrast"
  | "documentScan"
  | "grayscale"
  | "magicColor"
  | "original";

type CropRect = {
  height: number;
  width: number;
  x: number;
  y: number;
};

type CropViewMode = "edit" | "preview";

type CropCornerKey = "bottomLeft" | "bottomRight" | "topLeft" | "topRight";

type CropPoint = {
  x: number;
  y: number;
};

type CropCorners = Record<CropCornerKey, CropPoint>;

type QueuedImageFile = {
  file: File;
  index: number;
};

type CameraQualityTone = "danger" | "success" | "warning";

type CameraQualityHint = {
  detail: string;
  label: string;
  tone: CameraQualityTone;
};

const props = withDefaults(
  defineProps<{
    accept?: string;
    autoCapture?: boolean;
    autoCaptureDelayMs?: number;
    cameraSubtitle?: string;
    cameraTitle?: string;
    cropSubtitle?: string;
    cropTitle?: string;
    documentPreview?: string | null;
    enableScanner?: boolean;
    maxOutputHeight?: number;
    maxOutputWidth?: number;
    maxSourceHeight?: number;
    maxSourceWidth?: number;
    multiple?: boolean;
    quality?: number;
  }>(),
  {
    accept: "application/pdf,image/*,text/html",
    autoCapture: false,
    autoCaptureDelayMs: 1600,
    cameraSubtitle: "Arahkan kamera ke dokumen, lalu capture.",
    cameraTitle: "Scan Document",
    cropSubtitle: "Drag the box or resize from the corners, then apply crop.",
    cropTitle: "Adjust Document Crop",
    documentPreview: null,
    enableScanner: false,
    maxOutputHeight: 3000,
    maxOutputWidth: 2200,
    maxSourceHeight: 1800,
    maxSourceWidth: 1400,
    multiple: false,
    quality: 0.92,
  }
);

const emit = defineEmits<{
  (e: "error", message: string | null): void;
  (e: "select", event: Event): void;
  (e: "sourceFiles", files: File[]): void;
  (e: "update:canEdit", value: boolean): void;
}>();

const fileInput = ref<HTMLInputElement | null>(null);
const cameraInput = ref<HTMLInputElement | null>(null);
const scannerError = ref<string | null>(null);
const isProcessingImage = ref(false);
const isGeneratingCropPreview = ref(false);
const isDetectingCrop = ref(false);
const isRotatingImage = ref(false);
const isCameraStarting = ref(false);
const isAutoCapturing = ref(false);
const autoCaptureProgress = ref(0);
const isTorchSupported = ref(false);
const isTorchOn = ref(false);
const cameraDialogOpen = ref(false);
const cameraQualityHint = ref<CameraQualityHint>({
  detail: "Make sure the whole paper is visible",
  label: "Point camera at the document",
  tone: "warning",
});
const cameraVideoRef = ref<HTMLVideoElement | null>(null);
const cameraCanvasRef = ref<HTMLCanvasElement | null>(null);
const cropDialogOpen = ref(false);
const cropViewMode = ref<CropViewMode>("edit");
const canRetakeCropPhoto = ref(false);
const cropImageUrl = ref<string | null>(null);
const cropPreviewImageUrl = ref<string | null>(null);
const cropSourceFile = ref<File | null>(null);
const cropFrameRef = ref<HTMLDivElement | null>(null);
const cropRect = ref<CropRect>({ height: 90, width: 90, x: 5, y: 5 });
const cropCorners = ref<CropCorners>({
  bottomLeft: { x: 5, y: 95 },
  bottomRight: { x: 95, y: 95 },
  topLeft: { x: 5, y: 5 },
  topRight: { x: 95, y: 5 },
});
const selectedCropFilter = ref<CropFilter>("original");
const cropPreviewCacheKey = ref<string | null>(null);
const showCropDebug = ref(false);
const cropDetectionDebug = ref<CropDetectionDebug | null>(null);
const editableDocumentSourceFile = ref<File | null>(null);
const activeQueuedImageFile = ref<QueuedImageFile | null>(null);
const queuedImageFiles = ref<QueuedImageFile[]>([]);
const processedQueuedFiles = ref<Array<File | null>>([]);
const processedQueuedSourceFiles = ref<Array<File | null>>([]);
const lastAppliedCropRect = ref<CropRect>({
  height: 90,
  width: 90,
  x: 5,
  y: 5,
});
const lastAppliedCropCorners = ref<CropCorners>({
  bottomLeft: { x: 5, y: 95 },
  bottomRight: { x: 95, y: 95 },
  topLeft: { x: 5, y: 5 },
  topRight: { x: 95, y: 5 },
});
const lastAppliedCropFilter = ref<CropFilter>("original");

const cropFilterOptions: { icon: string; label: string; value: CropFilter }[] =
  [
    { icon: "pi-file-check", label: "Document Scan", value: "documentScan" },
    { icon: "pi-file", label: "Clean Paper", value: "cleanPaper" },
    { icon: "pi-sparkles", label: "Magic Color", value: "magicColor" },
    { icon: "pi-image", label: "Original", value: "original" },
    { icon: "pi-circle", label: "Gray Scale", value: "grayscale" },
    { icon: "pi-sun", label: "High Contrast", value: "contrast" },
    { icon: "pi-stop", label: "Black & White", value: "blackWhite" },
  ];

let cameraFrameTimer: number | null = null;
let cameraStream: MediaStream | null = null;
let cropPointerCleanup: (() => void) | null = null;
let autoCaptureStartedAt: number | null = null;
let cropPreviewGenerationId = 0;

const CAMERA_HIGHLIGHT_INTERVAL_MS = 500;
const CAMERA_HIGHLIGHT_MAX_WIDTH = 480;
const OPENCV_SCRIPT_ID = "opencv-js-document-cropper";
const OPENCV_SCRIPT_SRC = "https://docs.opencv.org/4.7.0/opencv.js";
const JSCANIFY_SCRIPT_ID = "jscanify-document-cropper";
const JSCANIFY_SCRIPT_SRC =
  "https://cdn.jsdelivr.net/gh/ColonelParrot/jscanify@master/src/jscanify.min.js";

const isProcessingQueue = computed(
  () => !!activeQueuedImageFile.value || queuedImageFiles.value.length > 0
);

const canEditSelectedImage = computed(
  () =>
    !isProcessingQueue.value &&
    !!props.documentPreview &&
    !!editableDocumentSourceFile.value
);

const autoCaptureCountdownText = computed(() => {
  if (!props.autoCapture || autoCaptureProgress.value <= 0) return "";

  const remainingMs = Math.max(
    0,
    props.autoCaptureDelayMs * (1 - autoCaptureProgress.value / 100)
  );

  return `Auto capture in ${Math.ceil(remainingMs / 1000)}s`;
});

const cameraQualityHintClass = computed(() => {
  switch (cameraQualityHint.value.tone) {
    case "danger":
      return "border-red-200 bg-red-50 text-red-700";
    case "success":
      return "border-green-200 bg-green-50 text-green-700";
    default:
      return "border-amber-200 bg-amber-50 text-amber-700";
  }
});

const cameraDialogMobileActions = computed<DialogMobileBottomBarAction[]>(
  () => [
    ...(isTorchSupported.value
      ? [
          {
            active: isTorchOn,
            disabled: isCameraStarting,
            icon: "pi-bolt",
            key: "toggle-torch",
            label: isTorchOn.value ? "Flash On" : "Flash",
            onClick: () => {
              void toggleTorch();
            },
            permissions: [],
            type: "button" as const,
          },
        ]
      : []),
    {
      active: true,
      disabled: computed(() => isCameraStarting.value || isAutoCapturing.value),
      icon: "pi-camera",
      key: "capture-camera",
      label: isAutoCapturing.value ? "Capturing" : "Capture",
      loading: isAutoCapturing,
      onClick: () => {
        void captureCameraFrame();
      },
      permissions: [],
      type: "button",
    },
  ]
);

const displayedCropImageUrl = computed(() =>
  cropViewMode.value === "preview" && cropPreviewImageUrl.value
    ? cropPreviewImageUrl.value
    : cropImageUrl.value
);

const getCropPolygonPoints = (corners: CropCorners) =>
  [corners.topLeft, corners.topRight, corners.bottomRight, corners.bottomLeft]
    .map((point) => `${point.x},${point.y}`)
    .join(" ");

const cropPolygonPoints = computed(() =>
  getCropPolygonPoints(cropCorners.value)
);

const rawDebugPolygonPoints = computed(() =>
  cropDetectionDebug.value
    ? getCropPolygonPoints(cropDetectionDebug.value.rawCorners)
    : ""
);

const refinedDebugPolygonPoints = computed(() =>
  cropDetectionDebug.value
    ? getCropPolygonPoints(cropDetectionDebug.value.refinedCorners)
    : ""
);

const cropFilterCss = computed(() => {
  switch (selectedCropFilter.value) {
    case "blackWhite":
    case "cleanPaper":
    case "documentScan":
      return "url(#doc-scan-svg-filter)";
    case "grayscale":
      return "grayscale(1)";
    case "magicColor":
      return "contrast(1.35) saturate(1.45) brightness(1.12)";
    default:
      return "none";
  }
});

const createCropPreviewCacheKey = (sourceFile: File) => {
  const roundedCorners = Object.entries(cropCorners.value)
    .map(([key, point]) => `${key}:${point.x.toFixed(2)},${point.y.toFixed(2)}`)
    .join("|");

  return [
    sourceFile.name,
    sourceFile.size,
    sourceFile.lastModified,
    selectedCropFilter.value,
    roundedCorners,
  ].join("::");
};

const hasValidCropPreviewCache = (sourceFile: File) =>
  !!cropPreviewImageUrl.value &&
  cropPreviewCacheKey.value === createCropPreviewCacheKey(sourceFile);

const updateCropFilter = (filter: CropFilter) => {
  if (selectedCropFilter.value === filter) return;

  selectedCropFilter.value = filter;
  clearCropPreview({ regeneratePreview: cropViewMode.value === "preview" });
};

const handleCropFilterChange = (event: Event) => {
  const target = event.target as HTMLSelectElement | null;
  if (!target) return;

  updateCropFilter(target.value as CropFilter);
};

const cropDialogMobileActions = computed<DialogMobileBottomBarAction[]>(() => [
  ...(canRetakeCropPhoto.value
    ? [
        {
          disabled: computed(
            () => isProcessingImage.value || isDetectingCrop.value
          ),
          icon: "pi-camera",
          key: "retake-photo",
          label: "Retake",
          onClick: retakePhoto,
          permissions: [],
          type: "button" as const,
        },
      ]
    : []),

  {
    disabled: computed(() => isDetectingCrop.value || isRotatingImage.value),
    icon: "pi-undo",
    key: "rotate-crop",
    label: "Rotate",
    loading: isRotatingImage,
    onClick: rotateCropImage,
    permissions: [],
    type: "button",
  },
  {
    active: computed(() => selectedCropFilter.value !== "original"),
    icon: "pi-sliders-h",
    items: cropFilterOptions.map((option) => ({
      icon: option.icon,
      key: `filter-${option.value}`,
      label:
        selectedCropFilter.value === option.value
          ? `${option.label} ✓`
          : option.label,
      onClick: () => updateCropFilter(option.value),
      permissions: [],
    })),
    key: "crop-filter",
    label: "Filter",
    panelTitle: "Crop Filter",
    permissions: [],
    type: "submenu",
  },
  {
    icon: "pi-image",
    key: "use-original",
    label: "Original",
    onClick: useOriginalImage,
    permissions: [],
    type: "button",
  },
  {
    active: true,
    icon: "pi-check",
    key: "apply-crop",
    label: "Apply",
    loading: isProcessingImage,
    onClick: applyCrop,
    permissions: [],
    type: "button",
  },
]);

const setError = (message: string | null) => {
  scannerError.value = message;
  emit("error", message);
};

const openFilePicker = () => fileInput.value?.click();
const triggerCameraPick = () => cameraInput.value?.click();

const isImageFile = (file: File) =>
  file.type.startsWith("image/") ||
  /\.(jpe?g|png|webp|gif|bmp|tiff?)$/i.test(file.name);

const createFilesEvent = (files: File[]) =>
  ({
    target: {
      files,
      value: "",
    },
  }) as unknown as Event;

const createFileEvent = (file: File) => createFilesEvent([file]);

const loadImageFromFile = (file: File) =>
  new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image();
    const objectUrl = URL.createObjectURL(file);

    image.onload = () => {
      URL.revokeObjectURL(objectUrl);
      resolve(image);
    };
    image.onerror = () => {
      URL.revokeObjectURL(objectUrl);
      reject(new Error("Failed to read image."));
    };
    image.src = objectUrl;
  });

const getExportCanvas = (canvas: HTMLCanvasElement) => {
  const maxOutputWidth = Math.max(0, props.maxOutputWidth || 0);
  const maxOutputHeight = Math.max(0, props.maxOutputHeight || 0);
  const widthRatio = maxOutputWidth ? maxOutputWidth / canvas.width : 1;
  const heightRatio = maxOutputHeight ? maxOutputHeight / canvas.height : 1;
  const scaleRatio = Math.min(1, widthRatio, heightRatio);

  if (scaleRatio >= 1) return canvas;

  const exportCanvas = document.createElement("canvas");
  const context = exportCanvas.getContext("2d");
  if (!context) return canvas;

  exportCanvas.width = Math.max(1, Math.round(canvas.width * scaleRatio));
  exportCanvas.height = Math.max(1, Math.round(canvas.height * scaleRatio));
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.fillStyle = "#ffffff";
  context.fillRect(0, 0, exportCanvas.width, exportCanvas.height);
  context.drawImage(canvas, 0, 0, exportCanvas.width, exportCanvas.height);

  return exportCanvas;
};

const resizeImageFileForProcessing = async (file: File) => {
  if (!isImageFile(file)) return file;

  const image = await loadImageFromFile(file);
  const sourceWidth = image.naturalWidth || image.width;
  const sourceHeight = image.naturalHeight || image.height;
  const maxSourceWidth = Math.max(0, props.maxSourceWidth || 0);
  const maxSourceHeight = Math.max(0, props.maxSourceHeight || 0);
  const widthRatio = maxSourceWidth ? maxSourceWidth / sourceWidth : 1;
  const heightRatio = maxSourceHeight ? maxSourceHeight / sourceHeight : 1;
  const scaleRatio = Math.min(1, widthRatio, heightRatio);

  if (scaleRatio >= 1 && file.size <= 1_600_000) return file;

  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");
  if (!context) return file;

  canvas.width = Math.max(1, Math.round(sourceWidth * scaleRatio));
  canvas.height = Math.max(1, Math.round(sourceHeight * scaleRatio));
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.fillStyle = "#ffffff";
  context.fillRect(0, 0, canvas.width, canvas.height);
  context.drawImage(image, 0, 0, canvas.width, canvas.height);

  return canvasToFile(canvas, file, null);
};

const canvasToFile = (
  canvas: HTMLCanvasElement,
  sourceFile: File,
  suffix: string | null = "document"
) =>
  new Promise<File>((resolve, reject) => {
    const exportCanvas = getExportCanvas(canvas);
    const outputQuality = clamp(props.quality, 0.72, 0.98);

    exportCanvas.toBlob(
      (blob) => {
        if (!blob) {
          reject(new Error("Failed to create document image."));
          return;
        }

        const baseName = sourceFile.name
          .replace(/\.[^.]+$/, "")
          .replace(/(?:-(?:document|rotated))+$/i, "");
        resolve(
          new File([blob], `${baseName}${suffix ? `-${suffix}` : ""}.jpg`, {
            lastModified: Date.now(),
            type: "image/jpeg",
          })
        );
      },
      "image/jpeg",
      outputQuality
    );
  });

const loadScript = (id: string, src: string) =>
  new Promise<void>((resolve, reject) => {
    const existing = document.getElementById(id) as HTMLScriptElement | null;

    if (existing?.dataset.loaded === "true") {
      resolve();
      return;
    }

    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener(
        "error",
        () => reject(new Error(`Failed to load ${src}.`)),
        { once: true }
      );
      return;
    }

    const script = document.createElement("script");
    script.id = id;
    script.src = src;
    script.async = true;
    script.onload = () => {
      script.dataset.loaded = "true";
      resolve();
    };
    script.onerror = () => reject(new Error(`Failed to load ${src}.`));

    document.head.appendChild(script);
  });

const waitForOpenCvRuntime = () =>
  new Promise<void>((resolve, reject) => {
    const startedAt = Date.now();
    const timeoutMs = 20000;

    const checkReady = () => {
      const cv = (window as any).cv;

      if (cv && typeof cv.Mat === "function") {
        resolve();
        return;
      }

      if (Date.now() - startedAt > timeoutMs) {
        reject(new Error("Auto crop detector is not ready."));
        return;
      }

      window.setTimeout(checkReady, 100);
    };

    const cv = (window as any).cv;
    const previousReady = cv?.onRuntimeInitialized;

    if (cv) {
      cv.onRuntimeInitialized = () => {
        previousReady?.();
        checkReady();
      };
    }

    checkReady();
  });

const loadScannerScripts = async () => {
  await loadScript(OPENCV_SCRIPT_ID, OPENCV_SCRIPT_SRC);
  await waitForOpenCvRuntime();
  await loadScript(JSCANIFY_SCRIPT_ID, JSCANIFY_SCRIPT_SRC);

  if (!(window as any).jscanify) {
    throw new Error("Auto crop detector failed to load.");
  }
};

const createDetectionCanvas = (image: HTMLImageElement, maxSize = 1000) => {
  const sourceWidth = image.naturalWidth || image.width;
  const sourceHeight = image.naturalHeight || image.height;
  const scale = Math.min(1, maxSize / Math.max(sourceWidth, sourceHeight));
  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");

  if (!context) throw new Error("Canvas context is not available.");

  canvas.width = Math.max(1, Math.round(sourceWidth * scale));
  canvas.height = Math.max(1, Math.round(sourceHeight * scale));
  context.drawImage(image, 0, 0, canvas.width, canvas.height);

  return canvas;
};

const clamp = (value: number, min: number, max: number) =>
  Math.min(Math.max(value, min), max);

const createDefaultCropCorners = (): CropCorners => ({
  bottomLeft: { x: 5, y: 95 },
  bottomRight: { x: 95, y: 95 },
  topLeft: { x: 5, y: 5 },
  topRight: { x: 95, y: 5 },
});

const cloneCropCorners = (corners: CropCorners): CropCorners => ({
  bottomLeft: { ...corners.bottomLeft },
  bottomRight: { ...corners.bottomRight },
  topLeft: { ...corners.topLeft },
  topRight: { ...corners.topRight },
});

const syncCropRectFromCorners = () => {
  const corners = Object.values(cropCorners.value);
  const minX = Math.min(...corners.map((point) => point.x));
  const maxX = Math.max(...corners.map((point) => point.x));
  const minY = Math.min(...corners.map((point) => point.y));
  const maxY = Math.max(...corners.map((point) => point.y));

  cropRect.value = {
    height: maxY - minY,
    width: maxX - minX,
    x: minX,
    y: minY,
  };
};

const distance = (from: CropPoint, to: CropPoint) =>
  Math.hypot(from.x - to.x, from.y - to.y);

const getNormalizedPerspectiveSize = (sourceCorners: CropCorners) => {
  const topWidth = distance(sourceCorners.topLeft, sourceCorners.topRight);
  const bottomWidth = distance(
    sourceCorners.bottomLeft,
    sourceCorners.bottomRight
  );
  const leftHeight = distance(sourceCorners.topLeft, sourceCorners.bottomLeft);
  const rightHeight = distance(
    sourceCorners.topRight,
    sourceCorners.bottomRight
  );
  const averageWidth = Math.max(1, (topWidth + bottomWidth) / 2);
  const averageHeight = Math.max(1, (leftHeight + rightHeight) / 2);
  const rawArea = averageWidth * averageHeight;
  const rawAspect = averageWidth / averageHeight;
  const aSeriesAspect = 1 / Math.SQRT2;
  const portraitTargetAspect = aSeriesAspect;
  const landscapeTargetAspect = 1 / aSeriesAspect;
  const isPortrait = rawAspect <= 1;
  const targetAspect = isPortrait
    ? portraitTargetAspect
    : landscapeTargetAspect;
  const aspectDelta = Math.abs(rawAspect - targetAspect) / targetAspect;
  const safeAspect =
    aspectDelta < 0.25 ? targetAspect : clamp(rawAspect, 0.5, 2.0);
  const targetWidth = Math.sqrt(rawArea * safeAspect);
  const targetHeight = targetWidth / safeAspect;
  const blendStrength = aspectDelta < 0.25 ? 0.92 : 0.45;
  const normalizedWidth =
    averageWidth + (targetWidth - averageWidth) * blendStrength;
  const normalizedHeight =
    averageHeight + (targetHeight - averageHeight) * blendStrength;

  return {
    height: Math.max(1, Math.round(normalizedHeight)),
    width: Math.max(1, Math.round(normalizedWidth)),
  };
};

const normalizeDetectedCorners = (
  corners: CropCorners,
  strength = 0.45
): CropCorners => {
  const leftX = (corners.topLeft.x + corners.bottomLeft.x) / 2;
  const rightX = (corners.topRight.x + corners.bottomRight.x) / 2;
  const topY = (corners.topLeft.y + corners.topRight.y) / 2;
  const bottomY = (corners.bottomLeft.y + corners.bottomRight.y) / 2;
  const lerp = (from: number, to: number) => from + (to - from) * strength;

  return {
    bottomLeft: {
      x: lerp(corners.bottomLeft.x, leftX),
      y: lerp(corners.bottomLeft.y, bottomY),
    },
    bottomRight: {
      x: lerp(corners.bottomRight.x, rightX),
      y: lerp(corners.bottomRight.y, bottomY),
    },
    topLeft: {
      x: lerp(corners.topLeft.x, leftX),
      y: lerp(corners.topLeft.y, topY),
    },
    topRight: {
      x: lerp(corners.topRight.x, rightX),
      y: lerp(corners.topRight.y, topY),
    },
  };
};

const getCropCornersArea = (corners: CropCorners) => {
  const points = [
    corners.topLeft,
    corners.topRight,
    corners.bottomRight,
    corners.bottomLeft,
  ];

  return Math.abs(
    points.reduce((total, point, index) => {
      const next = points[(index + 1) % points.length]!;
      return total + point.x * next.y - next.x * point.y;
    }, 0) / 2
  );
};

const orderCropCorners = (points: CropPoint[]): CropCorners | null => {
  if (points.length !== 4) return null;

  const sortedByY = [...points].sort((a, b) => a.y - b.y);
  const top = sortedByY.slice(0, 2).sort((a, b) => a.x - b.x);
  const bottom = sortedByY.slice(2).sort((a, b) => a.x - b.x);

  return {
    bottomLeft: { ...bottom[0]! },
    bottomRight: { ...bottom[1]! },
    topLeft: { ...top[0]! },
    topRight: { ...top[1]! },
  };
};

const getCornerAngle = (
  previous: CropPoint,
  current: CropPoint,
  next: CropPoint
) => {
  const previousVector = {
    x: previous.x - current.x,
    y: previous.y - current.y,
  };
  const nextVector = {
    x: next.x - current.x,
    y: next.y - current.y,
  };
  const dot = previousVector.x * nextVector.x + previousVector.y * nextVector.y;
  const lengths =
    Math.hypot(previousVector.x, previousVector.y) *
    Math.hypot(nextVector.x, nextVector.y);

  if (!lengths) return 0;

  return Math.acos(clamp(dot / lengths, -1, 1)) * (180 / Math.PI);
};

const getCropCornersBounds = (corners: CropCorners) => {
  const points = Object.values(corners);
  const minX = Math.min(...points.map((point) => point.x));
  const maxX = Math.max(...points.map((point) => point.x));
  const minY = Math.min(...points.map((point) => point.y));
  const maxY = Math.max(...points.map((point) => point.y));

  return { maxX, maxY, minX, minY };
};

const isStableCropCorners = (corners: CropCorners) => {
  const area = getCropCornersArea(corners);
  const topWidth = distance(corners.topLeft, corners.topRight);
  const bottomWidth = distance(corners.bottomLeft, corners.bottomRight);
  const leftHeight = distance(corners.topLeft, corners.bottomLeft);
  const rightHeight = distance(corners.topRight, corners.bottomRight);
  const averageWidth = (topWidth + bottomWidth) / 2;
  const averageHeight = (leftHeight + rightHeight) / 2;
  const aspectRatio = averageWidth / Math.max(1, averageHeight);
  const points = [
    corners.topLeft,
    corners.topRight,
    corners.bottomRight,
    corners.bottomLeft,
  ];
  const angles = points.map((point, index) =>
    getCornerAngle(
      points[(index + points.length - 1) % points.length]!,
      point,
      points[(index + 1) % points.length]!
    )
  );

  return (
    area >= 8 &&
    aspectRatio >= 0.34 &&
    aspectRatio <= 2.95 &&
    Math.min(topWidth, bottomWidth, leftHeight, rightHeight) >= 12 &&
    angles.every((angle) => angle >= 34 && angle <= 146)
  );
};

type CropSideKey = "bottom" | "left" | "right" | "top";

type CropLine = {
  end: CropPoint;
  start: CropPoint;
};

type CropDetectionDebug = {
  confidence: number;
  paperSurfaceRatio: number;
  rawAreaRatio: number;
  rawCorners: CropCorners;
  refinedCorners: CropCorners;
  sideConfidence: Record<CropSideKey, number>;
};

const getSidePoints = (
  corners: CropCorners,
  side: CropSideKey
): [CropPoint, CropPoint] => {
  const sidePoints: Record<CropSideKey, [CropPoint, CropPoint]> = {
    bottom: [corners.bottomLeft, corners.bottomRight],
    left: [corners.topLeft, corners.bottomLeft],
    right: [corners.topRight, corners.bottomRight],
    top: [corners.topLeft, corners.topRight],
  };

  return sidePoints[side];
};

const getOffsetSideLine = (
  corners: CropCorners,
  side: CropSideKey,
  offset: number
): CropLine => {
  const [start, end] = getSidePoints(corners, side);
  const center = Object.values(corners).reduce(
    (acc, point) => ({ x: acc.x + point.x / 4, y: acc.y + point.y / 4 }),
    { x: 0, y: 0 }
  );
  const midpoint = { x: (start.x + end.x) / 2, y: (start.y + end.y) / 2 };
  const inward = { x: center.x - midpoint.x, y: center.y - midpoint.y };
  const length = Math.max(0.001, Math.hypot(inward.x, inward.y));
  const unit = { x: inward.x / length, y: inward.y / length };

  return {
    end: {
      x: clamp(end.x + unit.x * offset, 0, 100),
      y: clamp(end.y + unit.y * offset, 0, 100),
    },
    start: {
      x: clamp(start.x + unit.x * offset, 0, 100),
      y: clamp(start.y + unit.y * offset, 0, 100),
    },
  };
};

const getLineIntersection = (first: CropLine, second: CropLine) => {
  const x1 = first.start.x;
  const y1 = first.start.y;
  const x2 = first.end.x;
  const y2 = first.end.y;
  const x3 = second.start.x;
  const y3 = second.start.y;
  const x4 = second.end.x;
  const y4 = second.end.y;
  const denominator = (x1 - x2) * (y3 - y4) - (y1 - y2) * (x3 - x4);

  if (Math.abs(denominator) < 0.001) return null;

  return {
    x: clamp(
      ((x1 * y2 - y1 * x2) * (x3 - x4) - (x1 - x2) * (x3 * y4 - y3 * x4)) /
        denominator,
      0,
      100
    ),
    y: clamp(
      ((x1 * y2 - y1 * x2) * (y3 - y4) - (y1 - y2) * (x3 * y4 - y3 * x4)) /
        denominator,
      0,
      100
    ),
  };
};

const rebuildCornersFromLines = (lines: Record<CropSideKey, CropLine>) => {
  const topLeft = getLineIntersection(lines.top, lines.left);
  const topRight = getLineIntersection(lines.top, lines.right);
  const bottomRight = getLineIntersection(lines.bottom, lines.right);
  const bottomLeft = getLineIntersection(lines.bottom, lines.left);

  if (!topLeft || !topRight || !bottomRight || !bottomLeft) return null;

  return orderCropCorners([topLeft, topRight, bottomRight, bottomLeft]);
};

const refineCornersByEdgeTransitions = (
  corners: CropCorners,
  detectionCanvas: HTMLCanvasElement
): {
  confidence: number;
  corners: CropCorners;
  sideConfidence: Record<CropSideKey, number>;
} => {
  const context = detectionCanvas.getContext("2d", {
    willReadFrequently: true,
  });
  if (!context) {
    return {
      confidence: 0,
      corners,
      sideConfidence: { bottom: 0, left: 0, right: 0, top: 0 },
    };
  }

  const imageData = context.getImageData(
    0,
    0,
    detectionCanvas.width,
    detectionCanvas.height
  );
  const { data } = imageData;
  const getStatsAt = (point: CropPoint) => {
    const x = clamp(
      Math.round((point.x / 100) * (detectionCanvas.width - 1)),
      0,
      detectionCanvas.width - 1
    );
    const y = clamp(
      Math.round((point.y / 100) * (detectionCanvas.height - 1)),
      0,
      detectionCanvas.height - 1
    );
    const index = (y * detectionCanvas.width + x) * 4;
    const red = data[index]!;
    const green = data[index + 1]!;
    const blue = data[index + 2]!;
    const maxChannel = Math.max(red, green, blue);
    const minChannel = Math.min(red, green, blue);

    return {
      luminance: red * 0.299 + green * 0.587 + blue * 0.114,
      saturation: maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0,
    };
  };
  const getInwardUnit = (start: CropPoint, end: CropPoint) => {
    const center = Object.values(corners).reduce(
      (acc, point) => ({ x: acc.x + point.x / 4, y: acc.y + point.y / 4 }),
      { x: 0, y: 0 }
    );
    const midpoint = { x: (start.x + end.x) / 2, y: (start.y + end.y) / 2 };
    const inward = { x: center.x - midpoint.x, y: center.y - midpoint.y };
    const length = Math.max(0.001, Math.hypot(inward.x, inward.y));

    return { x: inward.x / length, y: inward.y / length };
  };
  const getPointAtOffset = (
    point: CropPoint,
    unit: CropPoint,
    offset: number
  ) => ({ x: point.x + unit.x * offset, y: point.y + unit.y * offset });
  const median = (values: number[]) => {
    const sorted = [...values].sort((a, b) => a - b);
    return sorted[Math.floor(sorted.length / 2)] ?? 0;
  };
  const refineSide = (side: CropSideKey) => {
    const [start, end] = getSidePoints(corners, side);
    const inwardUnit = getInwardUnit(start, end);
    const offsets: number[] = [];
    let strongTransitions = 0;
    let textNearSide = 0;
    const sampleCount = 42;

    for (let sampleIndex = 1; sampleIndex <= sampleCount; sampleIndex += 1) {
      const ratio = sampleIndex / (sampleCount + 1);
      const sidePoint = {
        x: start.x + (end.x - start.x) * ratio,
        y: start.y + (end.y - start.y) * ratio,
      };
      let bestOffset = 0;
      let bestScore = 0;
      let previousStats = getStatsAt(
        getPointAtOffset(sidePoint, inwardUnit, -4.6)
      );

      for (let offset = -4.1; offset <= 12.4; offset += 0.55) {
        const currentStats = getStatsAt(
          getPointAtOffset(sidePoint, inwardUnit, offset)
        );
        const luminanceDiff = Math.abs(
          currentStats.luminance - previousStats.luminance
        );
        const saturationDiff = Math.abs(
          currentStats.saturation - previousStats.saturation
        );
        const score = luminanceDiff + saturationDiff * 95;

        if (score > bestScore) {
          bestScore = score;
          bestOffset = offset - 0.225;
        }

        previousStats = currentStats;
      }

      const insideStats = getStatsAt(
        getPointAtOffset(sidePoint, inwardUnit, 2.1)
      );
      if (insideStats.luminance < 104) textNearSide += 1;

      if (bestScore >= 24) {
        strongTransitions += 1;
        offsets.push(bestOffset);
      }
    }

    const confidence = strongTransitions / sampleCount;
    const contentRatio = textNearSide / sampleCount;
    let offset = median(offsets);

    if (contentRatio > 0.08 && offset > 0) offset = 0;
    if (confidence < 0.26) offset = 0;

    const maxInwardOffset =
      confidence > 0.54 ? 8.2 : confidence > 0.38 ? 5.4 : 2.2;

    return {
      confidence,
      line: getOffsetSideLine(
        corners,
        side,
        clamp(offset, -1.35, maxInwardOffset)
      ),
    };
  };
  const refinedSides = {
    bottom: refineSide("bottom"),
    left: refineSide("left"),
    right: refineSide("right"),
    top: refineSide("top"),
  };
  const lines = {
    bottom: refinedSides.bottom.line,
    left: refinedSides.left.line,
    right: refinedSides.right.line,
    top: refinedSides.top.line,
  };
  const rebuiltCorners = rebuildCornersFromLines(lines);
  const confidence =
    (refinedSides.top.confidence +
      refinedSides.right.confidence +
      refinedSides.bottom.confidence +
      refinedSides.left.confidence) /
    4;

  const sideConfidence = {
    bottom: refinedSides.bottom.confidence,
    left: refinedSides.left.confidence,
    right: refinedSides.right.confidence,
    top: refinedSides.top.confidence,
  };

  if (!rebuiltCorners || !isStableCropCorners(rebuiltCorners)) {
    return { confidence, corners, sideConfidence };
  }

  return { confidence, corners: rebuiltCorners, sideConfidence };
};

const detectPaperSurfaceBoundsCorners = (
  detectionCanvas: HTMLCanvasElement
): CropCorners | null => {
  const context = detectionCanvas.getContext("2d", {
    willReadFrequently: true,
  });
  if (!context) return null;

  const imageData = context.getImageData(
    0,
    0,
    detectionCanvas.width,
    detectionCanvas.height
  );
  const { data } = imageData;
  const gridWidth = 96;
  const gridHeight = Math.max(
    64,
    Math.round((detectionCanvas.height / detectionCanvas.width) * gridWidth)
  );
  const mask = Array<boolean>(gridWidth * gridHeight).fill(false);
  const visited = Array<boolean>(gridWidth * gridHeight).fill(false);
  const getGridIndex = (x: number, y: number) => y * gridWidth + x;

  for (let gridY = 0; gridY < gridHeight; gridY += 1) {
    for (let gridX = 0; gridX < gridWidth; gridX += 1) {
      const sourceX = clamp(
        Math.round(
          (gridX / Math.max(1, gridWidth - 1)) * (detectionCanvas.width - 1)
        ),
        0,
        detectionCanvas.width - 1
      );
      const sourceY = clamp(
        Math.round(
          (gridY / Math.max(1, gridHeight - 1)) * (detectionCanvas.height - 1)
        ),
        0,
        detectionCanvas.height - 1
      );
      const index = (sourceY * detectionCanvas.width + sourceX) * 4;
      const red = data[index]!;
      const green = data[index + 1]!;
      const blue = data[index + 2]!;
      const maxChannel = Math.max(red, green, blue);
      const minChannel = Math.min(red, green, blue);
      const luminance = red * 0.299 + green * 0.587 + blue * 0.114;
      const saturation =
        maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0;

      mask[getGridIndex(gridX, gridY)] = luminance > 112 && saturation < 0.42;
    }
  }

  const components: Array<{
    area: number;
    maxX: number;
    maxY: number;
    minX: number;
    minY: number;
    score: number;
    xs: number[];
    ys: number[];
  }> = [];

  for (let startY = 0; startY < gridHeight; startY += 1) {
    for (let startX = 0; startX < gridWidth; startX += 1) {
      const startIndex = getGridIndex(startX, startY);
      if (!mask[startIndex] || visited[startIndex]) continue;

      const queue: CropPoint[] = [{ x: startX, y: startY }];
      const xs: number[] = [];
      const ys: number[] = [];
      visited[startIndex] = true;

      for (let queueIndex = 0; queueIndex < queue.length; queueIndex += 1) {
        const point = queue[queueIndex]!;
        xs.push((point.x / Math.max(1, gridWidth - 1)) * 100);
        ys.push((point.y / Math.max(1, gridHeight - 1)) * 100);

        [
          { x: point.x + 1, y: point.y },
          { x: point.x - 1, y: point.y },
          { x: point.x, y: point.y + 1 },
          { x: point.x, y: point.y - 1 },
        ].forEach((next) => {
          if (
            next.x < 0 ||
            next.x >= gridWidth ||
            next.y < 0 ||
            next.y >= gridHeight
          ) {
            return;
          }

          const nextIndex = getGridIndex(next.x, next.y);
          if (!mask[nextIndex] || visited[nextIndex]) return;

          visited[nextIndex] = true;
          queue.push(next);
        });
      }

      if (xs.length < 60) continue;

      xs.sort((a, b) => a - b);
      ys.sort((a, b) => a - b);

      const minX = xs[0] ?? 0;
      const maxX = xs[xs.length - 1] ?? 0;
      const minY = ys[0] ?? 0;
      const maxY = ys[ys.length - 1] ?? 0;
      const width = maxX - minX;
      const height = maxY - minY;
      const area = (width * height) / 10000;
      const fillRatio =
        xs.length /
        Math.max(1, (width / 100) * gridWidth * (height / 100) * gridHeight);
      const touchesFrame = [
        minX <= 1.5,
        minY <= 1.5,
        maxX >= 98.5,
        maxY >= 98.5,
      ].filter(Boolean).length;
      const frontFacingScore = clamp(fillRatio / 0.52, 0, 1);
      const sizeScore = clamp((area - 0.12) / 0.38, 0, 1);
      const framePenalty =
        touchesFrame >= 2 ? 0.28 : touchesFrame === 1 ? 0.72 : 1;
      const score =
        (sizeScore * 1.8 + frontFacingScore * 1.6 + fillRatio) * framePenalty;

      components.push({ area, maxX, maxY, minX, minY, score, xs, ys });
    }
  }

  const bestComponent = components.sort((a, b) => b.score - a.score)[0];
  if (!bestComponent) return null;

  const percentile = (values: number[], ratio: number) =>
    values[
      clamp(Math.round((values.length - 1) * ratio), 0, values.length - 1)
    ] ?? 0;
  const xs = bestComponent.xs;
  const ys = bestComponent.ys;
  const minX = clamp(percentile(xs, 0.015) - 1.2, 0, 100);
  const maxX = clamp(percentile(xs, 0.985) + 1.2, 0, 100);
  const minY = clamp(percentile(ys, 0.015) - 1.2, 0, 100);
  const maxY = clamp(percentile(ys, 0.985) + 1.2, 0, 100);
  const corners = orderCropCorners([
    { x: minX, y: minY },
    { x: maxX, y: minY },
    { x: maxX, y: maxY },
    { x: minX, y: maxY },
  ]);

  if (!corners || !isStableCropCorners(corners)) return null;

  const areaRatio = getCropCornersArea(corners) / 10000;
  if (areaRatio < 0.16 || areaRatio > 0.86) return null;

  return corners;
};

const getPaperSurfaceRatioInCorners = (
  corners: CropCorners,
  detectionCanvas: HTMLCanvasElement
) => {
  const context = detectionCanvas.getContext("2d", {
    willReadFrequently: true,
  });
  if (!context) return 0;

  const imageData = context.getImageData(
    0,
    0,
    detectionCanvas.width,
    detectionCanvas.height
  );
  const { data } = imageData;
  const bounds = getCropCornersBounds(corners);
  const startX = Math.round((bounds.minX / 100) * detectionCanvas.width);
  const endX = Math.round((bounds.maxX / 100) * detectionCanvas.width);
  const startY = Math.round((bounds.minY / 100) * detectionCanvas.height);
  const endY = Math.round((bounds.maxY / 100) * detectionCanvas.height);
  const step = Math.max(
    1,
    Math.round(Math.max(endX - startX, endY - startY) / 90)
  );
  let sampledPixels = 0;
  let paperPixels = 0;

  for (let y = startY; y < endY; y += step) {
    for (let x = startX; x < endX; x += step) {
      const safeX = clamp(x, 0, detectionCanvas.width - 1);
      const safeY = clamp(y, 0, detectionCanvas.height - 1);
      const index = (safeY * detectionCanvas.width + safeX) * 4;
      const red = data[index]!;
      const green = data[index + 1]!;
      const blue = data[index + 2]!;
      const maxChannel = Math.max(red, green, blue);
      const minChannel = Math.min(red, green, blue);
      const luminance = red * 0.299 + green * 0.587 + blue * 0.114;
      const saturation =
        maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0;

      sampledPixels += 1;
      if (luminance > 118 && saturation < 0.36) paperPixels += 1;
    }
  }

  return paperPixels / Math.max(1, sampledPixels);
};

const applyPaperSurfaceFallbackCorners = (
  detectionCanvas: HTMLCanvasElement
) => {
  const surfaceCorners = detectPaperSurfaceBoundsCorners(detectionCanvas);
  if (!surfaceCorners) return false;

  const refinedResult = refineCornersByEdgeTransitions(
    surfaceCorners,
    detectionCanvas
  );
  const finalCorners = isStableCropCorners(refinedResult.corners)
    ? refinedResult.corners
    : surfaceCorners;

  cropCorners.value = finalCorners;
  cropDetectionDebug.value = {
    confidence: refinedResult.confidence,
    paperSurfaceRatio: getPaperSurfaceRatioInCorners(
      surfaceCorners,
      detectionCanvas
    ),
    rawAreaRatio: getCropCornersArea(surfaceCorners) / 10000,
    rawCorners: cloneCropCorners(surfaceCorners),
    refinedCorners: cloneCropCorners(finalCorners),
    sideConfidence: refinedResult.sideConfidence,
  };
  syncCropRectFromCorners();
  return true;
};

const applyDetectedContourCorners = (
  scanner: any,
  contour: any,
  detectionCanvas: HTMLCanvasElement,
  options: { normalizeStrength?: number; shrinkRatio?: number } = {}
) => {
  if (!contour) return false;

  const { bottomLeftCorner, bottomRightCorner, topLeftCorner, topRightCorner } =
    scanner.getCornerPoints(contour);

  if (
    !topLeftCorner ||
    !topRightCorner ||
    !bottomLeftCorner ||
    !bottomRightCorner
  ) {
    return false;
  }

  const orderedCorners = orderCropCorners(
    [topLeftCorner, topRightCorner, bottomRightCorner, bottomLeftCorner].map(
      (point) => ({
        x: clamp((point.x / detectionCanvas.width) * 100, 0, 100),
        y: clamp((point.y / detectionCanvas.height) * 100, 0, 100),
      })
    )
  );

  if (!orderedCorners || !isStableCropCorners(orderedCorners)) return false;

  const rawAreaRatio = getCropCornersArea(orderedCorners) / 10000;
  const rawBounds = getCropCornersBounds(orderedCorners);
  const frameTouchCount = [
    rawBounds.minX <= 2.5,
    rawBounds.minY <= 2.5,
    rawBounds.maxX >= 97.5,
    rawBounds.maxY >= 97.5,
  ].filter(Boolean).length;
  const paperSurfaceRatio = getPaperSurfaceRatioInCorners(
    orderedCorners,
    detectionCanvas
  );
  const refinedResult = refineCornersByEdgeTransitions(
    orderedCorners,
    detectionCanvas
  );

  cropDetectionDebug.value = {
    confidence: refinedResult.confidence,
    paperSurfaceRatio,
    rawAreaRatio,
    rawCorners: cloneCropCorners(orderedCorners),
    refinedCorners: cloneCropCorners(refinedResult.corners),
    sideConfidence: refinedResult.sideConfidence,
  };

  if (
    frameTouchCount >= 3 ||
    (rawAreaRatio > 0.82 && refinedResult.confidence < 0.28) ||
    (rawAreaRatio > 0.58 && paperSurfaceRatio < 0.62) ||
    (rawAreaRatio > 0.46 && paperSurfaceRatio < 0.54)
  ) {
    return false;
  }

  const shouldNormalize =
    !!options.normalizeStrength && refinedResult.confidence < 0.42;
  const normalizedCorners = shouldNormalize
    ? normalizeDetectedCorners(
        refinedResult.corners,
        Math.min(options.normalizeStrength ?? 0.25, 0.28)
      )
    : refinedResult.corners;

  const finalCorners = isStableCropCorners(normalizedCorners)
    ? normalizedCorners
    : orderedCorners;

  cropCorners.value = finalCorners;
  if (cropDetectionDebug.value) {
    cropDetectionDebug.value.refinedCorners = cloneCropCorners(finalCorners);
  }
  syncCropRectFromCorners();
  return true;
};

const findBestDocumentContour = (
  mat: any,
  detectionCanvas: HTMLCanvasElement
) => {
  const cv = (window as any).cv;
  const rgb = new cv.Mat();
  const hsv = new cv.Mat();
  const gray = new cv.Mat();
  const blurred = new cv.Mat();
  const edges = new cv.Mat();
  const mask = new cv.Mat();
  const morphed = new cv.Mat();
  const edgeMorphed = new cv.Mat();
  const adaptiveMask = new cv.Mat();
  const adaptiveMorphed = new cv.Mat();
  const contours = new cv.MatVector();
  const edgeContours = new cv.MatVector();
  const adaptiveContours = new cv.MatVector();
  const hierarchy = new cv.Mat();
  const edgeHierarchy = new cv.Mat();
  const adaptiveHierarchy = new cv.Mat();
  const kernel = cv.getStructuringElement(cv.MORPH_RECT, new cv.Size(11, 11));
  const edgeKernel = cv.getStructuringElement(cv.MORPH_RECT, new cv.Size(7, 7));
  const adaptiveKernel = cv.getStructuringElement(
    cv.MORPH_RECT,
    new cv.Size(9, 9)
  );

  const getLightPaperRatio = (rect: {
    height: number;
    width: number;
    x: number;
    y: number;
  }) => {
    const insetX = Math.max(1, Math.round(rect.width * 0.08));
    const insetY = Math.max(1, Math.round(rect.height * 0.08));
    const startX = clamp(rect.x + insetX, 0, mask.cols - 1);
    const endX = clamp(rect.x + rect.width - insetX, startX + 1, mask.cols);
    const startY = clamp(rect.y + insetY, 0, mask.rows - 1);
    const endY = clamp(rect.y + rect.height - insetY, startY + 1, mask.rows);
    const step = Math.max(
      1,
      Math.round(Math.max(rect.width, rect.height) / 120)
    );
    let sampled = 0;
    let lightPixels = 0;

    for (let y = startY; y < endY; y += step) {
      for (let x = startX; x < endX; x += step) {
        sampled += 1;
        if (mask.ucharPtr(Math.round(y), Math.round(x))[0] > 0) {
          lightPixels += 1;
        }
      }
    }

    return lightPixels / Math.max(1, sampled);
  };

  const getEdgeDensity = (rect: {
    height: number;
    width: number;
    x: number;
    y: number;
  }) => {
    const inset = Math.max(
      2,
      Math.round(Math.min(rect.width, rect.height) * 0.02)
    );
    const step = Math.max(
      1,
      Math.round(Math.max(rect.width, rect.height) / 180)
    );
    let sampled = 0;
    let hits = 0;
    const sample = (x: number, y: number) => {
      const nextX = clamp(Math.round(x), 0, edges.cols - 1);
      const nextY = clamp(Math.round(y), 0, edges.rows - 1);

      sampled += 1;
      if (edges.ucharPtr(nextY, nextX)[0] > 0) hits += 1;
    };

    for (let x = rect.x + inset; x <= rect.x + rect.width - inset; x += step) {
      sample(x, rect.y + inset);
      sample(x, rect.y + rect.height - inset);
    }

    for (let y = rect.y + inset; y <= rect.y + rect.height - inset; y += step) {
      sample(rect.x + inset, y);
      sample(rect.x + rect.width - inset, y);
    }

    return hits / Math.max(1, sampled);
  };

  try {
    cv.cvtColor(mat, rgb, cv.COLOR_RGBA2RGB);
    cv.cvtColor(rgb, hsv, cv.COLOR_RGB2HSV);
    cv.cvtColor(mat, gray, cv.COLOR_RGBA2GRAY);
    cv.GaussianBlur(gray, blurred, new cv.Size(5, 5), 0, 0);
    cv.Canny(blurred, edges, 60, 160);

    const lowLightPaper = new cv.Mat(
      hsv.rows,
      hsv.cols,
      hsv.type(),
      new cv.Scalar(0, 0, 138, 0)
    );
    const highLightPaper = new cv.Mat(
      hsv.rows,
      hsv.cols,
      hsv.type(),
      new cv.Scalar(180, 105, 255, 255)
    );

    try {
      cv.inRange(hsv, lowLightPaper, highLightPaper, mask);
      cv.morphologyEx(mask, morphed, cv.MORPH_CLOSE, kernel);
      cv.morphologyEx(morphed, morphed, cv.MORPH_OPEN, kernel);
      cv.findContours(
        morphed,
        contours,
        hierarchy,
        cv.RETR_EXTERNAL,
        cv.CHAIN_APPROX_SIMPLE
      );
      cv.dilate(edges, edgeMorphed, edgeKernel);
      cv.morphologyEx(edgeMorphed, edgeMorphed, cv.MORPH_CLOSE, edgeKernel);
      cv.findContours(
        edgeMorphed,
        edgeContours,
        edgeHierarchy,
        cv.RETR_EXTERNAL,
        cv.CHAIN_APPROX_SIMPLE
      );
      cv.adaptiveThreshold(
        blurred,
        adaptiveMask,
        255,
        cv.ADAPTIVE_THRESH_GAUSSIAN_C,
        cv.THRESH_BINARY,
        41,
        5
      );
      cv.bitwise_not(adaptiveMask, adaptiveMask);
      cv.morphologyEx(
        adaptiveMask,
        adaptiveMorphed,
        cv.MORPH_CLOSE,
        adaptiveKernel
      );
      cv.findContours(
        adaptiveMorphed,
        adaptiveContours,
        adaptiveHierarchy,
        cv.RETR_EXTERNAL,
        cv.CHAIN_APPROX_SIMPLE
      );
    } finally {
      lowLightPaper.delete();
      highLightPaper.delete();
    }

    const canvasArea = detectionCanvas.width * detectionCanvas.height;
    const canvasCenter = {
      x: detectionCanvas.width / 2,
      y: detectionCanvas.height / 2,
    };
    let bestContour: any = null;
    let bestScore = 0;

    const evaluateContour = (
      contour: any,
      source: "adaptive" | "edge" | "light"
    ) => {
      const perimeter = cv.arcLength(contour, true);
      const approx = new cv.Mat();

      try {
        cv.approxPolyDP(contour, approx, perimeter * 0.025, true);

        const area = cv.contourArea(contour);
        const areaRatio = area / canvasArea;
        const rect = cv.boundingRect(contour);
        const rectArea = rect.width * rect.height;
        const fillRatio = rectArea > 0 ? area / rectArea : 0;
        const aspectRatio = rect.height > 0 ? rect.width / rect.height : 0;
        const normalizedAspectRatio =
          aspectRatio >= 1 ? 1 / Math.max(aspectRatio, 0.001) : aspectRatio;
        const widthRatio = rect.width / detectionCanvas.width;
        const heightRatio = rect.height / detectionCanvas.height;
        const centerDistanceRatio =
          Math.hypot(
            rect.x + rect.width / 2 - canvasCenter.x,
            rect.y + rect.height / 2 - canvasCenter.y
          ) / Math.hypot(canvasCenter.x, canvasCenter.y);
        const frameInsetX = detectionCanvas.width * 0.025;
        const frameInsetY = detectionCanvas.height * 0.025;
        const touchesFrameCount = [
          rect.x <= frameInsetX,
          rect.y <= frameInsetY,
          rect.x + rect.width >= detectionCanvas.width - frameInsetX,
          rect.y + rect.height >= detectionCanvas.height - frameInsetY,
        ].filter(Boolean).length;
        const cornerCount = approx.rows;
        const isConvex = cornerCount >= 4 && cv.isContourConvex(approx);
        const edgeDensity = getEdgeDensity(rect);
        const lightPaperRatio = getLightPaperRatio(rect);
        const paperSurfaceScore = clamp(lightPaperRatio / 0.58, 0, 1);
        const edgeScore = clamp(edgeDensity / 0.18, 0, 1);
        const centerScore = clamp(1 - centerDistanceRatio, 0, 1);
        const a4Score = clamp(
          1 - Math.abs(normalizedAspectRatio - 0.707) / 0.42,
          0,
          1
        );
        const fillScore = clamp((fillRatio - 0.34) / 0.42, 0, 1);
        const cornerScore =
          cornerCount === 4
            ? 1
            : cornerCount >= 5 && cornerCount <= 8
              ? 0.62
              : 0.24;
        const sizeScore = clamp((areaRatio - 0.12) / 0.46, 0, 1);
        const dimensionScore = Math.min(
          clamp((widthRatio - 0.32) / 0.36, 0, 1),
          clamp((heightRatio - 0.32) / 0.36, 0, 1)
        );
        const tablePenalty =
          areaRatio < 0.22 || widthRatio < 0.42 || heightRatio < 0.42
            ? 0.72
            : 1;
        const oversizedWeakEdgePenalty =
          areaRatio > 0.86 && edgeScore < 0.42 ? 0.58 : 1;
        const oversizedFramePenalty =
          (widthRatio > 0.93 || heightRatio > 0.93) && centerScore < 0.9
            ? 0.72
            : 1;
        const lowPaperSurfacePenalty =
          areaRatio > 0.42 && lightPaperRatio < 0.56 ? 0.18 : 1;
        const darkFramePenalty =
          areaRatio > 0.58 && lightPaperRatio < 0.62 ? 0.08 : 1;
        const likelyBackgroundFramePenalty =
          areaRatio > 0.54 && lightPaperRatio < 0.6 ? 0.04 : 1;
        const frameTouchPenalty = touchesFrameCount >= 3 ? 0.02 : 1;
        const sourceBonus =
          source === "edge" ? 0.55 : source === "adaptive" ? 0.38 : 0.25;
        const sourceFillMinimum =
          source === "edge" ? 0.34 : source === "adaptive" ? 0.38 : 0.46;
        const shapeScore =
          cornerScore * 1.45 +
          edgeScore * 1.15 +
          dimensionScore * 1.2 +
          fillScore * 0.85 +
          paperSurfaceScore * 1.1;
        const looksLikePaper =
          areaRatio >= (source === "edge" ? 0.12 : 0.15) &&
          widthRatio >= 0.32 &&
          heightRatio >= 0.32 &&
          fillRatio >= sourceFillMinimum &&
          aspectRatio >= 0.38 &&
          aspectRatio <= 2.65 &&
          isConvex &&
          cornerCount <= 12 &&
          touchesFrameCount < 3 &&
          shapeScore >= 1.35 &&
          (areaRatio < 0.46 || lightPaperRatio >= 0.54) &&
          (areaRatio < 0.58 || lightPaperRatio >= 0.62) &&
          (source !== "edge" || areaRatio < 0.46 || lightPaperRatio >= 0.58);

        if (!looksLikePaper) return 0;

        return (
          (sizeScore * 3.0 +
            fillScore * (source === "light" ? 1.45 : 1.05) +
            dimensionScore * 1.5 +
            cornerScore * 1.5 +
            centerScore * 0.75 +
            a4Score * 0.55 +
            edgeScore * 1.25 +
            paperSurfaceScore * 1.85 +
            sourceBonus) *
          tablePenalty *
          oversizedWeakEdgePenalty *
          oversizedFramePenalty *
          lowPaperSurfacePenalty *
          darkFramePenalty *
          likelyBackgroundFramePenalty *
          frameTouchPenalty
        );
      } finally {
        approx.delete();
      }
    };

    const keepBestContour = (contour: any, score: number) => {
      if (score > bestScore) {
        bestContour?.delete?.();
        bestContour = contour;
        bestScore = score;
        return;
      }

      contour.delete();
    };

    for (let index = 0; index < contours.size(); index += 1) {
      const contour = contours.get(index);
      keepBestContour(contour, evaluateContour(contour, "light"));
    }

    for (let index = 0; index < edgeContours.size(); index += 1) {
      const contour = edgeContours.get(index);
      keepBestContour(contour, evaluateContour(contour, "edge"));
    }

    for (let index = 0; index < adaptiveContours.size(); index += 1) {
      const contour = adaptiveContours.get(index);
      keepBestContour(contour, evaluateContour(contour, "adaptive"));
    }

    return bestContour;
  } finally {
    rgb.delete();
    hsv.delete();
    gray.delete();
    blurred.delete();
    edges.delete();
    mask.delete();
    morphed.delete();
    edgeMorphed.delete();
    adaptiveMask.delete();
    adaptiveMorphed.delete();
    contours.delete();
    edgeContours.delete();
    adaptiveContours.delete();
    hierarchy.delete();
    edgeHierarchy.delete();
    adaptiveHierarchy.delete();
    kernel.delete();
    edgeKernel.delete();
    adaptiveKernel.delete();
  }
};

const findCameraHintContour = (scanner: any, mat: any) => {
  try {
    return scanner.findPaperContour(mat);
  } catch {
    return null;
  }
};

const detectInitialCrop = async (file: File) => {
  isDetectingCrop.value = true;

  try {
    await loadScannerScripts();

    const image = await loadImageFromFile(file);
    const detectionCanvas = createDetectionCanvas(image);
    const Scanner = (window as any).jscanify;
    const scanner = new Scanner();
    const cv = (window as any).cv;
    const mat = cv.imread(detectionCanvas);
    const bestDocumentContour = findBestDocumentContour(mat, detectionCanvas);
    let fallbackContour: any = null;

    try {
      const appliedBest = bestDocumentContour
        ? applyDetectedContourCorners(
            scanner,
            bestDocumentContour,
            detectionCanvas,
            {
              normalizeStrength: 0.45,
              shrinkRatio: 0.97,
            }
          )
        : false;

      if (!appliedBest) {
        fallbackContour = scanner.findPaperContour(mat);
        const appliedFallback = applyDetectedContourCorners(
          scanner,
          fallbackContour,
          detectionCanvas,
          {
            normalizeStrength: 0.25,
          }
        );

        if (!appliedFallback) {
          applyPaperSurfaceFallbackCorners(detectionCanvas);
        }
      }
    } finally {
      bestDocumentContour?.delete?.();
      fallbackContour?.delete?.();
      mat.delete();
    }
  } catch (err: any) {
    setError(err?.message || "Auto crop failed. You can adjust manually.");
  } finally {
    isDetectingCrop.value = false;

    if (cropViewMode.value === "preview") {
      clearCropPreview({ regeneratePreview: true });
    }
  }
};

const openCropEditor = (
  file: File,
  options: { restoreLastEdit?: boolean; source?: "camera" | "upload" } = {}
) => {
  if (isNativeMobilePlatform() || !props.enableScanner) {
    completeActiveImageFile(file);
    return;
  }

  if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value);
  if (cropPreviewImageUrl.value) URL.revokeObjectURL(cropPreviewImageUrl.value);

  cropPreviewGenerationId += 1;
  cropPreviewCacheKey.value = null;
  cropDetectionDebug.value = null;
  showCropDebug.value = false;
  cropSourceFile.value = file;
  cropImageUrl.value = URL.createObjectURL(file);
  cropPreviewImageUrl.value = null;
  cropViewMode.value = "preview";
  canRetakeCropPhoto.value = options.source === "camera";
  isGeneratingCropPreview.value = false;
  cropRect.value = options.restoreLastEdit
    ? { ...lastAppliedCropRect.value }
    : { height: 90, width: 90, x: 5, y: 5 };
  cropCorners.value = options.restoreLastEdit
    ? cloneCropCorners(lastAppliedCropCorners.value)
    : createDefaultCropCorners();
  selectedCropFilter.value = options.restoreLastEdit
    ? lastAppliedCropFilter.value
    : "original";
  cropDialogOpen.value = true;

  if (!options.restoreLastEdit) {
    void detectInitialCrop(file).then(() => generateCropPreview());
  } else {
    void generateCropPreview();
  }
};

const emitSelectedEvent = (event: Event) => {
  setError(null);
  emit("select", event);
};

const resetQueue = () => {
  activeQueuedImageFile.value = null;
  queuedImageFiles.value = [];
  processedQueuedFiles.value = [];
  processedQueuedSourceFiles.value = [];
};

const finishQueueIfComplete = () => {
  if (queuedImageFiles.value.length > 0) return false;

  const files = processedQueuedFiles.value.filter(
    (file): file is File => !!file
  );
  const sourceFiles = processedQueuedSourceFiles.value.filter(
    (file): file is File => !!file
  );
  const editableFile = sourceFiles.length === 1 ? sourceFiles[0] : null;

  resetQueue();
  closeCropDialog();

  if (files.length > 0) {
    editableDocumentSourceFile.value = editableFile ?? null;
    emit("sourceFiles", sourceFiles.length ? sourceFiles : files);
    emitSelectedEvent(createFilesEvent(files));
  }

  return true;
};

const openNextQueuedImage = async () => {
  const next = queuedImageFiles.value.shift();

  if (!next) {
    finishQueueIfComplete();
    return;
  }

  activeQueuedImageFile.value = next;
  const processedFile = await resizeImageFileForProcessing(next.file);
  openCropEditor(processedFile, { source: "upload" });
};

const completeActiveImageFile = (file: File) => {
  const activeItem = activeQueuedImageFile.value;

  if (!activeItem) {
    const sourceFile = cropSourceFile.value ?? file;

    editableDocumentSourceFile.value = sourceFile;
    emit("sourceFiles", [sourceFile]);
    emitSelectedEvent(createFileEvent(file));
    closeCropDialog();
    return;
  }

  processedQueuedFiles.value[activeItem.index] = file;
  processedQueuedSourceFiles.value[activeItem.index] =
    cropSourceFile.value ?? file;
  activeQueuedImageFile.value = null;
  void openNextQueuedImage();
};

const processSelectedFiles = async (files: File[], fallbackEvent?: Event) => {
  if (!files.length) return;

  if (!props.multiple || files.length === 1) {
    const file = files[0];

    if (!file || !isImageFile(file)) {
      editableDocumentSourceFile.value = null;
      emit("sourceFiles", files);
      emitSelectedEvent(fallbackEvent ?? createFilesEvent(files));
      return;
    }

    const processedFile = await resizeImageFileForProcessing(file);
    editableDocumentSourceFile.value = processedFile;
    openCropEditor(processedFile, { source: "upload" });
    return;
  }

  resetQueue();
  editableDocumentSourceFile.value = null;
  processedQueuedFiles.value = Array<File | null>(files.length).fill(null);
  processedQueuedSourceFiles.value = Array<File | null>(files.length).fill(
    null
  );

  files.forEach((file, index) => {
    if (isImageFile(file)) {
      queuedImageFiles.value.push({ file, index });
      return;
    }

    processedQueuedFiles.value[index] = file;
    processedQueuedSourceFiles.value[index] = file;
  });

  void openNextQueuedImage();
};

const handleFileChange = (event: Event) => {
  setError(null);

  const target = event.target as HTMLInputElement | null;
  const files = Array.from(target?.files ?? []);

  void processSelectedFiles(files, event);
  if (target) target.value = "";
};

const applyAdaptiveBlackWhiteFilter = (
  canvas: HTMLCanvasElement,
  options: { blockSize: number; blurSize?: number; constant: number }
) => {
  const cv = (window as any).cv;
  const sourceMat = cv.imread(canvas);
  const grayMat = new cv.Mat();
  const blurredMat = new cv.Mat();
  const thresholdMat = new cv.Mat();
  const resultCanvas = document.createElement("canvas");
  const blurSize = options.blurSize ?? 3;

  try {
    cv.cvtColor(sourceMat, grayMat, cv.COLOR_RGBA2GRAY);
    cv.GaussianBlur(grayMat, blurredMat, new cv.Size(blurSize, blurSize), 0, 0);
    cv.adaptiveThreshold(
      blurredMat,
      thresholdMat,
      255,
      cv.ADAPTIVE_THRESH_GAUSSIAN_C,
      cv.THRESH_BINARY,
      options.blockSize,
      options.constant
    );
    cv.imshow(resultCanvas, thresholdMat);
  } finally {
    sourceMat.delete();
    grayMat.delete();
    blurredMat.delete();
    thresholdMat.delete();
  }

  return resultCanvas;
};

const applyCleanPaperFilter = (canvas: HTMLCanvasElement) =>
  applyAdaptiveBlackWhiteFilter(canvas, {
    blockSize: 25,
    blurSize: 3,
    constant: 5,
  });

const applyBetterBlackWhiteFilter = (canvas: HTMLCanvasElement) =>
  applyAdaptiveBlackWhiteFilter(canvas, {
    blockSize: 31,
    blurSize: 5,
    constant: 4,
  });

const softenResidualBorderShadows = (canvas: HTMLCanvasElement) => {
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context || canvas.width < 80 || canvas.height < 80) return canvas;

  const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
  const { data } = imageData;
  const width = canvas.width;
  const height = canvas.height;
  const bandSize = Math.max(4, Math.round(Math.min(width, height) * 0.025));
  const getIndex = (x: number, y: number) => (y * width + x) * 4;

  for (let y = 0; y < height; y += 1) {
    for (let x = 0; x < width; x += 1) {
      const distanceToEdge = Math.min(x, y, width - 1 - x, height - 1 - y);
      if (distanceToEdge > bandSize) continue;

      const index = getIndex(x, y);
      const red = data[index]!;
      const green = data[index + 1]!;
      const blue = data[index + 2]!;
      const luminance = red * 0.299 + green * 0.587 + blue * 0.114;
      const maxChannel = Math.max(red, green, blue);
      const minChannel = Math.min(red, green, blue);
      const saturation =
        maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0;

      if (luminance < 92) continue;

      const edgeFactor = 1 - distanceToEdge / Math.max(1, bandSize);
      const shadowFactor = clamp((214 - luminance) / 112, 0, 1);
      const colorFactor = clamp((saturation - 0.08) / 0.26, 0, 1);
      const liftAmount =
        Math.max(shadowFactor, colorFactor) * edgeFactor * 0.72;

      if (liftAmount <= 0) continue;

      data[index] = clamp(red + (255 - red) * liftAmount, 0, 255);
      data[index + 1] = clamp(green + (255 - green) * liftAmount, 0, 255);
      data[index + 2] = clamp(blue + (255 - blue) * liftAmount, 0, 255);
    }
  }

  context.putImageData(imageData, 0, 0);
  return canvas;
};

const cleanupPerspectiveEdges = (canvas: HTMLCanvasElement) => {
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context || canvas.width < 80 || canvas.height < 80) return canvas;

  const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
  const { data } = imageData;
  const width = canvas.width;
  const height = canvas.height;
  const maxHorizontalScan = Math.round(width * 0.06);
  const maxVerticalScan = Math.round(height * 0.06);
  const baseInset = Math.max(1, Math.round(Math.min(width, height) * 0.012));

  const getPixelStatsAt = (x: number, y: number) => {
    const index = (y * width + x) * 4;
    const red = data[index]!;
    const green = data[index + 1]!;
    const blue = data[index + 2]!;
    const maxChannel = Math.max(red, green, blue);
    const minChannel = Math.min(red, green, blue);

    return {
      luminance: red * 0.299 + green * 0.587 + blue * 0.114,
      saturation: maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0,
    };
  };

  const getLuminanceAt = (x: number, y: number) =>
    getPixelStatsAt(x, y).luminance;

  const getPaperReferenceStats = () => {
    const startX = Math.round(width * 0.25);
    const endX = Math.round(width * 0.75);
    const startY = Math.round(height * 0.25);
    const endY = Math.round(height * 0.75);
    const step = Math.max(1, Math.round(Math.min(width, height) / 180));
    const samples: Array<{ luminance: number; saturation: number }> = [];

    for (let y = startY; y < endY; y += step) {
      for (let x = startX; x < endX; x += step) {
        const stats = getPixelStatsAt(x, y);
        if (stats.luminance > 95) samples.push(stats);
      }
    }

    if (!samples.length) return { luminance: 210, saturation: 0.12 };

    samples.sort((a, b) => b.luminance - a.luminance);
    const paperSamples = samples.slice(
      0,
      Math.max(8, Math.round(samples.length * 0.35))
    );
    const totals = paperSamples.reduce(
      (acc, sample) => ({
        luminance: acc.luminance + sample.luminance,
        saturation: acc.saturation + sample.saturation,
      }),
      { luminance: 0, saturation: 0 }
    );

    return {
      luminance: totals.luminance / paperSamples.length,
      saturation: totals.saturation / paperSamples.length,
    };
  };

  const paperReference = getPaperReferenceStats();

  const isDarkColumn = (x: number) => {
    let darkPixels = 0;
    const step = Math.max(1, Math.floor(height / 260));
    let sampledPixels = 0;

    for (let y = 0; y < height; y += step) {
      sampledPixels += 1;
      if (getLuminanceAt(x, y) < 48) darkPixels += 1;
    }

    return darkPixels / Math.max(1, sampledPixels) > 0.34;
  };

  const isDarkRow = (y: number) => {
    let darkPixels = 0;
    const step = Math.max(1, Math.floor(width / 260));
    let sampledPixels = 0;

    for (let x = 0; x < width; x += step) {
      sampledPixels += 1;
      if (getLuminanceAt(x, y) < 48) darkPixels += 1;
    }

    return darkPixels / Math.max(1, sampledPixels) > 0.34;
  };

  const isOutOfPaperPixel = (x: number, y: number) => {
    const stats = getPixelStatsAt(x, y);

    if (stats.luminance < 70) return false;

    const luminanceDiff = Math.abs(stats.luminance - paperReference.luminance);
    const saturationDiff = stats.saturation - paperReference.saturation;

    return luminanceDiff > 42 || saturationDiff > 0.2;
  };

  const isOutOfPaperColumn = (x: number) => {
    let outPixels = 0;
    const step = Math.max(1, Math.floor(height / 260));
    let sampledPixels = 0;

    for (let y = 0; y < height; y += step) {
      sampledPixels += 1;
      if (isOutOfPaperPixel(x, y)) outPixels += 1;
    }

    return outPixels / Math.max(1, sampledPixels) > 0.24;
  };

  const isOutOfPaperRow = (y: number) => {
    let outPixels = 0;
    const step = Math.max(1, Math.floor(width / 260));
    let sampledPixels = 0;

    for (let x = 0; x < width; x += step) {
      sampledPixels += 1;
      if (isOutOfPaperPixel(x, y)) outPixels += 1;
    }

    return outPixels / Math.max(1, sampledPixels) > 0.24;
  };

  const hasContentNearSide = (
    side: "bottom" | "left" | "right" | "top",
    detectedBorderInset: number
  ) => {
    const bandSize = Math.max(3, Math.round(Math.min(width, height) * 0.04));
    const step = Math.max(1, Math.round(Math.min(width, height) / 360));
    let sampledPixels = 0;
    let darkPixels = 0;

    const sample = (x: number, y: number) => {
      sampledPixels += 1;
      const luminance = getLuminanceAt(
        clamp(x, 0, width - 1),
        clamp(y, 0, height - 1)
      );
      if (luminance < 105) darkPixels += 1;
    };

    if (side === "left" || side === "right") {
      const startX =
        side === "left"
          ? detectedBorderInset
          : width - detectedBorderInset - bandSize;
      const endX = startX + bandSize;

      for (let x = startX; x < endX; x += step) {
        for (let y = 0; y < height; y += step) sample(x, y);
      }
    } else {
      const startY =
        side === "top"
          ? detectedBorderInset
          : height - detectedBorderInset - bandSize;
      const endY = startY + bandSize;

      for (let y = startY; y < endY; y += step) {
        for (let x = 0; x < width; x += step) sample(x, y);
      }
    }

    const darkRatio = darkPixels / Math.max(1, sampledPixels);

    return darkRatio > 0.012 && darkRatio < 0.28;
  };

  let leftInset = 0;
  let rightInset = 0;
  let topInset = 0;
  let bottomInset = 0;

  while (
    leftInset < maxHorizontalScan &&
    !hasContentNearSide("left", leftInset) &&
    (isDarkColumn(leftInset) || isOutOfPaperColumn(leftInset))
  ) {
    leftInset += 1;
  }

  while (
    rightInset < maxHorizontalScan &&
    !hasContentNearSide("right", rightInset) &&
    (isDarkColumn(width - 1 - rightInset) ||
      isOutOfPaperColumn(width - 1 - rightInset))
  ) {
    rightInset += 1;
  }

  while (
    topInset < maxVerticalScan &&
    !hasContentNearSide("top", topInset) &&
    (isDarkRow(topInset) || isOutOfPaperRow(topInset))
  ) {
    topInset += 1;
  }

  while (
    bottomInset < maxVerticalScan &&
    !hasContentNearSide("bottom", bottomInset) &&
    (isDarkRow(height - 1 - bottomInset) ||
      isOutOfPaperRow(height - 1 - bottomInset))
  ) {
    bottomInset += 1;
  }

  const safeLeftInset = hasContentNearSide("left", leftInset) ? 0 : baseInset;
  const safeRightInset = hasContentNearSide("right", rightInset)
    ? 0
    : baseInset;
  const safeTopInset = hasContentNearSide("top", topInset) ? 0 : baseInset;
  const safeBottomInset = hasContentNearSide("bottom", bottomInset)
    ? 0
    : baseInset;
  const cropX = clamp(leftInset + safeLeftInset, 0, width - 2);
  const cropY = clamp(topInset + safeTopInset, 0, height - 2);
  const cropRight = clamp(
    width - rightInset - safeRightInset,
    cropX + 1,
    width
  );
  const cropBottom = clamp(
    height - bottomInset - safeBottomInset,
    cropY + 1,
    height
  );
  const cropWidth = cropRight - cropX;
  const cropHeight = cropBottom - cropY;
  const cropAreaRatio = (cropWidth * cropHeight) / (width * height);

  if (cropAreaRatio < 0.84) {
    const fallbackX = safeLeftInset;
    const fallbackY = safeTopInset;
    const fallbackRight = width - safeRightInset;
    const fallbackBottom = height - safeBottomInset;
    const fallbackWidth = fallbackRight - fallbackX;
    const fallbackHeight = fallbackBottom - fallbackY;
    const safeCanvas = document.createElement("canvas");
    const safeContext = safeCanvas.getContext("2d");
    if (!safeContext) return canvas;

    safeCanvas.width = fallbackWidth;
    safeCanvas.height = fallbackHeight;
    safeContext.fillStyle = "#ffffff";
    safeContext.fillRect(0, 0, safeCanvas.width, safeCanvas.height);
    safeContext.drawImage(
      canvas,
      fallbackX,
      fallbackY,
      fallbackWidth,
      fallbackHeight,
      0,
      0,
      fallbackWidth,
      fallbackHeight
    );

    return softenResidualBorderShadows(safeCanvas);
  }

  const resultCanvas = document.createElement("canvas");
  const resultContext = resultCanvas.getContext("2d");
  if (!resultContext) return canvas;

  resultCanvas.width = cropWidth;
  resultCanvas.height = cropHeight;
  resultContext.fillStyle = "#ffffff";
  resultContext.fillRect(0, 0, cropWidth, cropHeight);
  resultContext.drawImage(
    canvas,
    cropX,
    cropY,
    cropWidth,
    cropHeight,
    0,
    0,
    cropWidth,
    cropHeight
  );

  return softenResidualBorderShadows(resultCanvas);
};

const computeFastBoxBlur = (
  src: Float32Array,
  width: number,
  height: number,
  radius: number
) => {
  const temp = new Float32Array(width * height);
  const dst = new Float32Array(width * height);

  // Horizontal pass
  for (let y = 0; y < height; y += 1) {
    let sum = 0;
    let count = 0;
    const rowOffset = y * width;

    for (let x = -radius; x < width + radius; x += 1) {
      if (x + radius < width) {
        sum += src[rowOffset + x + radius]!;
        count += 1;
      }
      if (x - radius - 1 >= 0) {
        sum -= src[rowOffset + x - radius - 1]!;
        count -= 1;
      }
      if (x >= 0 && x < width) {
        temp[rowOffset + x] = sum / count;
      }
    }
  }

  // Vertical pass
  for (let x = 0; x < width; x += 1) {
    let sum = 0;
    let count = 0;

    for (let y = -radius; y < height + radius; y += 1) {
      if (y + radius < height) {
        sum += temp[(y + radius) * width + x]!;
        count += 1;
      }
      if (y - radius - 1 >= 0) {
        sum -= temp[(y - radius - 1) * width + x]!;
        count -= 1;
      }
      if (y >= 0 && y < height) {
        dst[y * width + x] = sum / count;
      }
    }
  }

  return dst;
};

const applyDocumentScanFilter = (canvas: HTMLCanvasElement) => {
  const cv = (window as any).cv;
  if (cv && typeof cv.Mat === "function") {
    try {
      const sourceMat = cv.imread(canvas);
      const grayMat = new cv.Mat();
      const bgMat = new cv.Mat();
      const normMat = new cv.Mat();
      const resultMat = new cv.Mat();
      const resultCanvas = document.createElement("canvas");

      resultCanvas.width = canvas.width;
      resultCanvas.height = canvas.height;

      // 1. Convert to Grayscale
      cv.cvtColor(sourceMat, grayMat, cv.COLOR_RGBA2GRAY);

      // 2. Large Gaussian blur to compute local paper background illumination
      const ksize = Math.max(
        25,
        Math.round(Math.min(canvas.width, canvas.height) / 25) | 1
      );
      cv.GaussianBlur(grayMat, bgMat, new cv.Size(ksize, ksize), 0, 0);

      // 3. Divide grayscale image by background to flatten shadows & lighting
      cv.divide(grayMat, bgMat, normMat, 255);

      // 4. Adaptive thresholding for crisp black text on pure white paper
      cv.adaptiveThreshold(
        normMat,
        resultMat,
        255,
        cv.ADAPTIVE_THRESH_GAUSSIAN_C,
        cv.THRESH_BINARY,
        35,
        8
      );

      cv.imshow(resultCanvas, resultMat);

      sourceMat.delete();
      grayMat.delete();
      bgMat.delete();
      normMat.delete();
      resultMat.delete();

      return softenResidualBorderShadows(resultCanvas);
    } catch (e) {
      console.warn("OpenCV document scan filter failed, using JS fallback", e);
    }
  }

  // Pure JS Fallback
  const context = canvas.getContext("2d", { willReadFrequently: true });
  if (!context) throw new Error("Failed to apply document scan filter.");

  const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
  const { data } = imageData;
  const width = canvas.width;
  const height = canvas.height;

  // 1. Calculate luminance array
  const luminance = new Float32Array(width * height);
  for (let i = 0; i < data.length; i += 4) {
    luminance[i / 4] =
      data[i]! * 0.299 + data[i + 1]! * 0.587 + data[i + 2]! * 0.114;
  }

  // 2. Compute fast box-blurred background luminance map
  const blurRadius = Math.max(15, Math.round(Math.min(width, height) / 35));
  const bgLuminance = computeFastBoxBlur(luminance, width, height, blurRadius);

  // 3. Binarize/Threshold against local paper background
  for (let i = 0; i < luminance.length; i += 1) {
    const idx = i * 4;
    const lum = luminance[i]!;
    const bg = bgLuminance[i]!;
    const diff = bg - lum;

    let val = 255;
    if (diff > 6) {
      val = 0;
    }

    data[idx] = val;
    data[idx + 1] = val;
    data[idx + 2] = val;
  }

  const resultCanvas = document.createElement("canvas");
  const resultContext = resultCanvas.getContext("2d");
  if (!resultContext) throw new Error("Failed to create document scan result.");

  resultCanvas.width = width;
  resultCanvas.height = height;
  resultContext.putImageData(imageData, 0, 0);

  return softenResidualBorderShadows(resultCanvas);
};

const applyMagicColorFilter = (canvas: HTMLCanvasElement) => {
  const context = canvas.getContext("2d");
  if (!context) throw new Error("Failed to apply magic color filter.");

  const imageData = context.getImageData(0, 0, canvas.width, canvas.height);
  const { data } = imageData;
  let luminanceTotal = 0;
  let redTotal = 0;
  let greenTotal = 0;
  let blueTotal = 0;

  for (let index = 0; index < data.length; index += 4) {
    const red = data[index]!;
    const green = data[index + 1]!;
    const blue = data[index + 2]!;

    redTotal += red;
    greenTotal += green;
    blueTotal += blue;
    luminanceTotal += red * 0.299 + green * 0.587 + blue * 0.114;
  }

  const pixelCount = Math.max(1, data.length / 4);
  const averageRed = redTotal / pixelCount;
  const averageGreen = greenTotal / pixelCount;
  const averageBlue = blueTotal / pixelCount;
  const averageChannel = (averageRed + averageGreen + averageBlue) / 3;
  const redGain = clamp(averageChannel / Math.max(1, averageRed), 0.86, 1.18);
  const greenGain = clamp(
    averageChannel / Math.max(1, averageGreen),
    0.86,
    1.18
  );
  const blueGain = clamp(averageChannel / Math.max(1, averageBlue), 0.86, 1.18);
  const averageLuminance = luminanceTotal / pixelCount;
  const brightnessBoost = averageLuminance < 145 ? 30 : 18;
  const contrast = averageLuminance < 145 ? 1.52 : 1.4;
  const baseSaturation = 1.5;
  const contrastOffset = 128 * (1 - contrast);

  for (let index = 0; index < data.length; index += 4) {
    let red = data[index]! * redGain;
    let green = data[index + 1]! * greenGain;
    let blue = data[index + 2]! * blueGain;

    red = red * contrast + contrastOffset + brightnessBoost;
    green = green * contrast + contrastOffset + brightnessBoost;
    blue = blue * contrast + contrastOffset + brightnessBoost;

    const maxChannel = Math.max(red, green, blue);
    const minChannel = Math.min(red, green, blue);
    const currentSaturation =
      maxChannel > 0 ? (maxChannel - minChannel) / maxChannel : 0;
    const vibrance = 1 + (1 - currentSaturation) * 0.55;
    const saturation = baseSaturation * vibrance;
    const gray = red * 0.299 + green * 0.587 + blue * 0.114;

    red = gray + (red - gray) * saturation;
    green = gray + (green - gray) * saturation;
    blue = gray + (blue - gray) * saturation;

    const enhancedGray = red * 0.299 + green * 0.587 + blue * 0.114;

    if (enhancedGray > 205) {
      const paperWhitenAmount = clamp((enhancedGray - 205) / 50, 0, 0.42);
      red += (255 - red) * paperWhitenAmount;
      green += (255 - green) * paperWhitenAmount;
      blue += (255 - blue) * paperWhitenAmount;
    }

    if (enhancedGray < 95) {
      const textDarkenAmount = clamp((95 - enhancedGray) / 95, 0, 0.28);
      red *= 1 - textDarkenAmount;
      green *= 1 - textDarkenAmount;
      blue *= 1 - textDarkenAmount;
    }

    data[index] = clamp(red, 0, 255);
    data[index + 1] = clamp(green, 0, 255);
    data[index + 2] = clamp(blue, 0, 255);
  }

  const resultCanvas = document.createElement("canvas");
  const resultContext = resultCanvas.getContext("2d");
  if (!resultContext) throw new Error("Failed to create magic color result.");

  resultCanvas.width = canvas.width;
  resultCanvas.height = canvas.height;
  resultContext.putImageData(imageData, 0, 0);

  return resultCanvas;
};

const cropImageFile = async (file: File) => {
  await loadScannerScripts();

  const image = await loadImageFromFile(file);
  const sourceWidth = image.naturalWidth || image.width;
  const sourceHeight = image.naturalHeight || image.height;
  const sourceCanvas = document.createElement("canvas");
  const sourceContext = sourceCanvas.getContext("2d");

  if (!sourceContext) throw new Error("Failed to crop image.");

  sourceCanvas.width = sourceWidth;
  sourceCanvas.height = sourceHeight;
  sourceContext.drawImage(image, 0, 0, sourceWidth, sourceHeight);

  const corners = cropCorners.value;
  const sourceCorners = {
    bottomLeft: {
      x: (corners.bottomLeft.x / 100) * sourceWidth,
      y: (corners.bottomLeft.y / 100) * sourceHeight,
    },
    bottomRight: {
      x: (corners.bottomRight.x / 100) * sourceWidth,
      y: (corners.bottomRight.y / 100) * sourceHeight,
    },
    topLeft: {
      x: (corners.topLeft.x / 100) * sourceWidth,
      y: (corners.topLeft.y / 100) * sourceHeight,
    },
    topRight: {
      x: (corners.topRight.x / 100) * sourceWidth,
      y: (corners.topRight.y / 100) * sourceHeight,
    },
  };
  const { height: outputHeight, width: outputWidth } =
    getNormalizedPerspectiveSize(sourceCorners);
  const cv = (window as any).cv;
  const sourceMat = cv.imread(sourceCanvas);
  const outputMat = new cv.Mat();
  const outputSize = new cv.Size(outputWidth, outputHeight);
  const sourceTriangle = cv.matFromArray(4, 1, cv.CV_32FC2, [
    sourceCorners.topLeft.x,
    sourceCorners.topLeft.y,
    sourceCorners.topRight.x,
    sourceCorners.topRight.y,
    sourceCorners.bottomLeft.x,
    sourceCorners.bottomLeft.y,
    sourceCorners.bottomRight.x,
    sourceCorners.bottomRight.y,
  ]);
  const destinationTriangle = cv.matFromArray(4, 1, cv.CV_32FC2, [
    0,
    0,
    outputWidth - 1,
    0,
    0,
    outputHeight - 1,
    outputWidth - 1,
    outputHeight - 1,
  ]);
  const transform = cv.getPerspectiveTransform(
    sourceTriangle,
    destinationTriangle
  );
  const outputCanvas = document.createElement("canvas");

  try {
    cv.warpPerspective(
      sourceMat,
      outputMat,
      transform,
      outputSize,
      cv.INTER_LINEAR,
      cv.BORDER_CONSTANT,
      new cv.Scalar(255, 255, 255, 255)
    );
    cv.imshow(outputCanvas, outputMat);
  } finally {
    sourceMat.delete();
    outputMat.delete();
    sourceTriangle.delete();
    destinationTriangle.delete();
    transform.delete();
  }

  const cleanedOutputCanvas = cleanupPerspectiveEdges(outputCanvas);

  if (selectedCropFilter.value === "cleanPaper") {
    return canvasToFile(applyCleanPaperFilter(cleanedOutputCanvas), file);
  }

  if (selectedCropFilter.value === "blackWhite") {
    return canvasToFile(applyBetterBlackWhiteFilter(cleanedOutputCanvas), file);
  }

  if (selectedCropFilter.value === "documentScan") {
    return canvasToFile(applyDocumentScanFilter(cleanedOutputCanvas), file);
  }

  if (selectedCropFilter.value === "magicColor") {
    return canvasToFile(applyMagicColorFilter(cleanedOutputCanvas), file);
  }

  if (cropFilterCss.value === "none") {
    return canvasToFile(cleanedOutputCanvas, file);
  }

  const filteredCanvas = document.createElement("canvas");
  const filteredContext = filteredCanvas.getContext("2d");

  if (!filteredContext) throw new Error("Failed to apply image filter.");

  filteredCanvas.width = cleanedOutputCanvas.width;
  filteredCanvas.height = cleanedOutputCanvas.height;
  filteredContext.filter = cropFilterCss.value;
  filteredContext.drawImage(cleanedOutputCanvas, 0, 0);
  filteredContext.filter = "none";

  return canvasToFile(filteredCanvas, file);
};

const clearCropPreview = (options: { regeneratePreview?: boolean } = {}) => {
  cropPreviewGenerationId += 1;
  cropPreviewCacheKey.value = null;
  if (cropPreviewImageUrl.value) URL.revokeObjectURL(cropPreviewImageUrl.value);
  cropPreviewImageUrl.value = null;
  isGeneratingCropPreview.value = false;

  if (options.regeneratePreview) {
    void nextTick(() => generateCropPreview());
  }
};

const rotateCropImage = async () => {
  if (!cropSourceFile.value || isRotatingImage.value) return;

  isRotatingImage.value = true;
  setError(null);

  try {
    const image = await loadImageFromFile(cropSourceFile.value);
    const sourceWidth = image.naturalWidth || image.width;
    const sourceHeight = image.naturalHeight || image.height;
    const canvas = document.createElement("canvas");
    const context = canvas.getContext("2d");

    if (!context) throw new Error("Failed to rotate image.");

    canvas.width = sourceHeight;
    canvas.height = sourceWidth;
    context.translate(canvas.width, 0);
    context.rotate(Math.PI / 2);
    context.drawImage(image, 0, 0, sourceWidth, sourceHeight);

    const rotatedFile = await canvasToFile(canvas, cropSourceFile.value, null);

    clearCropPreview();
    if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value);
    cropSourceFile.value = rotatedFile;
    cropImageUrl.value = URL.createObjectURL(rotatedFile);
    cropRect.value = { height: 90, width: 90, x: 5, y: 5 };
    cropCorners.value = createDefaultCropCorners();
    void detectInitialCrop(rotatedFile);
  } catch (err: any) {
    setError(err?.message || "Failed to rotate image.");
  } finally {
    isRotatingImage.value = false;
  }
};

const closeCropDialog = () => {
  cropPreviewGenerationId += 1;
  cropPreviewCacheKey.value = null;
  cropDetectionDebug.value = null;
  showCropDebug.value = false;
  cropDialogOpen.value = false;
  cropSourceFile.value = null;
  cropViewMode.value = "edit";
  canRetakeCropPhoto.value = false;
  isGeneratingCropPreview.value = false;

  if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value);
  if (cropPreviewImageUrl.value) URL.revokeObjectURL(cropPreviewImageUrl.value);
  cropImageUrl.value = null;
  cropPreviewImageUrl.value = null;
};

const retakePhoto = async () => {
  closeCropDialog();
  await nextTick();
  await new Promise<void>((resolve) => window.setTimeout(resolve, 120));
  await openCamera();
};

const useOriginalImage = () => {
  if (!cropSourceFile.value) return;

  const sourceFile = cropSourceFile.value;

  editableDocumentSourceFile.value = sourceFile;
  lastAppliedCropRect.value = { ...cropRect.value };
  lastAppliedCropCorners.value = cloneCropCorners(cropCorners.value);
  lastAppliedCropFilter.value = selectedCropFilter.value;
  completeActiveImageFile(sourceFile);
};

const generateCropPreview = async () => {
  if (
    !cropDialogOpen.value ||
    !cropSourceFile.value ||
    isDetectingCrop.value ||
    isGeneratingCropPreview.value
  ) {
    return;
  }

  const sourceFile = cropSourceFile.value;

  if (hasValidCropPreviewCache(sourceFile)) {
    return;
  }

  const previewCacheKey = createCropPreviewCacheKey(sourceFile);
  const generationId = ++cropPreviewGenerationId;

  isGeneratingCropPreview.value = true;
  setError(null);

  try {
    const previewFile = await cropImageFile(sourceFile);
    const nextPreviewUrl = URL.createObjectURL(previewFile);

    if (
      generationId !== cropPreviewGenerationId ||
      !cropDialogOpen.value ||
      cropSourceFile.value !== sourceFile
    ) {
      URL.revokeObjectURL(nextPreviewUrl);
      return;
    }

    if (cropPreviewImageUrl.value)
      URL.revokeObjectURL(cropPreviewImageUrl.value);
    cropPreviewImageUrl.value = nextPreviewUrl;
    cropPreviewCacheKey.value = previewCacheKey;
  } catch (err: any) {
    if (generationId === cropPreviewGenerationId) {
      cropViewMode.value = "edit";
      setError(err?.message || "Failed to generate preview.");
    }
  } finally {
    isGeneratingCropPreview.value = false;
  }
};

const setCropViewMode = (mode: CropViewMode) => {
  cropViewMode.value = mode;

  if (mode === "preview" && cropSourceFile.value) {
    if (!hasValidCropPreviewCache(cropSourceFile.value)) {
      clearCropPreview();
    }

    void generateCropPreview();
  }
};

const applyCrop = async () => {
  if (!cropSourceFile.value) return;

  isProcessingImage.value = true;
  setError(null);

  try {
    const sourceFile = cropSourceFile.value;
    const croppedFile = await cropImageFile(sourceFile);

    editableDocumentSourceFile.value = sourceFile;
    lastAppliedCropRect.value = { ...cropRect.value };
    lastAppliedCropCorners.value = cloneCropCorners(cropCorners.value);
    lastAppliedCropFilter.value = selectedCropFilter.value;
    completeActiveImageFile(croppedFile);
  } catch (err: any) {
    setError(err?.message || "Failed to crop image.");
  } finally {
    isProcessingImage.value = false;
  }
};

const editLastImage = () => {
  if (!editableDocumentSourceFile.value) return;

  setError(null);
  openCropEditor(editableDocumentSourceFile.value, { restoreLastEdit: true });
};

const clearEditableImage = () => {
  editableDocumentSourceFile.value = null;
  lastAppliedCropRect.value = { height: 90, width: 90, x: 5, y: 5 };
  lastAppliedCropCorners.value = createDefaultCropCorners();
  lastAppliedCropFilter.value = "magicColor";
};

const stopCropPointer = () => {
  cropPointerCleanup?.();
  cropPointerCleanup = null;
};

type CropPointerMode = "move" | CropCornerKey;

const resetAutoCaptureState = () => {
  autoCaptureStartedAt = null;
  autoCaptureProgress.value = 0;
  isAutoCapturing.value = false;
};

const stopCameraPreviewLoop = () => {
  if (cameraFrameTimer !== null) {
    window.clearTimeout(cameraFrameTimer);
    cameraFrameTimer = null;
  }
};

const getCameraVideoTrack = () => cameraStream?.getVideoTracks?.()[0] ?? null;

const refreshTorchSupport = () => {
  const track = getCameraVideoTrack();
  const capabilities = track?.getCapabilities?.() as
    | { torch?: boolean }
    | undefined;

  isTorchSupported.value = !!capabilities?.torch;
  if (!isTorchSupported.value) {
    isTorchOn.value = false;
  }
};

const setTorch = async (enabled: boolean) => {
  const track = getCameraVideoTrack();

  if (!track || !isTorchSupported.value) return;

  try {
    await track.applyConstraints({
      advanced: [{ torch: enabled } as MediaTrackConstraintSet],
    });
    isTorchOn.value = enabled;
  } catch (err: any) {
    setError(err?.message || "Flash is not available on this device.");
    isTorchOn.value = false;
  }
};

const toggleTorch = async () => {
  await setTorch(!isTorchOn.value);
};

const stopCamera = () => {
  stopCameraPreviewLoop();
  resetAutoCaptureState();
  if (isTorchOn.value) {
    void setTorch(false);
  }
  isTorchSupported.value = false;
  isTorchOn.value = false;
  cameraQualityHint.value = {
    detail: "Make sure the whole paper is visible",
    label: "Point camera at the document",
    tone: "warning",
  };
  cameraStream?.getTracks().forEach((track) => track.stop());
  cameraStream = null;

  if (cameraVideoRef.value) {
    cameraVideoRef.value.srcObject = null;
  }

  const overlayCanvas = cameraCanvasRef.value;
  const overlayContext = overlayCanvas?.getContext("2d");
  if (overlayCanvas && overlayContext) {
    overlayContext.clearRect(0, 0, overlayCanvas.width, overlayCanvas.height);
  }
};

const closeCameraDialog = () => {
  cameraDialogOpen.value = false;
  stopCamera();
};

const startCameraPreviewLoop = () => {
  stopCameraPreviewLoop();

  const video = cameraVideoRef.value;
  const overlayCanvas = cameraCanvasRef.value;
  if (!video || !overlayCanvas) return;

  const Scanner = (window as any).jscanify;
  if (!Scanner) return;

  const scanner = new Scanner();
  const cv = (window as any).cv;
  const workCanvas = document.createElement("canvas");
  const workContext = workCanvas.getContext("2d");
  const overlayContext = overlayCanvas.getContext("2d");

  if (!cv || !workContext || !overlayContext) return;

  let previousCameraCorners: CropCorners | null = null;
  let pendingCameraQualityHint: CameraQualityHint | null = null;
  let pendingCameraQualityHintCount = 0;

  const getFrameBrightness = () => {
    const imageData = workContext.getImageData(
      0,
      0,
      workCanvas.width,
      workCanvas.height
    );
    const { data } = imageData;
    let total = 0;

    for (let index = 0; index < data.length; index += 16) {
      total +=
        data[index]! * 0.299 +
        data[index + 1]! * 0.587 +
        data[index + 2]! * 0.114;
    }

    return total / Math.max(1, data.length / 16);
  };

  const getFrameFocusScore = (mat: any) => {
    const gray = new cv.Mat();
    const laplacian = new cv.Mat();
    const mean = new cv.Mat();
    const stddev = new cv.Mat();

    try {
      cv.cvtColor(mat, gray, cv.COLOR_RGBA2GRAY);
      cv.Laplacian(gray, laplacian, cv.CV_64F);
      cv.meanStdDev(laplacian, mean, stddev);

      const deviation = stddev.doubleAt(0, 0);
      return deviation * deviation;
    } finally {
      gray.delete();
      laplacian.delete();
      mean.delete();
      stddev.delete();
    }
  };

  const mapCameraCorners = (corners: {
    bottomLeftCorner: CropPoint;
    bottomRightCorner: CropPoint;
    topLeftCorner: CropPoint;
    topRightCorner: CropPoint;
  }): CropCorners => ({
    bottomLeft: {
      x: (corners.bottomLeftCorner.x / workCanvas.width) * 100,
      y: (corners.bottomLeftCorner.y / workCanvas.height) * 100,
    },
    bottomRight: {
      x: (corners.bottomRightCorner.x / workCanvas.width) * 100,
      y: (corners.bottomRightCorner.y / workCanvas.height) * 100,
    },
    topLeft: {
      x: (corners.topLeftCorner.x / workCanvas.width) * 100,
      y: (corners.topLeftCorner.y / workCanvas.height) * 100,
    },
    topRight: {
      x: (corners.topRightCorner.x / workCanvas.width) * 100,
      y: (corners.topRightCorner.y / workCanvas.height) * 100,
    },
  });

  const updateAutoCapture = (isReady: boolean) => {
    if (!props.autoCapture || isAutoCapturing.value) return;

    if (!isReady) {
      autoCaptureStartedAt = null;
      autoCaptureProgress.value = 0;
      return;
    }

    const now = Date.now();
    autoCaptureStartedAt ??= now;
    const elapsed = now - autoCaptureStartedAt;
    autoCaptureProgress.value = clamp(
      (elapsed / props.autoCaptureDelayMs) * 100,
      0,
      100
    );

    if (elapsed < props.autoCaptureDelayMs) return;

    isAutoCapturing.value = true;
    autoCaptureProgress.value = 100;
    void captureCameraFrame().finally(() => {
      resetAutoCaptureState();
    });
  };

  const cameraQualityHintKey = (hint: CameraQualityHint) =>
    `${hint.tone}:${hint.label}`;

  const setSmoothedCameraQualityHint = (
    nextHint: CameraQualityHint,
    options: { immediate?: boolean } = {}
  ) => {
    const currentKey = cameraQualityHintKey(cameraQualityHint.value);
    const nextKey = cameraQualityHintKey(nextHint);

    if (currentKey === nextKey) {
      pendingCameraQualityHint = null;
      pendingCameraQualityHintCount = 0;
      return;
    }

    if (options.immediate) {
      cameraQualityHint.value = nextHint;
      pendingCameraQualityHint = null;
      pendingCameraQualityHintCount = 0;
      return;
    }

    if (
      pendingCameraQualityHint &&
      cameraQualityHintKey(pendingCameraQualityHint) === nextKey
    ) {
      pendingCameraQualityHintCount += 1;
    } else {
      pendingCameraQualityHint = nextHint;
      pendingCameraQualityHintCount = 1;
    }

    if (pendingCameraQualityHintCount >= 2) {
      cameraQualityHint.value = nextHint;
      pendingCameraQualityHint = null;
      pendingCameraQualityHintCount = 0;
    }
  };

  const smoothCameraCorners = (
    previous: CropCorners,
    next: CropCorners,
    strength = 0.35
  ): CropCorners => {
    const smoothPoint = (from: CropPoint, to: CropPoint) => ({
      x: from.x + (to.x - from.x) * strength,
      y: from.y + (to.y - from.y) * strength,
    });

    return {
      bottomLeft: smoothPoint(previous.bottomLeft, next.bottomLeft),
      bottomRight: smoothPoint(previous.bottomRight, next.bottomRight),
      topLeft: smoothPoint(previous.topLeft, next.topLeft),
      topRight: smoothPoint(previous.topRight, next.topRight),
    };
  };

  const rememberCameraCorners = (corners: CropCorners) => {
    previousCameraCorners = previousCameraCorners
      ? smoothCameraCorners(previousCameraCorners, corners)
      : cloneCropCorners(corners);
  };

  const updateCameraQualityHint = (
    contour: any,
    corners: CropCorners | null,
    brightness: number,
    focusScore: number
  ) => {
    if (brightness < 65) {
      setSmoothedCameraQualityHint(
        {
          detail: "Move to a brighter area or turn on more light",
          label: "Need more light",
          tone: "danger",
        },
        { immediate: true }
      );
      if (corners) rememberCameraCorners(corners);
      return false;
    }

    if (!contour || !corners) {
      setSmoothedCameraQualityHint({
        detail: "Make sure all four paper corners are visible",
        label: "Point camera at the document",
        tone: "warning",
      });
      previousCameraCorners = null;
      return false;
    }

    const areaRatio =
      cv.contourArea(contour) / (workCanvas.width * workCanvas.height);
    const points = Object.values(corners);
    const minX = Math.min(...points.map((point) => point.x));
    const maxX = Math.max(...points.map((point) => point.x));
    const minY = Math.min(...points.map((point) => point.y));
    const maxY = Math.max(...points.map((point) => point.y));
    const nearEdge = minX < 3 || minY < 3 || maxX > 97 || maxY > 97;

    if (areaRatio < 0.1) {
      setSmoothedCameraQualityHint({
        detail: "Fill more of the frame with the paper",
        label: "Move closer",
        tone: "warning",
      });
      rememberCameraCorners(corners);
      return false;
    }

    if (areaRatio > 0.92 || nearEdge) {
      setSmoothedCameraQualityHint({
        detail: "Keep all paper edges inside the frame",
        label: nearEdge ? "Center the document" : "Move farther",
        tone: "warning",
      });
      rememberCameraCorners(corners);
      return false;
    }

    if (focusScore < 45) {
      setSmoothedCameraQualityHint({
        detail: "Tap the document or hold the phone a bit steadier",
        label: "Improve focus",
        tone: "warning",
      });
      rememberCameraCorners(corners);
      return false;
    }

    if (previousCameraCorners) {
      const movement =
        Object.keys(corners).reduce((total, key) => {
          const cornerKey = key as CropCornerKey;
          return (
            total +
            distance(corners[cornerKey], previousCameraCorners![cornerKey])
          );
        }, 0) / 4;

      if (movement > 14) {
        setSmoothedCameraQualityHint({
          detail: "Keep the phone steady for a sharper scan",
          label: "Hold steady",
          tone: "warning",
        });
        rememberCameraCorners(corners);
        return false;
      }
    }

    rememberCameraCorners(corners);

    setSmoothedCameraQualityHint({
      detail: "Document is centered and stable",
      label: "Ready to capture",
      tone: "success",
    });
    return true;
  };

  const drawPaperContour = (contour: any) => {
    if (!contour) return null;

    const {
      bottomLeftCorner,
      bottomRightCorner,
      topLeftCorner,
      topRightCorner,
    } = scanner.getCornerPoints(contour);

    if (
      !topLeftCorner ||
      !topRightCorner ||
      !bottomLeftCorner ||
      !bottomRightCorner
    ) {
      return null;
    }

    overlayContext.beginPath();
    overlayContext.moveTo(topLeftCorner.x, topLeftCorner.y);
    overlayContext.lineTo(topRightCorner.x, topRightCorner.y);
    overlayContext.lineTo(bottomRightCorner.x, bottomRightCorner.y);
    overlayContext.lineTo(bottomLeftCorner.x, bottomLeftCorner.y);
    overlayContext.closePath();
    overlayContext.lineWidth = 4;
    overlayContext.lineJoin = "round";
    overlayContext.strokeStyle = "#22c55e";
    overlayContext.shadowBlur = 8;
    overlayContext.shadowColor = "rgba(34, 197, 94, 0.55)";
    overlayContext.stroke();
    overlayContext.shadowBlur = 0;

    return mapCameraCorners({
      bottomLeftCorner,
      bottomRightCorner,
      topLeftCorner,
      topRightCorner,
    });
  };

  const scanFrame = () => {
    if (!cameraDialogOpen.value) return;

    if (!video.videoWidth || !video.videoHeight) {
      cameraFrameTimer = window.setTimeout(
        scanFrame,
        CAMERA_HIGHLIGHT_INTERVAL_MS
      );
      return;
    }

    const scale = Math.min(1, CAMERA_HIGHLIGHT_MAX_WIDTH / video.videoWidth);
    const nextWidth = Math.max(1, Math.round(video.videoWidth * scale));
    const nextHeight = Math.max(1, Math.round(video.videoHeight * scale));

    if (workCanvas.width !== nextWidth || workCanvas.height !== nextHeight) {
      workCanvas.width = nextWidth;
      workCanvas.height = nextHeight;
      overlayCanvas.width = nextWidth;
      overlayCanvas.height = nextHeight;
    }

    workContext.drawImage(video, 0, 0, workCanvas.width, workCanvas.height);

    try {
      const brightness = getFrameBrightness();
      const mat = cv.imread(workCanvas);
      const focusScore = getFrameFocusScore(mat);
      const contour = findCameraHintContour(scanner, mat);

      try {
        overlayContext.clearRect(
          0,
          0,
          overlayCanvas.width,
          overlayCanvas.height
        );
        const detectedCorners = drawPaperContour(contour);
        const isReadyToCapture = updateCameraQualityHint(
          contour,
          detectedCorners,
          brightness,
          focusScore
        );
        updateAutoCapture(isReadyToCapture);
      } finally {
        contour?.delete?.();
        mat.delete();
      }
    } catch {
      overlayContext.clearRect(0, 0, overlayCanvas.width, overlayCanvas.height);
      updateAutoCapture(false);
      setSmoothedCameraQualityHint({
        detail: "Make sure the paper is visible and well lit",
        label: "Scanning document",
        tone: "warning",
      });
    }

    cameraFrameTimer = window.setTimeout(
      scanFrame,
      CAMERA_HIGHLIGHT_INTERVAL_MS
    );
  };

  scanFrame();
};

const getPreferredCameraConstraints = (): MediaStreamConstraints => ({
  audio: false,
  video: {
    facingMode: { ideal: "environment" },
    frameRate: { ideal: 30, max: 30 },
    height: { ideal: 1080, max: 1440, min: 720 },
    width: { ideal: 1920, max: 2560, min: 1280 },
  },
});

const getFallbackCameraConstraints = (): MediaStreamConstraints => ({
  audio: false,
  video: {
    facingMode: { ideal: "environment" },
    height: { ideal: 720 },
    width: { ideal: 1280 },
  },
});

const openPreferredCameraStream = async () => {
  try {
    return await navigator.mediaDevices.getUserMedia(
      getPreferredCameraConstraints()
    );
  } catch {
    return navigator.mediaDevices.getUserMedia(getFallbackCameraConstraints());
  }
};

const applyCameraQualityConstraints = async () => {
  const track = getCameraVideoTrack();
  if (!track?.getCapabilities || !track.applyConstraints) return;

  const capabilities = track.getCapabilities() as {
    exposureMode?: string[];
    focusMode?: string[];
    torch?: boolean;
    whiteBalanceMode?: string[];
  };
  const advanced: Record<string, string>[] = [];

  if (capabilities.focusMode?.includes("continuous")) {
    advanced.push({ focusMode: "continuous" });
  }

  if (capabilities.exposureMode?.includes("continuous")) {
    advanced.push({ exposureMode: "continuous" });
  }

  if (capabilities.whiteBalanceMode?.includes("continuous")) {
    advanced.push({ whiteBalanceMode: "continuous" });
  }

  if (!advanced.length) return;

  try {
    await track.applyConstraints({ advanced } as MediaTrackConstraints);
  } catch {
    // Some mobile browsers expose capabilities but reject advanced constraints.
  }
};

const openCamera = async () => {
  setError(null);

  if (!props.enableScanner && !isNativeMobilePlatform()) {
    triggerCameraPick();
    return;
  }

  if (!navigator.mediaDevices?.getUserMedia) {
    setError(
      "Realtime camera is not available in this browser. Opening native camera picker instead."
    );
    triggerCameraPick();
    return;
  }

  isCameraStarting.value = true;
  cameraDialogOpen.value = true;
  await nextTick();

  try {
    cameraStream = await openPreferredCameraStream();

    const video = cameraVideoRef.value;
    if (!video) throw new Error("Camera preview is not ready.");

    video.srcObject = cameraStream;
    await applyCameraQualityConstraints();
    refreshTorchSupport();

    if (isTorchSupported.value) {
      void setTorch(true);
    }

    await new Promise<void>((resolve) => {
      video.onloadedmetadata = () => resolve();
    });

    await video.play();
    isCameraStarting.value = false;

    try {
      await loadScannerScripts();
      startCameraPreviewLoop();
    } catch (scannerError: any) {
      setError(
        scannerError?.message ||
          "Document highlight is not available, but you can still capture manually."
      );
    }
  } catch (err: any) {
    setError(err?.message || "Failed to open camera.");
    stopCamera();
  } finally {
    isCameraStarting.value = false;
  }
};

const handleCameraChange = (event: Event) => {
  setError(null);

  const target = event.target as HTMLInputElement | null;
  const file = target?.files?.[0];
  if (!file) return;

  void resizeImageFileForProcessing(file).then((processedFile) => {
    resetQueue();
    editableDocumentSourceFile.value = processedFile;
    openCropEditor(processedFile, { source: "camera" });
  });
  if (target) target.value = "";
};

const waitForCameraDialogClose = async () => {
  await nextTick();
  await new Promise<void>((resolve) => window.setTimeout(resolve, 180));
};

const captureCameraFrame = async () => {
  const video = cameraVideoRef.value;
  if (!video || !video.videoWidth || !video.videoHeight) return;

  const canvas = document.createElement("canvas");
  const context = canvas.getContext("2d");
  if (!context) return;

  canvas.width = video.videoWidth;
  canvas.height = video.videoHeight;
  context.drawImage(video, 0, 0, canvas.width, canvas.height);

  const file = await canvasToFile(
    canvas,
    new File([], `document-camera-${Date.now()}.jpg`, {
      type: "image/jpeg",
    })
  );

  const processedFile = await resizeImageFileForProcessing(file);

  resetQueue();
  editableDocumentSourceFile.value = processedFile;
  closeCameraDialog();
  await waitForCameraDialogClose();
  openCropEditor(processedFile, { source: "camera" });
};

const startCropPointer = (mode: CropPointerMode, event: PointerEvent) => {
  if (cropViewMode.value !== "edit") return;

  const frame = cropFrameRef.value;
  if (!frame) return;

  event.preventDefault();
  stopCropPointer();

  const frameBounds = frame.getBoundingClientRect();
  const startX = event.clientX;
  const startY = event.clientY;
  const startCorners = cloneCropCorners(cropCorners.value);

  const onPointerMove = (moveEvent: PointerEvent) => {
    const deltaX = ((moveEvent.clientX - startX) / frameBounds.width) * 100;
    const deltaY = ((moveEvent.clientY - startY) / frameBounds.height) * 100;

    if (mode === "move") {
      const points = Object.values(startCorners);
      const minX = Math.min(...points.map((point) => point.x));
      const maxX = Math.max(...points.map((point) => point.x));
      const minY = Math.min(...points.map((point) => point.y));
      const maxY = Math.max(...points.map((point) => point.y));
      const nextDeltaX = clamp(deltaX, -minX, 100 - maxX);
      const nextDeltaY = clamp(deltaY, -minY, 100 - maxY);

      cropCorners.value = {
        bottomLeft: {
          x: startCorners.bottomLeft.x + nextDeltaX,
          y: startCorners.bottomLeft.y + nextDeltaY,
        },
        bottomRight: {
          x: startCorners.bottomRight.x + nextDeltaX,
          y: startCorners.bottomRight.y + nextDeltaY,
        },
        topLeft: {
          x: startCorners.topLeft.x + nextDeltaX,
          y: startCorners.topLeft.y + nextDeltaY,
        },
        topRight: {
          x: startCorners.topRight.x + nextDeltaX,
          y: startCorners.topRight.y + nextDeltaY,
        },
      };
      syncCropRectFromCorners();
      return;
    }

    cropCorners.value = {
      ...cloneCropCorners(startCorners),
      [mode]: {
        x: clamp(startCorners[mode].x + deltaX, 0, 100),
        y: clamp(startCorners[mode].y + deltaY, 0, 100),
      },
    };
    syncCropRectFromCorners();
  };

  const onPointerUp = () => stopCropPointer();

  window.addEventListener("pointermove", onPointerMove);
  window.addEventListener("pointerup", onPointerUp, { once: true });

  cropPointerCleanup = () => {
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
  };
};

const handleDrop = async (event: DragEvent) => {
  event.preventDefault();
  setError(null);

  const files = Array.from(event.dataTransfer?.files ?? []);
  await processSelectedFiles(files, createFilesEvent(files));
};

watch(selectedCropFilter, () => {
  if (
    cropDialogOpen.value &&
    cropViewMode.value === "preview" &&
    !isDetectingCrop.value
  ) {
    if (cropPreviewImageUrl.value)
      URL.revokeObjectURL(cropPreviewImageUrl.value);
    cropPreviewImageUrl.value = null;
    void generateCropPreview();
  }
});

watch(
  canEditSelectedImage,
  (value) => {
    emit("update:canEdit", value);
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  stopCamera();
  stopCropPointer();
  if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value);
  if (cropPreviewImageUrl.value) URL.revokeObjectURL(cropPreviewImageUrl.value);
});

defineExpose({
  canEditSelectedImage,
  clearEditableImage,
  editLastImage,
  handleDrop,
  isCameraStarting,
  isProcessingImage,
  openCamera,
  openCropEditor,
  openFilePicker,
  scannerError,
});
</script>

<template>
  <input
    ref="fileInput"
    type="file"
    :accept="accept"
    :multiple="multiple"
    class="hidden"
    @change="handleFileChange"
  />

  <input
    ref="cameraInput"
    type="file"
    accept="image/*"
    capture="environment"
    class="hidden"
    @change="handleCameraChange"
  />

  <AppDialog
    v-model="cameraDialogOpen"
    :title="cameraTitle"
    :subtitle="cameraSubtitle"
    width="min(920px, 96vw)"
    max-width="96vw"
    height="auto"
    max-height="92dvh"
    :mobile-actions="cameraDialogMobileActions"
    @hide="stopCamera"
  >
    <div class="space-y-3">
      <div class="relative overflow-hidden rounded-2xl bg-slate-950">
        <div
          class="absolute left-3 top-3 z-10 max-w-[calc(100%-1.5rem)] rounded-full border px-3 py-1.5 text-xs font-bold shadow-lg backdrop-blur-sm"
          :class="cameraQualityHintClass"
        >
          <span class="mr-1 inline-block h-2 w-2 rounded-full bg-current" />
          {{ cameraQualityHint.label }}
          <span class="ml-1 font-medium opacity-80">
            {{ cameraQualityHint.detail }}
          </span>
          <span
            v-if="autoCaptureCountdownText"
            class="ml-2 rounded-full bg-white/60 px-2 py-0.5"
          >
            {{ autoCaptureCountdownText }}
          </span>
          <span
            v-if="props.autoCapture && autoCaptureProgress > 0"
            class="absolute inset-x-3 bottom-0 h-0.5 overflow-hidden rounded-full bg-current/20"
          >
            <span
              class="block h-full rounded-full bg-current transition-all duration-200"
              :style="{ width: `${autoCaptureProgress}%` }"
            />
          </span>
        </div>
        <video
          ref="cameraVideoRef"
          autoplay
          muted
          playsinline
          class="block max-h-[68dvh] w-full object-contain"
        />
        <canvas
          ref="cameraCanvasRef"
          class="pointer-events-none absolute inset-0 h-full w-full object-contain"
        />
      </div>
      <p class="text-xs text-slate-500">
        Pastikan seluruh dokumen terlihat di frame. Garis highlight akan muncul
        jika dokumen terdeteksi. Setelah capture, kamu masih bisa adjust crop
        manual sebelum upload.
      </p>
      <p v-if="scannerError" class="text-xs font-medium text-red-600">
        {{ scannerError }}
      </p>
    </div>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <AppButton
          v-if="isTorchSupported"
          type="button"
          variant="secondary"
          icon="pi-bolt"
          :disabled="isCameraStarting"
          @click="toggleTorch"
        >
          {{ isTorchOn ? "Flash On" : "Flash" }}
        </AppButton>
        <AppButton
          icon="pi-camera"
          :disabled="isCameraStarting || isAutoCapturing"
          :loading="isAutoCapturing"
          @click="captureCameraFrame"
        >
          {{
            isCameraStarting
              ? "Opening Camera..."
              : isAutoCapturing
                ? "Capturing..."
                : "Capture"
          }}
        </AppButton>
      </div>
    </template>
  </AppDialog>

  <AppDialog
    v-model="cropDialogOpen"
    :title="cropTitle"
    :subtitle="cropSubtitle"
    width="min(920px, 96vw)"
    max-width="96vw"
    height="auto"
    max-height="92dvh"
    :mobile-actions="cropDialogMobileActions"
  >
    <div class="space-y-4">
      <div
        class="flex w-full overflow-hidden rounded-xl border border-slate-200 bg-slate-50 p-0.5"
      >
        <button
          type="button"
          class="inline-flex min-w-0 flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-xs font-bold transition"
          :class="
            cropViewMode === 'edit'
              ? 'bg-white text-[#dc2626] shadow-sm'
              : 'text-slate-500 hover:text-slate-800'
          "
          @click="setCropViewMode('edit')"
        >
          <span class="pi pi-pencil text-[11px]" />
          Edit
        </button>
        <button
          type="button"
          class="inline-flex min-w-0 flex-1 items-center justify-center gap-1.5 rounded-lg px-3 py-2 text-xs font-bold transition"
          :class="
            cropViewMode === 'preview'
              ? 'bg-white text-[#dc2626] shadow-sm'
              : 'text-slate-500 hover:text-slate-800'
          "
          @click="setCropViewMode('preview')"
        >
          <span class="pi pi-eye text-[11px]" />
          Preview
        </button>
      </div>
      <div
        class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700"
      >
        <span v-if="isDetectingCrop" class="font-semibold">
          Auto detecting document area...
        </span>
        <template v-else-if="cropViewMode === 'preview'">
          Preview mode menampilkan hasil final crop + filter yang akan
          di-upload. Kembali ke Edit untuk geser/resize area dokumen.
        </template>
        <template v-else>
          Area crop sudah diarahkan otomatis jika dokumen terdeteksi. Kamu tetap
          bisa geser/resize manual sebelum upload.
        </template>
      </div>

      <div
        class="max-h-[68dvh] overflow-auto rounded-2xl bg-slate-950 p-3 text-center"
      >
        <svg
          class="absolute w-0 h-0 pointer-events-none opacity-0"
          aria-hidden="true"
        >
          <filter id="doc-scan-svg-filter">
            <feColorMatrix
              type="matrix"
              values="0.299 0.587 0.114 0 0
                      0.299 0.587 0.114 0 0
                      0.299 0.587 0.114 0 0
                      0     0     0     1 0"
            />
            <feComponentTransfer>
              <feFuncR type="linear" slope="4.0" intercept="-1.35" />
              <feFuncG type="linear" slope="4.0" intercept="-1.35" />
              <feFuncB type="linear" slope="4.0" intercept="-1.35" />
            </feComponentTransfer>
          </filter>
        </svg>

        <div
          v-if="displayedCropImageUrl"
          ref="cropFrameRef"
          class="relative inline-block max-w-full select-none touch-none"
        >
          <img
            :src="displayedCropImageUrl"
            :alt="
              cropViewMode === 'preview' ? 'Crop result preview' : 'Crop source'
            "
            class="block max-h-[62dvh] max-w-full rounded-xl object-contain"
            :style="{
              filter: cropViewMode === 'preview' ? 'none' : cropFilterCss,
            }"
            draggable="false"
          />

          <div
            v-if="isGeneratingCropPreview"
            class="absolute inset-0 z-20 flex flex-col items-center justify-center gap-2 rounded-xl bg-slate-950/70 text-xs font-semibold text-white backdrop-blur-sm"
          >
            <span class="pi pi-spinner pi-spin text-lg" />
            Generating preview...
          </div>

          <div
            v-if="cropViewMode === 'edit'"
            class="pointer-events-none absolute inset-0 bg-black/35"
          />
          <svg
            v-if="cropViewMode === 'edit'"
            class="absolute inset-0 h-full w-full overflow-visible"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
          >
            <template v-if="showCropDebug && cropDetectionDebug">
              <polygon
                :points="rawDebugPolygonPoints"
                class="pointer-events-none fill-transparent stroke-amber-300/90"
                vector-effect="non-scaling-stroke"
                stroke-dasharray="4 3"
                stroke-width="2"
              />
              <polygon
                :points="refinedDebugPolygonPoints"
                class="pointer-events-none fill-transparent stroke-emerald-300/90"
                vector-effect="non-scaling-stroke"
                stroke-width="2"
              />
            </template>
            <polygon
              :points="cropPolygonPoints"
              class="cursor-move fill-white/10 stroke-white drop-shadow-[0_0_8px_rgba(0,0,0,0.7)]"
              vector-effect="non-scaling-stroke"
              stroke-width="2"
              @pointerdown="startCropPointer('move', $event)"
            />
            <line
              :x1="cropCorners.topLeft.x"
              :y1="cropCorners.topLeft.y"
              :x2="cropCorners.bottomRight.x"
              :y2="cropCorners.bottomRight.y"
              class="pointer-events-none stroke-white/40"
              vector-effect="non-scaling-stroke"
              stroke-width="1"
            />
            <line
              :x1="cropCorners.topRight.x"
              :y1="cropCorners.topRight.y"
              :x2="cropCorners.bottomLeft.x"
              :y2="cropCorners.bottomLeft.y"
              class="pointer-events-none stroke-white/40"
              vector-effect="non-scaling-stroke"
              stroke-width="1"
            />
          </svg>
          <div
            v-if="
              showCropDebug && cropDetectionDebug && cropViewMode === 'edit'
            "
            class="pointer-events-none absolute left-2 top-2 z-10 rounded-lg bg-slate-950/80 px-2 py-1 text-left text-[10px] font-semibold text-white shadow-lg"
          >
            <div>
              Edge confidence
              {{ Math.round(cropDetectionDebug.confidence * 100) }}%
            </div>
            <div class="text-white/80">
              Area {{ Math.round(cropDetectionDebug.rawAreaRatio * 100) }}% ·
              Paper
              {{ Math.round(cropDetectionDebug.paperSurfaceRatio * 100) }}%
            </div>
            <div class="text-amber-200">Amber: raw contour</div>
            <div class="text-emerald-200">Green: refined edge</div>
            <div class="mt-1 grid grid-cols-2 gap-x-2 text-white/80">
              <span
                >T
                {{
                  Math.round(cropDetectionDebug.sideConfidence.top * 100)
                }}%</span
              >
              <span
                >R
                {{
                  Math.round(cropDetectionDebug.sideConfidence.right * 100)
                }}%</span
              >
              <span
                >B
                {{
                  Math.round(cropDetectionDebug.sideConfidence.bottom * 100)
                }}%</span
              >
              <span
                >L
                {{
                  Math.round(cropDetectionDebug.sideConfidence.left * 100)
                }}%</span
              >
            </div>
          </div>
          <template v-if="cropViewMode === 'edit'">
            <button
              v-for="(point, key) in cropCorners"
              :key="key"
              type="button"
              class="absolute h-5 w-5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-white bg-[#dc2626] shadow-[0_0_0_4px_rgba(220,38,38,0.25)] touch-none"
              :style="{
                left: `${point.x}%`,
                top: `${point.y}%`,
              }"
              @pointerdown.stop="startCropPointer(key, $event)"
            />
          </template>
        </div>
      </div>
    </div>

    <template #footer>
      <div
        class="flex flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end"
      >
        <AppButton
          v-if="canRetakeCropPhoto"
          variant="secondary"
          icon="pi-camera"
          :disabled="isProcessingImage || isDetectingCrop"
          @click="retakePhoto"
        >
          Retake Photo
        </AppButton>
        <AppButton
          v-if="cropDetectionDebug"
          variant="secondary"
          icon="pi-eye"
          @click="showCropDebug = !showCropDebug"
        >
          {{ showCropDebug ? "Hide Debug" : "Debug Edge" }}
        </AppButton>
        <label
          class="flex items-center gap-2 text-xs font-semibold text-slate-600"
        >
          Filter
          <select
            :value="selectedCropFilter"
            class="h-9 rounded-lg border border-slate-200 bg-white px-3 text-xs font-semibold text-slate-700 outline-none focus:border-[#dc2626] focus:ring-2 focus:ring-red-100"
            @change="handleCropFilterChange"
          >
            <option
              v-for="option in cropFilterOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </option>
          </select>
        </label>

        <AppButton
          variant="secondary"
          icon="pi-undo"
          :loading="isRotatingImage"
          @click="rotateCropImage"
        >
          Rotate 90°
        </AppButton>
        <AppButton variant="secondary" @click="useOriginalImage">
          Use Original
        </AppButton>
        <AppButton
          icon="pi-check"
          :loading="isProcessingImage"
          @click="applyCrop"
        >
          Apply Crop
        </AppButton>
      </div>
    </template>
  </AppDialog>
</template>
