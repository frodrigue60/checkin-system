<script lang="ts">
  import { Menu } from "lucide-svelte";
  import { Button } from "$lib/components/ui/button";
  import { PUBLIC_APP_NAME } from "$env/static/public";

  let { onMenuClick } = $props();
</script>

<header class="lg:hidden fixed top-0 w-full z-50 bg-white border-b border-outline-variant/10">
    <div class="flex justify-between items-center px-6 py-4 w-full">
        <h1 class="font-display font-bold text-lg text-primary tracking-tight">{PUBLIC_APP_NAME}</h1>
        
        <Button variant="ghost" size="icon" onclick={onMenuClick} class="text-primary hover:bg-primary/5">
            <Menu class="h-6 w-6" />
        </Button>
    </div>
</header>

<style>
    :global(.font-display) {
        font-family: 'Public Sans', sans-serif;
    }
</style>
