<script>
  import { GradientButton, Label, Input, Alert } from "flowbite-svelte";
  import { Card, Listgroup, Avatar } from "flowbite-svelte";
  import {
    Connect,
    Disconnect,
    GetDataManually,
  } from "../wailsjs/go/main/App.js";
  import {
    ArrowDownToBracketOutline,
    ArrowRightOutline,
    ArrowsRepeatOutline,
    CartSolid,
    ExclamationCircleOutline,
  } from "flowbite-svelte-icons";
  import { Button, Dropdown, DropdownItem, Radio } from "flowbite-svelte";
  import {
    ChevronDownOutline,
    ServerOutline,
    CloudArrowUpOutline,
    CloseCircleOutline,
  } from "flowbite-svelte-icons";
  // import Fa from "svelte-fa/dist/fa.svelte";
  // import {
  //   faCaretDown,
  //   faCaretUp,
  // } from "@fortawesome/free-solid-svg-icons/index.es";
  import SvelteLogo from "virtual:icons/logos/svelte-icon";
  import HugeiconsMouse03 from "~icons/hugeicons/mouse-03";
  import LineMdCircleTwotoneToConfirmCircleTwotoneTransition from "~icons/line-md/circle-twotone-to-confirm-circle-twotone-transition";
  import LineMdCloseCircleTwotone from "~icons/line-md/close-circle-twotone";
  import LineMdCloudDownTwotone from "~icons/line-md/cloud-down-twotone";
  import LineMdDownloading from "~icons/line-md/downloading";
  import LucideServer from "~icons/lucide/server";
  import LucideListOrdered from "~icons/lucide/list-ordered";
  import LineMdListIndented from "~icons/line-md/list-indented";
  import LineMdDoubleArrowVertical from "~icons/line-md/double-arrow-vertical";
  import LineMdDoubleArrowHorizontal from "~icons/line-md/double-arrow-horizontal";
  import LucideMousePointerClick from "~icons/lucide/mouse-pointer-click";
  import LucideTimerReset from "~icons/lucide/timer-reset";
  import LucideInfo from "~icons/lucide/info";

  let host, port;
  let updateMethod = 0;
  let updateMethods = ["Manual", "Server events", "Client polling"];
  let updateMethodKeys = ["manual", "server_events", "client_polling"];
  let connected = false;
  let response;

  globalThis.runtime.EventsOn("processCollectorConnected", (state) => {
    connected = state;
    console.log("connected to mouse collector");
  });

  globalThis.runtime.EventsOn("processCollectorReceiveData", (data) => {
    console.log(
      "recieved data from process collector: " + JSON.stringify(data)
    );
    if (data?.Status && data.Status != "no changes") response = data;
  });
</script>

<div
  class="flex-grow bg-gray-700 text-gray-100 shadow-lg rounded-lg p-6 overflow-y-auto"
>
  <div class="flex justify-between items-center mb-4">
    <h1 class="text-xl font-bold text-left flex gap-1">
      <LineMdListIndented /> Process
    </h1>
  </div>

  <div class="bar flex items-center gap-2">
    <div>
      <Input placeholder="Collector IP" bind:value={host}>
        <LucideServer
          slot="left"
          class="w-5 h-5 text-gray-500 dark:text-gray-400"
        />
      </Input>
    </div>

    <div class="w-28">
      <Input placeholder="Port" type="number" bind:value={port}>
        <LucideListOrdered
          slot="left"
          class="w-5 h-5 text-gray-500 dark:text-gray-400"
        />
      </Input>
    </div>

    <Button>
      <LineMdDownloading class="w-6 h-6 me-2" />
      {updateMethods[updateMethod]}
      <ChevronDownOutline class="w-6 h-6 ms-2 text-white dark:text-white" />
    </Button>
    <Dropdown class="w-44 p-3 space-y-3 text-sm">
      {#each updateMethods as item, i}
        <li>
          <Radio name="group1" bind:group={updateMethod} value={i}>{item}</Radio
          >
        </li>
      {/each}
    </Dropdown>

    {#if !connected}
      <GradientButton
        color="green"
        on:click={() => {
          Connect("process", host, port, updateMethodKeys[updateMethod]);
        }}
        ><LineMdCircleTwotoneToConfirmCircleTwotoneTransition
          class="w-6 h-6 me-2"
        />Connect</GradientButton
      >
    {:else}
      <GradientButton color="red" on:click={() => Disconnect("process")}
        ><LineMdCloseCircleTwotone
          class="w-6 h-6 me-2"
        />Disconnect</GradientButton
      >
    {/if}

    <GradientButton color="blue" on:click={() => GetDataManually("process")}
      ><LineMdCloudDownTwotone class="w-6 h-6 me-2" />Get</GradientButton
    >
  </div>

  <ul class="text-left">
    {#if response?.Status && response?.Status != "error" && response?.Data}
      <p class="text-left mt-4 flex gap-1">
        <LucideInfo /> PID: {response.Data.Pid}
      </p>
      <p class="text-left flex gap-1">
        <LucideTimerReset /> Uptime: {response.Data.Uptime} ms
      </p>

      <p class="text-left mt-4">Status: {response.Status}</p>
      <p class="text-left">Request timestamp: {response.Timestamp}</p>
    {:else if response?.Status && response?.Status == "error"}
      <Alert border color="red">
        <ExclamationCircleOutline slot="icon" class="w-5 h-5" />
        <span class="font-medium">Error</span>
        {response.Error}
      </Alert>
    {/if}
  </ul>
</div>

<style>
</style>
