package feeds

import (
	"encoding/xml"
	"fmt"
	"time"

	"flec_blog/config"
	"flec_blog/internal/dto"
	"flec_blog/internal/service"

	"github.com/gin-gonic/gin"
)

// AtomController Atom控制器
type AtomController struct {
	articleService *service.ArticleService
	config         *config.Config
}

// NewAtomController 创建Atom控制器
func NewAtomController(articleService *service.ArticleService, config *config.Config) *AtomController {
	return &AtomController{
		articleService: articleService,
		config:         config,
	}
}

// GetAtomFeed 获取Atom订阅
//
//	@Summary		Atom订阅
//	@Description	生成博客文章的Atom 1.0订阅源
//	@Tags			订阅
//	@Accept			json
//	@Produce		xml
//	@Success		200	{string}	string	"Atom XML 订阅内容"
//	@Router			/atom.xml [get]
func (c *AtomController) GetAtomFeed(ctx *gin.Context) {
	// 获取所有已发布文章
	req := &dto.ListArticlesRequest{
		Page:     1,
		PageSize: 0,
	}

	articles, _, err := c.articleService.ListForWeb(ctx.Request.Context(), req)
	if err != nil {
		ctx.XML(500, gin.H{"error": "生成Atom失败"})
		return
	}

	// 获取网站URL配置
	baseURL := c.config.Basic.BlogURL

	// 获取网站标题
	siteName := c.config.Blog.Title

	// 构建Atom Feed
	atom := &Atom{
		XMLNS:   "http://www.w3.org/2005/Atom",
		Title:   siteName,
		ID:      baseURL,
		Updated: time.Now().Format(time.RFC3339),
		Author: &AtomAuthor{
			Name: siteName,
		},
		Link: []AtomLink{
			{Href: baseURL, Rel: "alternate"},
			{Href: fmt.Sprintf("%s/atom.xml", baseURL), Rel: "self"},
		},
		Entries: make([]AtomEntry, 0, len(articles)),
	}

	// 转换文章为Atom Entry
	for _, article := range articles {
		// 构建文章URL（article.URL 以 /posts/ 开头）
		articleURL := baseURL + article.URL

		// 构建链接数组，包含文章链接和封面
		links := []AtomLink{{Href: articleURL}}

		// 添加封面链接
		if article.Cover != "" {
			links = append(links, AtomLink{
				Href: article.Cover,
				Rel:  "enclosure",
			})
		}

		entry := AtomEntry{
			Title:   article.Title,
			ID:      articleURL,
			Link:    links,
			Summary: article.Summary,
		}

		// 添加发布时间和更新时间
		if article.PublishTime != nil && !article.PublishTime.IsZero() {
			entry.Published = article.PublishTime.Time.Format(time.RFC3339)
			entry.Updated = article.PublishTime.Time.Format(time.RFC3339)
		}

		if article.UpdateTime != nil && !article.UpdateTime.IsZero() {
			entry.Updated = article.UpdateTime.Time.Format(time.RFC3339)
		}

		// 添加分类
		if article.Category.Name != "" {
			entry.Category = []AtomCategory{{Term: article.Category.Name}}
		}

		atom.Entries = append(atom.Entries, entry)
	}

	// 手动构建XML，包含XML声明
	xmlData, err := xml.MarshalIndent(atom, "", "  ")
	if err != nil {
		ctx.XML(500, gin.H{"error": "生成Atom失败"})
		return
	}

	// 构建完整的XML文档（包含声明）
	xmlContent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + string(xmlData)

	// 设置正确的响应头并写入响应
	ctx.Header("Content-Type", "application/xml; charset=utf-8")
	ctx.String(200, xmlContent)
}

// Atom Feed 结构定义
type Atom struct {
	XMLName xml.Name    `xml:"feed"`
	XMLNS   string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Author  *AtomAuthor `xml:"author,omitempty"`
	Link    []AtomLink  `xml:"link"`
	Entries []AtomEntry `xml:"entry"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

type AtomEntry struct {
	Title     string         `xml:"title"`
	ID        string         `xml:"id"`
	Updated   string         `xml:"updated"`
	Published string         `xml:"published,omitempty"`
	Link      []AtomLink     `xml:"link"`
	Summary   string         `xml:"summary"`
	Category  []AtomCategory `xml:"category,omitempty"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

type AtomCategory struct {
	Term string `xml:"term,attr"`
}
