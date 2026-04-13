package mcp

import (
	"net/http"
	"strings"

	"flec_blog/internal/service"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type publicServer struct {
	articleService  *service.ArticleService
	categoryService *service.CategoryService
	tagService      *service.TagService
	commentService  *service.CommentService
	friendService   *service.FriendService
	rssFeedService  *service.RssFeedService
	momentService   *service.MomentService
	userService     *service.UserService
	statsService    *service.StatsService
}

func NewPublicHandler(
	articleService *service.ArticleService,
	categoryService *service.CategoryService,
	tagService *service.TagService,
	commentService *service.CommentService,
	friendService *service.FriendService,
	rssFeedService *service.RssFeedService,
	momentService *service.MomentService,
	userService *service.UserService,
	statsService *service.StatsService,
) http.Handler {
	implVersion := strings.TrimSpace(service.AppVersion)
	if implVersion == "" {
		implVersion = "dev"
	}

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "flecblog-public",
		Version: implVersion,
	}, nil)

	s := &publicServer{
		articleService:  articleService,
		categoryService: categoryService,
		tagService:      tagService,
		commentService:  commentService,
		friendService:   friendService,
		rssFeedService:  rssFeedService,
		momentService:   momentService,
		userService:     userService,
		statsService:    statsService,
	}

	// 注册 tools
	s.registerTools(server)

	return sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return server
	}, nil)
}
