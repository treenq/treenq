/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly APP_API_HOST: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
