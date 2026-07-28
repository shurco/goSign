<script lang="ts">
  import SvgIcon from "@/components/SvgIcon.svelte";
  import { apiUrl } from "@/services/api";
  import { toDate } from "@/utils/time";
  import type { Signer } from "@/models/index";

  interface VerifyData {
    verify: boolean;
    error: string | null;
    signers: Signer[];
  }

  interface VerifyInfo {
    success: boolean;
    data: VerifyData;
  }

  let verifyInfo = $state<VerifyInfo>({
    success: false,
    data: {
      verify: false,
      error: null,
      signers: []
    }
  });
  let status = $state(false);
  let fileEl = $state<HTMLInputElement | null>(null);

  const onChange = async () => {
    const files = fileEl?.files;
    if (!files || !files.length) {
      console.error("No file selected");
      return;
    }

    const formData = new FormData();
    for (const file of files) {
      formData.append("document", file);
    }

    try {
      status = true;
      const response = await fetch(apiUrl("/verify/pdf"), {
        method: "POST",
        body: formData
      });

      if (!response.ok) {
        throw new Error(`File upload failed: ${response.statusText}`);
      }

      verifyInfo = await response.json();
    } catch (error) {
      console.error("Error uploading file:", error);
      verifyInfo = {
        success: false,
        data: {
          verify: false,
          error: error instanceof Error ? error.message : "Unknown error occurred",
          signers: []
        }
      };
    } finally {
      status = false;
    }
  };

  const dragover = (event: DragEvent) => {
    event.preventDefault();
  };

  const drop = (event: DragEvent) => {
    event.preventDefault();
    event.stopPropagation();
    if (event.dataTransfer?.files && event.dataTransfer.files.length > 0 && fileEl) {
      fileEl.files = event.dataTransfer.files;
      onChange();
    }
  };
</script>

<div class="grid h-48 grid-cols-1 place-content-center gap-4">
  <div>
    <label
      for="assetsFieldHandle"
      class="relative block h-52 w-96 cursor-pointer rounded-md border-2 border-dashed border-[#343434] bg-[#efeae6]/70 hover:bg-[#efeae6]/0 {status
        ? 'bg-[#efeae6]0'
        : ''}"
      ondragover={dragover}
      ondrop={drop}
    >
      <div class="absolute top-0 right-0 bottom-0 left-0 flex items-center justify-center">
        <div class="flex flex-col items-center">
          {#if !status}
            <span data-target="file-dropzone.icon" class="flex flex-col items-center">
              <span>
                <SvgIcon name="cloud-upload" class="h-10 w-10" />
              </span>
              <div class="mb-1 font-medium">Verify Signed PDF</div>
              <div class="text-xs"><span class="font-medium">Click to upload</span> or drag and drop</div>
            </span>
          {:else}
            <span data-target="file-dropzone.loading" class="flex flex-col items-center">
              <SvgIcon name="upload" class="h-10 w-10 animate-spin" />
              <div class="mb-1 font-medium">Uploading...</div>
            </span>
          {/if}
        </div>

        <input
          id="assetsFieldHandle"
          bind:this={fileEl}
          name="fields[assetsFieldHandle][]"
          class="hidden"
          type="file"
          accept="application/pdf"
          onchange={onChange}
        />
      </div>
    </label>
  </div>

  {#if verifyInfo.success}
    <div class="w-96">
      {#if verifyInfo.data.verify}
        <div role="status" class="rounded border-s-4 border-green-500 bg-green-50 p-4">
          <div>
            {#if verifyInfo.data.signers.length > 1}
              <div class="space-y-1 font-medium">
                <span>Total signer: {verifyInfo.data.signers.length}</span>
              </div>
              <span class="flex items-center">
                <span class="my-2 h-px flex-1 bg-green-500"></span>
              </span>
            {/if}

            {#each verifyInfo.data.signers as item, index (index)}
              <div class="flex items-center space-x-1">
                {#if item.valid_signature}
                  <SvgIcon name="check-badge" class="h-5 w-5 text-green-500" />
                {:else}
                  <SvgIcon name="check-badge" class="h-5 w-5 text-red-500" />
                {/if}
                <span>Signature valid</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon
                  name="check-badge"
                  class="h-5 w-5 {item.time_stamp != null ? 'text-green-500' : 'text-gray-400'}"
                />
                <span class={item.time_stamp === null ? "text-gray-400" : ""}>Timestamp</span>
              </div>

              <div class="flex items-center space-x-1">
                {#if item.trusted_issuer.valid}
                  <SvgIcon name="check-badge" class="h-5 w-5 text-green-500" />
                {:else}
                  <SvgIcon name="x-circle" class="h-5 w-5 text-red-500" />
                {/if}
                <span>Trusted issuer</span>
              </div>

              <div class="flex items-center space-x-1">
                {#if !item.revoked_certificate}
                  <SvgIcon name="check-badge" class="h-5 w-5 text-green-500" />
                {:else}
                  <SvgIcon name="x-circle" class="h-5 w-5 text-red-500" />
                {/if}
                <span>Active certificate</span>
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon name="user" class="h-5 w-5" />
                <span class="max-w-80">{item.name ? item.name : item.reason}</span>
              </div>

              {#if item.time_stamp != null}
                <div class="flex items-center space-x-1">
                  <SvgIcon name="calendar" class="h-5 w-5" />
                  <span class="max-w-80">{toDate(item.time_stamp.time)} </span>
                </div>
              {/if}

              <div class="flex items-center space-x-1">
                <SvgIcon name="sign" class="h-5 w-5" />
                <span
                  >{item.cert_subject?.common_name
                    ? item.cert_subject.common_name
                    : item.cert_subject?.organization}</span
                >
              </div>

              <div class="flex items-center space-x-1">
                <SvgIcon name="lock-cube" class="h-5 w-5" />
                <span>{item.sig_format}</span>
              </div>

              {#if verifyInfo.data.signers.length - 1 > index}
                <span class="flex items-center">
                  <span class="my-2 h-px flex-1 bg-green-500"></span>
                </span>
              {/if}
            {/each}
          </div>
        </div>
      {:else}
        <div role="alert" class="rounded border-s-4 border-red-500 bg-red-50 p-4">
          <strong class="block font-medium text-red-800">PDF failed verification</strong>
          <p class="mt-2 text-sm text-red-700">
            {verifyInfo.data.error}
          </p>
        </div>
      {/if}
    </div>
  {/if}
</div>
