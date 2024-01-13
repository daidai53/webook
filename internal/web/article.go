// Copyright@daidai53 2023
package web

import (
	"github.com/daidai53/webook/internal/domain"
	"github.com/daidai53/webook/internal/service"
	"github.com/daidai53/webook/internal/web/jwt"
	"github.com/daidai53/webook/pkg/ginx"
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
	topSvc   service.TopArticlesService
	l        logger.LoggerV1
}

func NewArticleHandler(l logger.LoggerV1, svc service.ArticleService, interSvc service.InteractiveService,
	topSvc service.TopArticlesService) *ArticleHandler {
	return &ArticleHandler{
		svc:      svc,
		interSvc: interSvc,
		topSvc:   topSvc,
		l:        l,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
	g.POST("/publish", ginx.WrapBodyAndClaims(h.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndClaims(h.Withdraw))

	// 创作者接口
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", h.List)

	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true就是点赞，false就是取消点赞
	pub.POST("/like", ginx.WrapBodyAndClaims(h.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
	pub.POST("/top", ginx.WrapBody(h.TopArticles))
}

// Edit 接受 Article 输入，返回一个ID，文章的ID
func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq,
	uc jwt.UserClaim) (ginx.Result, error) {
	artId, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: artId,
	}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req ArticlePubReq,
	uc jwt.UserClaim) (ginx.Result, error) {
	artId, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: artId,
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq,
	uc jwt.UserClaim) (ginx.Result, error) {
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaim)
	arts, err := h.svc.GetByAuthor(ctx.Request.Context(), uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败", logger.Int64("uid", uc.Uid), logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit), logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
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
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "id 参数错误",
		})
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
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
		ctx.JSON(http.StatusOK, ginx.Result{
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
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: vo,
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
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
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查询文章失败,系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uid),
			logger.Error(err))
		return
	}
	//err = h.interSvc.IncrReadCnt(ctx, "article", art.BizId)
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
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: vo,
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req ArticleLikeReq,
	uc jwt.UserClaim) (ginx.Result, error) {
	var err error
	if req.Like {
		err = h.interSvc.Like(ctx, "article", req.Id, uc.Uid)
	} else {
		err = h.interSvc.CancelLike(ctx, "article", req.Id, uc.Uid)
	}
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req ArticleCollectReq,
	uc jwt.UserClaim) (ginx.Result, error) {
	err := h.interSvc.Collect(ctx, "article", req.Id, uc.Uid, req.Cid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, nil
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *ArticleHandler) TopArticles(ctx *gin.Context, req TopReq) (ginx.Result, error) {
	res, err := h.topSvc.GetTopArticles(ctx, req.N)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: res,
	}, nil

}

type Page struct {
	Limit  int
	Offset int
}
