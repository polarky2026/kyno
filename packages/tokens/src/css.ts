import { darkTheme, lightTheme } from "./themes";
import { typography } from "./typography";

function toKebabCase(value: string): string {
  return value.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`);
}

function toDeclarations(tokens: Record<string, string>, indent: string): string {
  return Object.entries(tokens)
    .map(([name, value]) => `${indent}--kyno-${toKebabCase(name)}: ${value};`)
    .join("\n");
}

export function createThemeCss(): string {
  return [
    ":root {",
    "  color-scheme: light dark;",
    toDeclarations(typography, "  "),
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
