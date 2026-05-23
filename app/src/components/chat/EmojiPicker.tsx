import { createSignal, For, onCleanup, onMount } from "solid-js";

const EMOJI_GROUPS: { name: string; emojis: string[] }[] = [
  {
    name: "表情",
    emojis: [
      "😀", "😂", "🤣", "😊", "😍", "🥰", "😘", "😜",
      "😎", "🤩", "😇", "🙂", "😏", "😌", "😋", "😴",
      "🤔", "🤗", "😅", "😓", "😥", "😰", "😢", "😭",
    ],
  },
  {
    name: "手势",
    emojis: [
      "👍", "👎", "👏", "🙌", "🤝", "👋", "✌️", "🤞",
      "💪", "🖐️", "👌", "🤏", "✍️", "🙏", "💅", "🤘",
    ],
  },
  {
    name: "物品",
    emojis: [
      "❤️", "🔥", "⭐", "🎉", "💯", "✅", "❌", "💡",
      "📌", "🎵", "🎶", "📷", "💻", "📱", "🎮", "💰",
    ],
  },
  {
    name: "动物",
    emojis: [
      "🐱", "🐶", "🐼", "🐨", "🦊", "🐰", "🐸", "🦄",
      "🐙", "🐬", "🦋", "🐝", "🦀", "🐳", "🦁", "🐯",
    ],
  },
];

interface EmojiPickerProps {
  onSelect: (emoji: string) => void;
  onClose: () => void;
}

export function EmojiPicker(props: EmojiPickerProps) {
  const [activeTab, setActiveTab] = createSignal(0);
  let pickerRef: HTMLDivElement | undefined;

  onMount(() => {
    const handler = (e: MouseEvent) => {
      if (pickerRef && !pickerRef.contains(e.target as Node)) {
        props.onClose();
      }
    };
    setTimeout(() => document.addEventListener("click", handler), 0);
    onCleanup(() => document.removeEventListener("click", handler));
  });

  return (
    <div
      ref={pickerRef}
      class="absolute bottom-full left-0 mb-2 w-80 bg-surface border border-border rounded-xl shadow-lg z-50 overflow-hidden"
    >
      {/* Tab bar */}
      <div class="flex border-b border-border">
        <For each={EMOJI_GROUPS}>
          {(group, i) => (
            <button
              onClick={() => setActiveTab(i())}
              class={`flex-1 py-2 text-xs font-medium transition-colors ${
                activeTab() === i()
                  ? "text-primary border-b-2 border-primary"
                  : "text-text-muted hover:text-text"
              }`}
            >
              {group.emojis[0]}
            </button>
          )}
        </For>
      </div>

      {/* Emoji grid */}
      <div class="p-3 max-h-56 overflow-y-auto">
        <div class="grid grid-cols-8 gap-1">
          <For each={EMOJI_GROUPS[activeTab()].emojis}>
            {(emoji) => (
              <button
                onClick={() => props.onSelect(emoji)}
                class="w-8 h-8 flex items-center justify-center text-lg hover:bg-surface-hover rounded-lg transition-colors cursor-pointer"
              >
                {emoji}
              </button>
            )}
          </For>
        </div>
      </div>
    </div>
  );
}