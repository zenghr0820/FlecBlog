<script lang="ts" setup></script>

<template>
  <footer id="footer">
    <LayoutsFooterContact />
    <LayoutsFooterNavigation />
    <LayoutsFooterCopyright />
  </footer>
</template>

<style lang="scss" scoped>
#footer {
  position: relative;
  padding-top: 1rem;
  display: flex;
  flex-direction: column;
  gap: 32px;
  align-items: center;
  width: 100%;
  background: var(--flec-footer-content-bg);
}
</style>
