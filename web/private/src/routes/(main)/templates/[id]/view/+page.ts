import type { PageLoad } from "./$types";
import { requireAuth } from "@/utils/guards";

export const load: PageLoad = ({ url }) => {
  requireAuth(url.pathname + url.search);
};
