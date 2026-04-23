<script lang="ts">
  interface Props {
    name:
      | 'plus'
      | 'minus'
      | 'undo'
      | 'trash'
      | 'refresh'
      | 'check'
      | 'close'
      | 'branch'
    size?: number
    class?: string
  }
  let { name, size = 16, class: klass = '' }: Props = $props()
</script>

<svg
  xmlns="http://www.w3.org/2000/svg"
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  class={klass}
  aria-hidden="true"
>
  {#if name === 'plus'}
    <path d="M12 5v14M5 12h14" />
  {:else if name === 'minus'}
    <path d="M5 12h14" />
  {:else if name === 'undo'}
    <path d="M3 7v6h6" />
    <path d="M21 17a9 9 0 0 0-15-6.7L3 13" />
  {:else if name === 'trash'}
    <path d="M3 6h18" />
    <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
    <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" />
  {:else if name === 'refresh'}
    <path d="M3 12a9 9 0 0 1 15-6.7L21 8" />
    <path d="M21 3v5h-5" />
    <path d="M21 12a9 9 0 0 1-15 6.7L3 16" />
    <path d="M3 21v-5h5" />
  {:else if name === 'check'}
    <path d="M20 6L9 17l-5-5" />
  {:else if name === 'close'}
    <path d="M18 6L6 18M6 6l12 12" />
  {:else if name === 'branch'}
    <circle cx="6" cy="6" r="2" />
    <circle cx="6" cy="18" r="2" />
    <circle cx="18" cy="9" r="2" />
    <path d="M6 8v8" />
    <path d="M6 11h6a3 3 0 0 0 3-3V7" />
  {/if}
</svg>
