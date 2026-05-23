import { save } from "@tauri-apps/plugin-dialog";
import { writeFile } from "@tauri-apps/plugin-fs";
import { getServerApiBase } from "./config";
import { getAccessToken } from "./api";

/**
 * Resolve a relative OSS URL to a full URL using the current server host.
 */
function resolveOssUrl(url: string): string {
  if (url.startsWith("http://") || url.startsWith("https://")) {
    return url;
  }
  const base = getServerApiBase();
  return base + url;
}

/**
 * Download a file from a URL with proper filename and progress tracking.
 * On Tauri: opens save dialog, downloads with progress, saves to disk.
 * On Web: opens the URL in a new tab for browser download.
 */
export async function downloadFile(
  url: string,
  options?: {
    filename?: string;
    onProgress?: (pct: number) => void;
  }
): Promise<void> {
  const filename = options?.filename || "";
  const onProgress = options?.onProgress;
  const fullUrl = resolveOssUrl(url);
  const token = getAccessToken();

  // Check if running in Tauri
  const isTauri = typeof window !== "undefined" && "__TAURI_INTERNALS__" in window;

  if (isTauri) {
    try {
      // Open save dialog with suggested filename
      const savedPath = filename
        ? await save({
            defaultPath: filename,
            title: "保存文件",
          })
        : await save({ title: "保存文件" });

      if (!savedPath) return; // User cancelled

      // Fetch file content with auth
      const headers: Record<string, string> = {};
      if (token) {
        headers["Authorization"] = `Bearer ${token}`;
      }

      const response = await fetch(fullUrl, { headers });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      // Read with progress tracking
      const contentLength = response.headers.get("content-length");
      const total = contentLength ? parseInt(contentLength, 10) : 0;
      const reader = response.body?.getReader();

      if (!reader) {
        throw new Error("Response body is not readable");
      }

      const chunks: Uint8Array[] = [];
      let received = 0;

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        chunks.push(value);
        received += value.length;
        if (onProgress && total > 0) {
          onProgress(Math.round((received / total) * 100));
        }
      }

      // Combine chunks and save via Tauri filesystem plugin
      const totalLength = chunks.reduce((acc, chunk) => acc + chunk.length, 0);
      const data = new Uint8Array(totalLength);
      let offset = 0;
      for (const chunk of chunks) {
        data.set(chunk, offset);
        offset += chunk.length;
      }

      await writeFile(savedPath, data);

      if (onProgress) onProgress(100);
    } catch (err) {
      // Don't throw for user cancel
      const msg = (err as Error).message?.toLowerCase() || "";
      if (!msg.includes("cancelled") && !msg.includes("canceled")) {
        throw err;
      }
    }
  } else {
    // Web fallback: open in new tab for browser download
    const a = document.createElement("a");
    a.href = fullUrl;
    if (filename) a.download = filename;
    a.target = "_blank";
    a.rel = "noopener noreferrer";
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    if (onProgress) onProgress(100);
  }
}