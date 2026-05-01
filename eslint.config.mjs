import tsParser from "@typescript-eslint/parser";
import tsPlugin from "@typescript-eslint/eslint-plugin";
import globals from "globals";

export default [
  {
    ignores: ["node_modules/**", "dist/**"],
  },
  {
    files: ["src/**/*.ts"],
    languageOptions: {
      parser: tsParser,
      ecmaVersion: "latest",
      sourceType: "module",
      globals: {
        ...globals.node,
        ...globals.browser,
      },
    },
    plugins: {
      "@typescript-eslint": tsPlugin,
    },
    rules: {
      "no-unused-vars": "error",
      "no-undef": "error",
      "linebreak-style": 0,
      "quotes": ["error", "double"],
      "semi": ["error", "always"],
      "no-empty": "error",
      "no-func-assign": "error",
      "no-case-declarations": "off",
      "no-unreachable": "error",
      "no-eval": "error",
      "no-global-assign": "error",
      "@typescript-eslint/no-explicit-any": "warn",
      "indent": ["error", 2],
    },
  },
];
