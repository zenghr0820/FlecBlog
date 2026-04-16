<script setup lang="ts">
const { aggregateMenus } = useMenus();

// 获取顶级聚合菜单（已包含子菜单）
const topAggregateMenus = computed(() => {
  return aggregateMenus.value.filter(menu => !menu.parent_id);
});

// 判断 icon 是否为图片 URL
const isImageIcon = (icon: string) => {
  return (
    icon && (icon.startsWith('http://') || icon.startsWith('https://') || icon.startsWith('/'))
  );
};
</script>

<template>
  <div v-if="topAggregateMenus.length > 0" class="nav-aggregate">
    <div class="aggregate-trigger brighten">
      <i class="ri-fingerprint-fill ri-lg"></i>
    </div>

    <!-- 聚合下拉菜单 -->
    <div class="aggregate-dropdown">
      <div class="aggregate-groups-container">
        <div
          v-for="menu in topAggregateMenus"
          :key="menu.id"
          v-show="menu.children && menu.children.length > 0"
          class="aggregate-group"
        >
          <!-- 主菜单标题 -->
          <div class="group-title">
            <span>{{ menu.title }}</span>
          </div>

          <!-- 子菜单列 -->
          <div class="group-children">
            <a
              v-for="child in menu.children"
              :key="child.id"
              :href="child.url"
              :aria-label="child.title"
            >
              <NuxtImg
                v-if="child.icon && isImageIcon(child.icon)"
                :src="child.icon"
                :alt="child.title"
                class="icon-img"
                loading="lazy"
              />
              <i v-else-if="child.icon" :class="child.icon + ' ri-lg'"></i>
              <span>{{ child.title }}</span>
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '@/assets/css/mixins' as *;

.nav-aggregate {
  position: relative;
  margin-right: 8px;

  .aggregate-trigger {
    cursor: pointer;
  }

  .aggregate-dropdown {
    position: absolute;
    top: 45px;
    transform: scale(.2);
    transform-origin: top left;
    left: 0;
    background-color: var(--flec-nav-bg);
    box-shadow: var(--flec-nav-shadow);
    border-radius: 12px;
    // border: var(--flec-card-border);
    flex-direction: column;
    font-size: 12px;
    transition: .5s;
    opacity: 0;
    pointer-events: none;
    


    &::before {
      position: absolute;
      top: -20px;
      left: 0;
      width: 100%;
      height: 30px;
      content: '';
    }

    .aggregate-groups-container{
      padding-bottom: 15px;
      width: 100%;
      max-width: calc(100vw - 40px);
      height: 100%;
      overflow-y: auto;
      max-height: calc(100vh - 80px);
      overflow-x: hidden;
    }

    .aggregate-group {
      display: flex;
      -webkit-box-orient: vertical;
      -moz-box-orient: vertical;
      -o-box-orient: vertical;
      -webkit-flex-direction: column;
      -ms-flex-direction: column;
      flex-direction: column;

      &:last-child {
        margin-bottom: 0;
      }

      &:hover {
        .group-title {
          transform: translateX(8px);
        }
      }

      .group-title {
        padding: 6px 10px;
        color: var(--flec-nav-fixed-font);
        margin: 8px 0 8px 20px;
        font-size: 1.6em;
        font-weight: 800;
        transition: .5s;

      }

      .group-children {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        padding: 0 16px 8px 16px;
        grid-gap: 8px;

        a {
          display: flex;
          align-items: center;
          padding: 8px 10px;
          color: var(--flec-nav-fixed-font);
          font-size: 0.9rem;
          opacity: 0;
          transform: translateY(-5px);
          transition: all 0.2s ease;
          align-items: center;
          width: 150px;
          padding: 6px 10px;
          border-radius: 8px;

          &:hover {
            background: var(--flec-nav-menu-bg-hover);
            border-radius: 8px;
            color: var(--white);

            span {
              color: var(--white);
            }

            i{
              transform: rotate(30deg);
            }
          }

          img {
            width: 20px;
            height: 20px;
            margin-right: 8px;
            border-radius: 4px;
            object-fit: cover;
          }

          i {
            margin-right: 8px;
            transition: all .5s;
            width: 24px;
            height: 24px;
            display: flex;
            align-items: center;
            justify-content: center;
          }

          span {
            transition: all .5s;
            color: var(--flec-nav-fixed-font);
            white-space: nowrap;
            font-size: 16px;
            font-weight: 600;
          }
        }
      }
    }
  }

  &:hover {
    .aggregate-dropdown {
      max-width: calc(100vw - 40px);
      display: flex;
      opacity: 1;
      filter: none;
      transition: .5s;
      top: 50px;
      pointer-events: auto;
      left: 0;
      transform: scale(1);
      // backdrop-filter: blur(10px);

      .group-children a {
        opacity: 1;
        transform: translateY(0);
      }
    }
  }
}

@media screen and (max-width: 768px) {
  .nav-aggregate {
    display: none;
  }
}
</style>
