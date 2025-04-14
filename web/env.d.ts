/// <reference types="vite/client" />

import 'vue-router'

declare module '*.vue' {
  import { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ImportMetaEnv {
  readonly APP_API_HOST: string
}

declare module 'vue-router' {
  interface RouteMeta {
    authRequired: boolean
  }
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
