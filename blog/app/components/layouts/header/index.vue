<script lang="ts" setup>
import Navbar from '@/components/layouts/navbar/index.vue';
import HomeHeader from './HomeHeader.vue'
import PostHeader from './PostHeader.vue'

const route = useRoute();
const type = computed(() => route.meta.typeHeader as string | undefined);
</script>

<template>
<header>
    <Navbar />
    <HomeHeader v-if="type === 'home'" />
    <PostHeader v-else-if="type === 'post'" />
</header>
  
</template>
