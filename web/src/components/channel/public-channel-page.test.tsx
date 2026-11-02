import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";
import { PublicChannelPage } from "./public-channel-page";

afterEach(() => {
  cleanup();
});

test("public channel shows name and description without email", () => {
  render(
    <PublicChannelPage
      channel={{
        name: "Canal Ada",
        handle: "ada-channel",
        description: "Notas de laboratório",
      }}
    />,
  );

  expect(screen.getByRole("heading", { name: "Canal Ada" })).toBeDefined();
  expect(screen.getByText("Notas de laboratório")).toBeDefined();
  expect(screen.queryByText(/@/)).toBeNull();
  expect(screen.queryByText(/e-mail/i)).toBeNull();
});
