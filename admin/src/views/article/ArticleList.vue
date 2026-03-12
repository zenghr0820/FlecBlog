<template>
  <common-list title="文章列表" :data="articleList" :loading="loading" :total="total" v-model:page="queryParams.page"
    v-model:page-size="queryParams.page_size" create-text="新增文章" @create="handleCreate" @refresh="fetchArticles"
    @update:page="fetchArticles" @update:pageSize="fetchArticles">
    <!-- 额外按钮 -->
    <template #toolbar-after>
      <el-button @click="categoryDialogVisible = true">
        分类管理
      </el-button>
      <el-button @click="tagDialogVisible = true">
        标签管理
      </el-button>
    </template>

    <!-- 额外组件 -->
    <template #extra>
      <category-manager v-model="categoryDialogVisible" />
      <tag-manager v-model="tagDialogVisible" />
    </template>

    <!-- 表格列 - 直接使用 el-table-column -->
    <el-table-column label="封面" width="120" align="center">
      <template #default="{ row }">
        <el-image :src="row.cover" fit="cover" style="width: 80px; height: 50px; border-radius: 4px" />
      </template>
    </el-table-column>

    <el-table-column label="标题" min-width="300">
      <template #default="{ row }">
        <span>{{ row.title }}</span>
        <el-tag v-if="row.is_top" type="primary" size="small" style="margin-left: 8px">置顶</el-tag>
        <el-tag v-if="row.is_essence" type="success" size="small" style="margin-left: 8px">精选</el-tag>
        <el-tag v-if="!row.is_publish" type="warning" size="small" style="margin-left: 8px">草稿</el-tag>
      </template>
    </el-table-column>

    <el-table-column label="分类" width="120" align="center">
      <template #default="{ row }">
        <span v-if="row.category">{{ row.category.name }}</span>
        <span v-else style="color: #999">-</span>
      </template>
    </el-table-column>

    <el-table-column label="标签" width="200" align="center">
      <template #default="{ row }">
        <el-tag v-for="tag in row.tags" :key="tag.id" size="small" type="info" style="margin: 2px">
          {{ tag.name }}
        </el-tag>
        <span v-if="!row.tags?.length" style="color: #999">-</span>
      </template>
    </el-table-column>

    <el-table-column label="发布地点" width="120" align="center">
      <template #default="{ row }">
        <span v-if="row.location">{{ row.location }}</span>
        <span v-else style="color: #999">-</span>
      </template>
    </el-table-column>

    <el-table-column label="统计" width="140" align="center">
      <template #default="{ row }">
        <div style="display: flex; align-items: center; justify-content: center; gap: 12px; font-size: 13px;">
          <div style="display: flex; align-items: center; gap: 4px;">
            <el-icon size="14" style="color: #409eff;">
              <View />
            </el-icon>
            <span>{{ row.view_count || 0 }}</span>
          </div>
          <div style="display: flex; align-items: center; gap: 4px;">
            <el-icon size="14" style="color: #67c23a;">
              <ChatDotRound />
            </el-icon>
            <span>{{ row.comment_count || 0 }}</span>
          </div>
        </div>
      </template>
    </el-table-column>

    <el-table-column label="发布时间" width="180" align="center">
      <template #default="{ row }">
        <div v-if="row.publish_time" style="font-size: 13px; line-height: 1.8;">
          <div style="display: flex; align-items: center; justify-content: center; gap: 4px;">
            <el-icon size="13" style="color: #67c23a;">
              <Upload />
            </el-icon>
            <span>{{ formatDateTime(row.publish_time) }}</span>
          </div>
          <div v-if="row.update_time && row.update_time !== row.publish_time"
            style="display: flex; align-items: center; justify-content: center; gap: 4px;">
            <el-icon size="13" style="color: #409eff;">
              <EditPen />
            </el-icon>
            <span>{{ formatDateTime(row.update_time) }}</span>
          </div>
        </div>
        <span v-else style="color: #999">未发布</span>
      </template>
    </el-table-column>

    <el-table-column label="操作" width="180" align="center" fixed="right">
      <template #default="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row.id)">编辑</el-button>
        <el-button type="success" link size="small" @click="openExportDialog(row.id)">导出</el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row.id)">删除</el-button>
      </template>
    </el-table-column>
  </common-list>

  <!-- 导出弹窗 -->
  <el-dialog v-model="exportDialogVisible" title="导出文章" width="480px" :close-on-click-modal="false">
    <div class="export-options">
      <div v-for="option in exportOptions" :key="option.key" class="export-option" :class="{ disabled: option.loading }"
        @click="handleExport(option.key)">
        <div class="option-icon">
          <i :class="option.icon"></i>
        </div>
        <div class="option-content">
          <div class="option-title">{{ option.title }}</div>
          <div class="option-desc">{{ option.desc }}</div>
        </div>
        <el-icon v-if="option.loading" class="is-loading">
          <Loading />
        </el-icon>
      </div>
    </div>
  </el-dialog>
</template>

<style scoped>
.export-options {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.export-option {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.export-option:hover {
  border-color: #409eff;
  background: #f5f7fa;
}

.export-option.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.option-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f9eb;
  border-radius: 8px;
  font-size: 20px;
  color: #67c23a;
}

.export-option:nth-child(1) .option-icon {
  background: #e6f7e6;
  color: #07c160;
}

.export-option:nth-child(2) .option-icon {
  background: #fef0f0;
  color: #f56c6c;
}

.export-option:nth-child(3) .option-icon {
  background: #ecf5ff;
  color: #409eff;
}

.export-option:nth-child(4) .option-icon {
  background: #f4f4f5;
  color: #909399;
}

.option-content {
  flex: 1;
}

.option-title {
  font-size: 15px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.option-desc {
  font-size: 12px;
  color: #909399;
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { View, ChatDotRound, Upload, EditPen, Loading } from '@element-plus/icons-vue'
import CommonList from '@/components/common/CommonList.vue'
import type { Article } from '@/types/article'
import type { PaginationQuery } from '@/types/request'
import { getArticles, getArticle, deleteArticle, exportToWeChat, getWeChatHTML } from '@/api/article'
import CategoryManager from './components/CategoryManager.vue'
import TagManager from './components/TagManager.vue'
import { formatDateTime } from '@/utils/date'
import { renderMarkdownWithStyles } from '@/utils/markdown'

const router = useRouter()
const loading = ref(false)
const categoryDialogVisible = ref(false)
const tagDialogVisible = ref(false)
const articleList = ref<Article[]>([])
const total = ref(0)
const queryParams = ref<PaginationQuery>({ page: 1, page_size: 20 })

const fetchArticles = async () => {
  loading.value = true
  try {
    const [result] = await Promise.all([
      getArticles(queryParams.value),
      new Promise(resolve => setTimeout(resolve, 300))
    ])
    articleList.value = result.list
    total.value = result.total
  } catch {
    ElMessage.error('获取文章列表失败')
  } finally {
    loading.value = false
  }
}

const handleCreate = () => router.push('/articles/create')
const handleEdit = (id: number) => router.push(`/articles/edit/${id}`)

const handleDelete = async (id: number) => {
  try {
    await ElMessageBox.confirm('确定要删除这篇文章吗？', '提示', { type: 'warning' })
    await deleteArticle(id)
    ElMessage.success('删除成功')
    fetchArticles()
  } catch (error) {
    if (error instanceof Error) ElMessage.error(error.message)
  }
}

// ==================== 导出功能 ====================

const exportDialogVisible = ref(false)
const exportArticleId = ref<number>(0)

const exportOptions = reactive([
  { key: 'wechat', title: '导出到微信公众平台', desc: '直接推送到微信公众号草稿箱', icon: 'ri-wechat-line', loading: false },
  { key: 'wechat-html', title: '复制为公众号图文', desc: '复制适配微信公众号的富文本', icon: 'ri-file-copy-line', loading: false },
  { key: 'html', title: '复制为 HTML', desc: '复制与预览一致的富文本', icon: 'ri-html5-line', loading: false },
  { key: 'markdown', title: '复制为 Markdown', desc: '非原生语法，包含特殊语法标记', icon: 'ri-markdown-line', loading: false }
])

const openExportDialog = (id: number) => {
  exportArticleId.value = id
  exportDialogVisible.value = true
}

const handleExport = async (key: string) => {
  const option = exportOptions.find(o => o.key === key)
  if (!option || option.loading) return

  option.loading = true

  try {
    switch (key) {
      case 'wechat':
        await handleExportToWeChat()
        break
      case 'wechat-html':
        await handleCopyWeChatHTML()
        break
      case 'html':
        await handleCopyHTML()
        break
      case 'markdown':
        await handleCopyMarkdown()
        break
    }
  } finally {
    option.loading = false
  }
}

// 导出到微信公众平台
const handleExportToWeChat = async () => {
  const result = await exportToWeChat(exportArticleId.value)
  ElMessage.success('已导出到微信公众平台草稿箱')
  if (result.warnings?.length) {
    result.warnings.forEach(w => ElMessage.warning(w))
  }
  exportDialogVisible.value = false
}

// 复制为公众号格式（富文本）
const handleCopyWeChatHTML = async () => {
  const result = await getWeChatHTML(exportArticleId.value)
  await copyRichText(result.html)
  ElMessage.success('已复制公众号格式到剪贴板')
  exportDialogVisible.value = false
}

// 复制为 HTML（与预览一致）
const handleCopyHTML = async () => {
  const article = await getArticle(exportArticleId.value)
  const html = renderMarkdownWithStyles(article.content)
  await copyRichText(html)
  ElMessage.success('已复制 HTML 到剪贴板')
  exportDialogVisible.value = false
}

// 复制为 Markdown（保留扩展语法）
const handleCopyMarkdown = async () => {
  const article = await getArticle(exportArticleId.value)
  await copyToClipboard(article.content)
  ElMessage.success('已复制 Markdown 到剪贴板')
  exportDialogVisible.value = false
}

// 复制到剪贴板（纯文本）
const copyToClipboard = async (text: string) => {
  if (navigator.clipboard) {
    await navigator.clipboard.writeText(text)
  } else {
    // 降级方案
    const textarea = document.createElement('textarea')
    textarea.value = text
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }
}

// 复制富文本到剪贴板（HTML 格式）
const copyRichText = async (html: string) => {
  try {
    // 使用 Clipboard API 写入富文本
    const blob = new Blob([html], { type: 'text/html' })
    const clipboardItem = new ClipboardItem({
      'text/html': blob,
      'text/plain': new Blob([html], { type: 'text/plain' })
    })
    await navigator.clipboard.write([clipboardItem])
  } catch {
    // 降级方案：通过临时元素复制
    const container = document.createElement('div')
    container.innerHTML = html
    container.style.position = 'fixed'
    container.style.left = '-9999px'
    container.style.whiteSpace = 'pre-wrap'
    document.body.appendChild(container)

    const range = document.createRange()
    range.selectNodeContents(container)
    const selection = window.getSelection()
    selection?.removeAllRanges()
    selection?.addRange(range)

    document.execCommand('copy')
    selection?.removeAllRanges()
    document.body.removeChild(container)
  }
}

onMounted(fetchArticles)
</script>
