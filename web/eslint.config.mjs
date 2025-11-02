import eslint from "@eslint/js";
import tseslint from "typescript-eslint";

import configPrettier from "eslint-config-prettier";

import vueParser from "vue-eslint-parser";

import pluginVue from "eslint-plugin-vue";
import pluginPrettier from "eslint-plugin-prettier";
import pluginImportPath from "eslint-plugin-no-relative-import-paths";
import pluginUnusedImports from "eslint-plugin-unused-imports";

export default [
  {
    ignores: [
      "*.config.{js,ts,mjs,cjs}",
      "*.json",
      "**/proto/",
      "*.d.ts",
      "dist/",
      "public/"
    ],
  },

  eslint.configs.recommended,
  ...tseslint.configs.strict,
  ...pluginVue.configs["flat/recommended"],
  configPrettier,

  {
    //files: ["./src/**/*.{vue,js,ts,jsx,tsx}"],

    languageOptions: {
      parser: vueParser,
      parserOptions: {
        parser: tseslint.parser,
        project: './tsconfig.json',
        extraFileExtensions: [".vue"],
      }
    },

    plugins: {
      'prettier': pluginPrettier,
      'no-relative-import-paths': pluginImportPath,
      'unused-imports': pluginUnusedImports,
    },

    rules: {
      "vue/multi-word-component-names": "off",
      "vue/singleline-html-element-content-newline": "off",
      "vue/no-mutating-props": "warn", // Warning instead of error
      "no-undef": "off", // Disable for browser globals (document, window, etc.)

      'vue/component-api-style': ['warn', ['script-setup']], // Use script setup
      'vue/component-name-in-template-casing': ['error', 'PascalCase'], // PascalCase component names
      'vue/v-for-delimiter-style': ['error', 'in'], // Use 'in' delimiter for v-for
      radix: ['error', 'always'], // Enforce radix when using parseInt()
      curly: 1, // Enforce curly braces for control statements
      '@typescript-eslint/explicit-function-return-type': [0], // Disable for now - too many violations
      '@typescript-eslint/prefer-ts-expect-error': [2], // Prefer @ts-expect-error over @ts-ignore
      '@typescript-eslint/ban-ts-comment': [0], // Allow @ts-comment
      'ordered-imports': [0], // Allow/disallow ordered imports
      'object-literal-sort-keys': [0], // Allow/disallow sorting of object literal keys
      'new-parens': 1, // Enforce parentheses with 'new'
      'no-bitwise': 1, // Disallow bitwise operators
      'no-cond-assign': 1, // Disallow assignment within conditionals
      'no-trailing-spaces': 0, // Allow/disallow trailing spaces
      'eol-last': 1, // Enforce newline at end of files
      'func-style': ['error', 'declaration', { allowArrowFunctions: true }], // Enforce function style
      'no-var': 2, // Disallow 'var' keyword
      'prettier/prettier': 'warn', // Integrate Prettier and warn about style discrepancies
      'no-void': ['error', { allowAsStatement: true }], // Disallow 'void' operator, except as a statement
      'no-relative-import-paths/no-relative-import-paths': ['warn', { allowSameFolder: true, rootDir: 'src', prefix: '@' }], // No relative imports

      "@typescript-eslint/no-explicit-any": "off", // Disallow 'any' type
      "@typescript-eslint/unified-signatures": "off", // Disable due to incompatibility with Vue SFC files
      "sort-imports": ["error", { "ignoreCase": true, "ignoreDeclarationSort": true}], // Enforce sorted import declarations within modules

      // Find and remove unused ES6 module imports.
      'no-unused-vars': 'off', // Disable ESLint's 'no-unused-vars'
      'unused-imports/no-unused-imports': 'error', // Disallow unused imports
      'unused-imports/no-unused-vars': ['error', { vars: 'all', varsIgnorePattern: '^_', args: 'after-used', argsIgnorePattern: '^_' }] // Disallow unused variables and arguments
    }
  },
  {
    // Disable vue/no-mutating-props for Field.vue - intentional mutation for performance
    files: ['src/components/field/Field.vue'],
    rules: {
      'vue/no-mutating-props': 'off'
    }
  }
];