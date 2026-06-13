/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Base URL for API calls. Defaults to "/api" when unset. */
  readonly VITE_API_BASE_URL?: string
  /** Dev-only: target Vite proxies "/api" requests to. */
  readonly VITE_API_PROXY_TARGET?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
