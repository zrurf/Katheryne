import { createSignal, createRoot } from "solid-js";

function createThemeStore() {
  const saved = localStorage.getItem("theme") || "dark";
  const isLight = saved === "light";
  document.documentElement.classList.toggle("light", isLight);

  const [theme, setTheme] = createSignal<"dark" | "light">(isLight ? "light" : "dark");

  function toggleTheme() {
    const next = theme() === "dark" ? "light" : "dark";
    setTheme(next);
    document.documentElement.classList.toggle("light", next === "light");
    localStorage.setItem("theme", next);
  }

  return { theme, toggleTheme };
}

export const themeStore = createRoot(createThemeStore);