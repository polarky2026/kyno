import { describe, expect, test } from "bun:test";
import { contrastRatio } from "./contrast";
import { darkTheme, lightTheme } from "./themes";

const WCAG_AA_TEXT = 4.5;

const themes = { light: lightTheme, dark: darkTheme };
const textTokens = ["textPrimary", "textSecondary", "brandText"] as const;
const backgroundTokens = ["background", "surface"] as const;

describe("contrastRatio", () => {
  test("returns 21 for black on white", () => {
    expect(contrastRatio("#000000", "#FFFFFF")).toBeCloseTo(21, 5);
  });

  test("returns 1 for identical colors", () => {
    expect(contrastRatio("#7B2032", "#7B2032")).toBe(1);
  });

  test("rejects invalid colors", () => {
    expect(() => contrastRatio("red", "#FFFFFF")).toThrow("Invalid hex color");
  });
});

for (const [name, theme] of Object.entries(themes)) {
  describe(`${name} theme`, () => {
    for (const text of textTokens) {
      for (const background of backgroundTokens) {
        test(`${text} on ${background} meets WCAG AA`, () => {
          expect(contrastRatio(theme[text], theme[background])).toBeGreaterThanOrEqual(
            WCAG_AA_TEXT,
          );
        });
      }
    }

    test("onBrand on brand meets WCAG AA", () => {
      expect(contrastRatio(theme.onBrand, theme.brand)).toBeGreaterThanOrEqual(WCAG_AA_TEXT);
    });
  });
}
