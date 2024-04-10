import globals from "globals";
import js from "@eslint/js";

export default [
  js.configs.recommended,
  {
    files: ["**/*.mjs"],
    ignores: ["**/*.min.js"],
  },
  {
    files: ["assets/js/**/*.js", "assets/js/**/*.mjs"],
    languageOptions: {
      ecmaVersion: "latest",
      globals: {
        ...globals.browser,
        bootstrap: "readonly",
        htmx: "readonly",
      },
    },
    rules: {},
  },
];
