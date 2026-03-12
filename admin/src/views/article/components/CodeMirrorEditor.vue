<template>
  <div class="codemirror-editor-wrapper" :class="{ 'is-fullscreen': isBrowserFullscreen || isPageFullscreen }">
    <!-- 工具栏 -->
    <div class="editor-toolbar">
      <template v-for="(item, index) in toolbarItems" :key="index">
        <div v-if="item.type === 'divider'" class="toolbar-divider"></div>

        <!-- 弹性空间 -->
        <div v-else-if="item.type === 'spacer'" class="toolbar-spacer"></div>

        <!-- 下载在线图片按钮 -->
        <template v-else-if="item.title === '下载在线图片'">
          <el-popover :width="350" trigger="click" placement="bottom">
            <template #reference>
              <button :title="item.title" class="toolbar-btn">
                <i :class="item.icon"></i>
              </button>
            </template>
            <div style="padding: 8px 0;">
              <el-input v-model="onlineImageUrl" placeholder="输入图片URL，按回车下载" size="small" clearable
                @keyup.enter="handleOnlineImageDownload" style="width: 100%;">
                <template #append>
                  <el-button @click="handleOnlineImageDownload" :loading="downloadingImage"
                    :disabled="!onlineImageUrl.trim()" size="small">
                    下载
                  </el-button>
                </template>
              </el-input>
            </div>
          </el-popover>
        </template>
        <!-- 表情选择器按钮 -->
        <template v-else-if="item.title === '表情'">
          <el-popover :width="320" trigger="click" placement="bottom" v-model:visible="emojiState.visible"
            @show="handleEmojiPickerShow">
            <template #reference>
              <button :title="item.title" class="toolbar-btn" :class="{ active: emojiState.visible }">
                <i :class="item.icon"></i>
              </button>
            </template>
            <!-- 表情内容 -->
            <div class="emoji-wrap">
              <div class="emoji-bar">
                <button v-for="(group, index) in emojiState.groups" :key="index" class="emoji-tab"
                  :class="{ active: emojiState.activeTab === index }" @click="emojiState.activeTab = index">
                  {{ group.name }}
                </button>
              </div>
              <div class="emoji-list">
                <div v-for="(group, index) in emojiState.groups" v-show="emojiState.activeTab === index" :key="index"
                  class="emoji-group" :class="{ 'emoji-text': group.type === 'emoticon' }">
                  <button v-for="item in group.items" :key="item.key" class="emoji-btn" :title="item.key"
                    @click="selectEmoji(item, group.type)">
                    <img v-if="group.type === 'image'" :src="item.val" :alt="item.key" />
                    <span v-else>{{ item.val }}</span>
                  </button>
                </div>
              </div>
            </div>
          </el-popover>
        </template>
        <!-- 普通按钮 -->
        <button v-else @click="item.action" :title="item.title" :class="{
          active: item.isActive?.(),
          'mobile-only': item.mobileOnly
        }" class="toolbar-btn">
          <i v-if="item.icon" :class="item.icon"></i>
          <span v-else>{{ item.label }}</span>
        </button>
      </template>

      <input ref="imageInputRef" type="file" accept="image/*" multiple style="display: none"
        @change="handleImageSelect" />
    </div>

    <!-- 编辑器主体 -->
    <div class="editor-container">
      <!-- 编辑器面板 -->
      <div ref="editorPaneRef" class="editor-pane" :class="{
        'full-width': viewMode === 'editor',
        'hidden': viewMode === 'preview'
      }" @scroll="handleEditorScroll" @mousedown="handleEditorPaneMouseDown">
        <div ref="editorRef" class="cm-host"></div>
      </div>

      <!-- 预览面板 -->
      <div v-show="viewMode !== 'editor'" ref="previewPaneRef" class="preview-pane" :class="{
        'full-width': viewMode === 'preview',
        'html-mode': viewMode === 'html'
      }" @scroll="handlePreviewScroll">
        <div v-if="viewMode === 'html'" class="html-code">
          <pre><code>{{ renderedHtml }}</code></pre>
        </div>
        <div v-else class="markdown-content" v-html="renderedHtml"></div>
      </div>

      <!-- 目录面板 -->
      <div v-if="showToc" class="toc-pane">
        <div class="toc-header">
          <span>目录</span>
          <button @click="showToc = false" class="toc-close">
            <i class="ri-close-line"></i>
          </button>
        </div>
        <div class="toc-content">
          <div v-for="(heading, index) in tableOfContents" :key="index" :class="`toc-item toc-level-${heading.level}`"
            @click="scrollToHeading(heading)">
            {{ heading.text }}
          </div>
          <div v-if="tableOfContents.length === 0" class="toc-empty">
            暂无目录
          </div>
        </div>
      </div>
    </div>

    <!-- 页脚 -->
    <div class="editor-footer">
      <div class="footer-left">
        <span class="word-count">字数：{{ wordCount }}</span>
        <span class="reading-time">阅读时长：{{ readingTime }} 分钟</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, shallowRef, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadFile } from '@/api/file'
import { getSettingGroup } from '@/api/sysconfig'
import {
  renderMarkdown,
  renderMarkdownWithSourceMap,
  countWords,
  extractToc,
  estimateReadingTime,
  type TocItem
} from '@/utils/markdown'
import { EditorView, keymap, showPanel } from '@codemirror/view'
import { EditorState, StateField, StateEffect, RangeSetBuilder } from '@codemirror/state'
import { Decoration } from '@codemirror/view'
import type { Panel, DecorationSet } from '@codemirror/view'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { markdown } from '@codemirror/lang-markdown'
import { SearchCursor } from '@codemirror/search'
import mermaid from 'mermaid'

// 简易搜索功能
const setSearchQuery = StateEffect.define<string>()
const setSearchIndex = StateEffect.define<number>()

const searchStateField = StateField.define<{ matches: { from: number; to: number }[]; idx: number }>({
  create: () => ({ matches: [], idx: 0 }),
  update: (v, tr) => {
    for (const e of tr.effects) {
      if (e.is(setSearchQuery)) {
        if (!e.value) return { matches: [], idx: 0 }
        const matches: { from: number; to: number }[] = []
        const cursor = new SearchCursor(tr.state.doc, e.value, 0, undefined, s => s.toLowerCase())
        while (!cursor.next().done) matches.push({ from: cursor.value.from, to: cursor.value.to })
        return { matches, idx: 0 }
      }
      if (e.is(setSearchIndex)) return { ...v, idx: e.value }
    }
    return v
  }
})

const searchDecorations = StateField.define<DecorationSet>({
  create: () => Decoration.none,
  update: (_, tr) => {
    const { matches, idx } = tr.state.field(searchStateField)
    if (!matches.length) return Decoration.none
    const builder = new RangeSetBuilder<Decoration>()
    matches.forEach((m, i) => builder.add(m.from, m.to, Decoration.mark({ class: i === idx ? 'cm-searchMatch-selected' : 'cm-searchMatch' })))
    return builder.finish()
  },
  provide: f => EditorView.decorations.from(f)
})

let searchPanel: { dom: HTMLElement; show: () => void } | null = null

function createSearchPanel(view: EditorView): Panel {
  const dom = document.createElement('div')
  dom.style.cssText = 'display:none;align-items:center;padding:8px;background:#f5f5f5;border-top:1px solid #ddd'
  dom.innerHTML = `
    <input placeholder="查找..." style="width:180px;padding:4px 8px;border:1px solid #ddd;border-radius:4px;outline:none">
    <span style="margin:0 8px;color:#666;font-size:13px"></span>
    <button style="padding:4px 8px;border:1px solid #ddd;border-radius:4px;background:#fff;cursor:pointer">↑</button>
    <button style="padding:4px 8px;border:1px solid #ddd;border-radius:4px;background:#fff;cursor:pointer;margin-left:4px">↓</button>
    <button style="padding:4px 8px;border:1px solid #ddd;border-radius:4px;background:#fff;cursor:pointer;margin-left:8px">×</button>
  `
  const [input, count, prev, next, close] = [dom.querySelector('input')!, dom.querySelector('span')!, ...dom.querySelectorAll('button')] as [HTMLInputElement, HTMLSpanElement, HTMLButtonElement, HTMLButtonElement, HTMLButtonElement]

  const update = () => {
    const { matches, idx } = view.state.field(searchStateField)
    count.textContent = matches.length ? `${idx + 1}/${matches.length}` : input.value ? '无匹配' : ''
  }

  const search = () => {
    view.dispatch({ effects: setSearchQuery.of(input.value) })
    update()
  }

  const go = (d: number) => {
    const { matches, idx } = view.state.field(searchStateField)
    if (!matches.length) return
    const i = (idx + d + matches.length) % matches.length
    view.dispatch({
      effects: setSearchIndex.of(i),
      selection: { anchor: matches[i]!.from, head: matches[i]!.to },
      scrollIntoView: true
    })
    update()
  }

  input.oninput = search
  input.onkeydown = e => {
    if (e.key === 'Enter') { e.preventDefault(); go(e.shiftKey ? -1 : 1) }
    if (e.key === 'Escape') { view.dispatch({ effects: setSearchQuery.of('') }); input.value = ''; update() }
  }
  prev.onclick = () => go(-1)
  next.onclick = () => go(1)
  close.onclick = () => {
    view.dispatch({ effects: setSearchQuery.of('') })
    input.value = ''
    dom.style.display = 'none'
  }

  searchPanel = { dom, show: () => { dom.style.display = 'flex'; input.focus(); input.select() } }
  return { dom, top: false }
}

function openSearchPanelCustom() {
  searchPanel?.show()
  return true
}

// 类型定义
interface ToolbarItem {
  type?: 'divider' | 'spacer'
  icon?: string
  label?: string
  title?: string
  action?: () => void
  isActive?: () => boolean
  mobileOnly?: boolean
}

interface ScrollNode {
  line: number
  previewTop: number
  editorTop: number
}

type ViewMode = 'split' | 'editor' | 'preview' | 'html'

// 常量
const SCROLL_DURATION = 100

const props = withDefaults(defineProps<{ modelValue: string }>(), { modelValue: '' })
const emit = defineEmits<{ 'update:modelValue': [value: string], 'save': [content: string] }>()

// Refs
const editorRef = ref<HTMLElement>()
const editorPaneRef = ref<HTMLElement>()
const previewPaneRef = ref<HTMLElement>()
const imageInputRef = ref<HTMLInputElement>()
const viewMode = ref<ViewMode>('split')
const isBrowserFullscreen = ref(false)
const isPageFullscreen = ref(false)
const showToc = ref(false)
const onlineImageUrl = ref('')
const downloadingImage = ref(false)

// 表情选择器状态
const emojiState = reactive({
  visible: false,
  groups: [] as Array<{ name: string; type: 'emoji' | 'image' | 'emoticon'; items: Array<{ key: string; val: string }> }>,
  activeTab: 0,
  emojiMap: new Map<string, string>()
})

// 编辑器实例
const editorViewRef = shallowRef<EditorView | null>(null)

// ==================== Mermaid 图表渲染 ====================
const initMermaid = () => {
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'loose'
  })
}

const renderMermaidDiagrams = async () => {
  const preview = previewPaneRef.value
  if (!preview) return

  const elements = preview.querySelectorAll('.mermaid:not(:has(svg))')

  for (const element of elements) {
    try {
      const { svg } = await mermaid.render(`mermaid-${Date.now()}`, element.textContent || '')
      element.innerHTML = svg
    } catch (error) {
      console.error('Mermaid 渲染失败:', error)
    }
  }
}

// ==================== 滚动同步 ====================
let scrollSource: 'editor' | 'preview' | null = null
let sourceResetTimer: ReturnType<typeof setTimeout> | null = null
let cachedNodes: ScrollNode[] | null = null
let currentAnimation: number | null = null

const getEditorScroller = () => editorViewRef.value?.scrollDOM ?? null

const setScrollSource = (source: 'editor' | 'preview') => {
  scrollSource = source
  if (sourceResetTimer) clearTimeout(sourceResetTimer)
  sourceResetTimer = setTimeout(() => {
    scrollSource = null
    sourceResetTimer = null
  }, SCROLL_DURATION + 200)
}

const cancelAnimation = () => {
  if (currentAnimation !== null) {
    cancelAnimationFrame(currentAnimation)
    currentAnimation = null
  }
}

const smoothScroll = (element: HTMLElement, target: number) => {
  cancelAnimation()
  const start = element.scrollTop
  const distance = target - start
  if (Math.abs(distance) < 2) {
    element.scrollTop = target
    return
  }
  const startTime = performance.now()
  const animate = (now: number) => {
    const elapsed = now - startTime
    const progress = Math.min(elapsed / SCROLL_DURATION, 1)
    const eased = 1 - (1 - progress) * (1 - progress)
    element.scrollTop = start + distance * eased
    if (progress < 1) {
      currentAnimation = requestAnimationFrame(animate)
    } else {
      currentAnimation = null
    }
  }
  currentAnimation = requestAnimationFrame(animate)
}

const invalidateScrollCache = () => {
  cachedNodes = null
}

const buildNodeMap = (): ScrollNode[] => {
  if (cachedNodes) return cachedNodes
  const editor = editorViewRef.value
  const preview = previewPaneRef.value
  if (!editor || !preview) return []

  const nodes: ScrollNode[] = []
  const previewStyle = getComputedStyle(preview)
  const previewPaddingTop = parseFloat(previewStyle.paddingTop) || 0

  nodes.push({ line: -1, previewTop: 0, editorTop: 0 })

  const elements = preview.querySelectorAll<HTMLElement>('[data-source-line]')
  elements.forEach((el) => {
    const line = parseInt(el.dataset.sourceLine || '0', 10)
    let previewTop = el.offsetTop
    let parent = el.offsetParent as HTMLElement | null
    while (parent && parent !== preview && preview.contains(parent)) {
      previewTop += parent.offsetTop
      parent = parent.offsetParent as HTMLElement | null
    }
    previewTop = Math.max(0, previewTop - previewPaddingTop)

    let editorTop = 0
    try {
      const docLine = editor.state.doc.line(line + 1)
      const block = editor.lineBlockAt(docLine.from)
      editorTop = block.top
    } catch {
      editorTop = line * 22
    }
    nodes.push({ line, previewTop, editorTop })
  })

  const editorScrollHeight = editor.scrollDOM?.scrollHeight || editor.contentHeight
  const previewScrollHeight = preview.scrollHeight
  const editorClientHeight = editor.scrollDOM?.clientHeight || 0
  const previewClientHeight = preview.clientHeight

  nodes.push({
    line: 999999,
    previewTop: Math.max(0, previewScrollHeight - previewClientHeight),
    editorTop: Math.max(0, editorScrollHeight - editorClientHeight)
  })

  nodes.sort((a, b) => a.line - b.line)
  const uniqueNodes: ScrollNode[] = []
  let lastLine = -999
  for (const node of nodes) {
    if (node.line !== lastLine) {
      uniqueNodes.push(node)
      lastLine = node.line
    }
  }
  cachedNodes = uniqueNodes
  return uniqueNodes
}

const mapEditorToPreview = (editorScrollTop: number, nodes: ScrollNode[]): number => {
  if (nodes.length === 0) return 0
  if (nodes.length === 1) return nodes[0]!.previewTop
  if (editorScrollTop <= 0) return 0

  let i = 0
  while (i < nodes.length - 1 && nodes[i + 1]!.editorTop <= editorScrollTop) i++

  const current = nodes[i]!
  const next = nodes[i + 1]
  if (!next) return current.previewTop

  const editorRange = next.editorTop - current.editorTop
  const previewRange = next.previewTop - current.previewTop
  if (editorRange <= 0) return current.previewTop

  const ratio = Math.max(0, Math.min(1, (editorScrollTop - current.editorTop) / editorRange))
  return current.previewTop + previewRange * ratio
}

const mapPreviewToEditor = (previewScrollTop: number, nodes: ScrollNode[]): number => {
  if (nodes.length === 0) return 0
  if (nodes.length === 1) return nodes[0]!.editorTop
  if (previewScrollTop <= 0) return 0

  let i = 0
  while (i < nodes.length - 1 && nodes[i + 1]!.previewTop <= previewScrollTop) i++

  const current = nodes[i]!
  const next = nodes[i + 1]
  if (!next) return current.editorTop

  const previewRange = next.previewTop - current.previewTop
  const editorRange = next.editorTop - current.editorTop
  if (previewRange <= 0) return current.editorTop

  const ratio = Math.max(0, Math.min(1, (previewScrollTop - current.previewTop) / previewRange))
  return current.editorTop + editorRange * ratio
}

const syncToPreview = () => {
  if (scrollSource === 'preview') return
  const editorScroller = getEditorScroller()
  const preview = previewPaneRef.value
  if (!editorScroller || !preview) return

  const nodes = buildNodeMap()
  if (nodes.length === 0) return

  const targetTop = mapEditorToPreview(editorScroller.scrollTop, nodes)
  setScrollSource('editor')
  smoothScroll(preview, targetTop)
}

const syncToEditor = () => {
  if (scrollSource === 'editor') return
  const editorScroller = getEditorScroller()
  const preview = previewPaneRef.value
  if (!editorScroller || !preview) return

  const nodes = buildNodeMap()
  if (nodes.length === 0) return

  const targetTop = mapPreviewToEditor(preview.scrollTop, nodes)
  setScrollSource('preview')
  smoothScroll(editorScroller, targetTop)
}

let editorScrollPending = false
let previewScrollPending = false

const handleEditorScroll = () => {
  if (viewMode.value !== 'split' || scrollSource === 'preview') return
  if (editorScrollPending) return
  editorScrollPending = true
  requestAnimationFrame(() => {
    editorScrollPending = false
    syncToPreview()
  })
}

const handlePreviewScroll = () => {
  if (viewMode.value !== 'split' || scrollSource === 'editor') return
  if (previewScrollPending) return
  previewScrollPending = true
  requestAnimationFrame(() => {
    previewScrollPending = false
    syncToEditor()
  })
}

const bindScrollEvents = () => {
  const editorScroller = getEditorScroller()
  const preview = previewPaneRef.value
  editorScroller?.addEventListener('scroll', handleEditorScroll, { passive: true })
  preview?.addEventListener('scroll', handlePreviewScroll, { passive: true })
}

const unbindScrollEvents = () => {
  const editorScroller = getEditorScroller()
  const preview = previewPaneRef.value
  editorScroller?.removeEventListener('scroll', handleEditorScroll)
  preview?.removeEventListener('scroll', handlePreviewScroll)
  cancelAnimation()
  if (sourceResetTimer) {
    clearTimeout(sourceResetTimer)
    sourceResetTimer = null
  }
}

// 使用带行号映射的渲染函数（用于滚动同步）
const renderedHtml = computed(() => {
  const html = viewMode.value === 'html'
    ? renderMarkdown(props.modelValue)
    : renderMarkdownWithSourceMap(props.modelValue)

  // 替换表情占位符为 img 标签
  if (emojiState.emojiMap.size > 0) {
    return html.replace(/:([^:\s]+):/g, (match, key) => {
      const url = emojiState.emojiMap.get(key)
      if (url) {
        return `<img src="${url}" alt="${key}" class="emoji-image" title="${key}" />`
      }
      return match
    })
  }

  return html
})

// 计算字数
const wordCount = computed(() => countWords(props.modelValue))

// 计算阅读时长
const readingTime = computed(() => estimateReadingTime(props.modelValue))

// 提取目录
const tableOfContents = computed<TocItem[]>(() => {
  return extractToc(props.modelValue)
})

// ==================== 编辑器操作 ====================

// 保存文章
const saveArticle = () => {
  const content = editorViewRef.value?.state.doc.toString() || '';

  if (!content.trim()) {
    ElMessage.warning('文章内容不能为空');
    return;
  }
  emit('save', content);

  ElMessage.success('文章保存成功');
}

// 插入文本到光标位置
const insertText = (before: string, after = '') => {
  if (!editorViewRef.value) return
  const { from, to } = editorViewRef.value.state.selection.main
  const text = editorViewRef.value.state.doc.sliceString(from, to)

  // 如果有选中文本，用语法包裹；否则只插入语法，光标定位在中间
  editorViewRef.value.dispatch({
    changes: { from, to, insert: `${before}${text}${after}` },
    // 如果有选中文本，保持选中状态；否则光标定位在中间
    selection: text ? { anchor: from + before.length, head: from + before.length + text.length } : { anchor: from + before.length, head: from + before.length }
  })
  editorViewRef.value.focus()
}

// 插入标题
const insertHeading = (level: string) => insertText(`${'#'.repeat(+level)} `)

// 滚动到指定标题
const scrollToHeading = (heading: TocItem) => {
  if (!editorViewRef.value) return
  const lines = editorViewRef.value.state.doc.toString().split('\n')
  const index = lines.findIndex(line => line.includes(heading.text) && line.startsWith('#'))

  if (index !== -1) {
    const pos = editorViewRef.value.state.doc.line(index + 1).from
    editorViewRef.value.dispatch({
      selection: { anchor: pos, head: pos },
      effects: EditorView.scrollIntoView(pos, { y: 'start' })
    })
    editorViewRef.value.focus()
  }
}

// 工具栏配置
const toolbarItems: ToolbarItem[] = [
  // 第一组：基本文本格式
  { icon: 'ri-bold', title: '粗体 (Ctrl+B)', action: () => insertText('**', '**') },
  { icon: 'ri-underline', title: '下划线', action: () => insertText('++', '++') },
  { icon: 'ri-italic', title: '斜体 (Ctrl+I)', action: () => insertText('*', '*') },
  { icon: 'ri-strikethrough', title: '删除线', action: () => insertText('~~', '~~') },
  { type: 'divider' },

  // 第二组：标题
  { label: 'H1', title: '一级标题', action: () => insertHeading('1') },
  { label: 'H2', title: '二级标题', action: () => insertHeading('2') },
  { label: 'H3', title: '三级标题', action: () => insertHeading('3') },
  { label: 'H4', title: '四级标题', action: () => insertHeading('4') },
  { label: 'H5', title: '五级标题', action: () => insertHeading('5') },
  { label: 'H6', title: '六级标题', action: () => insertHeading('6') },
  { type: 'divider' },
  { icon: 'ri-subscript', title: '下标', action: () => insertText('~', '~') },
  { icon: 'ri-superscript', title: '上标', action: () => insertText('^', '^') },
  { icon: 'ri-double-quotes-l', title: '引用', action: () => insertText('> ') },
  { icon: 'ri-list-unordered', title: '无序列表', action: () => insertText('- ') },
  { icon: 'ri-list-ordered', title: '有序列表', action: () => insertText('1. ') },
  { icon: 'ri-list-check', title: '任务列表', action: () => insertText('- [ ] ') },
  { type: 'divider' },

  // 第三组：代码和插入元素
  { icon: 'ri-code-line', title: '行内代码', action: () => insertText('`', '`') },
  { icon: 'ri-code-box-line', title: '块级代码', action: () => insertText('\n```', '\n```\n') },
  { icon: 'ri-link', title: '链接', action: () => insertText('[', '](https://)') },
  { icon: 'ri-image-add-line', title: '上传本地图片', action: () => imageInputRef.value?.click() },
  { icon: 'ri-image-download-line', title: '下载在线图片', action: () => { } },
  { icon: 'ri-emotion-line', title: '表情', action: () => toggleEmojiPicker() },
  { icon: 'ri-table-2', title: '表格', action: () => insertText('\n| 列1 | 列2 | 列3 |\n|:---:|:---:|:---:|\n|  ', '  |    |    |\n') },
  { icon: 'ri-mark-pen-line', title: '高亮', action: () => insertText('==', '==') },
  { type: 'divider' },

  // 第四组：自定义块
  { icon: 'ri-information-line', title: '提示框', action: () => insertText('\n:::note info 提示标题\ninfo/warning/success/error', '\n:::endnote\n') },
  { icon: 'ri-layout-grid-line', title: '标签页', action: () => insertText('\n:::tabs\n:::tab 标签1\n内容1', '\n:::endtab\n:::tab 标签2\n内容2\n:::endtab\n:::endtabs\n') },
  { icon: 'ri-increase-decrease-line', title: '折叠面板', action: () => insertText('\n:::fold 点击展开\n', '\n:::endfold\n') },
  { icon: 'ri-external-link-line', title: '链接卡片', action: () => insertText('\n:::link 标题', ' https://example.com 描述信息 :::\n') },
  { icon: 'ri-multi-image-line', title: '照片墙', action: () => insertText('\n:::photo\n图片1\n图片2\n:::n\n图片3\n图片4\n:::endphoto\n') },

  // 弹性空间，将后续按钮推到右侧
  { type: 'spacer' },

  // 第五组：视图控制（右侧）
  {
    icon: 'ri-fullscreen-line',
    title: '浏览器全屏',
    action: () => document.fullscreenElement ? document.exitFullscreen() : document.documentElement.requestFullscreen(),
    isActive: () => isBrowserFullscreen.value
  },
  {
    icon: 'ri-picture-in-picture-2-line',
    title: '页面全屏',
    action: () => isPageFullscreen.value = !isPageFullscreen.value,
    isActive: () => isPageFullscreen.value
  },
  {
    icon: 'ri-code-s-slash-line',
    title: 'HTML 代码预览',
    action: () => viewMode.value = viewMode.value === 'html' ? 'split' : 'html',
    isActive: () => viewMode.value === 'html'
  },
  {
    icon: 'ri-eye-line',
    title: '切换预览',
    action: () => viewMode.value = viewMode.value === 'preview' ? 'editor' : 'preview',
    isActive: () => viewMode.value === 'preview',
    mobileOnly: true
  },
  {
    icon: 'ri-list-unordered',
    title: '目录',
    action: () => showToc.value = !showToc.value,
    isActive: () => showToc.value
  },
]

// ==================== 图片上传 ====================
const handleImageSelect = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || []).filter(file => {
    if (!file.type.startsWith('image/')) {
      ElMessage.error(`${file.name} 不是图片格式`)
      return false
    }
    return true
  })

  if (!files.length) return

  const loading = ElMessage.info({ message: `正在上传 ${files.length} 张图片...`, duration: 0 })
  try {
    const results = await Promise.all(files.map(f => uploadFile(f, '文章图片')))
    insertText(results.map(r => `![图片](${r.file_url})`).join('\n'))
    ElMessage.success(`成功上传 ${files.length} 张图片`)
  } catch (error: any) {
    ElMessage.error(error.message || '图片上传失败')
  } finally {
    loading.close()
    input.value = ''
  }
}

// 处理粘贴图片
const handlePasteImage = async (files: File[]) => {
  const imageFiles = files.filter(file => {
    if (!file.type.startsWith('image/')) {
      ElMessage.error(`${file.name} 不是图片格式`)
      return false
    }
    return true
  })

  if (!imageFiles.length) return

  const loading = ElMessage.info({ message: `正在上传 ${imageFiles.length} 张图片...`, duration: 0 })
  try {
    const results = await Promise.all(imageFiles.map(f => uploadFile(f, '文章图片')))
    insertText(results.map(r => `![图片](${r.file_url})`).join('\n'))
    ElMessage.success(`成功上传 ${imageFiles.length} 张图片`)
  } catch (error: any) {
    ElMessage.error(error.message || '图片上传失败')
  } finally {
    loading.close()
  }
}

// 处理下载在线图片
const handleOnlineImageDownload = async () => {
  if (!onlineImageUrl.value.trim()) {
    ElMessage.error('请输入图片URL')
    return
  }

  const url = onlineImageUrl.value.trim()

  // 验证URL格式
  if (!url.match(/^https?:\/\/.+/)) {
    ElMessage.error('请输入有效的HTTP/HTTPS图片URL')
    return
  }

  downloadingImage.value = true
  try {
    // 导入下载图片API
    const { downloadImage } = await import('@/api/tools')

    // 下载图片
    const downloadResult = await downloadImage({ url })

    // 将base64数据转换为Blob
    const base64Data = downloadResult.data
    const byteCharacters = atob(base64Data)
    const byteNumbers = new Array(byteCharacters.length)
    for (let i = 0; i < byteCharacters.length; i++) {
      byteNumbers[i] = byteCharacters.charCodeAt(i)
    }
    const byteArray = new Uint8Array(byteNumbers)
    const blob = new Blob([byteArray], { type: downloadResult.content_type })

    // 创建文件对象并上传
    const file = new File([blob], 'image.jpg', { type: downloadResult.content_type })
    const uploadResult = await uploadFile(file, '文章图片')

    // 插入到编辑器
    insertText(`![图片](${uploadResult.file_url})`)

    // 清空输入
    onlineImageUrl.value = ''

    ElMessage.success('图片下载并上传成功')

    // 关闭 Popover
    document.body.click()
  } catch (error: any) {
    ElMessage.error(error.message || '图片下载失败')
  } finally {
    downloadingImage.value = false
  }
}

// 表情选择器
const loadEmojis = async () => {
  if (emojiState.groups.length) return

  const blogSettings = await getSettingGroup('blog')
  const emojisUrl = blogSettings.emojis || blogSettings['blog.emojis'] || ''
  if (!emojisUrl) return

  const response = await fetch(emojisUrl)
  const groups = await response.json()
  emojiState.groups = groups

  // 构建 image 类型表情映射
  for (const group of groups) {
    if (group.type === 'image') {
      for (const item of group.items) {
        emojiState.emojiMap.set(item.key, item.val)
      }
    }
  }
}

const selectEmoji = (item: { key: string; val: string }, type: string) => {
  const emoji = type === 'image' ? `:${item.key}:` : item.val
  insertText(emoji)
  emojiState.visible = false
}

// 表情选择器显示时加载数据
const handleEmojiPickerShow = () => {
  if (!emojiState.groups.length) {
    loadEmojis()
  }
}

const toggleEmojiPicker = () => {
  emojiState.visible = !emojiState.visible
  if (emojiState.visible && !emojiState.groups.length) {
    loadEmojis()
  }
}

// ==================== 编辑器初始化 ====================
const initEditor = () => {
  if (!editorRef.value) return

  // 创建粘贴事件处理器
  const pasteHandler = EditorView.domEventHandlers({
    paste: (event: ClipboardEvent, view) => {
      // 先检查是否有图片
      const items = event.clipboardData?.items
      if (items) {
        const files: File[] = []
        const textItems: DataTransferItem[] = []

        for (let i = 0; i < items.length; i++) {
          const item = items[i]
          if (item && item.type) {
            if (item.type.startsWith('image/')) {
              const file = item.getAsFile()
              if (file) {
                files.push(file)
              }
            } else if (item.kind === 'string' && item.type === 'text/plain') {
              textItems.push(item)
            }
          }
        }

        // 如果有图片，处理图片上传
        if (files.length > 0) {
          event.preventDefault()
          handlePasteImage(files)

          // 如果还有文本，在图片处理完后再处理文本
          if (textItems.length > 0) {
            textItems.forEach(item => {
              item.getAsString((text) => {
                // 使用默认的粘贴行为来正确替换选中文本
                view.dispatch({
                  changes: {
                    from: view.state.selection.main.from,
                    to: view.state.selection.main.to,
                    insert: text
                  }
                })
              })
            })
          }
          return
        }
      }

      // 如果没有图片，让默认行为处理（这样能正确替换选中文本）
      // 不调用 event.preventDefault()
    }
  })

  editorViewRef.value = new EditorView({
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        history(),
        markdown(),
        searchStateField,
        searchDecorations,
        showPanel.of(createSearchPanel),
        keymap.of([
          { key: 'Mod-b', run: () => (insertText('**', '**'), true), preventDefault: true },
          { key: 'Mod-i', run: () => (insertText('*', '*'), true), preventDefault: true },
          { key: 'Mod-s', run: () => { saveArticle(); return true; }, preventDefault: true },
          { key: 'Mod-f', run: openSearchPanelCustom, preventDefault: true },
          ...defaultKeymap,
          ...historyKeymap
        ]),
        EditorView.updateListener.of(update => {
          if (update.docChanged) {
            emit('update:modelValue', update.state.doc.toString())
            invalidateScrollCache()
          }
        }),
        EditorView.lineWrapping,
        pasteHandler
      ]
    }),
    parent: editorRef.value
  })

  // 编辑器初始化完成后，绑定滚动同步事件
  nextTick(() => {
    bindScrollEvents()
  })
}

// 监听外部内容变化
watch(() => props.modelValue, (newValue) => {
  if (editorViewRef.value && newValue !== editorViewRef.value.state.doc.toString()) {
    editorViewRef.value.dispatch({
      changes: { from: 0, to: editorViewRef.value.state.doc.length, insert: newValue }
    })
    invalidateScrollCache()
  }
})

// 监听预览区图片加载完成，使缓存失效
watch(renderedHtml, async () => {
  await nextTick()
  const preview = previewPaneRef.value
  if (!preview) return

  const images = preview.querySelectorAll('img')
  images.forEach((img) => {
    if (img.complete) return
    img.addEventListener('load', () => invalidateScrollCache(), { once: true })
  })
  invalidateScrollCache()

  // 渲染 Mermaid 图表
  await renderMermaidDiagrams()
})

// 监听视图模式变化
watch(viewMode, (newMode) => {
  if (newMode === 'split') {
    nextTick(() => {
      invalidateScrollCache()
      bindScrollEvents()
    })
  } else {
    unbindScrollEvents()
  }

  // 切换到非编辑模式时加载表情数据
  loadEmojis()
})

// ==================== 生命周期 ====================
const handleFullscreenChange = () => isBrowserFullscreen.value = !!document.fullscreenElement

const handleEditorPaneMouseDown = (event: MouseEvent) => {
  if (event.button !== 0) return
  if (!editorViewRef.value) return

  const target = event.target as HTMLElement | null
  // 点击发生在 Codemirror 内部时，让 Codemirror 自己处理（避免影响选择/光标）
  if (target?.closest('.cm-editor')) return

  // 空白处点击：把焦点交给编辑器
  editorViewRef.value.focus()
}

onMounted(() => {
  initMermaid()
  initEditor()
  loadEmojis()
  document.addEventListener('fullscreenchange', handleFullscreenChange)

  // 移动端默认为纯编辑模式
  if (window.innerWidth <= 768) {
    viewMode.value = 'editor'
  }
})

onBeforeUnmount(() => {
  // 解绑滚动同步事件
  unbindScrollEvents()
  // 销毁编辑器实例
  editorViewRef.value?.destroy()
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style lang="scss">
// 引入 Markdown 内容排版样式
@use '@/assets/css/prose';

// 引入代码高亮样式
@import 'highlight.js/styles/github.css';

// 搜索高亮样式
.cm-searchMatch {
  background-color: #ffeb3b80;
  border-radius: 2px;
}

.cm-searchMatch-selected {
  background-color: #ff9800;
  color: white;
}
</style>

<style scoped lang="scss">
.codemirror-editor-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;

  &.is-fullscreen {
    position: fixed;
    inset: 0;
    width: 100vw !important;
    height: 100vh !important;
    z-index: 9999;
    border-radius: 0;
  }

  .editor-toolbar {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    background: #f5f7fa;
    border-bottom: 1px solid #e4e7ed;
    flex-wrap: wrap;

    .toolbar-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      min-width: 28px;
      height: 28px;
      padding: 0 6px;
      background: transparent;
      border: none;
      border-radius: 4px;
      color: #606266;
      cursor: pointer;
      font-size: 13px;
      font-weight: 600;
      transition: all 0.2s;

      i {
        font-size: 15px;
      }

      &:hover {
        background: #e4e7ed;
        color: #409eff;
      }

      &.active {
        background: #409eff;
        color: #fff;
      }

      &.mobile-only {
        display: none;

        @media (max-width: 768px) {
          display: flex;
        }
      }
    }

    .toolbar-divider {
      width: 1px;
      height: 16px;
      background: #dcdfe6;
      margin: 0 4px;
    }

    .toolbar-spacer {
      flex: 1;
      min-width: 12px;
    }
  }

  .editor-container {
    flex: 1;
    display: flex;
    position: relative;
    overflow: hidden;

    .editor-pane {
      flex: 1;
      overflow: auto;
      border-right: 1px solid #e4e7ed;
      cursor: text;
      display: flex;
      flex-direction: column;
      min-height: 0;

      &.full-width {
        border-right: none;
      }

      &.hidden {
        display: none;
      }

      .cm-host {
        flex: 1;
        min-height: 0;
        display: flex;
      }

      :deep(.cm-editor) {
        width: 100%;
        flex: 1;
        min-height: 0;
        display: flex;
        flex-direction: column;
        font-size: 14px;
        font-family: Consolas, Monaco, monospace;

        &.cm-focused {
          outline: none;
        }

        .cm-content {
          padding: 16px;
          min-height: 100%;
          box-sizing: border-box;
        }

        .cm-line {
          line-height: 1.6;
        }

        .cm-cursor {
          border-left-color: #409eff;
        }

        .cm-selectionBackground {
          background: #409eff33 !important;
        }

        .cm-activeLine {
          background: #f5f7fa;
        }

        .cm-gutters {
          background: #fafafa;
          border-right: 1px solid #e4e7ed;
        }
      }
    }

    .preview-pane {
      flex: 1;
      overflow: auto;
      padding: 20px;

      &.html-mode {
        padding: 0;
        background: #282c34;

        pre {
          margin: 0;
          padding: 20px;
          height: 100%;

          code {
            color: #abb2bf;
            font-family: Consolas, Monaco, monospace;
            font-size: 14px;
            line-height: 1.6;
            white-space: pre-wrap;
            word-break: break-all;
          }
        }
      }

      // Mermaid 图表样式
      :deep(.markdown-content) {
        .mermaid {
          display: flex;
          justify-content: center;
          align-items: center;
          margin: 1.5rem 0;
          padding: 1rem;
          background: #f5f7fa;
          border-radius: 8px;
          overflow-x: auto;

          svg {
            max-width: 100%;
            height: auto;
          }
        }

        .mermaid-error {
          color: #f56c6c;
          padding: 1rem;
          background: #fef0f0;
          border-radius: 4px;
          border-left: 4px solid #f56c6c;
        }
      }
    }

    .toc-pane {
      position: absolute;
      right: 0;
      top: 0;
      bottom: 0;
      width: 260px;
      background: #fff;
      border-left: 1px solid #e4e7ed;
      display: flex;
      flex-direction: column;
      box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
      z-index: 10;

      .toc-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 12px 16px;
        border-bottom: 1px solid #e4e7ed;
        background: #f5f7fa;
        font-weight: 600;
        font-size: 14px;
        color: #303133;

        .toc-close {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 24px;
          height: 24px;
          border: none;
          border-radius: 4px;
          background: transparent;
          color: #909399;
          cursor: pointer;
          transition: all 0.2s;

          &:hover {
            background: #e4e7ed;
            color: #606266;
          }

          i {
            font-size: 18px;
          }
        }
      }

      .toc-content {
        flex: 1;
        overflow: auto;
        padding: 12px 0;

        .toc-item {
          padding: 8px 16px;
          cursor: pointer;
          font-size: 14px;
          line-height: 1.5;
          color: #606266;
          border-left: 3px solid transparent;
          transition: all 0.2s;

          &:hover {
            background: #f5f7fa;
            color: #409eff;
            border-left-color: #409eff;
          }

          @for $i from 1 through 6 {
            &.toc-level-#{$i} {
              padding-left: 16px + ($i - 1) * 12px;

              @if $i ==1 {
                font-weight: 600;
              }
            }
          }
        }

        .toc-empty {
          padding: 40px 16px;
          text-align: center;
          color: #909399;
          font-size: 14px;
        }
      }
    }
  }

  .editor-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 12px;
    background: #fafafa;
    border-top: 1px solid #e4e7ed;
    font-size: 12px;
    color: #909399;

    .footer-left {
      display: flex;
      align-items: center;
      gap: 16px;
    }

    .word-count,
    .reading-time {
      user-select: none;
    }
  }

  // 移动端适配
  @media (max-width: 767px) {
    .editor-toolbar {
      padding: 4px 8px;
      gap: 2px;
      overflow-x: auto;
      flex-wrap: nowrap;
      -webkit-overflow-scrolling: touch;

      &::-webkit-scrollbar {
        height: 3px;
      }

      .toolbar-btn {
        min-width: 32px;
        height: 32px;
        flex-shrink: 0;
      }

      .toolbar-divider {
        flex-shrink: 0;
      }
    }

    .editor-container {
      .editor-pane {
        :deep(.cm-editor) {
          .cm-content {
            padding: 12px;
          }
        }
      }

      .preview-pane {
        padding: 12px;
      }

      .toc-pane {
        width: 100%;
        max-width: 280px;
      }
    }

    .editor-footer {
      padding: 4px 8px;
      font-size: 11px;

      .footer-left {
        gap: 8px;
      }
    }
  }

  // 平板端适配
  @media (min-width: 768px) and (max-width: 991px) {
    .editor-toolbar {
      padding: 4px 10px;
      gap: 3px;
    }

    .editor-container {
      .toc-pane {
        width: 240px;
      }
    }
  }
}

// 表情选择器样式
.emoji-tip {
  padding: 40px 20px;
  text-align: center;
  color: #909399;
  font-size: 0.85rem;
}

.emoji-wrap {
  display: flex;
  flex-direction: column;
  height: 200px;
}

.emoji-bar {
  display: flex;
  border-bottom: 1px solid #eee;
  flex-shrink: 0;
}

.emoji-tab {
  flex: 1;
  padding: 8px 4px;
  border: none;
  background: transparent;
  color: #666;
  font-size: 0.75rem;
  cursor: pointer;

  &:hover {
    background: #f5f5f5;
  }

  &.active {
    color: #409eff;
  }
}

.emoji-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;

  &::-webkit-scrollbar {
    width: 0;
  }
}

.emoji-group {
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 2px;

  &.emoji-text {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }
}

.emoji-btn {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: none;
  background: transparent;
  border-radius: 4px;
  cursor: pointer;
  padding: 2px;
  overflow: hidden;

  span {
    font-size: 1.4rem;
  }

  img {
    width: 32px;
    height: 32px;
  }

  &:hover {
    background: #f0f0f0;
  }

  .emoji-text & {
    width: auto;
    height: auto;
    padding: 6px 10px;

    span {
      font-size: 0.85rem;
      white-space: nowrap;
      overflow: hidden;
      max-width: 100%;
    }
  }
}
</style>
