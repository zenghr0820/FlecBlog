<script setup lang="ts">
const { basicConfig, blogConfig } = useSysConfig();
const { total: articlesTotal, fetchArticles } = useArticles();
const avatarUrl = computed(() => basicConfig.value.author_avatar || '/avatar.webp');
const { total: categoriesTotal } = useCategories();
const { total: tagsTotal } = useTags();

// 作者卡片背景图片URL
const authorCardBgUrl = computed(() => {
  return blogConfig.value.author_card_bg || '/author_bg.webp'
});

const parseJSON = <T = any,>(jsonStr: string | undefined, fallback: T): T => {
  try {
    return jsonStr ? JSON.parse(jsonStr) : fallback;
  } catch {
    return fallback;
  }
};

const contacts = computed(() => {
  const socialList = parseJSON<Array<{ name: string; url: string; icon: string }>>(
    blogConfig.value.sidebar_social,
    []
  );
  return socialList.filter(item => item.url && item.url.trim() !== '');
});

onMounted(async () => {
  if (articlesTotal.value === 0) {
    await fetchArticles({ page: 1, page_size: 1 });
  }
});
</script>

<template>
  <div class="card-widget card-info is-center" :style="{ '--author-card-bg-url': authorCardBgUrl ? `url(${authorCardBgUrl})` : 'none' }">
    <div class="author-info-detail">
      <p class="author-info-hello">👋 欢迎光临！</p>
      <p class="author-info-desc">{{ basicConfig.author_desc }}</p>
    </div>
    <div class="avatar-img">
      <NuxtImg :src="avatarUrl" alt="头像" loading="lazy" />
    </div>
    <div class="author-info-name">{{ basicConfig.author }}</div>
    <div class="site-data">
      <ClientOnly>
        <router-link to="/archive" :aria-label="`查看全部 ${articlesTotal} 篇文章`">
          <div class="headline">文章</div>
          <div class="num">{{ articlesTotal }}</div>
        </router-link>
        <router-link to="/categories" :aria-label="`查看全部 ${categoriesTotal} 个分类`">
          <div class="headline">分类</div>
          <div class="num">{{ categoriesTotal }}</div>
        </router-link>
        <router-link to="/tags" :aria-label="`查看全部 ${tagsTotal} 个标签`">
          <div class="headline">标签</div>
          <div class="num">{{ tagsTotal }}</div>
        </router-link>
      </ClientOnly>
    </div>
    <a id="card-info-btn" target="_blank" rel="noopener" href="https://github.com/zenghr0820">
      <i class="fab fa-github"></i>
      <span>Follow Me 🛫</span>
    </a>
    <div class="card-info-icons">
      <a
        v-for="contact in contacts"
        :key="contact.name"
        :href="contact.url"
        class="icon"
        target="_blank"
        :aria-label="`访问 ${contact.name}`"
        rel="noopener noreferrer"
      >
        <i :class="'ri-' + contact.icon" aria-hidden="true"></i>
      </a>
    </div>
  </div>
</template>

<style lang="scss">
@use '@/assets/css/card.scss'  as *;
</style>
