<template>
  <div class="uploads-page">
    <!-- Header -->
    <div class="mb-6 flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold">{{ $t('uploads.title') }}</h1>
        <p class="mt-1 text-sm text-gray-600">{{ $t('uploads.description') }}</p>
      </div>
    </div>

    <label
      for="assetsFieldHandle"
      class="relative block h-52 w-64 cursor-pointer rounded-xl border-2 border-dashed border-[#e7e2df] hover:bg-[#efeae6]/30"
      :class="{ 'bg-[#efeae6]/30': status }"
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
            <div class="mb-1 font-medium">{{ $t('uploads.uploadNewDocument') }}</div>
            <div class="text-xs"><span class="font-medium">{{ $t('uploads.clickToUpload') }}</span> {{ $t('uploads.dragAndDrop') }}</div>
          </span>
          <span v-else data-target="file-dropzone.loading" class="flex flex-col items-center">
            <SvgIcon name="upload" class="h-10 w-10 animate-spin" />
            <div class="mb-1 font-medium">{{ $t('uploads.uploading') }}</div>
          </span>
        </div>

        <input
          id="assetsFieldHandle"
          ref="file"
          name="fields[assetsFieldHandle][]"
          class="hidden"
          type="file"
          accept="image/png, image/jpeg, image/tiff, application/pdf, .docx, .doc, .xlsx, .xls"
          multiple
        />
      </div>
    </label>
  </div>
</template>

<script setup lang="ts">
import { getCurrentInstance, ref } from "vue";
import { useI18n } from "vue-i18n";
import { fetchWithAuth } from "@/utils/auth";

const { t } = useI18n();

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
    const response = await fetchWithAuth("/api/v1/upload", {
      method: "POST",
      body: formData
    });

    if (response.ok) {
      const data = await response.json();
      // Reset file input
      if (instance?.refs.file) {
        instance.refs.file.value = "";
      }
    } else {
      const errorData = await response.json().catch(() => ({ message: "Failed to upload file" }));
      throw new Error(errorData.message || "Failed to upload file");
    }
  } catch (error) {
    console.error("Error uploading file:", error);
    alert(`Error: ${error instanceof Error ? error.message : "Failed to upload file"}`);
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
