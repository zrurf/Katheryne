import { For, Show, createSignal, onCleanup, createEffect } from "solid-js";
import { X, Pause, Play, Square, Check, AlertTriangle, ChevronUp, ChevronDown } from "lucide-solid";
import {
  downloadManager,
  type DownloadTask,
  type DownloadStatus,
} from "../../services/download";

// Shared signal: toggled by sidebar button, auto-shows when downloads are active
const [panelVisible, setPanelVisible] = createSignal(false);
export function toggleDownloadPanel() {
  setPanelVisible((v) => !v);
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

function formatSpeed(bytesPerSec: number): string {
  if (bytesPerSec <= 0) return "";
  return formatBytes(bytesPerSec) + "/s";
}

function formatETA(seconds: number): string {
  if (seconds <= 0) return "";
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
  return `${Math.floor(seconds / 3600)}时${Math.floor((seconds % 3600) / 60)}分`;
}

function statusIcon(status: DownloadStatus) {
  switch (status) {
    case "completed":
      return <Check size={14} class="text-green-400" />;
    case "error":
      return <AlertTriangle size={14} class="text-red-400" />;
    case "downloading":
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="animate-spin text-primary">
          <path d="M21 12a9 9 0 1 1-6.219-8.56" />
        </svg>
      );
    case "verifying":
      return (
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="animate-spin text-yellow-400">
          <path d="M21 12a9 9 0 1 1-6.219-8.56" />
        </svg>
      );
    case "paused":
      return <Pause size={14} class="text-yellow-400" />;
    default:
      return null;
  }
}

function statusText(status: DownloadStatus): string {
  switch (status) {
    case "downloading":
      return "下载中";
    case "paused":
      return "已暂停";
    case "completed":
      return "已完成";
    case "error":
      return "失败";
    case "verifying":
      return "校验中";
    default:
      return "";
  }
}

export function DownloadProgress() {
  const [tasks, setTasks] = createSignal<DownloadTask[]>([]);
  const [expanded, setExpanded] = createSignal(true);

  const unsub = downloadManager.onChange((t) => {
    setTasks([...t]);
  });

  onCleanup(() => unsub());

  const activeCount = () =>
    tasks().filter((t) => t.status === "downloading" || t.status === "verifying").length;
  const hasTasks = () => tasks().length > 0;

  // Auto-cleanup: remove completed/error tasks after 5 seconds
  const doneTasks = () =>
    tasks().filter((t) => t.status === "completed" || t.status === "error");

  createEffect(() => {
    const done = doneTasks();
    if (done.length > 0) {
      const timer = setTimeout(() => {
        for (const t of done) {
          downloadManager.cancel(t.id);
        }
      }, 5000);
      return () => clearTimeout(timer);
    }
  });

  // Show when: (a) sidebar toggled on, or (b) there are active tasks
  const shouldShow = () => panelVisible() || hasTasks();

  return (
    <Show when={shouldShow()}>
      <div class="fixed bottom-4 right-4 z-50 w-80 bg-bg-secondary border border-border rounded-xl shadow-2xl overflow-hidden">
      {/* Header */}
      <button
        class="w-full flex items-center justify-between px-4 py-2.5 hover:bg-surface transition-colors"
        onClick={() => setExpanded((e) => !e)}
      >
        <div class="flex items-center gap-2 text-sm font-medium text-text">
          <span>下载</span>
          <Show when={activeCount() > 0}>
            <span class="text-xs bg-primary/20 text-primary px-1.5 py-0.5 rounded-full">
              {activeCount()}
            </span>
          </Show>
        </div>
        <div class="flex items-center gap-1 text-text-muted">
          {expanded() ? <ChevronDown size={16} /> : <ChevronUp size={16} />}
        </div>
      </button>

      {/* Task list */}
        <Show when={expanded()}>
          <div class="max-h-72 overflow-y-auto border-t border-border">
            <Show when={hasTasks()} fallback={
              <div class="px-4 py-8 text-center text-xs text-text-muted">
                暂无下载任务
              </div>
            }>
              <For each={tasks()}>
                {(task) => (
              <div class="px-4 py-3 border-b border-border last:border-b-0">
                {/* Top row: name + status + controls */}
                <div class="flex items-center gap-2 mb-1.5">
                  {statusIcon(task.status)}
                  <span class="flex-1 text-xs text-text truncate" title={task.filename}>
                    {task.filename || "文件"}
                  </span>
                  <span class="text-xs text-text-muted shrink-0">
                    {statusText(task.status)}
                  </span>
                  {/* Controls */}
                  <Show when={task.status === "downloading"}>
                    <button
                      onClick={() => downloadManager.pause(task.id)}
                      class="p-0.5 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                      title="暂停"
                    >
                      <Pause size={14} />
                    </button>
                  </Show>
                  <Show when={task.status === "paused"}>
                    <button
                      onClick={() => downloadManager.resume(task.id)}
                      class="p-0.5 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                      title="继续"
                    >
                      <Play size={14} />
                    </button>
                  </Show>
                  <Show when={task.status !== "completed"}>
                    <button
                      onClick={() => downloadManager.cancel(task.id)}
                      class="p-0.5 hover:bg-surface rounded transition-colors text-text-muted hover:text-red-400"
                      title="取消"
                    >
                      <Square size={14} />
                    </button>
                  </Show>
                  <Show when={task.status === "completed"}>
                    <button
                      onClick={() => downloadManager.cancel(task.id)}
                      class="p-0.5 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                      title="清除"
                    >
                      <X size={14} />
                    </button>
                  </Show>
                  <Show when={task.status === "error"}>
                    <button
                      onClick={() => downloadManager.cancel(task.id)}
                      class="p-0.5 hover:bg-surface rounded transition-colors text-text-muted hover:text-text"
                      title="清除"
                    >
                      <X size={14} />
                    </button>
                  </Show>
                </div>

                {/* Progress bar */}
                <div class="w-full h-1.5 bg-surface rounded-full overflow-hidden mb-1">
                  <div
                    class="h-full rounded-full transition-all duration-300"
                    classList={{
                      "bg-primary": task.status === "downloading" || task.status === "verifying",
                      "bg-yellow-400": task.status === "paused",
                      "bg-green-400": task.status === "completed",
                      "bg-red-400": task.status === "error",
                    }}
                    style={{
                      width:
                        task.progress >= 0
                          ? `${Math.max(2, task.progress)}%`
                          : task.status === "downloading"
                            ? "30%"
                            : "0%",
                      animation:
                        task.progress < 0 && task.status === "downloading"
                          ? "indeterminate-progress 1.5s ease-in-out infinite"
                          : "none",
                    }}
                  />
                </div>

                {/* Bottom row: progress / speed / ETA */}
                <div class="flex items-center justify-between text-xs text-text-muted">
                  <span>
                    {task.totalBytes > 0
                      ? `${formatBytes(task.receivedBytes)} / ${formatBytes(task.totalBytes)} · ${Math.max(0, task.progress)}%`
                      : task.receivedBytes > 0
                        ? `${formatBytes(task.receivedBytes)}`
                        : task.status === "downloading"
                          ? "连接中..."
                          : ""}
                  </span>
                  <Show when={task.status === "downloading"}>
                    <span>
                      {formatSpeed(task.speed)}
                      {task.eta > 0 ? ` · 剩余 ${formatETA(task.eta)}` : ""}
                    </span>
                  </Show>
                </div>

                {/* Error message */}
                <Show when={task.status === "error" && task.error}>
                  <p class="text-xs text-red-400 mt-1 truncate">{task.error}</p>
                </Show>

                {/* Checksum */}
                <Show when={task.status === "completed" && task.checksum}>
                  <p class="text-xs text-text-muted mt-1 truncate font-mono">
                    SHA256: {(task.checksum || "").substring(0, 16)}...
                  </p>
                </Show>
              </div>
            )}
          </For>
            </Show>
        </div>
      </Show>

      {/* Indeterminate progress animation keyframes */}
      <style>{`
        @keyframes indeterminate-progress {
          0% { transform: translateX(-100%); }
          100% { transform: translateX(350%); }
        }
      `}</style>
    </div>
    </Show>
  );
}