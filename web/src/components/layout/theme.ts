export const themeStorageKey = "theme";

export const themeBootScript = `(function(){try{var stored=localStorage.getItem(${JSON.stringify(themeStorageKey)});var dark=stored!=="light";document.documentElement.classList.toggle("dark",dark);document.documentElement.dataset.theme=dark?"dark":"light";}catch(e){}})();`;

export function toggleDocumentTheme(
  root: HTMLElement,
  storage: Pick<Storage, "setItem">,
): "light" | "dark" {
  const next = root.classList.contains("dark") ? "light" : "dark";
  root.classList.toggle("dark", next === "dark");
  root.dataset.theme = next;
  storage.setItem(themeStorageKey, next);
  return next;
}
