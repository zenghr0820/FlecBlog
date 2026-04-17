package service

import (
	"fmt"
	"strings"

	"flec_blog/internal/dto"
	"flec_blog/internal/model"
	"flec_blog/internal/repository"
)

// MetaMappingService 元数据映射服务（用于导入/导出 Markdown meta）
type MetaMappingService struct {
	repo         *repository.MetaMappingRepository
	templateRepo *repository.MetaMappingTemplateRepository
}

func NewMetaMappingService(repo *repository.MetaMappingRepository, templateRepo *repository.MetaMappingTemplateRepository) *MetaMappingService {
	return &MetaMappingService{repo: repo, templateRepo: templateRepo}
}

func (s *MetaMappingService) GetMappingsByTemplateKey(templateKey string) ([]model.MetaMapping, error) {
	return s.repo.GetByTemplateKey(templateKey)
}

func (s *MetaMappingService) ListTemplates() ([]dto.MetaMappingTemplateItem, error) {
	templates, err := s.templateRepo.List()
	if err != nil {
		return nil, err
	}
	items := make([]dto.MetaMappingTemplateItem, 0, len(templates))
	for _, t := range templates {
		cnt, _ := s.templateRepo.CountMappings(t.TemplateKey)
		items = append(items, dto.MetaMappingTemplateItem{
			ID:           t.ID,
			TemplateKey:  t.TemplateKey,
			TemplateName: t.TemplateName,
			Description:  t.Description,
			MappingCount: cnt,
		})
	}
	return items, nil
}

func (s *MetaMappingService) CreateTemplate(req *dto.CreateMetaMappingTemplateRequest) (*model.MetaMappingTemplate, error) {
	t := &model.MetaMappingTemplate{
		TemplateKey:  strings.TrimSpace(req.TemplateKey),
		TemplateName: strings.TrimSpace(req.TemplateName),
		Description:  req.Description,
	}
	if t.TemplateKey == "" || t.TemplateName == "" {
		return nil, fmt.Errorf("template_key 与 template_name 不能为空")
	}

	// 根据模版key中是否存在模版
	if _, err := s.templateRepo.GetByKey(t.TemplateKey); err == nil {
		return nil, fmt.Errorf("模版 %s 已存在", t.TemplateKey)
	}

	if err := s.templateRepo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *MetaMappingService) UpdateTemplate(id uint, req *dto.UpdateMetaMappingTemplateRequest) (*model.MetaMappingTemplate, error) {
	t, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.TemplateName) != "" {
		t.TemplateName = strings.TrimSpace(req.TemplateName)
	}
	if req.Description != "" {
		t.Description = req.Description
	}
	if err := s.templateRepo.Update(t); err != nil {
		return nil, err
	}
	_ = s.repo.UpdateTemplateNameByKey(t.TemplateKey, t.TemplateName)
	return t, nil
}

func (s *MetaMappingService) DeleteTemplate(id uint) error {
	t, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.templateRepo.DeleteMappingsByKey(t.TemplateKey); err != nil {
		return err
	}
	return s.templateRepo.Delete(id)
}

func (s *MetaMappingService) CreateMapping(req *dto.CreateMetaMappingRequest) (*model.MetaMapping, error) {
	t, err := s.templateRepo.GetByKey(req.TemplateKey)
	if err != nil {
		return nil, fmt.Errorf("映射模版不存在，请先创建模版")
	}
	exists, err := s.repo.ExistsByTemplateKeyAndSourceField(req.TemplateKey, req.SourceField)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("模版 %s 的源字段 %s 已存在", req.TemplateKey, req.SourceField)
	}

	mapping := &model.MetaMapping{
		TemplateKey:   req.TemplateKey,
		TemplateName:  t.TemplateName,
		SourceField:   req.SourceField,
		TargetField:   req.TargetField,
		FieldType:     req.FieldType,
		TransformRule: req.TransformRule,
		SortOrder:     req.SortOrder,
		IsActive:      true,
		IsSystem:      false,
	}
	if err := s.repo.Create(mapping); err != nil {
		return nil, err
	}
	return mapping, nil
}

func (s *MetaMappingService) UpdateMapping(id uint, req *dto.UpdateMetaMappingRequest) (*model.MetaMapping, error) {
	mapping, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.SourceField != "" && req.SourceField != mapping.SourceField {
		exists, err := s.repo.ExistsByTemplateKeyAndSourceField(mapping.TemplateKey, req.SourceField, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("源字段 %s 已存在", req.SourceField)
		}
		mapping.SourceField = req.SourceField
	}
	if req.TargetField != "" {
		mapping.TargetField = req.TargetField
	}
	if req.FieldType != "" {
		mapping.FieldType = req.FieldType
	}
	if req.IsActive != nil {
		mapping.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		mapping.SortOrder = *req.SortOrder
	}
	mapping.TransformRule = req.TransformRule
	if err := s.repo.Update(mapping); err != nil {
		return nil, err
	}
	return mapping, nil
}

func (s *MetaMappingService) DeleteMapping(id uint) error {
	return s.repo.Delete(id)
}

func (s *MetaMappingService) ToggleMappingStatus(id uint) (bool, error) {
	mapping, err := s.repo.GetByID(id)
	if err != nil {
		return false, err
	}
	mapping.IsActive = !mapping.IsActive
	if err := s.repo.Update(mapping); err != nil {
		return false, err
	}
	return mapping.IsActive, nil
}

// func (s *MetaMappingService) ApplyMappingsToFrontMatter(templateKey string, frontMatter map[string]interface{}) (*dto.ApplyMappingResult, error) {
// 	result := &dto.ApplyMappingResult{
// 		Tags:       []string{},
// 		IsPublish:  false,
// 		CustomData: make(map[string]interface{}),
// 	}

// 	mappings, err := s.GetMappingsByTemplateKey(templateKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, mapping := range mappings {
// 		sourceValue, exists := frontMatter[mapping.SourceField]
// 		if !exists {
// 			continue
// 		}
// 		switch mapping.TargetField {
// 		case "title":
// 			result.Title = s.convertToString(sourceValue)
// 		case "slug":
// 			result.Slug = s.convertToString(sourceValue)
// 		case "content":
// 			result.Content = s.convertToString(sourceValue)
// 		case "summary":
// 			result.Summary = s.convertToString(sourceValue)
// 		case "cover":
// 			result.Cover = s.convertToString(sourceValue)
// 		case "category":
// 			result.Category = s.convertToString(sourceValue)
// 		case "location":
// 			result.Location = s.convertToString(sourceValue)
// 		case "is_publish":
// 			result.IsPublish = s.convertToBoolean(sourceValue, mapping.TransformRule)
// 		case "is_top":
// 			result.IsTop = s.convertToBoolean(sourceValue, mapping.TransformRule)
// 		case "is_essence":
// 			result.IsEssence = s.convertToBoolean(sourceValue, mapping.TransformRule)
// 		case "is_outdated":
// 			result.IsOutdated = s.convertToBoolean(sourceValue, mapping.TransformRule)
// 		case "publish_time":
// 			if timeStr := s.convertToString(sourceValue); timeStr != "" {
// 				result.PublishTime = &timeStr
// 			}
// 		case "update_time":
// 			if timeStr := s.convertToString(sourceValue); timeStr != "" {
// 				result.UpdateTime = &timeStr
// 			}
// 		case "tags":
// 			result.Tags = s.convertToArray(sourceValue)
// 		default:
// 			result.CustomData[mapping.TargetField] = sourceValue
// 		}
// 	}

// 	for key, value := range frontMatter {
// 		isMapped := false
// 		for _, mapping := range mappings {
// 			if mapping.SourceField == key {
// 				isMapped = true
// 				break
// 			}
// 		}
// 		if !isMapped {
// 			result.CustomData[key] = value
// 		}
// 	}

// 	return result, nil
// }

// func (s *MetaMappingService) convertToString(value interface{}) string {
// 	switch v := value.(type) {
// 	case string:
// 		return v
// 	case int, int32, int64:
// 		return fmt.Sprintf("%d", v)
// 	case float32, float64:
// 		return fmt.Sprintf("%f", v)
// 	case bool:
// 		return strconv.FormatBool(v)
// 	default:
// 		return fmt.Sprintf("%v", v)
// 	}
// }

// func (s *MetaMappingService) convertToBoolean(value interface{}, transformRule string) bool {
// 	var rule dto.TransformRule
// 	if transformRule != "" {
// 		_ = json.Unmarshal([]byte(transformRule), &rule)
// 	}

// 	strValue := strings.ToLower(s.convertToString(value))

// 	if rule.TrueValue != "" || rule.FalseValue != "" {
// 		if strValue == strings.ToLower(rule.TrueValue) {
// 			return true
// 		}
// 		if strValue == strings.ToLower(rule.FalseValue) {
// 			return false
// 		}
// 	}

// 	if rule.Comparison != "" {
// 		numValue, err := strconv.ParseFloat(strValue, 64)
// 		if err == nil {
// 			compareValue, err := strconv.ParseFloat(rule.Value, 64)
// 			if err == nil {
// 				switch rule.Comparison {
// 				case ">":
// 					return numValue > compareValue
// 				case "<":
// 					return numValue < compareValue
// 				case "=":
// 					return numValue == compareValue
// 				case "!=":
// 					return numValue != compareValue
// 				}
// 			}
// 		}
// 	}

// 	return strValue == "true" || strValue == "1" || strValue == "yes" || strValue == "on"
// }

// func (s *MetaMappingService) convertToArray(value interface{}) []string {
// 	switch v := value.(type) {
// 	case []string:
// 		return v
// 	case []interface{}:
// 		result := make([]string, 0, len(v))
// 		for _, item := range v {
// 			result = append(result, s.convertToString(item))
// 		}
// 		return result
// 	case string:
// 		if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
// 			var result []string
// 			if err := json.Unmarshal([]byte(v), &result); err == nil {
// 				return result
// 			}
// 		}
// 		if strings.Contains(v, ",") {
// 			return strings.Split(v, ",")
// 		}
// 		return []string{v}
// 	default:
// 		return []string{s.convertToString(v)}
// 	}
// }

func (s *MetaMappingService) ApplyTransformRule(value string, transformRule string) string {
	if transformRule == "" {
		return value
	}
	// var rule dto.TransformRule
	// if err := json.Unmarshal([]byte(transformRule), &rule); err != nil {
	// 	return value
	// }
	// result := value
	// if rule.Prefix != "" {
	// 	result = rule.Prefix + result
	// }
	// if rule.Suffix != "" {
	// 	result = result + rule.Suffix
	// }
	// if rule.Regex != "" && rule.Replace != "" {
	// 	re, err := regexp.Compile(rule.Regex)
	// 	if err == nil {
	// 		result = re.ReplaceAllString(result, rule.Replace)
	// 	}
	// }
	return value
}
