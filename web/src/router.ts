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
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Settings.vue")
    },
    {
      path: "/view",
      name: "view",
      meta: { layout: "Main" },
      component: () => import("@/pages/View.vue")
    },
    {
      path: "/edit",
      name: "edit",
      meta: { layout: "Sidebar", requiresAuth: true },
      component: () => import("@/pages/Edit.vue")
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
      path: "/:pathMatch(.*)*",
      name: "404",
      meta: { layout: "Blank" },
      component: () => import("@/pages/404.vue")
    }
  ]
});

router.beforeEach((to, _from, next) => {
  NProgress.start();
  loadLayoutMiddleware(to);

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
});

router.afterEach(() => {
  NProgress.done();
});

async function loadLayoutMiddleware(route: RouteLocationNormalized): Promise<void> {
  try {
    const layoutComponent = await import(`@/layouts/${route.meta.layout}.vue`);
    route.meta.layoutComponent = layoutComponent.default;
  } catch {
    const layoutComponent = await import(`@/layouts/Blank.vue`);
    route.meta.layoutComponent = layoutComponent.default;
  }
}

export default router;
