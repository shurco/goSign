import { createRouter, createWebHistory, type RouteLocationNormalized } from "vue-router";
import * as NProgress from "nprogress";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      meta: { layout: "Main" },
      component: () => import("@/pages/Home.vue")
    },
    {
      path: "/auth/signup",
      name: "signup",
      meta: { layout: "Blank" },
      component: () => import("@/pages/SignUp.vue")
    },
    {
      path: "/auth/signin",
      name: "signin",
      meta: { layout: "Blank" },
      component: () => import("@/pages/SignIn.vue")
    },
    {
      path: "/auth/password/forgot",
      name: "password-forgot",
      meta: { layout: "Blank" },
      component: () => import("@/pages/ForgotPassword.vue")
    },
    {
      path: "/auth/password/reset",
      name: "password-reset",
      meta: { layout: "Blank" },
      component: () => import("@/pages/ResetPassword.vue")
    },
    {
      path: "/auth/verify-email",
      name: "verify-email",
      meta: { layout: "Blank" },
      component: () => import("@/pages/VerifyEmail.vue")
    },
    {
      path: "/dashboard",
      name: "dashboard",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Dashboard.vue")
    },
    {
      path: "/submissions",
      name: "submissions",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Submissions.vue")
    },
    {
      path: "/settings",
      name: "settings",
      meta: { layout: "SettingsSidebar", requiresAuth: true },
      component: () => import("@/pages/Settings.vue"),
      redirect: { name: "settings-general" },
      children: [
        {
          path: "general",
          name: "settings-general",
          component: () => import("@/pages/settings/SettingsGeneral.vue")
        },
        {
          path: "email/smtp",
          name: "settings-smtp",
          component: () => import("@/pages/settings/SettingsSmtp.vue")
        },
        {
          path: "storage",
          name: "settings-storage",
          component: () => import("@/pages/settings/SettingsStorage.vue")
        },
        {
          path: "webhooks",
          name: "settings-webhooks",
          component: () => import("@/pages/settings/SettingsWebhooks.vue")
        },
        {
          path: "api-keys",
          name: "settings-api-keys",
          component: () => import("@/pages/settings/SettingsApiKeys.vue")
        },
        {
          path: "branding",
          name: "settings-branding",
          component: () => import("@/pages/settings/SettingsBranding.vue")
        },
        {
          path: "email/templates",
          name: "settings-email-templates",
          component: () => import("@/pages/settings/SettingsEmailTemplates.vue")
        }
      ]
    },
    {
      path: "/templates",
      name: "templates",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/TemplateLibrary.vue")
    },
    {
      path: "/templates/:id/edit",
      name: "template-edit",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Edit.vue")
    },
    {
      path: "/templates/:id/folder",
      name: "template-folder",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/TemplateLibrary.vue")
    },
    {
      path: "/view",
      name: "view",
      meta: { layout: "Main" },
      component: () => import("@/pages/View.vue")
    },
    {
      path: "/uploads",
      name: "uploads",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Uploads.vue")
    },
    {
      path: "/sign",
      name: "sign",
      meta: { layout: "Main" },
      component: () => import("@/pages/Sign.vue")
    },
    {
      path: "/s/:slug",
      name: "submitter-sign",
      meta: { layout: "Blank" },
      component: () => import("@/pages/SubmitterSign.vue")
    },
    {
      path: "/verify",
      name: "verify",
      meta: { layout: "Main" },
      component: () => import("@/pages/Verify.vue")
    },
    {
      path: "/organizations",
      name: "organizations",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Organizations.vue")
    },
    {
      path: "/organizations/:organization_id/members",
      name: "organization-members",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/OrganizationMembers.vue")
    },
    {
      path: "/:pathMatch(.*)*",
      name: "404",
      meta: { layout: "Blank" },
      component: () => import("@/pages/404.vue")
    }
  ]
});

let pendingNavigation: Promise<void> | null = null;

router.beforeEach(async (to, from, next) => {
  // Cancel any pending navigation if route changed
  if (pendingNavigation) {
    // If route changed while loading, wait for current load to complete
    try {
      await pendingNavigation;
    } catch {
      // Navigation was cancelled, continue with new one
    }
  }

  NProgress.start();

  // Create navigation promise
  pendingNavigation = (async () => {
    try {
      // Wait for layout to load before proceeding
      await loadLayoutMiddleware(to);

      // Check if route requires authentication
      if (to.meta.requiresAuth) {
        const token = localStorage.getItem("access_token");
        if (!token) {
          next({ name: "signin", query: { redirect: to.fullPath } });
          return;
        }
      }

      // Redirect to dashboard if already logged in and trying to access auth pages
      const authPages = ["signin", "signup"];
      if (authPages.includes(to.name as string)) {
        const token = localStorage.getItem("access_token");
        if (token) {
          next({ name: "dashboard" });
          return;
        }
      }

      next();
    } finally {
      pendingNavigation = null;
    }
  })();

  await pendingNavigation;
});

router.afterEach(() => {
  NProgress.done();
});

// Cache for loaded layout components to avoid reloading
const layoutCache = new Map<string, any>();

async function loadLayoutMiddleware(route: RouteLocationNormalized): Promise<void> {
  const layoutName = route.meta.layout as string;

  // Check cache first
  if (layoutCache.has(layoutName)) {
    route.meta.layoutComponent = layoutCache.get(layoutName);
    return;
  }

  // Don't reload layout if it's already loaded and we're navigating to the same layout
  if (route.meta.layoutComponent && route.meta.layout === layoutName) {
    layoutCache.set(layoutName, route.meta.layoutComponent);
    return;
  }

  try {
    const layoutComponent = await import(`@/layouts/${layoutName}.vue`);
    route.meta.layoutComponent = layoutComponent.default;
    layoutCache.set(layoutName, layoutComponent.default);
  } catch (error) {
    console.error(`Failed to load layout ${layoutName}:`, error);
    // Fallback to Blank layout
    try {
      const layoutComponent = await import(`@/layouts/Blank.vue`);
      route.meta.layoutComponent = layoutComponent.default;
      layoutCache.set("Blank", layoutComponent.default);
    } catch (fallbackError) {
      console.error("Failed to load Blank layout:", fallbackError);
      // Use a minimal fallback component to prevent white screen
      route.meta.layoutComponent = {
        name: "FallbackLayout",
        template: "<div><router-view /></div>"
      };
    }
  }
}

export default router;
