import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import { createSvgIconsPlugin } from "vite-plugin-svg-icons";
import VueDevTools from 'vite-plugin-vue-devtools'
import VueI18nPlugin from '@intlify/unplugin-vue-i18n/vite'
import path from "path";

export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
      "@public": fileURLToPath(new URL("./public", import.meta.url)),
    },
  },
  server: {
    // Configure SPA fallback - serve index.html for all routes
    fs: {
      strict: false,
    },
    proxy: {
      "/api": {
        target: "http://localhost:8088/",
      },
      "/verify": {
        target: "http://localhost:8088/",
        // Keep /verify as a frontend SPA route (GET), but proxy verification API calls (POST).
        bypass(req, _res, _options) {
          const method = req.method || "";
          if (method === "GET") {
            return "/index.html"; // Bypass proxy, Vite will serve SPA for /verify
          }
          return null; // Proxy non-GET requests (e.g. POST /verify/pdf)
        },
      },
      // Proxy auth API endpoints - only POST/PUT/DELETE requests, not GET (except OAuth callbacks)
      "/auth": {
        target: "http://localhost:8088/",
        bypass(req, _res, _options) {
          const method = req.method || "";
          const path = req.url || "";
          
          // Allow OAuth callbacks (GET requests to /auth/oauth/*/callback)
          if (method === "GET" && path.includes("/oauth/") && path.includes("/callback")) {
            return null; // Proxy this request
          }
          
          // Allow GET /auth/verify-email (backend endpoint)
          if (method === "GET" && path.includes("/verify-email")) {
            return null; // Proxy this request
          }
          
          // Block GET requests to /auth/* (except OAuth callbacks and verify-email)
          // Return index.html path to bypass proxy and let Vite handle as SPA route
          if (method === "GET") {
            return "/index.html"; // Bypass proxy, Vite will serve index.html for SPA
          }
          
          // Proxy all POST, PUT, DELETE requests
          return null;
        },
      },
      "/sign": {
        target: "http://localhost:8088/",
        // /sign is not a frontend route; keep SPA 404 for GET,
        // but proxy signing API calls (POST /sign).
        bypass(req, _res, _options) {
          const method = req.method || "";
          if (method === "GET") {
            return "/index.html";
          }
          return null;
        },
      },
      "/drive": {
        target: "http://localhost:8088/",
      },
      "/public": {
        target: "http://localhost:8088/",
      },
    },
  },
  plugins: [
    VueDevTools(),
    vue(),
    tailwindcss(),
    VueI18nPlugin({
      include: [path.resolve(__dirname, './src/i18n/locales/**')],
    }),
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), "./src/assets/svg")],
      symbolId: "icon-[dir]-[name]",
    }),
  ],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes("node_modules")) {
            return id.toString().split("node_modules/")[1].split("/")[0].toString();
          }
        },
      },
    },
  },
});
