import type { LayoutLoad } from "./$types";
import { requireAuth } from "@/utils/guards";

export const load: LayoutLoad = ({ url }) => {
  requireAuth(url.pathname + url.search);
};
