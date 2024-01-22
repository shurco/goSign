import { createRouter, createWebHistory, type RouteLocationNormalized } from "vue-router";
import * as NProgress from "nprogress";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "home",
      meta: { layout: "Main" },
      component: () => import("@/pages/Home.vue"),
    },
    {
      path: "/view",
      name: "view",
      meta: { layout: "Main" },
      component: () => import("@/pages/View.vue"),
    },
    {
      path: "/edit",
      name: "edit",
      meta: { layout: "Profile" },
      component: () => import("@/pages/Edit.vue"),
    },
    {
      path: "/uploads",
      name: "uploads",
      meta: { layout: "Main" },
      component: () => import("@/pages/Uploads.vue"),
    },
    {
      path: "/sign",
      name: "sign",
      meta: { layout: "Main" },
      component: () => import("@/pages/Sign.vue"),
    },
    {
      path: "/verify",
      name: "verify",
      meta: { layout: "Main" },
      component: () => import("@/pages/Verify.vue"),
    },
    {
      path: "/:pathMatch(.*)*",
      name: "404",
      meta: { layout: "Blank" },
      component: () => import("@/pages/404.vue"),
    },
  ],
});

router.beforeEach((to, _from, next) => {
  NProgress.start();
  loadLayoutMiddleware(to);
  next();
});

router.afterEach(() => {
  NProgress.done();
});

async function loadLayoutMiddleware(route: RouteLocationNormalized) {
  let layoutComponent;
  try {
    layoutComponent = await import(`@/layouts/${route.meta.layout}.vue`);
  } catch (e) {
    console.error("Error occurred in processing of layout: ", e);
    console.log("Mounted default layout `Blank`");
    layoutComponent = await import(`@/layouts/Blank.vue`);
  }
  route.meta.layoutComponent = layoutComponent.default;
}

export default router;
