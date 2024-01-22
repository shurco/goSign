<template>
  <label for="assetsFieldHandle" class="w-64 block h-52 relative hover:bg-[#efeae6]/30 rounded-xl border-2 border-[#e7e2df] border-dashed cursor-pointer"
    :class="{ 'bg-[#efeae6]/30': status }" @dragover="dragover" @drop="drop" @change="onChange">
    <div class="absolute top-0 right-0 left-0 bottom-0 flex items-center justify-center">
      <div class="flex flex-col items-center">
        <span data-target="file-dropzone.icon" class="flex flex-col items-center" v-if="!status">
          <span>
            <SvgIcon name="cloud-upload" class="w-10 h-10" />
          </span>
          <div class="font-medium mb-1">
            Upload New Document
          </div>
          <div class="text-xs">
            <span class="font-medium">Click to upload</span> or drag and drop
          </div>
        </span>
        <span data-target="file-dropzone.loading" class="flex flex-col items-center" v-else>
          <SvgIcon name="upload" class="w-10 h-10 animate-spin" />
          <div class="font-medium mb-1">
            Uploading...
          </div>
        </span>
      </div>

      <input id="assetsFieldHandle" name="fields[assetsFieldHandle][]" class="hidden" type="file" ref="file"
        accept="image/png, image/jpeg, image/tiff, application/pdf, .docx, .doc, .xlsx, .xls" multiple>
    </div>
  </label>
</template>

<script setup lang="ts">
import { ref, getCurrentInstance } from "vue";

const status = ref(false);
const instance: any = getCurrentInstance();
const emits = defineEmits(["added"]);

const onChange = async () => {
  const files = instance?.refs.file.files;
  if (!files || !files.length) {
    console.error('No file selected');
    return;
  }

  const formData = new FormData();
  for (const file of files) {
    formData.append("document", file);
  }

  try {
    status.value = true;
    const response = await fetch('/api/upload', {
      method: 'POST',
      body: formData
    });

    //if (response.ok) {
    //  const data = await response.json();
    //  console.log('Server response:', data);
    //} else {
    //  throw new Error('File upload failed');
    //}
  } catch (error) {
    console.error('Error uploading file:', error);
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