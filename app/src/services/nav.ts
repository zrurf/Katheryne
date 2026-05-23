import { createSignal } from "solid-js";

type Page = "chat" | "settings";

const [_page, setPage] = createSignal<Page>("chat");

export const appNav = {
  /** Current page signal (read-only from outside) */
  page: () => _page(),
  /** Navigate to settings (hides chat, shows settings) */
  goSettings: () => setPage("settings"),
  /** Navigate back to chat (hides settings, shows chat) */
  goChat: () => setPage("chat"),
};