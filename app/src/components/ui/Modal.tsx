import { cn } from "../../lib/utils";
import { createSignal, JSX, splitProps, type ComponentProps } from "solid-js";
import { Portal, Show } from "solid-js/web";

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: JSX.Element;
  class?: string;
}

export function Modal(props: ModalProps) {
  return (
    <Show when={props.open}>
      <Portal>
        <div class="fixed inset-0 z-50 flex items-center justify-center">
          <div
            class="absolute inset-0 bg-black/60 backdrop-blur-sm"
            onClick={props.onClose}
          />
          <div
            class={cn(
              "relative bg-bg-secondary border border-border rounded-2xl shadow-2xl w-full max-w-md mx-4 p-6 animate-in",
              props.class
            )}
          >
            {props.title && (
              <div class="flex items-center justify-between mb-4">
                <h2 class="text-lg font-semibold text-text">{props.title}</h2>
                <button
                  onClick={props.onClose}
                  class="text-text-muted hover:text-text transition-colors p-1"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 6L6 18M6 6l12 12" />
                  </svg>
                </button>
              </div>
            )}
            {props.children}
          </div>
        </div>
      </Portal>
    </Show>
  );
}