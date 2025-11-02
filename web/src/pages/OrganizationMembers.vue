<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Header -->
    <div class="border-b border-gray-200 bg-white">
      <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between py-6">
          <div>
            <h1 class="text-2xl font-bold text-gray-900">{{ organization?.name || "Organization" }} Members</h1>
            <p class="mt-1 text-sm text-gray-600">Manage team members and their permissions</p>
          </div>
          <div class="flex items-center space-x-3">
            <button
              class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
              @click="showInviteModal = true"
            >
              <UserPlusIcon class="mr-2 h-5 w-5" />
              Invite Member
            </button>
            <button
              class="inline-flex items-center rounded-md border border-gray-200 bg-white px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-50"
              @click="$router.back()"
            >
              Back
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Content -->
    <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <!-- Members List -->
      <div class="overflow-hidden rounded-md border border-gray-200 bg-white transition-colors hover:border-gray-300">
        <ul role="list" class="divide-y divide-gray-200">
          <li v-for="member in members" :key="member.id" class="px-6 py-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center">
                <div class="h-10 w-10 flex-shrink-0">
                  <div class="flex h-10 w-10 items-center justify-center rounded-full bg-gray-300">
                    <UserIcon class="h-6 w-6 text-gray-600" />
                  </div>
                </div>
                <div class="ml-4">
                  <div class="flex items-center">
                    <h3 class="text-sm font-medium text-gray-900">{{ member.user_id }}</h3>
                    <span
                      v-if="member.role === 'owner'"
                      class="ml-2 inline-flex items-center rounded-full bg-green-100 px-2.5 py-0.5 text-xs font-medium text-green-800"
                    >
                      Owner
                    </span>
                    <span
                      v-else-if="member.role === 'admin'"
                      class="ml-2 inline-flex items-center rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800"
                    >
                      Admin
                    </span>
                    <span
                      v-else-if="member.role === 'member'"
                      class="ml-2 inline-flex items-center rounded-full bg-gray-100 px-2.5 py-0.5 text-xs font-medium text-gray-800"
                    >
                      Member
                    </span>
                    <span
                      v-else
                      class="ml-2 inline-flex items-center rounded-full bg-yellow-100 px-2.5 py-0.5 text-xs font-medium text-yellow-800"
                    >
                      {{ member.role }}
                    </span>
                  </div>
                  <p class="text-sm text-gray-500">{{ member.user_id }}</p>
                </div>
              </div>

              <div class="flex items-center space-x-4">
                <div class="text-sm text-gray-500">Joined {{ formatDate(member.joined_at) }}</div>

                <!-- Role selector -->
                <div v-if="canChangeRole(member)">
                  <select
                    :value="member.role"
                    class="block w-full rounded-md border-gray-300 py-2 pr-10 pl-3 text-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
                    @change="changeMemberRole(member, $event)"
                  >
                    <option value="viewer">Viewer</option>
                    <option value="member">Member</option>
                    <option v-if="currentUserRole === 'owner'" value="admin">Admin</option>
                  </select>
                </div>

                <!-- Remove member button -->
                <button
                  v-if="canRemoveMember(member)"
                  class="p-1 text-red-600 hover:text-red-900"
                  title="Remove member"
                  @click="removeMember(member)"
                >
                  <TrashIcon class="h-5 w-5" />
                </button>
              </div>
            </div>
          </li>
        </ul>

        <!-- Empty state -->
        <div v-if="members.length === 0 && !loading" class="py-12 text-center">
          <UserIcon class="mx-auto h-12 w-12 text-gray-400" />
          <h3 class="mt-2 text-sm font-medium text-gray-900">No members</h3>
          <p class="mt-1 text-sm text-gray-500">Get started by inviting team members.</p>
        </div>

        <!-- Loading state -->
        <div v-if="loading" class="flex justify-center py-12">
          <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-blue-600"></div>
        </div>
      </div>

      <!-- Invitations Section -->
      <div class="mt-8">
        <h2 class="mb-4 text-lg font-medium text-gray-900">Pending Invitations</h2>
        <div class="overflow-hidden rounded-md border border-gray-200 bg-white transition-colors hover:border-gray-300">
          <ul role="list" class="divide-y divide-gray-200">
            <li v-for="invitation in invitations" :key="invitation.id" class="px-6 py-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center">
                  <div class="h-10 w-10 flex-shrink-0">
                    <div class="flex h-10 w-10 items-center justify-center rounded-full bg-yellow-100">
                      <EnvelopeIcon class="h-6 w-6 text-yellow-600" />
                    </div>
                  </div>
                  <div class="ml-4">
                    <h3 class="text-sm font-medium text-gray-900">{{ invitation.email }}</h3>
                    <p class="text-sm text-gray-500">
                      Invited as {{ invitation.role }} • Expires {{ formatDate(invitation.expires_at) }}
                    </p>
                  </div>
                </div>

                <button
                  v-if="canRevokeInvitation"
                  class="p-1 text-red-600 hover:text-red-900"
                  title="Revoke invitation"
                  @click="revokeInvitation(invitation)"
                >
                  <XMarkIcon class="h-5 w-5" />
                </button>
              </div>
            </li>
          </ul>

          <div v-if="invitations.length === 0" class="py-8 text-center">
            <p class="text-sm text-gray-500">No pending invitations</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Invite Member Modal -->
    <InviteMemberModal
      v-if="showInviteModal"
      :organization="organization"
      @close="showInviteModal = false"
      @invited="onMemberInvited"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { EnvelopeIcon, TrashIcon, UserIcon, UserPlusIcon, XMarkIcon } from "@heroicons/vue/24/outline";
import { apiDelete, apiGet, apiPut } from "@/services/api";
import { Organization, OrganizationInvitation, OrganizationMember } from "@/models";
import InviteMemberModal from "@/components/organization/InviteMemberModal.vue";

const route = useRoute();
const organization = ref<Organization | null>(null);
const members = ref<OrganizationMember[]>([]);
const invitations = ref<OrganizationInvitation[]>([]);
const loading = ref(true);
const showInviteModal = ref(false);
const currentUserRole = ref("");

const orgId = computed(() => route.params.organization_id as string);

const loadMembers = async () => {
  try {
    const response = await apiGet(`/api/v1/organizations/${orgId.value}/members`);
    members.value = response.data.members || [];
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
  if (member.user_id === "current-user-id") {
    return false;
  } // TODO: Get current user ID

  // Only owners can change roles to admin
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
  if (member.user_id === "current-user-id") {
    return false;
  } // TODO: Get current user ID

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

const changeMemberRole = async (member: OrganizationMember, event: Event) => {
  const target = event.target as HTMLSelectElement;
  const newRole = target.value;

  try {
    await apiPut(`/api/v1/organizations/${orgId.value}/members/${member.id}/role`, {
      role: newRole
    });

    member.role = newRole as any; // TODO: Fix type casting
  } catch (error) {
    console.error("Failed to change member role:", error);
    target.value = member.role; // Revert on error
  }
};

const removeMember = async (member: OrganizationMember) => {
  // Note: API returns user info, but we only have user_id
  if (!confirm(`Are you sure you want to remove this member from the organization?`)) {
    return;
  }

  try {
    await apiDelete(`/api/v1/organizations/${orgId.value}/members/${member.id}`);
    members.value = members.value.filter((m) => m.id !== member.id);
  } catch (error) {
    console.error("Failed to remove member:", error);
  }
};

const revokeInvitation = async (invitation: OrganizationInvitation) => {
  if (!confirm(`Are you sure you want to revoke the invitation for ${invitation.email}?`)) {
    return;
  }

  try {
    await apiDelete(`/api/v1/organizations/${orgId.value}/invitations/${invitation.id}`);
    invitations.value = invitations.value.filter((i) => i.id !== invitation.id);
  } catch (error) {
    console.error("Failed to revoke invitation:", error);
  }
};

const onMemberInvited = () => {
  loadInvitations();
  showInviteModal.value = false;
};

const formatDate = (dateString: string | undefined) => {
  if (!dateString) {
    return "";
  }
  return new Date(dateString).toLocaleDateString();
};

onMounted(async () => {
  await Promise.all([loadOrganization(), loadMembers(), loadInvitations()]);
  loading.value = false;

  // TODO: Get current user role from organization context
  currentUserRole.value = "owner"; // Replace with actual role
});
</script>
