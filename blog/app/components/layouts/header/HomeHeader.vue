<script lang="ts" setup>

const { basicConfig, blogConfig } = useSysConfig()
const displayText = ref('')
const typingSpeed = 150 // 打字速度（毫秒）
const deletingSpeed = 80 // 删除速度（毫秒）
const pauseTime = 2000 // 暂停时间（毫秒）
let typingTimer: number | null = null


const avatarUrl = computed(
  () => basicConfig.value.author_avatar || '/avatar.webp'
)

// 获取打字机文本列表
const getTypingTexts = (): string[] => {
  try {
    const parsed = JSON.parse(blogConfig.value.typing_texts || '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

const scrollToContent = () => {
  window.scrollTo({
    top: window.innerHeight - 64,
    behavior: 'smooth'
  })
}

const typeWriter = () => {
  const texts = getTypingTexts()
  if (texts.length === 0) return

  let textIndex = 0 // 当前显示的文本索引
  let charIndex = 0 // 当前字符索引
  let isDeleting = false

  const animate = () => {
    const currentText = texts[textIndex]
    if (!currentText) return

    if (!isDeleting) {
      // 打字阶段
      if (charIndex < currentText.length) {
        displayText.value += currentText.charAt(charIndex)
        charIndex++
        typingTimer = window.setTimeout(animate, typingSpeed)
      } else {
        // 打完后停留一会再删除
        isDeleting = true
        typingTimer = window.setTimeout(animate, pauseTime)
      }
    } else {
      // 删除阶段
      if (charIndex > 0) {
        displayText.value = currentText.substring(0, charIndex - 1)
        charIndex--
        typingTimer = window.setTimeout(animate, deletingSpeed)
      } else {
        // 删除完成，切换到下一条文本
        isDeleting = false
        textIndex = (textIndex + 1) % texts.length // 循环到下一条
        typingTimer = window.setTimeout(animate, typingSpeed)
      }
    }
  }

  animate()
}

onMounted(() => {
  // 延迟一点开始打字效果
  setTimeout(typeWriter, 500)
})

onUnmounted(() => {
  if (typingTimer) {
    clearTimeout(typingTimer)
  }
})
</script>

<template>
  <div  class="home-header">
    <div class="site-info">
      <div class="info_title_avatar">
        <NuxtImg :src="avatarUrl" alt="头像" loading="lazy" id="site-info-avatar"  />
        </div>
      <h1>{{ blogConfig.title }}</h1>
      <div class="site-subtitle">
        <span id="subtitle">{{ displayText }}</span>
        <span class="cursor">|</span>
      </div>
    </div>
    <div class="scroll-indicator" @click="scrollToContent">
      <i class="ri-arrow-down-s-line ri-2x"></i>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.home-header {
  position: relative;
  width: 100%;
  height: 100vh;
  background-attachment: fixed;
  background-position: center center;
  background-size: cover;
  background-repeat: no-repeat;
  -webkit-transition: all .5s;
  -moz-transition: all .5s;
  -o-transition: all .5s;
  -ms-transition: all .5s;
  transition: all .5s;

  &::before{
    position: absolute;
    width: 100%;
    height: 100%;
    background-color: var(--mark-bg);
    content: '';
  }

  .site-info {
    position: absolute;
    top: 35%;
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;

    .info_title_avatar {
      #site-info-avatar {
        text-align: center;
        border: 4px solid var(--white);
        padding: 4px;
        height: 130px;
        width: 130px;
        border-radius: 50%;
        -webkit-box-sizing: border-box;
        -moz-box-sizing: border-box;
        box-sizing: border-box; 
      }

    }

    

    h1 {
      font-size: 2.6rem;
      color: var(--white);
    }

    .site-subtitle {
      font-size: 1.7rem;
      color: var(--light-grey);

      .cursor {
        display: inline-block;
        margin-left: 4px;
        animation: blink 1s infinite;
      }
    }
  }

  .scroll-indicator {
    position: absolute;
    bottom: 10px;
    width: 100%;
    animation: bounce 1.5s infinite;
    cursor: pointer;

    i {
      color: #eee;
      position: relative;
      text-align: center;
      width: 100%;
    }
  }
}

@keyframes bounce {
  0% {
    opacity: 0.4;
    transform: translate(0, 0);
  }

  50% {
    opacity: 1;
    transform: translate(0, -16px);
  }

  100% {
    opacity: 0.4;
    transform: translate(0, 0);
  }
}

@keyframes blink {
  0%,
  49% {
    opacity: 1;
  }

  50%,
  100% {
    opacity: 0;
  }
}

// 响应式设计
@media screen and (max-width: 768px) {
  .home-header {
    .site-info {
      h1 {
        font-size: 2.2rem;
      }

      .site-subtitle {
        font-size: 1.4rem;
      }
    }
  }
}
</style>
