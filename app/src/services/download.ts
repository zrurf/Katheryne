import { save } from "@tauri-apps/plugin-dialog";
import { writeFile } from "@tauri-apps/plugin-fs";
import { getServerApiBase } from "./config";
import { getAccessToken } from "./api";

// ─────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────

export type DownloadStatus =
  | "idle"
  | "downloading"
  | "paused"
  | "completed"
  | "error"
  | "verifying";

export interface DownloadTask {
  id: string;
  url: string;
  filename: string;
  savedPath: string;
  totalBytes: number;
  receivedBytes: number;
  progress: number; // 0-100, -1 means indeterminate
  status: DownloadStatus;
  speed: number; // bytes per second
  eta: number; // seconds remaining
  checksum?: string; // SHA-256 hex after verification
  error?: string;
}

export type DownloadCallback = (tasks: DownloadTask[]) => void;

// ─────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────

function resolveOssUrl(url: string): string {
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }
  const base = getServerApiBase();
  return base + url;
}

let _taskIdCounter = 0;
function nextTaskId(): string {
  _taskIdCounter++;
  return `dl_${Date.now()}_${_taskIdCounter}`;
}

const isTauri = typeof window !== "undefined" && "__TAURI_INTERNALS__" in window;

function buildAuthHeaders(): Record<string, string> {
  const headers: Record<string, string> = {};
  const token = getAccessToken();
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

/** Delete this dead function - HEAD not supported by OssProxy */
// Previously fetchFileSize used HEAD which the OssProxy doesn't handle (returns 404).
// Instead we get Content-Length from the GET response headers directly in startDownload.

// SHA-256 for downloaded data
async function computeSHA256(data: Uint8Array): Promise<string> {
  try {
    const buf = data.buffer.slice(data.byteOffset, data.byteOffset + data.byteLength) as ArrayBuffer;
    const hashBuffer = await crypto.subtle.digest("SHA-256", buf);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
  } catch {
    return "";
  }
}

// ─────────────────────────────────────────────────────────────
// Download Manager
// ─────────────────────────────────────────────────────────────

export class DownloadManager {
  private tasks: Map<string, DownloadTask> = new Map();
  private controllers: Map<string, AbortController> = new Map();
  private callbacks: Set<DownloadCallback> = new Set();

  /** Subscribe to task updates. Returns unsubscribe function. */
  onChange(cb: DownloadCallback): () => void {
    this.callbacks.add(cb);
    return () => this.callbacks.delete(cb);
  }

  private emit() {
    const list = Array.from(this.tasks.values());
    for (const cb of this.callbacks) {
      try {
        cb(list);
      } catch {
        // ignore
      }
    }
  }

  getTask(id: string): DownloadTask | undefined {
    return this.tasks.get(id);
  }

  getAllTasks(): DownloadTask[] {
    return Array.from(this.tasks.values());
  }

  pause(id: string) {
    const controller = this.controllers.get(id);
    if (controller) {
      controller.abort();
      this.controllers.delete(id);
    }
    const task = this.tasks.get(id);
    if (task && task.status === "downloading") {
      task.status = "paused";
      task.speed = 0;
      this.emit();
    }
  }

  resume(id: string) {
    const task = this.tasks.get(id);
    if (!task || task.status !== "paused") return;
    this.startDownload(task);
  }

  cancel(id: string) {
    const controller = this.controllers.get(id);
    if (controller) {
      controller.abort();
      this.controllers.delete(id);
    }
    this.tasks.delete(id);
    this.emit();
  }

  /** Start a new download. Opens save dialog on Tauri. Returns task ID. */
  async start(
    url: string,
    options?: {
      filename?: string;
      savedPath?: string;
    }
  ): Promise<string> {
    const filename = options?.filename || "";
    const fullUrl = resolveOssUrl(url);
    const id = nextTaskId();

    let savedPath = options?.savedPath || "";

    if (isTauri && !savedPath) {
      const chosen = filename
        ? await save({ defaultPath: filename, title: "保存文件" })
        : await save({ title: "保存文件" });
      if (!chosen) {
        return ""; // User cancelled
      }
      savedPath = chosen;
    }

    // Content-Length will be determined from GET response headers in startDownload()
    const task: DownloadTask = {
      id,
      url: fullUrl,
      filename,
      savedPath,
      totalBytes: 0, // will be set from GET response
      receivedBytes: 0,
      progress: -1, // indeterminate until we know the size
      status: "downloading",
      speed: 0,
      eta: 0,
    };
    this.tasks.set(id, task);
    console.log(`[DownloadManager] starting: ${filename}, url=${fullUrl}`);
    this.emit();

    this.startDownload(task);
    return id;
  }

  private async startDownload(task: DownloadTask) {
    const controller = new AbortController();
    this.controllers.set(task.id, controller);

    task.status = "downloading";
    task.error = undefined;
    this.emit();

    const isResuming = task.receivedBytes > 0;

    try {
      const headers = buildAuthHeaders();

      // Resume: request range from current received bytes
      if (isResuming) {
        headers["Range"] = `bytes=${task.receivedBytes}-`;
      }

      const response = await fetch(task.url, {
        headers,
        signal: controller.signal,
      });

      if (!response.ok && response.status !== 206) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Determine total size
      if (response.status === 206) {
        const cr = response.headers.get("Content-Range");
        if (cr) {
          const m = cr.match(/bytes \d+-\d+\/(\d+)/);
          if (m) task.totalBytes = parseInt(m[1], 10);
        }
      }

      // Fallback: get Content-Length from GET response if HEAD didn't work
      if (task.totalBytes <= 0) {
        const cl = response.headers.get("Content-Length");
        if (cl) {
          task.totalBytes = parseInt(cl, 10) || 0;
        }
      }

      const etag = response.headers.get("ETag");
      console.log(
        `[DownloadManager] GET response: status=${response.status}, total=${task.totalBytes}, etag=${etag}`
      );

      // Read response body
      if (!response.body) {
        throw new Error("Response body is null");
      }

      const reader = response.body.getReader();

      // Speed tracking (2-second sliding window)
      const speedWindowMs = 2000;
      let windowStart = Date.now();
      let windowBytes = 0;
      let totalReceived = task.receivedBytes;

      const chunks: Uint8Array[] = [];

      while (true) {
        const { done, value } = await reader.read();
        if (done) {
          console.log(`[DownloadManager] stream done, total received=${totalReceived}`);
          break;
        }

        chunks.push(value);
        const len = value.length;
        totalReceived += len;
        windowBytes += len;

        // Update speed every window
        const now = Date.now();
        const elapsed = now - windowStart;
        if (elapsed >= speedWindowMs) {
          task.speed = Math.round(windowBytes / (elapsed / 1000));
          windowStart = now;
          windowBytes = 0;
        }

        // Update progress
        task.receivedBytes = totalReceived;
        if (task.totalBytes > 0) {
          task.progress = Math.min(99, Math.round((totalReceived / task.totalBytes) * 100));
          if (task.speed > 0) {
            task.eta = Math.round((task.totalBytes - totalReceived) / task.speed);
          }
        } else {
          // Indeterminate: show as -1 (animated bar in UI)
          task.progress = -1;
        }

        this.emit();
      }

      // Combine chunks
      const totalLength = chunks.reduce((acc, c) => acc + c.length, 0);
      console.log(`[DownloadManager] combining ${chunks.length} chunks, total ${totalLength} bytes`);
      const data = new Uint8Array(totalLength);
      let offset = 0;
      for (const chunk of chunks) {
        data.set(chunk, offset);
        offset += chunk.length;
      }

      // Compute SHA-256 BEFORE writing to disk (writeFile may invalidate buffer on some platforms)
      let checksum = "";
      if (etag && totalLength < 50 * 1024 * 1024) {
        checksum = await computeSHA256(data);
        task.checksum = checksum;
        console.log(`[DownloadManager] SHA256=${checksum}`);
      } else if (etag) {
        console.log(`[DownloadManager] skipping SHA-256: file too large (${totalLength} bytes)`);
      }

      // Save to disk (Tauri)
      if (isTauri && task.savedPath) {
        console.log(`[DownloadManager] saving to ${task.savedPath}, append=${isResuming}`);
        try {
          await writeFile(task.savedPath, data, { append: isResuming });
          console.log(`[DownloadManager] saved OK`);
        } catch (writeErr) {
          throw new Error(`文件写入失败: ${(writeErr as Error).message || JSON.stringify(writeErr)}`);
        }
      }

      // Checksum verification
      if (checksum && etag) {
        const cleanEtag = etag.replace(/^"/, "").replace(/"$/, "");
        if (cleanEtag.toLowerCase() !== checksum.toLowerCase()) {
          console.warn(
            `[DownloadManager] checksum mismatch: expected=${cleanEtag}, got=${checksum}`
          );
        }
      }

      task.status = "completed";
      task.progress = 100;
      task.speed = 0;
      task.eta = 0;
      console.log(`[DownloadManager] completed: ${task.filename}`);
      this.emit();
    } catch (err) {
      const errMsg = (err as Error).message || "";
      if ((err as Error).name === "AbortError") {
        task.status = "paused";
        task.speed = 0;
        console.log(`[DownloadManager] paused: ${task.filename}, received=${task.receivedBytes}`);
      } else if (errMsg.toLowerCase().includes("cancelled") || errMsg.toLowerCase().includes("canceled")) {
        this.tasks.delete(task.id);
        this.emit();
        return;
      } else {
        task.status = "error";
        task.error = errMsg;
        console.error(`[DownloadManager] error: ${errMsg}`);
      }
      this.emit();
    } finally {
      this.controllers.delete(task.id);
    }
  }
}

/** Singleton instance */
export const downloadManager = new DownloadManager();

/** Whether the download progress panel is visible (toggled by sidebar button) */
export const downloadPanelVisible = { value: false };

// ─────────────────────────────────────────────────────────────
// Convenience wrapper (backward-compatible with existing usage)
// ─────────────────────────────────────────────────────────────

export async function downloadFile(
  url: string,
  options?: {
    filename?: string;
    onProgress?: (pct: number) => void;
  }
): Promise<string> {
  const taskId = await downloadManager.start(url, {
    filename: options?.filename,
  });

  if (!taskId) return "";

  if (options?.onProgress) {
    const unsub = downloadManager.onChange((tasks) => {
      const t = tasks.find((x) => x.id === taskId);
      if (t && t.progress >= 0) {
        options.onProgress?.(t.progress);
      }
    });
    const checkDone = () => {
      const t = downloadManager.getTask(taskId);
      if (!t || t.status === "completed" || t.status === "error") {
        unsub();
      } else {
        setTimeout(checkDone, 500);
      }
    };
    setTimeout(checkDone, 500);
  }

  return taskId;
}