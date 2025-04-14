import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import pluginOxlint from 'eslint-plugin-oxlint'
import pluginPrettier from 'eslint-plugin-prettier'
import pluginVue from 'eslint-plugin-vue'
import { globalIgnores } from 'eslint/config'

import globals from 'globals'

export default defineConfigWithVueTs(
  {
    files: ['**/*.{ts,mts,tsx,vue}'],
  },

  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**', 'env.d.ts']),

  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  ...pluginOxlint.configs['flat/recommended'],
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.node,
        myCustomGlobal: 'readonly',
      },
      parserOptions: {
        project: ['tsconfig.eslint.json', 'tsconfig.node.json', 'tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
  {
    plugins: { prettier: pluginPrettier },
    rules: {
      'prettier/prettier': 'error',
      'vue/multi-word-component-names': 'off',
    },
  },
  {
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: 'Literal[value=null]',
          message: 'Use undefined instead of null',
        },
        {
          selector: 'TSNullKeyword',
          message: 'Use undefined instead of null (TypeScript type)',
        },
      ],
      'no-unused-vars': [
        'error',
        {
          vars: 'all',
          args: 'none',
          caughtErrors: 'all',
          ignoreRestSiblings: false,
          reportUsedIgnorePattern: false,
        },
      ],
    },
  },
)
