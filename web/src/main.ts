import { createApp } from "vue";
import App from "@/App.vue";
import { createPinia } from "pinia";
import router from "@/router";
import svgIcon from "@/components/SvgIcon.vue";
import "virtual:svg-icons-register";
import "@/assets/app.css";

const pinia = createPinia();
const app = createApp(App);
app.use(pinia);
app.use(router);
app.component("SvgIcon", svgIcon);
app.mount("#app");
