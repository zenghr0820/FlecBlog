package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"flec_blog/config"
	"flec_blog/internal/dto"
	"flec_blog/internal/model"
	"flec_blog/internal/repository"
	"flec_blog/pkg/logger"
	"flec_blog/pkg/random"
	"flec_blog/pkg/utils"
	"flec_blog/pkg/wechatmp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gorm.io/gorm"
)

// ArticleService 文章服务
type ArticleService struct {
	articleRepo       *repository.ArticleRepository
	tagRepo           *repository.TagRepository
	categoryRepo      *repository.CategoryRepository
	commentRepo       *repository.CommentRepository
	fileService       *FileService
	subscriberService *SubscriberService
	db                *gorm.DB
	config            *config.Config // 配置对象（支持热重载）
	md                goldmark.Markdown
	httpClient        *http.Client
}

// NewArticleService 创建文章服务实例
func NewArticleService(articleRepo *repository.ArticleRepository, tagRepo *repository.TagRepository, categoryRepo *repository.CategoryRepository, commentRepo *repository.CommentRepository, fileService *FileService, db *gorm.DB, cfg *config.Config) *ArticleService {
	// 初始化 goldmark（用于微信导出时渲染 Markdown）
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	return &ArticleService{
		articleRepo:  articleRepo,
		tagRepo:      tagRepo,
		categoryRepo: categoryRepo,
		commentRepo:  commentRepo,
		fileService:  fileService,
		db:           db,
		config:       cfg,
		md:           md,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// SetSubscriberService 设置订阅者服务（避免循环依赖）
func (s *ArticleService) SetSubscriberService(subscriberService *SubscriberService) {
	s.subscriberService = subscriberService
}

// ============ 通用服务 ============

// Get 获取文章详情
func (s *ArticleService) Get(_ context.Context, id uint) (*dto.ArticleAdminDetailResponse, error) {
	article, err := s.articleRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("获取文章失败: %w", err)
	}

	response := &dto.ArticleAdminDetailResponse{
		ID:          article.ID,
		Title:       article.Title,
		Content:     article.Content,
		Summary:     article.Summary,
		AISummary:   article.AISummary,
		Cover:       article.Cover,
		Location:    article.Location,
		IsPublish:   article.IsPublish,
		IsTop:       article.IsTop,
		IsEssence:   article.IsEssence,
		IsOutdated:  article.IsOutdated,
		PublishTime: utils.ToJSONTime(article.PublishTime),
		UpdateTime:  utils.ToJSONTime(article.UpdateTime),
	}

	// 填充分类信息
	response.Category.ID = article.Category.ID
	response.Category.Name = article.Category.Name

	// 填充标签信息
	for _, tag := range article.Tags {
		response.Tags = append(response.Tags, struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
		}{tag.ID, tag.Name})
	}

	return response, nil
}

// ============ 前台服务 ============

// ListForWeb 获取前台文章列表
func (s *ArticleService) ListForWeb(ctx context.Context, req *dto.ListArticlesRequest) ([]dto.ArticleWebResponse, int64, error) {
	articles, total, err := s.articleRepo.ListForWeb(req.Page, req.PageSize, req.Year, req.Month, req.Category, req.Tag)
	if err != nil {
		return nil, 0, err
	}

	// 批量获取文章评论数
	articleSlugs := make([]string, len(articles))
	for i, article := range articles {
		articleSlugs[i] = article.Slug
	}

	commentCounts := make(map[string]int64)
	if len(articleSlugs) > 0 && s.commentRepo != nil {
		commentCounts, err = s.commentRepo.CountByTargetKeys(ctx, "article", articleSlugs)
		if err != nil {
			// 如果获取评论数失败，不影响主流程，只记录错误
			commentCounts = make(map[string]int64)
		}
	}

	// 转换为前台响应格式
	var response []dto.ArticleWebResponse
	for _, article := range articles {
		item := dto.ArticleWebResponse{
			ID:           article.ID,
			Title:        article.Title,
			Summary:      article.Summary,
			Cover:        article.Cover,
			Location:     article.Location,
			IsTop:        article.IsTop,
			IsEssence:    article.IsEssence,
			IsOutdated:   article.IsOutdated,
			URL:          fmt.Sprintf("/posts/%s", article.Slug),
			CommentCount: commentCounts[article.Slug],
			PublishTime:  utils.ToJSONTime(article.PublishTime),
			UpdateTime:   utils.ToJSONTime(article.UpdateTime),
		}

		// 填充分类信息
		if article.Category.ID > 0 {
			item.Category.ID = article.Category.ID
			item.Category.Name = article.Category.Name
			item.Category.URL = fmt.Sprintf("/category/%s", article.Category.Slug)
		}

		// 填充标签信息
		for _, tag := range article.Tags {
			item.Tags = append(item.Tags, struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
				URL  string `json:"url"`
			}{
				ID:   tag.ID,
				Name: tag.Name,
				URL:  fmt.Sprintf("/tag/%s", tag.Slug),
			})
		}

		response = append(response, item)
	}

	return response, total, nil
}

// Search 搜索文章
func (s *ArticleService) Search(ctx context.Context, req *dto.SearchArticlesRequest) ([]dto.ArticleWebResponse, int64, error) {
	offset := (req.Page - 1) * req.PageSize
	articles, total, err := s.articleRepo.Search(req.Keyword, offset, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 批量获取文章评论数
	articleSlugs := make([]string, len(articles))
	for i, article := range articles {
		articleSlugs[i] = article.Slug
	}

	commentCounts := make(map[string]int64)
	if len(articleSlugs) > 0 && s.commentRepo != nil {
		commentCounts, err = s.commentRepo.CountByTargetKeys(ctx, "article", articleSlugs)
		if err != nil {
			// 如果获取评论数失败，不影响主流程，只记录错误
			commentCounts = make(map[string]int64)
		}
	}

	var response []dto.ArticleWebResponse
	for _, article := range articles {
		item := dto.ArticleWebResponse{
			ID:           article.ID,
			Title:        article.Title,
			Summary:      article.Summary,
			Cover:        article.Cover,
			Location:     article.Location,
			IsTop:        article.IsTop,
			IsEssence:    article.IsEssence,
			URL:          fmt.Sprintf("/posts/%s", article.Slug),
			Excerpt:      utils.GenerateExcerpt(article.Content, req.Keyword, 40), // 生成包含关键词的摘录
			CommentCount: commentCounts[article.Slug],
			PublishTime:  utils.ToJSONTime(article.PublishTime),
			UpdateTime:   utils.ToJSONTime(article.UpdateTime),
		}

		if article.Category.ID > 0 {
			item.Category.ID = article.Category.ID
			item.Category.Name = article.Category.Name
			item.Category.URL = fmt.Sprintf("/category/%s", article.Category.Slug)
		}

		for _, tag := range article.Tags {
			item.Tags = append(item.Tags, struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
				URL  string `json:"url"`
			}{
				ID:   tag.ID,
				Name: tag.Name,
				URL:  fmt.Sprintf("/tag/%s", tag.Slug),
			})
		}

		response = append(response, item)
	}

	return response, total, nil
}

// GetBySlug 通过slug获取文章详情
func (s *ArticleService) GetBySlug(ctx context.Context, slug string) (*dto.ArticleDetailResponse, error) {
	article, err := s.articleRepo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}

	// 异步增加浏览数
	go func() {
		_ = s.articleRepo.IncrementViewCount(article.ID)
	}()

	// 获取文章评论数
	var commentCount int64
	if s.commentRepo != nil {
		commentCounts, err := s.commentRepo.CountByTargetKeys(ctx, "article", []string{article.Slug})
		if err == nil {
			commentCount = commentCounts[article.Slug]
		}
	}

	response := &dto.ArticleDetailResponse{
		ID:           article.ID,
		Title:        article.Title,
		Slug:         article.Slug,
		Content:      article.Content,
		Summary:      article.Summary,
		AISummary:    article.AISummary,
		Cover:        article.Cover,
		Location:     article.Location,
		IsTop:        article.IsTop,
		IsEssence:    article.IsEssence,
		IsOutdated:   article.IsOutdated,
		ViewCount:    article.ViewCount,
		CommentCount: commentCount,
		URL:          fmt.Sprintf("/posts/%s", article.Slug),
		PublishTime:  utils.ToJSONTime(article.PublishTime),
		UpdateTime:   utils.ToJSONTime(article.UpdateTime),
	}

	// 填充分类信息
	if article.Category.ID > 0 {
		response.Category.ID = article.Category.ID
		response.Category.Name = article.Category.Name
		response.Category.URL = fmt.Sprintf("/category/%s", article.Category.Slug)
	}

	// 填充标签信息
	for _, tag := range article.Tags {
		response.Tags = append(response.Tags, struct {
			ID   uint   `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		}{
			ID:   tag.ID,
			Name: tag.Name,
			URL:  fmt.Sprintf("/tag/%s", tag.Slug),
		})
	}

	// 查询上一篇文章
	if prevArticle, err := s.articleRepo.GetPrevArticle(article.PublishTime); err == nil {
		response.Prev = &struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}{
			Title: prevArticle.Title,
			URL:   fmt.Sprintf("/posts/%s", prevArticle.Slug),
		}
	}

	// 查询下一篇文章
	if nextArticle, err := s.articleRepo.GetNextArticle(article.PublishTime); err == nil {
		response.Next = &struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}{
			Title: nextArticle.Title,
			URL:   fmt.Sprintf("/posts/%s", nextArticle.Slug),
		}
	}

	return response, nil
}

// ============ 后台管理服务 ============

// List 获取文章列表
func (s *ArticleService) List(ctx context.Context, req *dto.ListArticlesRequest) ([]dto.ArticleListResponse, int64, error) {
	offset := (req.Page - 1) * req.PageSize
	articles, total, err := s.articleRepo.List(offset, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 批量获取文章评论数
	articleSlugs := make([]string, len(articles))
	for i, article := range articles {
		articleSlugs[i] = article.Slug
	}

	commentCounts := make(map[string]int64)
	if len(articleSlugs) > 0 && s.commentRepo != nil {
		commentCounts, err = s.commentRepo.CountByTargetKeys(ctx, "article", articleSlugs)
		if err != nil {
			// 如果获取评论数失败，不影响主流程
			commentCounts = make(map[string]int64)
		}
	}

	// 转换为后台列表响应格式
	var response []dto.ArticleListResponse
	for _, article := range articles {
		item := dto.ArticleListResponse{
			ID:           article.ID,
			Title:        article.Title,
			Cover:        article.Cover,
			Location:     article.Location,
			IsPublish:    article.IsPublish,
			IsTop:        article.IsTop,
			IsEssence:    article.IsEssence,
			IsOutdated:   article.IsOutdated,
			ViewCount:    article.ViewCount,
			CommentCount: commentCounts[article.Slug],
			PublishTime:  utils.ToJSONTime(article.PublishTime),
			UpdateTime:   utils.ToJSONTime(article.UpdateTime),
		}

		item.Category.ID = article.Category.ID
		item.Category.Name = article.Category.Name

		for _, tag := range article.Tags {
			item.Tags = append(item.Tags, struct {
				ID   uint   `json:"id"`
				Name string `json:"name"`
			}{tag.ID, tag.Name})
		}

		response = append(response, item)
	}

	return response, total, nil
}

// Create 创建文章
func (s *ArticleService) Create(ctx context.Context, req *dto.CreateArticleRequest) (*dto.ArticleAdminDetailResponse, error) {
	// 验证分类是否存在
	if req.CategoryID != nil && *req.CategoryID > 0 {
		_, err := s.categoryRepo.Get(ctx, *req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("分类不存在: %w", err)
		}
	}

	article := &model.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		Cover:      req.Cover,
		Location:   req.Location,
		CategoryID: req.CategoryID,
	}

	// 设置置顶状态
	if req.IsTop != nil {
		article.IsTop = *req.IsTop
	}

	// 设置精选状态
	if req.IsEssence != nil {
		article.IsEssence = *req.IsEssence
	}

	// 设置过时状态
	if req.IsOutdated != nil {
		article.IsOutdated = *req.IsOutdated
	}

	// 设置发布状态
	if req.IsPublish != nil {
		article.IsPublish = *req.IsPublish
	}

	// 如果是发布状态，自动设置发布时间
	if article.IsPublish {
		now := utils.Now().Time
		article.PublishTime = &now
	}

	// 生成唯一slug
	generatedSlug, err := random.UniqueCode(8, s.articleRepo.CheckSlugExists)
	if err != nil {
		return nil, fmt.Errorf("生成 slug 失败: %w", err)
	}
	article.Slug = generatedSlug

	// 创建文章并关联标签
	if err := s.articleRepo.Create(article, req.TagIDs); err != nil {
		return nil, err
	}

	// 如果是发布状态，增加分类和标签计数
	if article.IsPublish {
		s.incrementCounts(ctx, article)
	}

	// 标记封面为使用中
	if req.Cover != "" && s.fileService != nil {
		_ = s.fileService.MarkAsUsed(req.Cover)
	}

	// 标记内容中的图片为使用中
	s.markContentImagesAsUsed(req.Content)

	// 如果是发布状态，异步发送订阅推送
	if article.IsPublish && s.subscriberService != nil {
		go func() {
			if err := s.subscriberService.SendArticleNotification(context.Background(), article); err != nil {
				logger.Warn("发送文章推送失败 (文章ID: %d): %v", article.ID, err)
			}
		}()
	}

	return s.Get(ctx, article.ID)
}

// Update 更新文章
func (s *ArticleService) Update(ctx context.Context, id uint, req *dto.UpdateArticleRequest) (*dto.ArticleAdminDetailResponse, error) {
	article, err := s.articleRepo.Get(id)
	if err != nil {
		return nil, err
	}

	// 验证新分类是否存在
	if req.CategoryID != nil && *req.CategoryID > 0 {
		if _, err := s.categoryRepo.Get(ctx, *req.CategoryID); err != nil {
			return nil, fmt.Errorf("分类不存在: %w", err)
		}
	}

	// 保存旧值用于后续处理
	oldCategoryID := article.CategoryID
	oldTagIDs := extractTagIDs(article.Tags)
	oldCover := article.Cover
	oldContent := article.Content
	oldIsPublish := article.IsPublish

	// 更新字段
	if req.Title != "" {
		article.Title = req.Title
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	article.Summary = req.Summary
	article.AISummary = req.AISummary
	article.Cover = req.Cover
	article.Location = req.Location
	article.CategoryID = req.CategoryID
	if req.IsTop != nil {
		article.IsTop = *req.IsTop
	}

	// 处理精选状态
	if req.IsEssence != nil {
		article.IsEssence = *req.IsEssence
	}

	// 处理过时状态
	if req.IsOutdated != nil {
		article.IsOutdated = *req.IsOutdated
	}

	// 处理发布状态
	if req.IsPublish != nil {
		article.IsPublish = *req.IsPublish
	}

	// 先处理请求中的 PublishTime（仅当传入非空时间时才更新）
	if req.PublishTime != nil && !req.PublishTime.IsZero() {
		article.PublishTime = utils.FromJSONTime(req.PublishTime)
	}

	// 如果是发布状态且没有发布时间，自动设置发布时间
	if article.IsPublish && article.PublishTime == nil {
		now := utils.Now().Time
		article.PublishTime = &now
	}
	if req.UpdateTime != nil {
		article.UpdateTime = utils.FromJSONTime(req.UpdateTime)
	}

	if err := s.articleRepo.Update(article, req.TagIDs); err != nil {
		return nil, err
	}

	// 处理发布状态变化的计数
	if oldIsPublish != article.IsPublish {
		if article.IsPublish {
			// 草稿 -> 已发布：增加计数
			s.incrementCounts(ctx, article)
		} else {
			// 已发布 -> 草稿：减少计数
			s.decrementCounts(ctx, article)
		}
	} else if article.IsPublish {
		// 如果一直是已发布状态，更新分类和标签计数（处理分类/标签变更）
		s.updateCountsOnChange(ctx, oldCategoryID, req.CategoryID, oldTagIDs, req.TagIDs)
	}

	// 处理封面变化
	if s.fileService != nil && oldCover != req.Cover {
		if oldCover != "" {
			_ = s.fileService.MarkAsUnused(oldCover)
		}
		if req.Cover != "" {
			_ = s.fileService.MarkAsUsed(req.Cover)
		}
	}

	// 处理内容图片变化
	if req.Content != "" {
		s.updateContentFileStatus(oldContent, req.Content)
	}

	// 如果从草稿变为发布状态，异步发送订阅推送
	if !oldIsPublish && article.IsPublish && s.subscriberService != nil {
		go func() {
			if err := s.subscriberService.SendArticleNotification(context.Background(), article); err != nil {
				logger.Warn("发送文章推送失败 (文章ID: %d): %v", article.ID, err)
			}
		}()
	}

	return s.Get(ctx, id)
}

// Delete 删除文章
func (s *ArticleService) Delete(ctx context.Context, id uint) error {
	article, err := s.articleRepo.Get(id)
	if err != nil {
		return err
	}

	// 如果是已发布文章，减少计数
	if article.IsPublish {
		s.decrementCounts(ctx, article)
	}

	// 标记封面为未使用
	if s.fileService != nil && article.Cover != "" {
		_ = s.fileService.MarkAsUnused(article.Cover)
	}

	// 标记内容中的图片为未使用
	s.markContentImagesAsUnused(article.Content)

	return s.articleRepo.Delete(id)
}

// ============ 辅助方法 ============

// extractTagIDs 提取标签ID列表
func extractTagIDs(tags []model.Tag) []uint {
	if len(tags) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}

// incrementCounts 增加分类和标签的文章计数
func (s *ArticleService) incrementCounts(ctx context.Context, article *model.Article) {
	if article.CategoryID != nil && *article.CategoryID > 0 {
		_ = s.categoryRepo.IncrementCount(ctx, *article.CategoryID)
	}
	if tagIDs := extractTagIDs(article.Tags); len(tagIDs) > 0 {
		_ = s.tagRepo.IncrementCountBatch(ctx, tagIDs)
	}
}

// decrementCounts 减少分类和标签的文章计数
func (s *ArticleService) decrementCounts(ctx context.Context, article *model.Article) {
	if article.CategoryID != nil && *article.CategoryID > 0 {
		_ = s.categoryRepo.DecrementCount(ctx, *article.CategoryID)
	}
	if tagIDs := extractTagIDs(article.Tags); len(tagIDs) > 0 {
		_ = s.tagRepo.DecrementCountBatch(ctx, tagIDs)
	}
}

// diffTagIDs 比较标签ID列表差异
func diffTagIDs(oldIDs, newIDs []uint) (removed, added []uint) {
	oldMap := make(map[uint]bool, len(oldIDs))
	for _, id := range oldIDs {
		oldMap[id] = true
	}

	newMap := make(map[uint]bool, len(newIDs))
	for _, id := range newIDs {
		newMap[id] = true
		if !oldMap[id] {
			added = append(added, id)
		}
	}

	for _, id := range oldIDs {
		if !newMap[id] {
			removed = append(removed, id)
		}
	}
	return
}

// updateCountsOnChange 更新文章时处理计数变化
func (s *ArticleService) updateCountsOnChange(ctx context.Context, oldCategoryID, newCategoryID *uint, oldTagIDs, newTagIDs []uint) {
	// 处理分类计数变化
	oldID := getIDValue(oldCategoryID)
	newID := getIDValue(newCategoryID)
	if oldID != newID {
		if oldID > 0 {
			_ = s.categoryRepo.DecrementCount(ctx, oldID)
		}
		if newID > 0 {
			_ = s.categoryRepo.IncrementCount(ctx, newID)
		}
	}

	// 处理标签计数变化
	if newTagIDs != nil {
		removed, added := diffTagIDs(oldTagIDs, newTagIDs)
		if len(removed) > 0 {
			_ = s.tagRepo.DecrementCountBatch(ctx, removed)
		}
		if len(added) > 0 {
			_ = s.tagRepo.IncrementCountBatch(ctx, added)
		}
	}
}

// getIDValue 安全获取指针ID的值
func getIDValue(id *uint) uint {
	if id == nil {
		return 0
	}
	return *id
}

// extractContentImages 从 Markdown/HTML 内容中提取所有图片 URL
func extractContentImages(content string) []string {
	var urls []string
	seen := make(map[string]bool)

	// 提取 Markdown 图片: ![alt](url)
	mdImageRe := regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)
	matches := mdImageRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			url := strings.TrimSpace(match[1])
			if url != "" && !seen[url] {
				seen[url] = true
				urls = append(urls, url)
			}
		}
	}

	// 提取 HTML img 标签: <img src="url" />
	htmlImageRe := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	matches = htmlImageRe.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 1 {
			url := strings.TrimSpace(match[1])
			if url != "" && !seen[url] {
				seen[url] = true
				urls = append(urls, url)
			}
		}
	}

	return urls
}

// markContentImagesAsUsed 标记内容中的图片为已使用
func (s *ArticleService) markContentImagesAsUsed(content string) {
	if s.fileService == nil {
		return
	}
	for _, url := range extractContentImages(content) {
		_ = s.fileService.MarkAsUsed(url)
	}
}

// markContentImagesAsUnused 标记内容中的图片为未使用
func (s *ArticleService) markContentImagesAsUnused(content string) {
	if s.fileService == nil {
		return
	}
	for _, url := range extractContentImages(content) {
		_ = s.fileService.MarkAsUnused(url)
	}
}

// updateContentFileStatus 对比新旧内容，更新图片文件状态
func (s *ArticleService) updateContentFileStatus(oldContent, newContent string) {
	if s.fileService == nil {
		return
	}

	oldImages := make(map[string]bool)
	for _, url := range extractContentImages(oldContent) {
		oldImages[url] = true
	}

	newImages := make(map[string]bool)
	for _, url := range extractContentImages(newContent) {
		newImages[url] = true
		// 新增的图片标记为使用中
		if !oldImages[url] {
			_ = s.fileService.MarkAsUsed(url)
		}
	}

	// 移除的图片标记为未使用
	for url := range oldImages {
		if !newImages[url] {
			_ = s.fileService.MarkAsUnused(url)
		}
	}
}

// ============ 数据导入导出方法 ============

// ImportFromHexo 从Hexo格式导入文章
func (s *ArticleService) ImportFromHexo(ctx context.Context, files map[string]string) (*dto.ImportArticlesResult, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("没有找到有效的文章数据")
	}

	result := &dto.ImportArticlesResult{
		Total: len(files),
	}

	// 缓存已创建的分类和标签
	categoryCache := make(map[string]*model.Category)
	tagCache := make(map[string]*model.Tag)

	// 处理每篇文章
	for filename, content := range files {
		if err := s.importSingleHexoArticle(ctx, content, categoryCache, tagCache); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, dto.ImportArticleError{
				Filename: filename,
				Title:    extractTitle(content),
				Error:    err.Error(),
			})
		} else {
			result.Success++
		}
	}

	result.CategoriesAdded = len(categoryCache)
	result.TagsAdded = len(tagCache)

	return result, nil
}

// importSingleHexoArticle 导入单篇Hexo文章
func (s *ArticleService) importSingleHexoArticle(
	ctx context.Context,
	content string,
	categoryCache map[string]*model.Category,
	tagCache map[string]*model.Tag,
) error {
	// 解析Hexo文章
	parsed, err := parseHexoArticle(content)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	// 处理分类
	var categoryID *uint
	if parsed.Category != "" {
		category, err := s.getOrCreateCategory(ctx, parsed.Category, categoryCache)
		if err != nil {
			return fmt.Errorf("分类处理失败: %w", err)
		}
		categoryID = &category.ID
	}

	// 处理标签
	var tagIDs []uint
	for _, tagName := range parsed.Tags {
		tag, err := s.getOrCreateTag(ctx, tagName, tagCache)
		if err != nil {
			return fmt.Errorf("标签处理失败: %w", err)
		}
		tagIDs = append(tagIDs, tag.ID)
	}

	// 处理 slug：优先使用原有的，否则生成新的
	articleSlug := parsed.Slug
	if articleSlug != "" {
		if exists, _ := s.articleRepo.CheckSlugExists(articleSlug); exists {
			articleSlug = "" // slug 已存在，需要生成新的
		}
	}
	if articleSlug == "" {
		articleSlug, _ = random.UniqueCode(8, s.articleRepo.CheckSlugExists)
	}

	// 创建文章
	article := &model.Article{
		Title:       parsed.Title,
		Slug:        articleSlug,
		Content:     parsed.Content,
		Summary:     parsed.Summary,
		Cover:       parsed.Cover,
		IsPublish:   false, // 导入的文章默认为草稿
		IsTop:       false,
		CategoryID:  categoryID,
		PublishTime: parsed.PublishTime,
		UpdateTime:  parsed.UpdateTime,
	}

	if err := s.articleRepo.Create(article, tagIDs); err != nil {
		return fmt.Errorf("保存失败: %w", err)
	}

	// 增加分类和标签计数
	s.incrementCounts(ctx, article)

	return nil
}

// getOrCreateCategory 获取或创建分类
func (s *ArticleService) getOrCreateCategory(ctx context.Context, name string, cache map[string]*model.Category) (*model.Category, error) {
	// 检查缓存
	if category, exists := cache[name]; exists {
		return category, nil
	}

	// 尝试从数据库获取
	category, err := s.categoryRepo.GetBySlug(ctx, name)
	if err == nil {
		cache[name] = category
		return category, nil
	}

	// 不存在则创建
	category = &model.Category{
		Name:        name,
		Slug:        name,
		Description: "",
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	cache[name] = category
	return category, nil
}

// getOrCreateTag 获取或创建标签
func (s *ArticleService) getOrCreateTag(ctx context.Context, name string, cache map[string]*model.Tag) (*model.Tag, error) {
	// 检查缓存
	if tag, exists := cache[name]; exists {
		return tag, nil
	}

	// 尝试从数据库获取
	tag, err := s.tagRepo.GetBySlug(ctx, name)
	if err == nil {
		cache[name] = tag
		return tag, nil
	}

	// 不存在则创建
	tag = &model.Tag{
		Name:        name,
		Slug:        name,
		Description: "",
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	cache[name] = tag
	return tag, nil
}

// extractTitle 从内容中提取标题
func extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "title:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "title:"))
		}
	}
	return "未知标题"
}

// HexoParsedArticle 解析后的Hexo文章
type HexoParsedArticle struct {
	Title       string
	Slug        string
	Content     string
	Summary     string
	Cover       string
	Category    string
	Tags        []string
	PublishTime *time.Time
	UpdateTime  *time.Time
}

// parseHexoArticle 解析Hexo文章格式（Front Matter + Markdown）
func parseHexoArticle(content string) (*HexoParsedArticle, error) {
	// 检查是否包含Front Matter标记
	if !strings.HasPrefix(content, "---") {
		return nil, fmt.Errorf("无效的Hexo格式：缺少Front Matter")
	}

	// 分割Front Matter和内容
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("无效的Hexo格式：Front Matter格式错误")
	}

	frontMatter := parts[1]
	markdown := strings.TrimSpace(parts[2])

	// 解析Front Matter
	parsed := &HexoParsedArticle{
		Content: markdown,
	}

	lines := strings.Split(frontMatter, "\n")
	var tagLines []string
	inTags := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理标签数组
		if inTags {
			if strings.HasPrefix(line, "-") {
				tagValue := strings.TrimSpace(strings.TrimPrefix(line, "-"))
				tagValue = strings.Trim(tagValue, "\"'")
				if tagValue != "" {
					tagLines = append(tagLines, tagValue)
				}
			} else {
				inTags = false
			}
		}

		// 解析键值对
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "-") {
			parts := strings.SplitN(line, ":", 2)
			key := strings.TrimSpace(parts[0])
			value := ""
			if len(parts) > 1 {
				value = strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")
			}

			switch key {
			case "title":
				parsed.Title = value
			case "date":
				if t, err := parseHexoDate(value); err == nil {
					parsed.PublishTime = t
				}
			case "updated":
				if t, err := parseHexoDate(value); err == nil {
					parsed.UpdateTime = t
				}
			case "categories", "category":
				if value != "" {
					parsed.Category = value
				}
				// 如果value为空，可能是数组格式，下一行开始
			case "tags":
				if value != "" {
					// 内联格式: tags: [tag1, tag2]
					value = strings.Trim(value, "[]")
					for _, tag := range strings.Split(value, ",") {
						tag = strings.TrimSpace(tag)
						tag = strings.Trim(tag, "\"'")
						if tag != "" {
							parsed.Tags = append(parsed.Tags, tag)
						}
					}
				} else {
					// 数组格式
					inTags = true
				}
			case "cover", "thumbnail":
				parsed.Cover = value
			case "description", "excerpt":
				parsed.Summary = value
			case "slug", "abbrlink":
				parsed.Slug = value
			}
		}
	}

	// 添加收集的标签
	if len(tagLines) > 0 {
		parsed.Tags = append(parsed.Tags, tagLines...)
	}

	// 验证必需字段
	if parsed.Title == "" {
		return nil, fmt.Errorf("文章缺少标题")
	}

	// 如果没有摘要，从内容中生成
	if parsed.Summary == "" {
		parsed.Summary = generateSummary(parsed.Content, 200)
	}

	return parsed, nil
}

// parseHexoDate 解析Hexo日期格式
func parseHexoDate(dateStr string) (*time.Time, error) {
	// 支持多种日期格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("无法解析日期: %s", dateStr)
}

// generateSummary 从内容生成摘要
func generateSummary(content string, maxLen int) string {
	// 移除Markdown标记
	content = strings.ReplaceAll(content, "#", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "`", "")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.TrimSpace(content)

	// 截取指定长度
	runes := []rune(content)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return content
}

// ============ 微信公众号导出 ============

// ExportToWeChatDraft 导出文章到微信公众号草稿箱
func (s *ArticleService) ExportToWeChatDraft(ctx context.Context, id uint) (*dto.WeChatExportResult, error) {
	// 检查微信配置
	if s.config.WeChat.AppID == "" || s.config.WeChat.AppSecret == "" {
		return nil, fmt.Errorf("微信公众号配置不完整，请先在设置中配置 AppID 和 AppSecret")
	}

	// 获取文章
	article, err := s.articleRepo.Get(id)
	if err != nil {
		return nil, fmt.Errorf("文章不存在: %w", err)
	}

	// 创建微信客户端
	client, err := wechatmp.NewClient(wechatmp.Config{
		AppID:     s.config.WeChat.AppID,
		AppSecret: s.config.WeChat.AppSecret,
		BaseURL:   s.config.WeChat.TokenURL,
	})
	if err != nil {
		return nil, fmt.Errorf("创建微信客户端失败: %w", err)
	}

	// 预处理 Markdown 扩展语法
	processedContent := wechatmp.ConvertCustomBlocks(article.Content)
	processedContent = wechatmp.ConvertLinksToFootnotes(processedContent)
	processedContent = wechatmp.PreprocessMarkdown(processedContent)

	// 渲染 Markdown 为 HTML
	var htmlBuf bytes.Buffer
	if err := s.md.Convert([]byte(processedContent), &htmlBuf); err != nil {
		return nil, fmt.Errorf("渲染 Markdown 失败: %w", err)
	}

	// 转换为微信兼容格式
	result, err := wechatmp.ConvertMarkdownToWeChatHTML(htmlBuf.String())
	if err != nil {
		return nil, fmt.Errorf("转换 HTML 失败: %w", err)
	}

	htmlContent := result.HTML

	// 处理图片上传
	var uploadErrors []string
	for _, img := range result.Images {
		newURL, err := s.uploadImageToWeChat(ctx, client, img.OriginalURL)
		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Sprintf("图片 %s 上传失败: %v", img.OriginalURL, err))
			continue
		}
		htmlContent = wechatmp.ReplaceImageURL(htmlContent, img.OriginalURL, newURL)
	}

	// 处理封面图（微信草稿必须有封面）
	coverURL := article.Cover
	if coverURL == "" {
		// 使用 Bing 每日一图作为默认封面
		coverURL = "https://api.pearktrue.cn/api/bing/"
	}
	thumbMediaID, err := s.uploadCoverToWeChat(ctx, client, coverURL)
	if err != nil {
		return nil, fmt.Errorf("封面图上传失败: %w", err)
	}

	// 获取作者名称（从基本配置）
	author := s.config.Basic.Author
	if author == "" {
		author = s.config.Blog.Title
	}

	// 构建草稿
	draftArticle := wechatmp.DraftArticle{
		Title:            article.Title,
		Author:           author,
		Content:          htmlContent,
		Digest:           truncateString(article.Summary, 120),
		ContentSourceURL: s.buildArticleURL(article),
		ThumbMediaID:     thumbMediaID,
		NeedOpenComment:  1,
	}

	// 创建草稿
	draftResult, err := client.CreateDraft(ctx, []wechatmp.DraftArticle{draftArticle})
	if err != nil {
		return nil, fmt.Errorf("创建草稿失败: %w", err)
	}

	return &dto.WeChatExportResult{
		MediaID:  draftResult.MediaID,
		Warnings: uploadErrors,
	}, nil
}

// GetWeChatHTML 获取文章的微信公众号 HTML 格式
func (s *ArticleService) GetWeChatHTML(_ context.Context, id uint) (string, error) {
	// 获取文章
	article, err := s.articleRepo.Get(id)
	if err != nil {
		return "", fmt.Errorf("文章不存在: %w", err)
	}

	// 预处理 Markdown 扩展语法
	// 处理顺序：自定义块 → 链接转脚注 → 其他扩展语法
	processedContent := wechatmp.ConvertCustomBlocks(article.Content)
	processedContent = wechatmp.ConvertLinksToFootnotes(processedContent)
	processedContent = wechatmp.PreprocessMarkdown(processedContent)

	// 渲染 Markdown 为 HTML
	var htmlBuf bytes.Buffer
	if err := s.md.Convert([]byte(processedContent), &htmlBuf); err != nil {
		return "", fmt.Errorf("渲染 Markdown 失败: %w", err)
	}

	// 转换为微信兼容格式
	result, err := wechatmp.ConvertMarkdownToWeChatHTML(htmlBuf.String())
	if err != nil {
		return "", fmt.Errorf("转换 HTML 失败: %w", err)
	}

	return result.HTML, nil
}

// uploadImageToWeChat 上传文章内图片到微信
func (s *ArticleService) uploadImageToWeChat(ctx context.Context, client *wechatmp.Client, imgURL string) (string, error) {
	data, filename, err := s.fetchImage(ctx, imgURL)
	if err != nil {
		return "", err
	}

	result, err := client.UploadImage(ctx, filename, data)
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

// uploadCoverToWeChat 上传封面图到微信素材库
func (s *ArticleService) uploadCoverToWeChat(ctx context.Context, client *wechatmp.Client, coverURL string) (string, error) {
	data, contentType, err := s.fetchImage(ctx, coverURL)
	if err != nil {
		return "", fmt.Errorf("下载封面图失败: %w", err)
	}

	// 微信 image 类型素材限制
	const maxImageSize = 10 * 1024 * 1024 // 10MB
	if len(data) > maxImageSize {
		return "", fmt.Errorf("封面图片过大（%d MB），微信限制最大 10MB", len(data)/1024/1024)
	}

	// 根据 Content-Type 确定文件扩展名
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/jpeg", "image/jpg":
		ext = ".jpg"
	}

	result, err := client.AddThumbMaterial(ctx, "cover"+ext, data)
	if err != nil {
		return "", fmt.Errorf("上传封面到微信失败: %w", err)
	}

	return result.MediaID, nil
}

// fetchImage 下载图片
func (s *ArticleService) fetchImage(ctx context.Context, imgURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	filename := path.Base(imgURL)
	if filename == "" || filename == "." || filename == "/" {
		filename = "image.jpg"
	}

	return data, filename, nil
}

// buildArticleURL 构建文章链接
func (s *ArticleService) buildArticleURL(article *model.Article) string {
	if s.config.Basic.BlogURL != "" {
		return s.config.Basic.BlogURL + "/posts/" + article.Slug
	}
	return ""
}

// truncateString 截断字符串
func truncateString(str string, maxLen int) string {
	runes := []rune(str)
	if len(runes) <= maxLen {
		return str
	}
	return string(runes[:maxLen-3]) + "..."
}
