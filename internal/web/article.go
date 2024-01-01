// Copyright@daidai53 2023
package web

import (
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc      service.ArticleService
	interSvc service.InteractiveService
	l        logger.LoggerV1
}

func NewArticleHandler(l logger.LoggerV1, svc service.ArticleService, interSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		l:        l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)

	// 创作者接口
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", h.List)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true就是点赞，false就是取消点赞
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)
}

// Edit 接受 Article 输入，返回一个ID，文章的ID
func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uid := ctx.MustGet("user-id").(int64)
	artId, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.l.Error("保存文章数据失败", logger.Error(err), logger.Int64("uid", uid))
	}
	ctx.JSON(http.StatusOK, Result{
		Data: artId,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uid := ctx.MustGet("user-id").(int64)
	artId, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("发表文章数据失败", logger.Error(err), logger.Int64("uid", uid))
	}
	ctx.JSON(http.StatusOK, Result{
		Data: artId,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uid := ctx.MustGet("user-id").(int64)
	err := h.svc.Withdraw(ctx, uid, req.Id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("撤回文章权限失败", logger.Error(err), logger.Int64("uid", uid), logger.Int64("aid", req.Id))
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaim)
	arts, err := h.svc.GetByAuhtor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败", logger.Int64("uid", uc.Uid), logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit), logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:         src.Id,
				Title:      src.Title,
				Abstract:   src.Abstract(),
				Content:    src.Content,
				AuthorId:   src.Author.Id,
				AuthorName: src.Author.Name,
				Status:     src.Status.ToUint8(),
				CTime:      src.CTime.Format(time.DateTime),
				UTime:      src.UTime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "id 参数错误",
		})
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败,id 格式不对",
			logger.String("id", idstr),
			logger.Error(err),
		)
		return
	}
	uid := ctx.MustGet("user-id").(int64)
	if art.Author.Id != uid {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("非法查询文章",
			logger.String("id", idstr),
			logger.Int64("uid", uid),
		)
		return
	}
	vo := ArticleVo{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Status:  art.Status.ToUint8(),
		CTime:   art.CTime.Format(time.DateTime),
		UTime:   art.UTime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{
		Data: vo,
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "id 参数错误",
		})
		return
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	uid := ctx.MustGet("user-id").(int64)
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id, uid)
		return er
	})

	eg.Go(func() error {
		var er error
		intr, er = h.interSvc.Get(ctx, "article", id, uid)
		return er
	})

	err = eg.Wait()

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败,系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uid),
			logger.Error(err))
		return
	}
	//err = h.interSvc.IncrReadCnt(ctx, "article", art.Id)
	//if err != nil {
	//	h.l.Error("更新阅读数失败",
	//		logger.Int64("aid", id),
	//		logger.Error(err))
	//}
	vo := ArticleVo{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,

		AuthorId:   art.Author.Id,
		AuthorName: art.Author.Name,

		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,

		Status: art.Status.ToUint8(),
		CTime:  art.CTime.Format(time.DateTime),
		UTime:  art.UTime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, Result{
		Data: vo,
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uid := ctx.MustGet("user-id").(int64)
	var err error
	if req.Like {
		err = h.interSvc.Like(ctx, "article", req.Id, uid)
	} else {
		err = h.interSvc.CancelLike(ctx, "article", req.Id, uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("aid", req.Id),
			logger.Int64("uid", uid),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uid := ctx.MustGet("user-id").(int64)

	err := h.interSvc.Collect(ctx, "article", req.Id, uid, req.Cid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("收藏失败",
			logger.Error(err),
			logger.Int64("aid", req.Id),
			logger.Int64("uid", uid),
		)
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

type Page struct {
	Limit  int
	Offset int
}
