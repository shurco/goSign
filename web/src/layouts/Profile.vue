<template>
  <div class="relative flex h-screen overflow-hidden">
    <div class="flex h-screen w-16 flex-col justify-between border-e bg-white">
      <div>
        <div class="inline-flex h-16 w-16 items-center justify-center">
          <SvgIcon name="logo" stroke="currentColor" class="h-8 w-8" />
        </div>
      </div>

      <div class="sticky inset-x-0 bottom-0 border-t border-gray-100 bg-white px-2">
        <ul class="py-1.5">
          <li>
            <a href="/" target="_blank" class="bg-white hover:bg-red-50 hover:text-red-500">
              <div class="group cursor-pointer relative flex justify-center rounded px-2 py-1.5 text-gray-500 hover:bg-red-100 hover:text-red-700">
                <SvgIcon name="exit" stroke="currentColor" class="h-6 w-6" />
              </div>
            </a>
          </li>
        </ul>
      </div>
    </div>

    <div class="overflow-x-hidden overflow-y-auto relative block px-5">
      <slot />
    </div>
  </div>
</template>
