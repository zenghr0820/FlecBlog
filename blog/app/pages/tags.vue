<script lang="ts" setup>
definePageMeta({
  showSidebar: false,
});

const { tags } = useTags();

// 容器引用
const containerRef = ref(null)

// 随机颜色函数
const getRandomColor = () => {
  const colors = [
    '#ff6b6b', '#4ecdc4', '#45b7d1', '#96ceb4',
    '#ffeaa7', '#dda0dd', '#f8b195', '#6c5ce7'
  ]
  return colors[Math.floor(Math.random() * colors.length)]
}

onMounted(() => {
  if (!containerRef.value) return
  
  // 获取 div 里所有 a 标签
  const links = containerRef.value.querySelectorAll('a')
  
  // 每个都随机颜色
  links.forEach(link => {
    link.style.color = getRandomColor()
  })
})

useSeoMeta({
  title: '标签',
  description: '浏览所有文章标签，快速找到感兴趣的内容',
});
</script>

<template>
  <!-- 内容区域 -->
  <div id="page">
    <h1 class="page-title">标签</h1>
    <div ref="containerRef" class="tag-cloud-list">
      <router-link v-for="tag in tags" :key="tag.id" :to="tag.url" :title="tag.name">
        {{ tag.name }}
      </router-link>
    </div>
  </div>
</template>

<style lang="scss">
@use '@/assets/css/mixins' as *;

#page {
  @extend .cardHover;
  align-self: flex-start;
  padding: 40px;

  .page-title {
    margin: 0 0 10px;
    font-weight: bold;
    font-size: 2rem;
  }

  .tag-cloud-list {
    text-align: center;

    a {
      display: inline-block;
      margin: 2px;
      padding: 2px 7px;
      line-height: 1.7;
      transition: all 0.3s;
      font-size: 1.2em;
      border-radius: 5px;

      &:hover { 
      background: var(--flec-nav-menu-bg-hover) !important;
      box-shadow: 2px 2px 6px rgba(0, 0, 0, .2);
      color: var(--white) !important;
      }
    }
  }
}

// 响应式设计
@media screen and (max-width: 1024px) {
  #page {
    padding: 30px;

    .page-title {
      font-size: 1.75rem;
    }

    .tag-cloud-list {
      a {
        font-size: 1.1em;
      }
    }
  }
}

@media screen and (max-width: 768px) {
  #page {
    padding: 18px;

    .page-title {
      font-size: 1.4rem;
    }

    .tag-cloud-list {
      a {
        font-size: 0.95em;
        padding: 2px 6px;
        margin: 1.5px;
      }
    }
  }
}
</style>
