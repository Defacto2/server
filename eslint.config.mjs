// eslint.config.mjs
// ESLint flat configuration for JavaScript and JavaScript modules.
import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import js from "@eslint/js";

export default defineConfig([
  globalIgnores(["public/**", "dist/**", "**/*.min.js"]),
  js.configs.recommended,
  {
    files: ["assets/js/**/*.{js,mjs}"],
    languageOptions: {
      globals: {
        ...globals.browser,
        bootstrap: "readonly",
        htmx: "readonly",
      },
    },
  },
]);
