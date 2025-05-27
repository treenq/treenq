import js from '@eslint/js'
import pluginPrettier from 'eslint-plugin-prettier'
import { globalIgnores } from 'eslint/config'
import ts from 'typescript-eslint'

import globals from 'globals'

/** @type {import('eslint').Linter.Config[]} */
export default ts.config(
  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**', 'src/env.d.ts']),
  js.configs.recommended,
  ...ts.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },
  },
  {
    plugins: { prettier: pluginPrettier },
    rules: {
      'prettier/prettier': 'error',
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
      'prettier/prettier': [
        'error',
        {
          endOfLine: 'auto',
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
