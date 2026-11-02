import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import Loading from "./loading";

test("catalog navigation shows a loading state", () => {
  render(<Loading />);

  expect(screen.getByText(/carregando o catálogo/i)).toBeDefined();
});
