import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import tseslint from 'typescript-eslint'

export default tseslint.config(
  { ignores: ['dist', 'wailsjs', '**/*.d.ts'] },
  {
    extends: [js.configs.recommended, ...tseslint.configs.recommended],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      // Wails desktop apps rely on data-fetching in useEffect — this rule is designed for RSC/SSR
      'react-hooks/set-state-in-effect': 'off',
      // TanStack Virtual's useVirtualizer is incompatible with React Compiler — this is expected
      'react-hooks/incompatible-library': 'off',
      // Date.now() for session IDs and freshness is acceptable in desktop GUI patterns
      'react-hooks/purity': 'off',
      // Wails bridge (window.go) requires dynamic types
      '@typescript-eslint/no-explicit-any': 'off',
    },
  },
)
