<template>
  <div class="grid h-48 grid-cols-1 place-content-center gap-4">
    <div>
      <label
        for="assetsFieldHandle"
        class="relative block h-52 w-96 cursor-pointer rounded-xl border-2 border-dashed border-[#e7e2df] hover:bg-[#efeae6]/30"
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
              <div class="mb-1 font-medium">Upload New Document</div>
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
    <div v-if="signInfo.data.file_name_signed" class="w-96">
      <div role="info" class="rounded border-s-4 border-green-500 bg-green-50 p-4">
        <a target="_blank" :href="`/drive/signed/${signInfo.data.file_name_signed}`">signed PDF</a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getCurrentInstance, ref } from "vue";

const signPDF = {
  success: false,
  data: {
    file_name_signed: null
  }
};

const signInfo = ref(signPDF);
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
    const response = await fetch("/api/sign", {
      method: "POST",
      body: formData
    });

    signInfo.value = await response.json();
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
