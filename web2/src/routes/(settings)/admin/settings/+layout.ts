import type { LayoutLoad } from "./$types";
import { requireAdmin } from "@/utils/guards";

export const load: LayoutLoad = async ({ url }) => {
  await requireAdmin(url.pathname + url.search);
};
