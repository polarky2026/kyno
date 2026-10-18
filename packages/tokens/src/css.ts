import { darkTheme, lightTheme, type ThemeTokens } from "./themes";

function toKebabCase(value: string): string {
  return value.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`);
}

function toDeclarations(theme: ThemeTokens, indent: string): string {
  return Object.entries(theme)
    .map(([name, value]) => `${indent}--kyno-${toKebabCase(name)}: ${value};`)
    .join("\n");
}

export function createThemeCss(): string {
  return [
    ":root {",
    "  color-scheme: light dark;",
    toDeclarations(lightTheme, "  "),
    "}",
    "",
    "@media (prefers-color-scheme: dark) {",
    "  :root {",
    toDeclarations(darkTheme, "    "),
    "  }",
    "}",
  ].join("\n");
}
