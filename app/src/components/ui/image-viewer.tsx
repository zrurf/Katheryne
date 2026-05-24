import { Show, createSignal, onMount, onCleanup } from "solid-js";
import { X, ChevronLeft, ChevronRight, ZoomIn, ZoomOut } from "lucide-solid";
import { downloadFile } from "../../services/download";
import { getServerApiBase } from "../../services/config";

interface ImageViewerProps {
  images: string[];
  initialIndex?: number;
  onClose: () => void;
}

function resolveUrl(url: string): string {
  if (url.startsWith("http://") || url.startsWith("https://")) return url;
  return getServerApiBase() + url;
}

export function ImageViewer(props: ImageViewerProps) {
  const [index, setIndex] = createSignal(props.initialIndex || 0);
  const [scale, setScale] = createSignal(1);
  const [posX, setPosX] = createSignal(0);
  const [posY, setPosY] = createSignal(0);
  const [dragging, setDragging] = createSignal(false);
  let dragStartX = 0;
  let dragStartY = 0;
  let dragStartPosX = 0;
  let dragStartPosY = 0;

  const currentSrc = () => props.images[index()] || "";

  const handlePrev = (e: Event) => {
    e.stopPropagation();
    if (index() > 0) {
      setIndex((i) => i - 1);
      setScale(1);
      setPosX(0);
      setPosY(0);
    }
  };

  const handleNext = (e: Event) => {
    e.stopPropagation();
    if (index() < props.images.length - 1) {
      setIndex((i) => i + 1);
      setScale(1);
      setPosX(0);
      setPosY(0);
    }
  };

  const handleZoomIn = (e: Event) => {
    e.stopPropagation();
    setScale((s) => Math.min(s + 0.5, 5));
  };

  const handleZoomOut = (e: Event) => {
    e.stopPropagation();
    setScale((s) => Math.max(s - 0.5, 0.5));
  };

  const handleMouseDown = (e: MouseEvent) => {
    if (scale() <= 1) return;
    e.preventDefault();
    setDragging(true);
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    dragStartPosX = posX();
    dragStartPosY = posY();
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (!dragging()) return;
    setPosX(dragStartPosX + (e.clientX - dragStartX));
    setPosY(dragStartPosY + (e.clientY - dragStartY));
  };

  const handleMouseUp = () => {
    setDragging(false);
  };

  const handleWheel = (e: WheelEvent) => {
    e.preventDefault();
    if (e.deltaY < 0) setScale((s) => Math.min(s + 0.2, 5));
    else setScale((s) => Math.max(s - 0.2, 0.5));
  };

  const handleKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case "Escape":
        props.onClose();
        break;
      case "ArrowLeft":
        handlePrev(e);
        break;
      case "ArrowRight":
        handleNext(e);
        break;
    }
  };

  onMount(() => {
    document.addEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "hidden";
  });

  onCleanup(() => {
    document.removeEventListener("keydown", handleKeyDown);
    document.body.style.overflow = "";
  });

  return (
    <div
      class="fixed inset-0 z-[100] bg-black/90 flex flex-col"
      onClick={props.onClose}
    >
      {/* Top bar */}
      <div class="flex items-center justify-between px-4 py-3 shrink-0">
        <span class="text-white/70 text-sm">
          {index() + 1} / {props.images.length}
        </span>
        <div class="flex items-center gap-2">
          <button
            onClick={handleZoomOut}
            class="p-2 rounded-lg hover:bg-white/10 text-white/70 hover:text-white transition-colors"
            title="缩小"
          >
            <ZoomOut size={18} />
          </button>
          <button
            onClick={handleZoomIn}
            class="p-2 rounded-lg hover:bg-white/10 text-white/70 hover:text-white transition-colors"
            title="放大"
          >
            <ZoomIn size={18} />
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              downloadFile(currentSrc(), { filename: `image_${Date.now()}` });
            }}
            class="p-2 rounded-lg hover:bg-white/10 text-white/70 hover:text-white transition-colors inline-flex items-center gap-1 text-sm"
            title="下载"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>
          </button>
          <button
            onClick={props.onClose}
            class="p-2 rounded-lg hover:bg-white/10 text-white/70 hover:text-white transition-colors"
            title="关闭"
          >
            <X size={20} />
          </button>
        </div>
      </div>

      {/* Image area */}
      <div class="flex-1 flex items-center justify-center min-h-0 px-4" onWheel={handleWheel}>
        <Show when={props.images.length > 1}>
          <button
            onClick={handlePrev}
            disabled={index() === 0}
            class="shrink-0 p-2 rounded-full hover:bg-white/10 text-white/70 hover:text-white transition-colors disabled:opacity-30 disabled:cursor-default mr-2"
          >
            <ChevronLeft size={28} />
          </button>
        </Show>

        <div
          class="flex items-center justify-center overflow-hidden"
          style="flex: 1 1 auto; height: 100%;"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        >
          <img
            src={resolveUrl(currentSrc())}
            alt="预览"
            class="max-w-full max-h-full object-contain rounded-lg select-none"
            style={`transform: translate(${posX()}px, ${posY()}px) scale(${scale()}); transition: ${dragging() ? 'none' : 'transform 0.2s ease'}; cursor: ${scale() > 1 ? (dragging() ? 'grabbing' : 'grab') : 'zoom-in'}`}
            onClick={(e) => {
              if (dragging()) return;
              e.stopPropagation();
              setScale((s) => (s > 1 ? 1 : 2));
              setPosX(0);
              setPosY(0);
            }}
            draggable={false}
          />
        </div>

        <Show when={props.images.length > 1}>
          <button
            onClick={handleNext}
            disabled={index() === props.images.length - 1}
            class="shrink-0 p-2 rounded-full hover:bg-white/10 text-white/70 hover:text-white transition-colors disabled:opacity-30 disabled:cursor-default ml-2"
          >
            <ChevronRight size={28} />
          </button>
        </Show>
      </div>
    </div>
  );
}