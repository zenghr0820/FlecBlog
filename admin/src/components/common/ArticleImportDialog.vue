<template>
  <el-dialog
    v-model="visible"
    title="导入文章"
    width="600px"
    :close-on-click-modal="false"
  >
    <el-form label-width="100px">
      <el-form-item label="映射模版">
        <el-select
          v-model="articleMappingTemplate"
          placeholder="请选择映射模版"
          style="width: 100%"
          :loading="templateLoading"
        >
          <el-option
            v-for="template in templateList"
            :key="template.template_key"
            :label="template.template_name"
            :value="template.template_key"
          >
            <span>{{ template.template_name }}</span>
            <span style="float: right; color: #8492a6; font-size: 13px">
              {{ template.mapping_count }} 个映射
            </span>
          </el-option>
        </el-select>
        <div class="form-tip">
          选择映射模版后，系统将自动识别文章元数据字段
        </div>
      </el-form-item>

      <el-form-item label="上传文件">
        <el-upload
          :auto-upload="false"
          :file-list="articleFileList"
          :on-change="handleArticleFileChange"
          :on-remove="handleArticleFileRemove"
          accept=".md,.markdown"
          :limit="100"
          multiple
          drag
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">拖拽或点击选择文件</div>
          <template #tip>
            <div class="el-upload__tip">最多添加 100 个文件，如遇上传失败请减少数量后重试</div>
          </template>
        </el-upload>
      </el-form-item>

      <el-form-item label="图片处理">
        <el-switch v-model="articleUploadImages" />
        <div class="form-tip" style="margin: 0 15px">开启后将自动下载并上传文章中的图片</div>
      </el-form-item>
    </el-form>

    <el-alert
      v-if="articleImportResult"
      :type="articleImportResult.failed > 0 ? 'warning' : 'success'"
      :closable="false"
      style="margin-top: 16px"
    >
      <div>成功 {{ articleImportResult.success }} 篇，失败 {{ articleImportResult.failed }} 篇</div>
      <div
        v-if="articleImportResult.errors?.length"
        style="margin-top: 8px; font-size: 12px; color: #909399"
      >
        <div v-for="(err, i) in articleImportResult.errors" :key="i">
          {{ err.filename }}: {{ err.error }}
        </div>
      </div>
    </el-alert>

    <template #footer>
      <el-button @click="handleCancel">取消</el-button>
      <el-button
        type="primary"
        :loading="articleUploading"
        :disabled="articleFileList.length === 0"
        @click="handleArticleImport"
      >
        {{ articleUploading ? '导入中...' : '开始导入' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { UploadFilled } from '@element-plus/icons-vue';
import type { UploadUserFile, UploadFile } from 'element-plus';
import { importArticles } from '@/api/article';
import { getTemplates } from '@/api/metaMapping';
import type { ImportArticlesResult } from '@/types/article';

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  'import-success': [];
}>();

// 映射模版相关
interface MappingTemplate {
  id: number;
  template_key: string;
  template_name: string;
  description: string;
  mapping_count: number;
}

const templateList = ref<MappingTemplate[]>([]);
const templateLoading = ref(false);
const articleMappingTemplate = ref<string>('');

// 加载映射模版列表
const loadTemplates = async () => {
  try {
    templateLoading.value = true;
    const response = await getTemplates();
    const templates = response.data || [];
    templateList.value = templates;
    
    // 如果有模版且当前未选择，默认选择第一个
    if (templateList.value.length > 0 && !articleMappingTemplate.value) {
      articleMappingTemplate.value = templateList.value[0]?.template_key || '';
    }
  } catch (error: any) {
    console.error('获取映射模版列表失败:', error);
    ElMessage.warning('获取映射模版列表失败');
  } finally {
    templateLoading.value = false;
  }
};

// 文章导入相关
const visible = ref(props.modelValue);
const articleFileList = ref<UploadUserFile[]>([]);
const articleUploading = ref(false);
const articleImportResult = ref<ImportArticlesResult | undefined>();
const articleUploadImages = ref(false);

const handleArticleFileChange = (file: UploadFile, files: UploadUserFile[]) => {
  articleFileList.value = files;
};

const handleArticleFileRemove = (file: UploadFile, files: UploadUserFile[]) => {
  articleFileList.value = files;
};

const handleArticleImport = async () => {
  if (articleFileList.value.length === 0) {
    ElMessage.warning('请选择要导入的文件');
    return;
  }

  if (!articleMappingTemplate.value) {
    ElMessage.warning('请选择映射模版');
    return;
  }

  try {
    articleUploading.value = true;
    articleImportResult.value = undefined;

    const formData = new FormData();
    formData.append('source_type', articleMappingTemplate.value);
    formData.append('upload_images', String(articleUploadImages.value));
    
    articleFileList.value.forEach(file => {
      if (file.raw) formData.append('files', file.raw);
    });

    const result = await importArticles(formData);
    articleImportResult.value = result;

    if (result.failed === 0) {
      ElMessage.success(`成功导入 ${result.success} 篇文章`);
      emit('import-success');
    } else if (result.success > 0) {
      ElMessage.warning(`导入完成：成功 ${result.success} 篇，失败 ${result.failed} 篇`);
      emit('import-success');
    } else {
      ElMessage.error('导入失败');
    }
  } catch (error: any) {
    ElMessage.error(error.message || '导入失败');
  } finally {
    articleUploading.value = false;
  }
};

const handleCancel = () => {
  visible.value = false;
};

watch(() => props.modelValue, (val) => {
  visible.value = val;
});

watch(visible, val => {
  emit('update:modelValue', val);
  if (!val) {
    setTimeout(() => {
      articleFileList.value = [];
      articleImportResult.value = undefined;
      articleUploadImages.value = false;
      // 重置映射模版选择（保留已加载的列表）
      articleMappingTemplate.value = templateList.value.length > 0 
        ? templateList.value[0]?.template_key || ''
        : '';
    }, 300);
  } else {
    // 对话框打开时，如果模版列表为空则加载
    if (templateList.value.length === 0) {
      loadTemplates();
    }
  }
});

onMounted(() => {
  loadTemplates();
});
</script>

<style lang="scss" scoped>
:deep(.el-icon--upload) {
  font-size: 40px;
  color: #409eff;
  margin-bottom: 12px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: 8px;
}
</style>
