import type { PageLoad } from "./$types";
import { requireAdmin } from "@/utils/guards";

export const load: PageLoad = async ({ url }) => {
  await requireAdmin(url.pathname + url.search);
};
