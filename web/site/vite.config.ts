import type { IncomingMessage } from "node:http";
import "vitest/config";
import type { UserConfig } from "vite";
import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import adapter from "@sveltejs/adapter-static";
import { sveltekit } from "@sveltejs/kit/vite";
import { createSvgIconsPlugin } from "vite-plugin-svg-icons";

type AppConfig = UserConfig & {
  test?: unknown;
};

const config: AppConfig = {
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8088/"
      },
      "/verify": {
        target: "http://localhost:8088/",
        // Keep /verify as a frontend SPA route (GET), but proxy verification API calls (POST).
        bypass(req: IncomingMessage) {
          const method = req.method || "";
          if (method === "GET") {
            return req.url; // Bypass proxy, SvelteKit will serve the SPA route
          }
          return null; // Proxy non-GET requests (e.g. POST /verify/pdf)
        }
      },
      // Proxy auth API endpoints - only POST/PUT/DELETE requests, not GET (except OAuth callbacks)
      "/auth": {
        target: "http://localhost:8088/",
        bypass(req: IncomingMessage) {
          const method = req.method || "";
          const reqPath = req.url || "";

          // Allow OAuth callbacks (GET requests to /auth/oauth/*/callback)
          if (method === "GET" && reqPath.includes("/oauth/") && reqPath.includes("/callback")) {
            return null; // Proxy this request
          }

          // Allow GET /auth/verify-email (backend endpoint)
          if (method === "GET" && reqPath.includes("/verify-email")) {
            return null; // Proxy this request
          }

          // Block GET requests to /auth/* (except OAuth callbacks and verify-email):
          // let SvelteKit handle them as SPA routes.
          if (method === "GET") {
            return req.url;
          }

          // Proxy all POST, PUT, DELETE requests
          return null;
        }
      },
      "/sign": {
        target: "http://localhost:8088/",
        // /sign is not a frontend route; keep SPA 404 for GET,
        // but proxy signing API calls (POST /sign).
        bypass(req: IncomingMessage) {
          const method = req.method || "";
          if (method === "GET") {
            return req.url;
          }
          return null;
        }
      },
      "/drive": {
        target: "http://localhost:8088/"
      },
      "/public": {
        target: "http://localhost:8088/"
      }
    }
  },
  plugins: [
    tailwindcss(),
    createSvgIconsPlugin({
      iconDirs: [path.resolve(process.cwd(), "src/lib/assets/svg")],
      symbolId: "icon-[dir]-[name]"
    }),
    sveltekit({
      compilerOptions: {
        // Force runes mode for the project, except for libraries. Can be removed in svelte 6.
        runes: ({ filename }) => (filename.split(/[/\\]/).includes("node_modules") ? undefined : true)
      },
      adapter: adapter({
        pages: "dist",
        assets: "dist",
        fallback: "index.html"
      }),
      alias: {
        "@": "src/lib"
      }
    })
  ],
  test: {
    expect: { requireAssertions: true },
    projects: [
      {
        extends: "./vite.config.ts",
        test: {
          name: "server",
          environment: "node",
          include: ["src/**/*.{test,spec}.{js,ts}"],
          exclude: ["src/**/*.svelte.{test,spec}.{js,ts}"]
        }
      },
      {
        extends: "./vite.config.ts",
        test: {
          name: "client",
          environment: "jsdom",
          include: ["src/**/*.svelte.{test,spec}.{js,ts}"]
        },
        resolve: {
          conditions: ["browser"]
        }
      }
    ]
  }
};

export default config;
