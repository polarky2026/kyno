import { palette } from "./palette";

export type ThemeTokens = {
  background: string;
  surface: string;
  border: string;
  textPrimary: string;
  textSecondary: string;
  brand: string;
  onBrand: string;
  brandText: string;
  accent: string;
  highlight: string;
};

export const lightTheme = {
  background: palette.champagne,
  surface: palette.white,
  border: palette.sand,
  textPrimary: palette.espresso,
  textSecondary: palette.mocha,
  brand: palette.burgundy,
  onBrand: palette.champagne,
  brandText: palette.burgundy,
  accent: palette.berry,
  highlight: palette.gold,
} satisfies ThemeTokens;

export const darkTheme = {
  background: palette.espresso,
  surface: palette.cabernet,
  border: palette.plum,
  textPrimary: palette.champagne,
  textSecondary: palette.dustyMauve,
  brand: palette.burgundy,
  onBrand: palette.champagne,
  brandText: palette.roseWine,
  accent: palette.berry,
  highlight: palette.gold,
} satisfies ThemeTokens;
