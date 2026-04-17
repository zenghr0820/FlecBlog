package repository

import (
	"flec_blog/internal/model"

	"gorm.io/gorm"
)

type MetaMappingTemplateRepository struct {
	db *gorm.DB
}

func NewMetaMappingTemplateRepository(db *gorm.DB) *MetaMappingTemplateRepository {
	return &MetaMappingTemplateRepository{db: db}
}

func (r *MetaMappingTemplateRepository) List() ([]model.MetaMappingTemplate, error) {
	var templates []model.MetaMappingTemplate
	err := r.db.Order("id ASC").Find(&templates).Error
	return templates, err
}

func (r *MetaMappingTemplateRepository) GetByID(id uint) (*model.MetaMappingTemplate, error) {
	var t model.MetaMappingTemplate
	if err := r.db.First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MetaMappingTemplateRepository) GetByKey(templateKey string) (*model.MetaMappingTemplate, error) {
	var t model.MetaMappingTemplate
	if err := r.db.Where("template_key = ?", templateKey).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *MetaMappingTemplateRepository) Create(t *model.MetaMappingTemplate) error {
	return r.db.Create(t).Error
}

func (r *MetaMappingTemplateRepository) Update(t *model.MetaMappingTemplate) error {
	return r.db.Save(t).Error
}

func (r *MetaMappingTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&model.MetaMappingTemplate{}, id).Error
}

func (r *MetaMappingTemplateRepository) CountMappings(templateKey string) (int, error) {
	var count int64
	err := r.db.Model(&model.MetaMapping{}).Where("template_key = ?", templateKey).Count(&count).Error
	return int(count), err
}

func (r *MetaMappingTemplateRepository) DeleteMappingsByKey(templateKey string) error {
	return r.db.Where("template_key = ?", templateKey).Delete(&model.MetaMapping{}).Error
}

