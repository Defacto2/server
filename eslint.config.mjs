// eslint.config.mjs
// ESLint flat configuration for JavaScript and JavaScript modules.
import globals from "globals";
import js from "@eslint/js";

export default [
  js.configs.recommended,
  {
    ignores: ["**/*.min.js"],
  },
  {
    files: ["**/*.mjs"],
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
