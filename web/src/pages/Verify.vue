// @ts-nocheck
<template>
  <div class="grid h-48 grid-cols-1 place-content-center gap-4">
    <div>
      <label
        for="assetsFieldHandle"
        class="relative block h-52 w-96 cursor-pointer rounded-md border-2 border-dashed border-[#343434] bg-[#efeae6]/70 hover:bg-[#efeae6]/0"
        :class="{ 'bg-[#efeae6]0': status }"
        @dragover="dragover"
        @drop="drop"
        @change="onChange"
      >
        <div class="absolute top-0 right-0 bottom-0 left-0 flex items-center justify-center">
          <div class="flex flex-col items-center">
            <span v-if="!status" data-target="file-dropzone.icon" class="flex flex-col items-center">
              <span>
                <SvgIcon name="cloud-upload" class="h-10 w-10" />
              </span>
              <div class="mb-1 font-medium">Verify Signed PDF</div>
              <div class="text-xs"><span class="font-medium">Click to upload</span> or drag and drop</div>
            </span>
            <span v-else data-target="file-dropzone.loading" class="flex flex-col items-center">
              <SvgIcon name="upload" class="h-10 w-10 animate-spin" />
              <div class="mb-1 font-medium">Uploading...</div>
            </span>
          </div>

          <input
            id="assetsFieldHandle"
            ref="file"
            name="fields[assetsFieldHandle][]"
            class="hidden"
            type="file"
            accept="application/pdf"
          />
        </div>
      </label>
    </div>

    <div v-if="verifyInfo.success" class="w-96">
      <template v-if="verifyInfo.data.verify">
        <div role="info" class="rounded border-s-4 border-green-500 bg-green-50 p-4">
          <div>
            <div v-if="verifyInfo.data.signers.length > 1">
              <div class="space-y-1 font-medium">
                <span>Total signer: {{ verifyInfo.data.signers.length }}</span>
              </div>
              <span class="flex items-center">
                <span class="my-2 h-px flex-1 bg-green-500"></span>
              </span>
            </div>

            <div v-for="(item, index) in verifyInfo.data.signers" :key="index">
              <div class="flex items-center space-x-1">
                <SvgIcon v-if="item.valid_signature" name="check-badge" class="h-5 w-5 text-green-500" />
                <SvgIcon v-else name="check-badge" class="h-5 w-5 text-red-500" />
                <span>Signature valid</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon
                  name="check-badge"
                  class="h-5 w-5"
                  :class="item.time_stamp != null ? `text-green-500` : `text-gray-400`"
                />
                <span :class="{ 'text-gray-400': item.time_stamp === null }">Timestamp</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon v-if="item.trusted_issuer.valid" name="check-badge" class="h-5 w-5 text-green-500" />
                <SvgIcon v-else name="x-circle" class="h-5 w-5 text-red-500" />
                <span>Trusted issuer</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon v-if="!item.revoked_certificate" name="check-badge" class="h-5 w-5 text-green-500" />
                <SvgIcon v-else name="x-circle" class="h-5 w-5 text-red-500" />
                <span>Active certificate</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon name="user" class="h-5 w-5" />
                <span class="max-w-80">{{ item.name ? item.name : item.reason }}</span>
              </div>

              <div v-if="item.time_stamp != null" class="flex items-center space-x-1">
                <SvgIcon name="calendar" class="h-5 w-5" />
                <span class="max-w-80">{{ toDate(item.time_stamp.time) }} </span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon name="sign" class="h-5 w-5" />
                <span>{{
                  item.cert_subject.common_name ? item.cert_subject.common_name : item.cert_subject.organization
                }}</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon name="lock-cube" class="h-5 w-5" />
                <span>{{ item.sig_format }}</span>
              </div>

              <span v-if="verifyInfo.data.signers.length - 1 > index" class="flex items-center">
                <span class="my-2 h-px flex-1 bg-green-500"></span>
              </span>
            </div>
          </div>
        </div>
      </template>
      <template v-else>
        <div role="alert" class="rounded border-s-4 border-red-500 bg-red-50 p-4">
          <strong class="block font-medium text-red-800">PDF failed verification</strong>
          <p class="mt-2 text-sm text-red-700">
            {{ verifyInfo.data.error }}
          </p>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getCurrentInstance, ref } from "vue";
import { toDate } from "@/utils/time";

const initialVerifyInfo = {
  success: false,
  data: {
    verify: false,
    error: null,
    signers: []
  }
};

const verifyInfo = ref(initialVerifyInfo);
const status = ref(false);
const instance: any = getCurrentInstance();

const onChange = async () => {
  const files = instance?.refs.file.files;
  if (!files || !files.length) {
    console.error("No file selected");
    return;
  }

  const formData = new FormData();
  for (const file of files) {
    formData.append("document", file);
  }

  try {
    status.value = true;
    const response = await fetch("/verify/pdf", {
      method: "POST",
      body: formData
    });

    verifyInfo.value = await response.json();
    if (!response.ok) {
      throw new Error("File upload failed");
    }
  } catch (error) {
    console.error("Error uploading file:", error);
  } finally {
    status.value = false;
  }
};

const dragover = (event: any) => {
  event.preventDefault();
};

const drop = (event: any) => {
  event.preventDefault();
  instance.refs.file.files = event.dataTransfer.files;
  onChange();
};
</script>
