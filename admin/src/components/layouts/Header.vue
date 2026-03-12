<template>
  <div class="header">
    <div class="left">
      <!-- 移动端：label 触发 checkbox -->
      <label for="sidebar-toggle" class="toggle-sidebar mobile-only">
        <i class="ri-menu-line ri-lg"></i>
      </label>
      <!-- 桌面端：折叠按钮 -->
      <div class="toggle-sidebar desktop-only" @click="handleToggleSidebar">
        <i class="ri-menu-fold-3-line ri-lg" v-if="!sidebarCollapsed"></i>
        <i class="ri-menu-unfold-3-line ri-lg" v-else></i>
      </div>
    </div>
    <div class="right">
      <NotificationBell />
      <el-dropdown trigger="click">
        <span class="user-info">
          <el-avatar :src="userAvatar" />
          <span class="nickname hide-on-mobile">{{ nickname }}</span>
          <el-icon class="arrow-icon"><ArrowDown /></el-icon>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item disabled>
              <el-icon><User /></el-icon>
              <span>{{ nickname }}</span>
            </el-dropdown-item>
            <el-dropdown-item divided @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>
              <span>退出登录</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { User, SwitchButton, ArrowDown } from '@element-plus/icons-vue'
import NotificationBell from '@/components/common/NotificationBell.vue'
import { logout as logoutApi } from '@/api/user'
import { removeTokens } from '@/utils/auth'

const router = useRouter()

const userInfoStr = localStorage.getItem('userInfo')
const userInfo = userInfoStr ? JSON.parse(userInfoStr) : {}
const nickname = ref(userInfo.nickname || 'Admin')
const userAvatar = ref(userInfo.avatar || 'https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png')

// 接收 props
interface Props {
  layoutMode: 'drawer' | 'fixed'
  sidebarCollapsed: boolean
}

defineProps<Props>()

// 定义事件
const emit = defineEmits(['toggle-sidebar'])

const handleToggleSidebar = () => {
  emit('toggle-sidebar')
}

const handleLogout = async () => {
  try {
    await ElMessageBox.confirm('确定要退出登录吗？', '提示', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })

    // 调用后端登出 API，将 token 加入黑名单
    try {
      await logoutApi()
    } catch (error) {
      console.error('登出 API 调用失败:', error)
      // 即使后端 API 失败，也要清除本地 token
    }

    // 清除所有本地存储的认证信息
    removeTokens()
    localStorage.removeItem('userInfo')

    ElMessage.success('已退出登录')
    router.push('/login')
  } catch {}
}
</script>

<style scoped lang="scss">
.header {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;

  // 移动端减小内边距
  @media (max-width: 767px) {
    padding: 0 12px;
  }

  .left {
    display: flex;
    align-items: center;

    .toggle-sidebar {
      margin-right: 20px;
      font-size: 20px;
      cursor: pointer;
      padding: 8px;
      border-radius: 4px;
      transition: background-color 0.3s;

      &:hover {
        background-color: #f5f7fa;
      }

      // 移动端增大触摸区域
      @media (max-width: 767px) {
        margin-right: 12px;
        padding: 10px;
      }
    }

    // 移动端显示/隐藏
    .mobile-only {
      display: none;
      @media (max-width: 768px) {
        display: block;
      }
    }

    .desktop-only {
      display: block;
      @media (max-width: 768px) {
        display: none;
      }
    }
  }

  .right {
    display: flex;
    align-items: center;

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: 4px;
      transition: background-color 0.3s;
      outline: none;

      &:hover {
        background-color: #f5f7fa;
      }

      &:focus {
        outline: none;
      }

      .nickname {
        font-size: 14px;
        color: #303133;
        font-weight: 500;
      }

      .arrow-icon {
        font-size: 12px;
        color: #909399;
      }
    }
  }
}

.hide-on-mobile {
  @media (max-width: 768px) {
    display: none !important;
  }
}
</style>