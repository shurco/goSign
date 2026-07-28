<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { t } from "@/i18n/index.svelte";
  import { apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
  import type { Organization, OrganizationInvitation, OrganizationMember } from "@/models";
  import ResourceTable from "@/components/common/ResourceTable.svelte";
  import FormModal from "@/components/common/FormModal.svelte";
  import Button from "@/components/ui/Button.svelte";
  import Input from "@/components/ui/Input.svelte";
  import Select from "@/components/ui/Select.svelte";
  import Badge from "@/components/ui/Badge.svelte";
  import SvgIcon from "@/components/SvgIcon.svelte";

  let organization = $state<Organization | null>(null);
  let members = $state<OrganizationMember[]>([]);
  let invitations = $state<OrganizationInvitation[]>([]);
  let loading = $state(true);
  let showInviteModal = $state(false);
  let currentUserRole = $state("");
  let currentUserId = $state("");

  const orgId = $derived(page.params.organization_id as string);

  const memberColumns = $derived([
    { key: "user_id", label: t("organizationMembers.member"), sortable: true },
    { key: "role", label: t("organizationMembers.role"), sortable: true },
    {
      key: "joined_at",
      label: t("organizationMembers.joined"),
      sortable: true,
      formatter: (value: unknown): string => (value ? formatDate(value as string) : "")
    }
  ]);

  const invitationColumns = $derived([
    { key: "email", label: t("organizationMembers.email"), sortable: true },
    { key: "role", label: t("organizationMembers.role"), sortable: true },
    {
      key: "expires_at",
      label: t("organizationMembers.expiresAt"),
      sortable: true,
      formatter: (value: unknown): string => (value ? formatDate(value as string) : "")
    }
  ]);

  const canRevokeInvitation = $derived(currentUserRole === "owner" || currentUserRole === "admin");

  const loadMembers = async () => {
    try {
      const response = await apiGet(`/api/v1/organizations/${orgId}/members`);
      members = response.data.members || [];

      // Get current user's role and ID from members list
      // Find current user by matching with stored user data or token
      const storedOrg = localStorage.getItem("current_organization");
      if (storedOrg) {
        try {
          const orgData = JSON.parse(storedOrg);
          currentUserRole = orgData.role || "";
        } catch (e) {
          console.error("Failed to parse current organization:", e);
        }
      }

      // Get current user ID from API
      try {
        const userResponse = await apiGet("/api/v1/users/me");

        if (userResponse?.data) {
          // user_id in members table is actually account_id
          // Use account_id from API response if available, otherwise fallback to user id
          const accountId = userResponse.data.account_id || userResponse.data.id;

          // Find member with matching user_id (which is account_id in the database)
          const currentMember = members.find((m) => m.user_id === accountId);

          if (currentMember) {
            currentUserId = currentMember.user_id;
            if (!currentUserRole) {
              currentUserRole = currentMember.role;
            }
          } else {
            // Fallback: if role is set from localStorage, use it
            if (!currentUserRole && storedOrg) {
              try {
                const orgData = JSON.parse(storedOrg);
                currentUserRole = orgData.role || "owner"; // Default to owner if not found
              } catch (e) {
                console.error("Failed to parse stored organization:", e);
              }
            }
          }
        }
      } catch (error) {
        console.error("Failed to load current user:", error);
        // Fallback: use role from localStorage
        if (!currentUserRole && storedOrg) {
          try {
            const orgData = JSON.parse(storedOrg);
            currentUserRole = orgData.role || "owner";
          } catch (e) {
            console.error("Failed to parse stored organization:", e);
          }
        }
      }
    } catch (error) {
      console.error("Failed to load members:", error);
    }
  };

  const loadInvitations = async () => {
    try {
      const response = await apiGet(`/api/v1/organizations/${orgId}/invitations`);
      invitations = response.data.invitations || [];
    } catch (error) {
      console.error("Failed to load invitations:", error);
    }
  };

  const loadOrganization = async () => {
    try {
      const response = await apiGet(`/api/v1/organizations/${orgId}`);
      organization = response.data.organization;
    } catch (error) {
      console.error("Failed to load organization:", error);
    }
  };

  const canChangeRole = (member: OrganizationMember) => {
    // Can't change owner's role
    if (member.role === "owner") {
      return false;
    }

    // Can't change your own role
    if (member.user_id === currentUserId) {
      return false;
    }

    // Only owners and admins can change roles
    if (currentUserRole === "owner") {
      return true;
    }

    // Admins can change roles but not to admin
    return currentUserRole === "admin" && member.role !== "admin";
  };

  const canRemoveMember = (member: OrganizationMember) => {
    // Can't remove owner
    if (member.role === "owner") {
      return false;
    }

    // Can't remove yourself
    if (member.user_id === currentUserId) {
      return false;
    }

    // Owners can remove anyone except other owners
    if (currentUserRole === "owner") {
      return true;
    }

    // Admins can remove members and viewers
    return currentUserRole === "admin" && (member.role === "member" || member.role === "viewer");
  };

  const changeMemberRole = async (member: OrganizationMember, newRole: string) => {
    try {
      await apiPut(`/api/v1/organizations/${orgId}/members/${member.id}/role`, {
        role: newRole
      });

      member.role = newRole as any; // TODO: Fix type casting
    } catch (error) {
      console.error("Failed to change member role:", error);
      // Revert on error - reload members
      await loadMembers();
    }
  };

  const removeMember = async (member: OrganizationMember) => {
    if (!confirm(t("organizationMembers.removeMemberConfirm"))) {
      return;
    }

    try {
      await apiDelete(`/api/v1/organizations/${orgId}/members/${member.id}`);
      members = members.filter((m) => m.id !== member.id);
    } catch (error) {
      console.error("Failed to remove member:", error);
      alert(t("organizationMembers.removeMemberError"));
    }
  };

  const revokeInvitation = async (invitation: OrganizationInvitation) => {
    if (!confirm(t("organizationMembers.revokeInvitationConfirm", { email: invitation.email }))) {
      return;
    }

    try {
      await apiDelete(`/api/v1/organizations/${orgId}/invitations/${invitation.id}`);
      invitations = invitations.filter((i) => i.id !== invitation.id);
    } catch (error) {
      console.error("Failed to revoke invitation:", error);
      alert(t("organizationMembers.revokeInvitationError"));
    }
  };

  const handleInviteMember = async (formData: Record<string, unknown>) => {
    if (!organization) {
      console.error("Organization is null");
      alert(t("organizationMembers.inviteError"));
      return;
    }

    const email = String(formData.email || "").trim();
    const role = String(formData.role || "member");

    if (!email) {
      alert(t("organizationMembers.emailRequired"));
      return;
    }

    try {
      await apiPost(`/api/v1/organizations/${orgId}/members/invite`, {
        email,
        role
      });

      // API returns { data: {...}, message: "..." } on success
      // If we get here without error, the invitation was created
      showInviteModal = false;
      await loadInvitations();
      // Show success message
      alert(t("organizationMembers.invitationSent"));
    } catch (error: any) {
      console.error("Failed to invite member:", error);
      let errorMessage = t("organizationMembers.inviteError");

      // Handle ApiError from apiPost
      // ApiError has structure: { status: number, message: string }
      // API response structure: { success: boolean, message: string, data: any }
      if (error?.message) {
        errorMessage = error.message;
      } else if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === "string") {
        errorMessage = error;
      } else if (error?.response?.data?.message) {
        // Handle response data structure from webutil.Response
        errorMessage = error.response.data.message;
      } else if (error?.data?.message) {
        // Alternative response structure
        errorMessage = error.data.message;
      }

      console.error("Error details:", {
        error,
        message: errorMessage,
        status: error?.status,
        response: error?.response,
        data: error?.data
      });

      alert(errorMessage);
      // Don't close modal on error so user can retry
    }
  };

  const formatDate = (dateString: string | undefined) => {
    if (!dateString) {
      return "";
    }
    return new Date(dateString).toLocaleDateString();
  };

  const getRoleLabel = (role: string) => {
    switch (role) {
      case "owner":
        return t("organizationMembers.owner");
      case "admin":
        return t("organizationMembers.admin");
      case "member":
        return t("organizationMembers.member");
      case "viewer":
        return t("organizationMembers.viewer");
      default:
        return role;
    }
  };

  const getRoleBadgeVariant = (role: string): "success" | "primary" | "ghost" | "warning" => {
    switch (role) {
      case "owner":
        return "success";
      case "admin":
        return "primary";
      case "member":
        return "ghost";
      default:
        return "warning";
    }
  };

  // Initialize form data when modal opens
  $effect(() => {
    if (showInviteModal) {
      // FormModal will initialize formData automatically
      // Default values will be set through v-model bindings
    }
  });

  onMount(async () => {
    await Promise.all([loadOrganization(), loadMembers(), loadInvitations()]);
    loading = false;
  });
</script>

<div class="organization-members-page">
  <!-- Header -->
  <div class="mb-6 flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold">
        {organization?.name || t("organizations.title")} - {t("organizationMembers.title")}
      </h1>
      <p class="mt-1 text-sm text-gray-600">{t("organizationMembers.description")}</p>
    </div>
    <div class="flex items-center gap-3">
      <Button variant="primary" onclick={() => (showInviteModal = true)}>
        <SvgIcon name="user-plus" class="mr-2 h-5 w-5" />
        {t("organizationMembers.inviteMember")}
      </Button>
    </div>
  </div>

  <!-- Members Table -->
  {#snippet cellUserId(item: unknown, _value: string)}
    <div class="flex items-center gap-2">
      <div class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-300">
        <SvgIcon name="user" class="h-4 w-4 text-gray-600" />
      </div>
      <div class="flex flex-col">
        <span class="font-medium text-gray-900">
          {(item as OrganizationMember).email ||
            (item as OrganizationMember).user_name ||
            (item as OrganizationMember).user_id}
        </span>
        {#if (item as OrganizationMember).first_name || (item as OrganizationMember).last_name}
          <span class="text-xs text-gray-500">
            {(item as OrganizationMember).first_name}
            {(item as OrganizationMember).last_name}
          </span>
        {/if}
      </div>
    </div>
  {/snippet}

  {#snippet cellRole(item: unknown, _value: string)}
    <Badge variant={getRoleBadgeVariant((item as OrganizationMember).role)}>
      {getRoleLabel((item as OrganizationMember).role)}
    </Badge>
  {/snippet}

  {#snippet cellJoinedAt(_item: unknown, value: string)}
    <span class="text-sm text-gray-500">{formatDate(value)}</span>
  {/snippet}

  {#snippet memberActions(item: unknown)}
    <div class="flex items-center justify-end gap-2">
      <!-- Role selector -->
      {#if canChangeRole(item as OrganizationMember)}
        <div class="w-32">
          <Select
            value={(item as OrganizationMember).role}
            onchange={(e) => changeMemberRole(item as OrganizationMember, String(e.currentTarget.value))}
          >
            <option value="viewer">{t("organizationMembers.viewer")}</option>
            <option value="member">{t("organizationMembers.member")}</option>
            {#if currentUserRole === "owner"}
              <option value="admin">{t("organizationMembers.admin")}</option>
            {/if}
          </Select>
        </div>
      {/if}

      <!-- Remove member button -->
      {#if canRemoveMember(item as OrganizationMember)}
        <button
          class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
          title={t("organizationMembers.removeMember")}
          onclick={(e) => {
            e.stopPropagation();
            removeMember(item as OrganizationMember);
          }}
        >
          <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
        </button>
      {/if}
    </div>
  {/snippet}

  <ResourceTable
    data={members}
    columns={memberColumns}
    isLoading={loading}
    searchable={true}
    searchKeys={["user_id", "role"]}
    searchPlaceholder={t("organizationMembers.searchMembers")}
    emptyMessage={t("organizationMembers.noMembers")}
    hasActions={true}
    showEdit={false}
    showDelete={false}
    cellSnippets={{ user_id: cellUserId, role: cellRole, joined_at: cellJoinedAt }}
    actions={memberActions}
  />

  <!-- Invitations Section -->
  <div class="mt-8">
    <h2 class="mb-4 text-lg font-semibold text-gray-900">{t("organizationMembers.pendingInvitations")}</h2>

    {#snippet cellInvitationEmail(item: unknown, _value: string)}
      <div class="flex items-center gap-2">
        <div class="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-100">
          <SvgIcon name="user" class="h-4 w-4 text-yellow-600" />
        </div>
        <span class="font-medium text-gray-900">{(item as OrganizationInvitation).email}</span>
      </div>
    {/snippet}

    {#snippet cellInvitationRole(item: unknown, _value: string)}
      <Badge variant="ghost">
        {getRoleLabel((item as OrganizationInvitation).role)}
      </Badge>
    {/snippet}

    {#snippet cellExpiresAt(_item: unknown, value: string)}
      <span class="text-sm text-gray-500">{t("organizationMembers.expires")} {formatDate(value)}</span>
    {/snippet}

    {#snippet invitationActions(item: unknown)}
      <div class="flex items-center justify-end gap-2">
        {#if canRevokeInvitation}
          <button
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
            title={t("organizationMembers.revokeInvitation")}
            onclick={(e) => {
              e.stopPropagation();
              revokeInvitation(item as OrganizationInvitation);
            }}
          >
            <SvgIcon name="x" class="h-5 w-5 stroke-[2]" />
          </button>
        {/if}
      </div>
    {/snippet}

    <ResourceTable
      data={invitations}
      columns={invitationColumns}
      isLoading={false}
      searchable={false}
      emptyMessage={t("organizationMembers.noInvitations")}
      hasActions={true}
      showEdit={false}
      showDelete={false}
      cellSnippets={{
        email: cellInvitationEmail,
        role: cellInvitationRole,
        expires_at: cellExpiresAt
      }}
      actions={invitationActions}
    />
  </div>

  <!-- Invite Member Modal -->
  <FormModal
    bind:open={showInviteModal}
    title={t("organizationMembers.inviteMember")}
    submitText={t("organizationMembers.sendInvitation")}
    onSubmit={handleInviteMember}
    onCancel={() => (showInviteModal = false)}
    onClose={() => (showInviteModal = false)}
  >
    {#snippet children(formData, errors)}
      <div class="space-y-4">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("organizationMembers.email")} *</label>
          <Input
            bind:value={formData.email}
            type="email"
            placeholder={t("organizationMembers.emailPlaceholder")}
            error={!!errors.email}
            required
          />
          {#if errors.email}
            <span class="mt-1 text-sm text-red-600">{errors.email}</span>
          {/if}
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700">{t("organizationMembers.role")} *</label>
          <Select bind:value={formData.role} error={!!errors.role}>
            <option value="viewer">
              {t("organizationMembers.viewer")} - {t("organizationMembers.viewerDescription")}
            </option>
            <option value="member">
              {t("organizationMembers.member")} - {t("organizationMembers.memberDescription")}
            </option>
            {#if currentUserRole === "owner"}
              <option value="admin">
                {t("organizationMembers.admin")} - {t("organizationMembers.adminDescription")}
              </option>
            {/if}
          </Select>
          {#if errors.role}
            <span class="mt-1 text-sm text-red-600">{errors.role}</span>
          {/if}
        </div>
      </div>
    {/snippet}
  </FormModal>
</div>

<style>
  .organization-members-page {
    @apply min-h-full;
  }
</style>
