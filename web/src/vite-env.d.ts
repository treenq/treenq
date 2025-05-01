/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly APP_API_HOST: string
  readonly APP_GITHUB_APP_NAME: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
