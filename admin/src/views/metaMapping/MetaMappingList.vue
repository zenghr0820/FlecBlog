<template>
  <common-list
    title="Meta映射"
    :data="currentMappings"
    :loading="loading"
    :show-pagination="false"
    create-text="新增映射"
    @create="handleCreate"
    @refresh="reloadAll"
    row-key="id"
  >
    <template #toolbar-before>
      <el-segmented
        v-model="selectedTemplate"
        :options="templateOptions"
        :disabled="templateOptions.length === 0"
        @change="handlePlatformChange"
      />
    </template>

    <template #toolbar-after >
      <el-button type="primary" @click="openTemplateDialog('add')">新建模版</el-button>
      <el-button
        v-if="selectedTemplate"
        type="primary"
        plain
        @click="openTemplateDialog('edit')"
      >
        编辑模版
      </el-button>
      <el-button
        v-if="selectedTemplate"
        type="danger"
        plain
        @click="handleDeleteTemplate"
      >
        删除模版
      </el-button>
    </template>

    <el-table-column prop="source_field" label="源字段" min-width="180" align="center" />
    <el-table-column prop="target_field" label="目标字段" min-width="180" align="center" />

    <el-table-column prop="field_type" label="类型" width="120" align="center">
      <template #default="{ row }">
        <el-tag size="small" :type="getFieldTypeTagType(row.field_type)">
          {{ getFieldTypeLabel(row.field_type) }}
        </el-tag>
      </template>
    </el-table-column>

    <el-table-column label="转换规则" min-width="260" align="center">
      <template #default="{ row }">
        <span v-if="row.transform_rule" class="rule-text">{{ row.transform_rule }}</span>
        <span v-else class="no-rule">-</span>
      </template>
    </el-table-column>

    <el-table-column prop="sort_order" label="排序" width="90" align="center" />

    <el-table-column label="状态" width="110" align="center">
      <template #default="{ row }">
        <el-switch
          :model-value="row.is_active"
          size="small"
          @change="(val) => handleToggle(row, Boolean(val))"
        />
      </template>
    </el-table-column>

    <el-table-column label="操作" width="160" align="center" fixed="right">
      <template #default="{ row }">
        <el-button type="primary" link size="small" @click="handleEdit(row)">
          编辑
        </el-button>
        <el-button type="danger" link size="small" @click="handleDelete(row)">
          删除
        </el-button>
      </template>
    </el-table-column>

    <template #extra>
      <div v-if="selectedTemplateInfo" class="template-hint">
        <el-text type="info">当前模版：{{ selectedTemplateInfo.template_name }}</el-text>
        <el-divider direction="vertical" />
        <el-text type="info">映射字段数：{{ currentMappings.length }} 个</el-text>
      </div>

      <el-dialog
        v-model="templateDialogVisible"
        :title="templateDialogMode === 'add' ? '新建映射模版' : '编辑映射模版'"
        width="560px"
      >
        <el-form
          ref="templateFormRef"
          :model="templateForm"
          :rules="templateRules"
          label-width="110px"
        >
          <el-form-item
            v-if="templateDialogMode === 'add'"
            label="模版 Key"
            prop="template_key"
          >
            <el-input
              v-model="templateForm.template_key"
              placeholder="例如：my_legacy_blog（建议字母/数字/下划线）"
            />
          </el-form-item>
          <el-form-item label="模版名称" prop="template_name">
            <el-input
              v-model="templateForm.template_name"
              placeholder="例如：旧站点导入模版"
            />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="templateForm.description" type="textarea" :rows="3" />
          </el-form-item>
        </el-form>

        <template #footer>
          <span class="dialog-footer">
            <el-button @click="templateDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="templateSaving" @click="saveTemplate">
              确定
            </el-button>
          </span>
        </template>
      </el-dialog>

      <el-dialog
        v-model="mappingDialogVisible"
        :title="mappingDialogMode === 'add' ? '新增映射' : '编辑映射'"
        width="640px"
      >
        <el-form
          ref="mappingFormRef"
          :model="mappingForm"
          :rules="mappingRules"
          label-width="100px"
        >
          <el-form-item label="模版" v-if="mappingDialogMode === 'add'">
            <el-input :model-value="selectedTemplateInfo?.template_name || ''" disabled />
          </el-form-item>

          <el-form-item label="源字段" prop="source_field">
            <el-input v-model="mappingForm.source_field" />
          </el-form-item>

          <el-form-item label="目标字段" prop="target_field">
            <el-select v-model="mappingForm.target_field" style="width: 100%">
              <el-option label="标题 (title)" value="title" />
              <el-option label="别名 (slug)" value="slug" />
              <el-option label="内容 (content)" value="content" />
              <el-option label="摘要 (summary)" value="summary" />
              <el-option label="封面 (cover)" value="cover" />
              <el-option label="分类 (category)" value="category" />
              <el-option label="标签 (tags)" value="tags" />
              <el-option label="发布时间 (publish_time)" value="publish_time" />
              <el-option label="更新时间 (update_time)" value="update_time" />
              <el-option label="是否发布 (is_publish)" value="is_publish" />
              <el-option label="是否置顶 (is_top)" value="is_top" />
              <el-option label="是否精选 (is_essence)" value="is_essence" />
              <el-option label="是否过时 (is_outdated)" value="is_outdated" />
              <el-option label="发布地点 (location)" value="location" />
            </el-select>
          </el-form-item>

          <el-form-item label="字段类型" prop="field_type">
            <el-select v-model="mappingForm.field_type" style="width: 100%">
              <el-option label="字符串 (string)" value="string" />
              <el-option label="布尔值 (boolean)" value="boolean" />
              <el-option label="日期 (date)" value="date" />
              <el-option label="数组 (array)" value="array" />
            </el-select>
          </el-form-item>

          <el-form-item label="转换规则">
            <el-input v-model="mappingForm.transform_rule" type="textarea" :rows="3" />
          </el-form-item>

          <el-form-item label="排序" prop="sort_order">
            <el-input-number v-model="mappingForm.sort_order" :min="0" />
          </el-form-item>
        </el-form>

        <template #footer>
          <span class="dialog-footer">
            <el-button @click="mappingDialogVisible = false">取消</el-button>
            <el-button type="primary" :loading="saving" @click="saveMapping">
              确定
            </el-button>
          </span>
        </template>
      </el-dialog>
    </template>
  </common-list>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import CommonList from '@/components/common/CommonList.vue'
import {
  createMapping,
  createTemplate,
  deleteTemplate,
  deleteMapping,
  getMappingsByTemplate,
  getTemplates,
  toggleMappingStatus,
  updateTemplate,
  updateMapping
} from '@/api/metaMapping'

type TemplateItem = {
  id: number
  template_key: string
  template_name: string
  description: string
  mapping_count: number
}

type MappingItem = {
  id: number
  source_field: string
  target_field: string
  field_type: string
  transform_rule: string
  is_active: boolean
  is_system: boolean
  sort_order: number
}

const loading = ref(false)
const saving = ref(false)
const templates = ref<TemplateItem[]>([])
const selectedTemplate = ref<string>('')
const mappings = ref<Record<string, MappingItem[]>>({})

const templateOptions = computed(() =>
  templates.value.map((t) => ({ label: t.template_name, value: t.template_key }))
)

const selectedTemplateInfo = computed(
  () => templates.value.find((t) => t.template_key === selectedTemplate.value) || null
)

const currentMappings = computed(() => {
  const list = mappings.value[selectedTemplate.value] || []
  return [...list].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
})

const templateDialogVisible = ref(false)
const templateDialogMode = ref<'add' | 'edit'>('add')
const templateFormRef = ref()
const templateSaving = ref(false)
const templateForm = ref({
  id: 0,
  template_key: '',
  template_name: '',
  description: ''
})
const templateRules = {
  template_key: [{ required: true, message: '请输入模版 Key', trigger: 'blur' }],
  template_name: [{ required: true, message: '请输入模版名称', trigger: 'blur' }]
}

const mappingDialogVisible = ref(false)
const mappingDialogMode = ref<'add' | 'edit'>('add')
const mappingFormRef = ref()
const mappingForm = ref({
  id: 0,
  source_field: '',
  target_field: '',
  field_type: 'string',
  transform_rule: '',
  sort_order: 0
})
const mappingRules = {
  source_field: [{ required: true, message: '请输入源字段', trigger: 'blur' }],
  target_field: [{ required: true, message: '请选择目标字段', trigger: 'change' }],
  field_type: [{ required: true, message: '请选择字段类型', trigger: 'change' }]
}

const reloadTemplates = async () => {
  const data = (await getTemplates()) as TemplateItem[]
  templates.value = Array.isArray(data) ? data : []
  if (!selectedTemplate.value && templates.value.length > 0) {
    selectedTemplate.value = templates.value[0].template_key
  }
}

const reloadMappings = async (templateKey: string) => {
  if (!templateKey) return
  const data = (await getMappingsByTemplate(templateKey)) as {
    template_key: string
    template_name: string
    mappings: MappingItem[]
  }
  mappings.value = {
    ...mappings.value,
    [templateKey]: Array.isArray(data?.mappings) ? data.mappings : []
  }
}

const reloadAll = async () => {
  loading.value = true
  try {
    await reloadTemplates()
    if (selectedTemplate.value) {
      await reloadMappings(selectedTemplate.value)
    }
  } finally {
    loading.value = false
  }
}

const handlePlatformChange = async () => {
  if (!selectedTemplate.value) return
  await reloadMappings(selectedTemplate.value)
}

const openTemplateDialog = (mode: 'add' | 'edit') => {
  templateDialogMode.value = mode
  if (mode === 'add') {
    templateForm.value = { id: 0, template_key: '', template_name: '', description: '' }
  } else {
    const t = selectedTemplateInfo.value
    if (!t) return
    templateForm.value = {
      id: t.id,
      template_key: t.template_key,
      template_name: t.template_name,
      description: t.description || ''
    }
  }
  templateDialogVisible.value = true
}

const saveTemplate = async () => {
  if (!templateFormRef.value) return
  await templateFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    templateSaving.value = true
    try {
      if (templateDialogMode.value === 'add') {
        await createTemplate({
          template_key: templateForm.value.template_key,
          template_name: templateForm.value.template_name,
          description: templateForm.value.description
        })
        ElMessage.success('模版创建成功')
        templateDialogVisible.value = false
        await reloadTemplates()
        selectedTemplate.value = templateForm.value.template_key
        await reloadMappings(selectedTemplate.value)
      } else {
        await updateTemplate(templateForm.value.id, {
          template_name: templateForm.value.template_name,
          description: templateForm.value.description
        })
        ElMessage.success('模版更新成功')
        templateDialogVisible.value = false
        await reloadTemplates()
      }
    } catch (e) {
      ElMessage.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      templateSaving.value = false
    }
  })
}

const handleDeleteTemplate = async () => {
  const t = selectedTemplateInfo.value
  if (!t) return
  try {
    await ElMessageBox.confirm(
      `确定要删除模版“${t.template_name}”吗？该模版下的全部映射字段也会一并删除。`,
      '提示',
      { type: 'warning' }
    )
    await deleteTemplate(t.id)
    ElMessage.success('模版删除成功')
    selectedTemplate.value = ''
    mappings.value = {}
    await reloadAll()
  } catch {
    // ignore
  }
}

const handleCreate = () => {
  if (!selectedTemplate.value) {
    ElMessage.warning('请先选择模版')
    return
  }
  mappingDialogMode.value = 'add'
  mappingForm.value = {
    id: 0,
    source_field: '',
    target_field: '',
    field_type: 'string',
    transform_rule: '',
    sort_order: currentMappings.value.length
  }
  mappingDialogVisible.value = true
}

const handleEdit = (row: MappingItem) => {
  mappingDialogMode.value = 'edit'
  mappingForm.value = {
    id: row.id,
    source_field: row.source_field,
    target_field: row.target_field,
    field_type: row.field_type,
    transform_rule: row.transform_rule || '',
    sort_order: row.sort_order ?? 0
  }
  mappingDialogVisible.value = true
}

const handleDelete = async (row: MappingItem) => {
  try {
    await ElMessageBox.confirm('确定要删除该映射吗？', '提示', { type: 'warning' })
    await deleteMapping(row.id)
    ElMessage.success('删除成功')
    await reloadMappings(selectedTemplate.value)
  } catch {
    // ignore
  }
}

const handleToggle = async (row: MappingItem, isActive: boolean) => {
  try {
    await toggleMappingStatus(row.id)
    ElMessage.success(isActive ? '已启用' : '已禁用')
    await reloadMappings(selectedTemplate.value)
  } catch (e) {
    row.is_active = !isActive
    ElMessage.error(e instanceof Error ? e.message : '操作失败')
  }
}

const saveMapping = async () => {
  if (!mappingFormRef.value) return
  if (!selectedTemplate.value) return
  if (!selectedTemplateInfo.value) return

  await mappingFormRef.value.validate(async (valid: boolean) => {
    if (!valid) return
    saving.value = true
    try {
      const payloadBase = {
        source_field: mappingForm.value.source_field,
        target_field: mappingForm.value.target_field,
        field_type: mappingForm.value.field_type,
        transform_rule: mappingForm.value.transform_rule || '',
        sort_order: mappingForm.value.sort_order
      }

      if (mappingDialogMode.value === 'add') {
        await createMapping({
          ...payloadBase,
          template_key: selectedTemplate.value,
          template_name: selectedTemplateInfo.value.template_name
        })
        ElMessage.success('创建成功')
      } else {
        await updateMapping(mappingForm.value.id, payloadBase)
        ElMessage.success('更新成功')
      }

      mappingDialogVisible.value = false
      await reloadMappings(selectedTemplate.value)
    } catch (e) {
      ElMessage.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      saving.value = false
    }
  })
}

const getFieldTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    string: '字符串',
    boolean: '布尔值',
    date: '日期',
    array: '数组'
  }
  return labels[type] || type
}

const getFieldTypeTagType = (type: string) => {
  const types: Record<string, any> = {
    string: 'primary',
    boolean: 'success',
    date: 'warning',
    array: 'info'
  }
  return types[type] || 'primary'
}

onMounted(() => {
  reloadAll()
})
</script>

<style scoped lang="scss">
.template-hint {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}
.rule-text {
  color: #606266;
  font-size: 12px;
  word-break: break-all;
}
.no-rule {
  color: #999;
}
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>

