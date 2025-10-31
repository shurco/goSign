/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly BASE_URL: string
  readonly MODE: string
  readonly DEV: boolean
  readonly PROD: boolean
  readonly SSR: boolean
  // Add more env variables here if needed
  // readonly VITE_API_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

