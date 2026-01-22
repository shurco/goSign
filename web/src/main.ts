import { createApp } from "vue";
import App from "@/App.vue";
import { createPinia } from "pinia";
import router from "@/router";
import svgIcon from "@/components/SvgIcon.vue";
import { setAuthRouter } from "@/utils/auth";
import i18n from "@/i18n";
import "virtual:svg-icons-register";
import "@/assets/app.css";

const pinia = createPinia();
const app = createApp(App);
app.use(pinia);
app.use(router);
app.use(i18n);
// Set router instance for auth redirects
setAuthRouter(router);
app.component("SvgIcon", svgIcon);
app.mount("#app");
