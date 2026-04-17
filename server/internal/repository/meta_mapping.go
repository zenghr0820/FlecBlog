package repository

import (
	"flec_blog/internal/model"

	"gorm.io/gorm"
)

// MetaMappingRepository 元数据映射仓库
type MetaMappingRepository struct {
	db *gorm.DB
}

func NewMetaMappingRepository(db *gorm.DB) *MetaMappingRepository {
	return &MetaMappingRepository{db: db}
}

func (r *MetaMappingRepository) GetByTemplateKey(templateKey string) ([]model.MetaMapping, error) {
	var mappings []model.MetaMapping
	err := r.db.Where("template_key = ?", templateKey).
		Order("sort_order ASC, id ASC").
		Find(&mappings).Error
	return mappings, err
}

func (r *MetaMappingRepository) Create(mapping *model.MetaMapping) error {
	return r.db.Create(mapping).Error
}

func (r *MetaMappingRepository) Update(mapping *model.MetaMapping) error {
	return r.db.Save(mapping).Error
}

func (r *MetaMappingRepository) Delete(id uint) error {
	return r.db.Delete(&model.MetaMapping{}, id).Error
}

func (r *MetaMappingRepository) GetByID(id uint) (*model.MetaMapping, error) {
	var mapping model.MetaMapping
	if err := r.db.First(&mapping, id).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *MetaMappingRepository) ExistsByTemplateKeyAndSourceField(templateKey, sourceField string, excludeID ...uint) (bool, error) {
	var count int64
	query := r.db.Model(&model.MetaMapping{}).
		Where("template_key = ? AND source_field = ?", templateKey, sourceField)
	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MetaMappingRepository) UpdateTemplateNameByKey(templateKey string, templateName string) error {
	return r.db.Model(&model.MetaMapping{}).
		Where("template_key = ?", templateKey).
		Update("template_name", templateName).Error
}

