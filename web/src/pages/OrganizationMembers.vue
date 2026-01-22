<template>
  <div class="organization-members-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ organization?.name || $t('organizations.title') }} - {{ $t('organizationMembers.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('organizationMembers.description') }}</p>
      </div>
      <div class="flex items-center gap-3">
        <Button variant="primary" @click="showInviteModal = true">
          <SvgIcon name="user-plus" class="mr-2 h-5 w-5" />
          {{ $t('organizationMembers.inviteMember') }}
        </Button>
      </div>
    </div>

    <!-- Members Table -->
    <ResourceTable
      :data="members"
      :columns="memberColumns"
      :is-loading="loading"
      :searchable="true"
      :search-keys="['user_id', 'role']"
      :search-placeholder="$t('organizationMembers.searchMembers')"
      :empty-message="$t('organizationMembers.noMembers')"
      :has-actions="true"
      :show-edit="false"
      :show-delete="false"
    >
      <template #cell-user_id="{ item }">
        <div class="flex items-center gap-2">
          <div class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-300">
            <SvgIcon name="user" class="h-4 w-4 text-gray-600" />
          </div>
          <div class="flex flex-col">
            <span class="font-medium text-gray-900">
              {{ (item as OrganizationMember).email || (item as OrganizationMember).user_name || (item as OrganizationMember).user_id }}
            </span>
            <span v-if="(item as OrganizationMember).first_name || (item as OrganizationMember).last_name" class="text-xs text-gray-500">
              {{ (item as OrganizationMember).first_name }} {{ (item as OrganizationMember).last_name }}
            </span>
          </div>
        </div>
      </template>

      <template #cell-role="{ item }">
        <Badge
          :variant="getRoleBadgeVariant((item as OrganizationMember).role)"
        >
          {{ getRoleLabel((item as OrganizationMember).role) }}
        </Badge>
      </template>

      <template #cell-joined_at="{ value }">
        <span class="text-sm text-gray-500">{{ formatDate(value) }}</span>
      </template>

      <template #actions="{ item }">
        <div class="flex items-center justify-end gap-2">
          <!-- Role selector -->
          <div v-if="canChangeRole(item as OrganizationMember)" class="w-32">
            <Select
              :model-value="(item as OrganizationMember).role"
              @update:model-value="(val: string | number) => changeMemberRole(item as OrganizationMember, String(val))"
            >
              <option value="viewer">{{ $t('organizationMembers.viewer') }}</option>
              <option value="member">{{ $t('organizationMembers.member') }}</option>
              <option v-if="currentUserRole === 'owner'" value="admin">{{ $t('organizationMembers.admin') }}</option>
            </Select>
          </div>

          <!-- Remove member button -->
          <button
            v-if="canRemoveMember(item as OrganizationMember)"
            class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
            :title="$t('organizationMembers.removeMember')"
            @click.stop="removeMember(item as OrganizationMember)"
          >
            <SvgIcon name="trash-x" class="h-5 w-5 stroke-[2]" />
          </button>
        </div>
      </template>
    </ResourceTable>

    <!-- Invitations Section -->
    <div class="mt-8">
      <h2 class="mb-4 text-lg font-semibold text-gray-900">{{ $t('organizationMembers.pendingInvitations') }}</h2>
      <ResourceTable
        :data="invitations"
        :columns="invitationColumns"
        :is-loading="false"
        :searchable="false"
        :empty-message="$t('organizationMembers.noInvitations')"
        :has-actions="true"
        :show-edit="false"
        :show-delete="false"
      >
        <template #cell-email="{ item }">
          <div class="flex items-center gap-2">
            <div class="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-100">
              <SvgIcon name="user" class="h-4 w-4 text-yellow-600" />
            </div>
            <span class="font-medium text-gray-900">{{ (item as OrganizationInvitation).email }}</span>
          </div>
        </template>

        <template #cell-role="{ item }">
          <Badge variant="ghost">
            {{ getRoleLabel((item as OrganizationInvitation).role) }}
          </Badge>
        </template>

        <template #cell-expires_at="{ value }">
          <span class="text-sm text-gray-500">{{ $t('organizationMembers.expires') }} {{ formatDate(value) }}</span>
        </template>

        <template #actions="{ item }">
          <div class="flex items-center justify-end gap-2">
            <button
              v-if="canRevokeInvitation"
              class="cursor-pointer rounded-full p-1.5 text-gray-400 transition-colors hover:text-red-600"
              :title="$t('organizationMembers.revokeInvitation')"
              @click.stop="revokeInvitation(item as OrganizationInvitation)"
            >
              <SvgIcon name="x" class="h-5 w-5 stroke-[2]" />
            </button>
          </div>
        </template>
      </ResourceTable>
    </div>

    <!-- Invite Member Modal -->
    <FormModal
      v-model="showInviteModal"
      :title="$t('organizationMembers.inviteMember')"
      :submit-text="$t('organizationMembers.sendInvitation')"
      @submit="handleInviteMember"
      @cancel="showInviteModal = false"
      @close="showInviteModal = false"
    >
      <template #default="{ formData, errors }">
        <div class="space-y-4">
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('organizationMembers.email') }} *</label>
            <Input
              v-model="formData.email"
              type="email"
              :placeholder="$t('organizationMembers.emailPlaceholder')"
              :error="!!errors.email"
              required
            />
            <span v-if="errors.email" class="mt-1 text-sm text-red-600">{{ errors.email }}</span>
          </div>

          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700">{{ $t('organizationMembers.role') }} *</label>
            <Select
              v-model="formData.role"
              :error="!!errors.role"
            >
              <option value="viewer">{{ $t('organizationMembers.viewer') }} - {{ $t('organizationMembers.viewerDescription') }}</option>
              <option value="member">{{ $t('organizationMembers.member') }} - {{ $t('organizationMembers.memberDescription') }}</option>
              <option v-if="currentUserRole === 'owner'" value="admin">{{ $t('organizationMembers.admin') }} - {{ $t('organizationMembers.adminDescription') }}</option>
            </Select>
            <span v-if="errors.role" class="mt-1 text-sm text-red-600">{{ errors.role }}</span>
          </div>
        </div>
      </template>
    </FormModal>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { useI18n } from "vue-i18n";
import { apiDelete, apiGet, apiPost, apiPut } from "@/services/api";
import { Organization, OrganizationInvitation, OrganizationMember } from "@/models";
import ResourceTable from "@/components/common/ResourceTable.vue";
import FormModal from "@/components/common/FormModal.vue";
import Button from "@/components/ui/Button.vue";
import Input from "@/components/ui/Input.vue";
import Select from "@/components/ui/Select.vue";
import Badge from "@/components/ui/Badge.vue";
import SvgIcon from "@/components/SvgIcon.vue";

const route = useRoute();
const { t } = useI18n();
const organization = ref<Organization | null>(null);
const members = ref<OrganizationMember[]>([]);
const invitations = ref<OrganizationInvitation[]>([]);
const loading = ref(true);
const showInviteModal = ref(false);
const currentUserRole = ref("");
const currentUserId = ref("");

const orgId = computed(() => route.params.organization_id as string);

const memberColumns = computed(() => [
  { key: "user_id", label: t('organizationMembers.member'), sortable: true },
  { key: "role", label: t('organizationMembers.role'), sortable: true },
  {
    key: "joined_at",
    label: t('organizationMembers.joined'),
    sortable: true,
    formatter: (value: unknown): string => value ? formatDate(value as string) : ""
  }
]);

const invitationColumns = computed(() => [
  { key: "email", label: t('organizationMembers.email'), sortable: true },
  { key: "role", label: t('organizationMembers.role'), sortable: true },
  {
    key: "expires_at",
    label: t('organizationMembers.expiresAt'),
    sortable: true,
    formatter: (value: unknown): string => value ? formatDate(value as string) : ""
  }
]);

const loadMembers = async () => {
  try {
    const response = await apiGet(`/api/v1/organizations/${orgId.value}/members`);
    members.value = response.data.members || [];
    
    // Get current user's role and ID from members list
    // Find current user by matching with stored user data or token
    const storedOrg = localStorage.getItem("current_organization");
    if (storedOrg) {
      try {
        const orgData = JSON.parse(storedOrg);
        currentUserRole.value = orgData.role || "";
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
        const currentMember = members.value.find(m => m.user_id === accountId);
        
        if (currentMember) {
          currentUserId.value = currentMember.user_id;
          if (!currentUserRole.value) {
            currentUserRole.value = currentMember.role;
          }
          console.log("Set currentUserId:", currentUserId.value, "currentUserRole:", currentUserRole.value);
        } else {
          console.warn("Current user not found in members list. Account ID:", accountId, "Members:", members.value);
          // Fallback: if role is set from localStorage, use it
          if (!currentUserRole.value && storedOrg) {
            try {
              const orgData = JSON.parse(storedOrg);
              currentUserRole.value = orgData.role || "owner"; // Default to owner if not found
            } catch (e) {
              console.error("Failed to parse stored organization:", e);
            }
          }
        }
      }
    } catch (error) {
      console.error("Failed to load current user:", error);
      // Fallback: use role from localStorage
      if (!currentUserRole.value && storedOrg) {
        try {
          const orgData = JSON.parse(storedOrg);
          currentUserRole.value = orgData.role || "owner";
          console.log("Using role from localStorage (fallback):", currentUserRole.value);
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
    const response = await apiGet(`/api/v1/organizations/${orgId.value}/invitations`);
    invitations.value = response.data.invitations || [];
  } catch (error) {
    console.error("Failed to load invitations:", error);
  }
};

const loadOrganization = async () => {
  try {
    const response = await apiGet(`/api/v1/organizations/${orgId.value}`);
    organization.value = response.data.organization;
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
  if (member.user_id === currentUserId.value) {
    return false;
  }

  // Only owners and admins can change roles
  if (currentUserRole.value === "owner") {
    return true;
  }

  // Admins can change roles but not to admin
  return currentUserRole.value === "admin" && member.role !== "admin";
};

const canRemoveMember = (member: OrganizationMember) => {
  // Can't remove owner
  if (member.role === "owner") {
    return false;
  }

  // Can't remove yourself
  if (member.user_id === currentUserId.value) {
    return false;
  }

  // Owners can remove anyone except other owners
  if (currentUserRole.value === "owner") {
    return true;
  }

  // Admins can remove members and viewers
  return currentUserRole.value === "admin" && (member.role === "member" || member.role === "viewer");
};

const canRevokeInvitation = computed(() => {
  return currentUserRole.value === "owner" || currentUserRole.value === "admin";
});

const changeMemberRole = async (member: OrganizationMember, newRole: string) => {
  try {
    await apiPut(`/api/v1/organizations/${orgId.value}/members/${member.id}/role`, {
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
  if (!confirm(t('organizationMembers.removeMemberConfirm'))) {
    return;
  }

  try {
    await apiDelete(`/api/v1/organizations/${orgId.value}/members/${member.id}`);
    members.value = members.value.filter((m) => m.id !== member.id);
  } catch (error) {
    console.error("Failed to remove member:", error);
    alert(t('organizationMembers.removeMemberError'));
  }
};

const revokeInvitation = async (invitation: OrganizationInvitation) => {
  if (!confirm(t('organizationMembers.revokeInvitationConfirm', { email: invitation.email }))) {
    return;
  }

  try {
    await apiDelete(`/api/v1/organizations/${orgId.value}/invitations/${invitation.id}`);
    invitations.value = invitations.value.filter((i) => i.id !== invitation.id);
  } catch (error) {
    console.error("Failed to revoke invitation:", error);
    alert(t('organizationMembers.revokeInvitationError'));
  }
};

const handleInviteMember = async (formData: Record<string, unknown>) => {
  if (!organization.value) {
    console.error("Organization is null");
    alert(t('organizationMembers.inviteError'));
    return;
  }

  const email = String(formData.email || "").trim();
  const role = String(formData.role || "member");

  if (!email) {
    alert(t('organizationMembers.emailRequired'));
    return;
  }

  try {
    const response = await apiPost(`/api/v1/organizations/${orgId.value}/members/invite`, {
      email,
      role
    });


    // API returns { data: {...}, message: "..." } on success
    // If we get here without error, the invitation was created
    showInviteModal.value = false;
    await loadInvitations();
    // Show success message
    alert(t('organizationMembers.invitationSent'));
  } catch (error: any) {
    console.error("Failed to invite member:", error);
    let errorMessage = t('organizationMembers.inviteError');
    
    // Handle ApiError from apiPost
    // ApiError has structure: { status: number, message: string }
    // API response structure: { success: boolean, message: string, data: any }
    if (error?.message) {
      errorMessage = error.message;
    } else if (error instanceof Error) {
      errorMessage = error.message;
    } else if (typeof error === 'string') {
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
      return t('organizationMembers.owner');
    case "admin":
      return t('organizationMembers.admin');
    case "member":
      return t('organizationMembers.member');
    case "viewer":
      return t('organizationMembers.viewer');
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
watch(showInviteModal, (isOpen) => {
  if (isOpen) {
    // FormModal will initialize formData automatically
    // Default values will be set through v-model bindings
  }
});

onMounted(async () => {
  await Promise.all([loadOrganization(), loadMembers(), loadInvitations()]);
  loading.value = false;
});
</script>

<style scoped>
.organization-members-page {
  @apply min-h-full;
}
</style>
