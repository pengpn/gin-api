package v1

import (
	"gin-api/models"
	"gin-api/pkg/app"
	"gin-api/pkg/e"
	"gin-api/pkg/logging"
	"gin-api/pkg/setting"
	"gin-api/pkg/util"
	"gin-api/service/article_service"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
	"net/http"
)

// 获取单个文章
func GetArticle(c *gin.Context) {
	appG := app.Gin{c}
	id := com.StrTo(c.Param("id")).MustInt()
	valid := validation.Validation{}
	valid.Min(id, 1, "id").Message("ID必须大于0")

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusOK, e.INVALID_PARAMS, nil)
		return
	}

	articleService := article_service.Article{ID: id}
	exists, err := articleService.ExistByID()
	if err != nil {
		appG.Response(http.StatusOK, e.ERROR_CHECK_EXIST_ARTICLE_FAIL, nil)
		return
	}
	if !exists {
		appG.Response(http.StatusOK, e.ERROR_NOT_EXIST_ARTICLE, nil)
		return
	}

	article, err := articleService.Get()
	if err != nil {
		appG.Response(http.StatusOK, e.ERROR_GET_ARTICLE_FAIL, nil)
		return
	}

	appG.Response(http.StatusOK, e.SUCCESS, article)

}

// 获取多个文章
func GetArticles(c *gin.Context) {
	appG := app.Gin{C: c}
	data := make(map[string]interface{})
	valid := validation.Validation{}

	state := -1
	if arg := c.PostForm("state"); arg != "" {
		state = com.StrTo(arg).MustInt()
		valid.Range(state, 0, 1, "state")
	}

	tagId := -1
	if arg := c.PostForm("tag_id"); arg != "" {
		tagId = com.StrTo(arg).MustInt()
		valid.Min(tagId, 1, "tag_id")
	}

	if valid.HasErrors() {
		app.MarkErrors(valid.Errors)
		appG.Response(http.StatusBadRequest, e.INVALID_PARAMS, nil)
		return
	}

	articleService := article_service.Article{
		TagID:    tagId,
		State:    state,
		PageNum:  util.GetPage(c),
		PageSize: setting.AppSetting.PageSize,
	}

	total, err := articleService.Count()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_COUNT_ARTICLE_FAIL, nil)
		return
	}

	articles, err := articleService.GetAll()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_GET_ARTICLES_FAIL, nil)
		return
	}

	data = make(map[string]interface{})
	data["lists"] = articles
	data["total"] = total

	appG.Response(http.StatusOK, e.SUCCESS, data)

}

//新增文章
func AddArticle(c *gin.Context) {
	tagId := com.StrTo(c.Query("tag_id")).MustInt()
	title := c.Query("title")
	desc := c.Query("desc")
	content := c.Query("content")
	createdBy := c.Query("created_by")
	coverImageUrl := c.Query("cover_image_url")
	state := com.StrTo(c.DefaultQuery("state", "0")).MustInt()

	valid := validation.Validation{}
	valid.Min(tagId, 1, "tag_id").Message("标签ID必须大于0")
	valid.Required(title, "title").Message("标题不能为空")
	valid.Required(desc, "desc").Message("简述不能为空")
	valid.Required(content, "content").Message("内容不能为空")
	valid.Required(createdBy, "created_by").Message("创建人不能为空")
	valid.Required(coverImageUrl, "cover_image_url").Message("图片封面不能为空")
	valid.Range(state, 0, 1, "state").Message("状态只允许0或1")
	valid.MaxSize(coverImageUrl, 255, "cover_image_url").Message("图片封面最长长度只能是255")

	code := e.INVALID_PARAMS
	if !valid.HasErrors() {
		if models.ExistTagByID(tagId) {
			data := make(map[string]interface{})
			data["tag_id"] = tagId
			data["title"] = title
			data["desc"] = desc
			data["content"] = content
			data["created_by"] = createdBy
			data["state"] = state
			data["cover_image_url"] = coverImageUrl

			models.AddArticle(data)
			code = e.SUCCESS
		} else {
			code = e.ERROR_NOT_EXIST_TAG
		}
	} else {
		for _, err := range valid.Errors {
			logging.Info("err.key: %s, err.message: %s", err.Key, err.Message)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": make(map[string]interface{}),
	})
}

//修改文章
//func EditArticle(c *gin.Context) {
//	valid := validation.Validation{}
//
//	id := com.StrTo(c.Param("id")).MustInt()
//	tagId := com.StrTo(c.Query("tag_id")).MustInt()
//	title := c.Query("title")
//	desc := c.Query("desc")
//	content := c.Query("content")
//	modifiedBy := c.Query("modified_by")
//	coverImageUrl := c.Query("cover_image_url")
//
//	var state int = -1
//	if arg := c.Query("state"); arg != "" {
//		state = com.StrTo(arg).MustInt()
//		valid.Range(state, 0, 1, "state").Message("状态只允许0或1")
//	}
//
//	valid.Min(id, 1, "id").Message("ID必须大于0")
//	valid.MaxSize(title, 100, "title").Message("标题最长为100字符")
//	valid.MaxSize(desc, 255, "desc").Message("简述最长为255字符")
//	valid.MaxSize(content, 65535, "content").Message("内容最长为65535字符")
//	valid.Required(modifiedBy, "modified_by").Message("修改人不能为空")
//	valid.Required(coverImageUrl, "cover_image_url").Message("图片地址不能为空")
//	valid.MaxSize(modifiedBy, 100, "modified_by").Message("修改人最长为100字符")
//	valid.MaxSize(coverImageUrl, 255, "cover_image_url").Message("图片长度最长为255字符")
//
//	code := e.INVALID_PARAMS
//	if !valid.HasErrors() {
//		if models.ExistArticleByID(id) {
//			if models.ExistTagByID(tagId) {
//				data := make(map[string]interface{})
//				if tagId > 0 {
//					data["tag_id"] = tagId
//				}
//				if title != "" {
//					data["title"] = title
//				}
//				if desc != "" {
//					data["desc"] = desc
//				}
//				if content != "" {
//					data["content"] = content
//				}
//
//				if coverImageUrl != "" {
//					data["cover_image_url"] = coverImageUrl
//				}
//
//				data["modified_by"] = modifiedBy
//
//				models.EditArticle(id, data)
//				code = e.SUCCESS
//			} else {
//				code = e.ERROR_NOT_EXIST_TAG
//			}
//		} else {
//			code = e.ERROR_NOT_EXIST_ARTICLE
//		}
//	} else {
//		for _, err := range valid.Errors {
//			logging.Info("err.key: %s, err.message: %s", err.Key, err.Message)
//		}
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code": code,
//		"msg":  e.GetMsg(code),
//		"data": make(map[string]string),
//	})
//}

//删除文章
//func DeleteArticle(c *gin.Context) {
//	id := com.StrTo(c.Param("id")).MustInt()
//
//	valid := validation.Validation{}
//	valid.Min(id, 1, "id").Message("ID必须大于0")
//
//	code := e.INVALID_PARAMS
//	if !valid.HasErrors() {
//		if models.ExistArticleByID(id) {
//			models.DeleteArticle(id)
//			code = e.SUCCESS
//		} else {
//			code = e.ERROR_NOT_EXIST_ARTICLE
//		}
//	} else {
//		for _, err := range valid.Errors {
//			logging.Info("err.key: %s, err.message: %s", err.Key, err.Message)
//		}
//	}
//
//	c.JSON(http.StatusOK, gin.H{
//		"code": code,
//		"msg":  e.GetMsg(code),
//		"data": make(map[string]string),
//	})
//}
