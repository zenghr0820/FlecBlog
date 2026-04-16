<script lang="ts" setup>
import { getArticleBySlug } from '@/composables/api/article';

const router = useRouter()
const route = useRoute()

// 使用 useAsyncData 确保服务端渲染时数据已加载
const { data: article } = await useAsyncData(
  'post-header-article',
  async () => {
    const slug = route.params.slug as string
    return await getArticleBySlug(slug)
  }
)

// 计算文章字数（去除 Markdown 标记后的准确字数）
const wordCount = computed(() => {
  if (!article.value?.content) return 0
  return countWords(article.value.content)
})

// 计算阅读时长（按每分钟300字计算）
const readingTime = computed(() => {
  if (!article.value?.content) return 0
  return estimateReadingTime(article.value.content, 300)
})

// 文章评论总数
const commentCount = computed(() => {
  return article.value?.comment_count || 0
})

// 跳转到分类详情页
const goToCategory = () => {
  if (article.value?.category?.url) {
    router.push(article.value.category.url)
  }
}
</script>

<template>
  <div
    class="post-header"
    :style="{
      backgroundImage: article?.cover ? `url(${article.cover})` : 'none'
    }"
  >
    <div v-if="article" class="post-info">
      <h1 class="post-title">{{ article.title }}</h1>

      <!-- 移动端：合并为一行 -->
      <div class="post-meta post-meta-mobile">
        <span class="post-meta-item">
          <i class="ri-calendar-line"></i>
          <span>发表于 {{ formatFriendly(article.publish_time) }}</span>
        </span>
        <span v-if="article.update_time" class="post-meta-item">
          <i class="ri-refresh-line"></i>
          <span>更新于 {{ formatFriendly(article.update_time) }}</span>
        </span>
        <span v-if="article.location" class="post-meta-item">
          <i class="ri-map-pin-line"></i>
          <span>{{ article.location }}</span>
        </span>
        <span
          v-if="article.category"
          class="post-meta-item clickable"
          @click="goToCategory"
        >
          <i class="ri-folder-line"></i>
          <span>{{ article.category.name }}</span>
        </span>
        <span class="post-meta-item">
          <i class="ri-file-word-line"></i>
          <span>总字数: {{ wordCount }}</span>
        </span>
        <span class="post-meta-item">
          <i class="ri-time-line"></i>
          <span>阅读时长: {{ readingTime }}分钟</span>
        </span>
        <span class="post-meta-item">
          <i class="ri-eye-line"></i>
          <span>浏览量: {{ article.view_count }}</span>
        </span>
        <span
          class="post-meta-item clickable"
          @click="scrollToElement('.comment-input')"
        >
          <i class="ri-message-3-line"></i>
          <span>评论数: {{ commentCount }}</span>
        </span>
      </div>

      <!-- 桌面端：分两行显示 -->
      <div class="post-meta-desktop">
        <div class="post-meta">
          <span class="post-meta-item">
            <i class="ri-calendar-line"></i>
            <span>发表于 {{ formatFriendly(article.publish_time) }}</span>
          </span>
          <span v-if="article.update_time" class="post-meta-item">
            <i class="ri-refresh-line"></i>
            <span>更新于 {{ formatFriendly(article.update_time) }}</span>
          </span>
          <span v-if="article.location" class="post-meta-item">
            <i class="ri-map-pin-line"></i>
            <span>{{ article.location }}</span>
          </span>
          <span
            v-if="article.category"
            class="post-meta-item clickable"
            @click="goToCategory"
          >
            <i class="ri-folder-line"></i>
            <span>{{ article.category.name }}</span>
          </span>
        </div>
        <div class="post-meta">
          <span class="post-meta-item">
            <i class="ri-file-word-line"></i>
            <span>总字数: {{ wordCount }}</span>
          </span>
          <span class="post-meta-item">
            <i class="ri-time-line"></i>
            <span>阅读时长: {{ readingTime }}分钟</span>
          </span>
          <span class="post-meta-item">
            <i class="ri-eye-line"></i>
            <span>浏览量: {{ article.view_count }}</span>
          </span>
          <span
            class="post-meta-item clickable"
            @click="scrollToElement('.comment-input')"
          >
            <i class="ri-message-3-line"></i>
            <span>评论数: {{ commentCount }}</span>
          </span>
        </div>
      </div>
    </div>
    <section class="main-hero-waves-area waves-area">
      <svg class="waves-svg" xmlns="http://www.w3.org/2000/svg" xlink="http://www.w3.org/1999/xlink" viewBox="0 24 150 28" preserveAspectRatio="none" shape-rendering="auto">
        <defs>
          <path id="gentle-wave" d="M-160 44c30 0 58-18 88-18s58 18 88 18 58-18 88-18 58 18 88 18v44h-352Z"></path>
        </defs>
        <g class="parallax">
          <use href="#gentle-wave" x="48" y="0"></use>
          <use href="#gentle-wave" x="48" y="3"></use>
          <use href="#gentle-wave" x="48" y="5"></use>
          <use href="#gentle-wave" x="48" y="7"></use>
          </g>
      </svg>
    </section>
  </div>
</template>

<style lang="scss" scoped>
.post-header {
  position: relative;
  height: 400px;
  width: 100%;
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  // 添加遮罩层，确保文字清晰可见
  &::before {
    position: absolute;
    width: 100%;
    height: 100%;
    background-color: var(--mark-bg);
    content: '';
  }

  .post-info {
    position: absolute;
    padding: 0 8%;
    width: 100%;
    text-align: center;
    height: 100%;
    display: -webkit-box;
    display: -moz-box;
    display: -webkit-flex;
    display: -ms-flexbox;
    display: flex;
    -webkit-box-pack: center;
    -moz-box-pack: center;
    -o-box-pack: center;
    -ms-flex-pack: center;
    -webkit-justify-content: center;
    justify-content: center;
    -webkit-flex-wrap: npwrap;
    -ms-flex-wrap: npwrap;
    flex-wrap: npwrap;
    -webkit-box-orient: vertical;
    -moz-box-orient: vertical;
    -o-box-orient: vertical;
    -webkit-flex-direction: column;
    -ms-flex-direction: column;
    flex-direction: column;
    backdrop-filter: blur(15px);

    .post-title {
      margin-bottom: 8px;
      color: var(--white);
      font-weight: 400;
      font-size: 2.5em;
      line-height: 1.5;
      -webkit-line-clamp: 3;
    }

    // 默认显示桌面端，隐藏移动端
    .post-meta-mobile {
      display: none !important;
    }

    .post-meta-desktop {
      display: block !important;
    }

    .post-meta {
      display: flex;
      align-items: center;
      flex-wrap: wrap;
      justify-content: center;
      color: var(--light-grey);
      font-size: 95%;

      .post-meta-item {
        display: flex;
        align-items: center;

        &:not(:last-child)::after {
          content: '|';
          color: rgba(255, 255, 255, 0.6);
          margin: 0 0.5rem;
        }

        i {
          font-size: 1rem;
          margin-right: 0.3rem;
        }

        &.clickable {
          cursor: pointer;
          transition: all 0.3s ease;

          &:hover {
            color: rgba(255, 255, 255, 1);
          }
        }
      }
    }
  }

 

}

 /* 波浪css */
.main-hero-waves-area {
  width: 100%;
  position: absolute;
  left: 0;
  bottom: -11px;
  z-index: 5;
}
.waves-area .waves-svg {
  width: 100%;
  height: 5rem;
}
/* Animation */

.parallax > use {
  animation: move-forever 25s cubic-bezier(0.55, 0.5, 0.45, 0.5) infinite;
}
.parallax > use:nth-child(1) {
  animation-delay: -2s;
  animation-duration: 7s;
  fill: #f7f9febd;
}
.parallax > use:nth-child(2) {
  animation-delay: -3s;
  animation-duration: 10s;
  fill: #f7f9fe82;
}
.parallax > use:nth-child(3) {
  animation-delay: -4s;
  animation-duration: 13s;
  fill: #f7f9fe36;
}
.parallax > use:nth-child(4) {
  animation-delay: -5s;
  animation-duration: 20s;
  fill: #f7f9fe;
}
/* 黑色模式背景 */
[data-theme="dark"] .parallax > use:nth-child(1) {
  animation-delay: -2s;
  animation-duration: 7s;
  fill: #0f172ab3
}
[data-theme="dark"] .parallax > use:nth-child(2) {
  animation-delay: -3s;
  animation-duration: 10s;
  fill: #0f172a80;
}
[data-theme="dark"] .parallax > use:nth-child(3) {
  animation-delay: -4s;
  animation-duration: 13s;
  fill: #0f172a4d;
}
[data-theme="dark"] .parallax > use:nth-child(4) {
  animation-delay: -5s;
  animation-duration: 20s;
  fill: rgba(66, 249, 251, 0.3);
}

@keyframes move-forever {
  0% {
    transform: translate3d(-90px, 0, 0);
  }
  100% {
    transform: translate3d(85px, 0, 0);
  }
}
/*Shrinking for mobile*/
@media (max-width: 768px) {
  .waves-area .waves-svg {
    height: 40px;
    min-height: 40px;
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  .post-header {
    height: 350px;

    .post-info {
      padding: 0 1.25rem;

      .post-meta {
        font-size: 0.8rem;
        line-height: 1.6;
      }
    }
  }
}

@media screen and (max-width: 500px) {
  .post-header {
    .post-info {
      // 移动端显示移动版，隐藏桌面版
      .post-meta-mobile {
        display: flex !important;
      }

      .post-meta-desktop {
        display: none !important;
      }
    }
  }
}
</style>
