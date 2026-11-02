import { fireEvent, render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { ThemeToggle } from "./theme-toggle";
import { themeBootScript, themeStorageKey } from "./theme";

test("boot script applies a stored theme before interaction", () => {
  localStorage.setItem(themeStorageKey, "dark");
  const run = new Function(themeBootScript);
  run();

  expect(document.documentElement.classList.contains("dark")).toBe(true);
  expect(document.documentElement.dataset.theme).toBe("dark");
});

test("theme control keeps a stable name and persists the next choice", () => {
  document.documentElement.classList.remove("dark");
  render(<ThemeToggle />);
  const button = screen.getByRole("button", {
    name: "Alternar tema claro e escuro",
  });

  expect(button.textContent).toBe("Tema");
  fireEvent.click(button);

  expect(document.documentElement.classList.contains("dark")).toBe(true);
  expect(localStorage.getItem(themeStorageKey)).toBe("dark");
  expect(button.textContent).toBe("Tema");
});
