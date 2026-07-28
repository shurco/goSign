import type { PageLoad } from "./$types";
import { redirectIfAuthenticated } from "@/utils/guards";

export const load: PageLoad = () => {
  redirectIfAuthenticated();
};
