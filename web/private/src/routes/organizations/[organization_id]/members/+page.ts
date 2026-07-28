import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = ({ params }) => {
  redirect(307, `/admin/organizations/${params.organization_id}/members`);
};
