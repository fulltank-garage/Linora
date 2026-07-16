package controllers

import (
	"net/http"
	"net/url"

	"github.com/fulltank-garage/linora/apps/api/internal/middleware"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

type FacebookController struct {
	page    *services.PageService
	service *services.FacebookService
}

func NewFacebookController(service *services.FacebookService, page *services.PageService) *FacebookController {
	return &FacebookController{page: page, service: service}
}

func (c *FacebookController) Begin(cxt *gin.Context) {
	if !c.service.Configured() {
		cxt.JSON(http.StatusServiceUnavailable, gin.H{"error": "Facebook Login ยังไม่ได้ตั้งค่าในระบบ"})
		return
	}

	authorizationURL, err := c.service.StartAuthorization(cxt.Request.Context(), middleware.LineUserID(cxt))
	if err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่ม Facebook Login ได้"})
		return
	}
	cxt.JSON(http.StatusOK, gin.H{"authorizationUrl": authorizationURL})
}

func (c *FacebookController) Callback(cxt *gin.Context) {
	if cxt.Query("error") != "" {
		c.redirectWithError(cxt, "Facebook Login ถูกยกเลิกหรือไม่สำเร็จ")
		return
	}

	ownerID, err := c.service.ConsumeAuthorizationState(cxt.Request.Context(), cxt.Query("state"))
	if err != nil || ownerID == "" {
		c.redirectWithError(cxt, "ไม่สามารถยืนยัน Facebook Login ได้")
		return
	}
	code := cxt.Query("code")
	if code == "" {
		c.redirectWithError(cxt, "Facebook ไม่ส่งรหัสยืนยันกลับมา")
		return
	}
	handoff, err := c.service.CompleteLogin(cxt.Request.Context(), code, ownerID)
	if err != nil {
		c.redirectWithError(cxt, "ไม่สามารถเชื่อมต่อหรืออ่านรายการ Facebook Page ได้")
		return
	}

	redirectURL, err := url.Parse(c.service.AppURL())
	if err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "ตั้งค่า APP_URL ไม่ถูกต้อง"})
		return
	}
	redirectURL.Path = "/connect-facebook"
	query := redirectURL.Query()
	query.Set("facebook_connect", handoff)
	redirectURL.RawQuery = query.Encode()
	cxt.Redirect(http.StatusFound, redirectURL.String())
}

func (c *FacebookController) Session(cxt *gin.Context) {
	code := cxt.Query("code")
	if code == "" {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": "ไม่พบรหัสยืนยัน Facebook Login"})
		return
	}
	pages, err := c.service.RedeemHandoff(code, middleware.LineUserID(cxt))
	if err != nil {
		cxt.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	cxt.JSON(http.StatusOK, gin.H{"pages": pages})
}

func (c *FacebookController) DataDeletion(cxt *gin.Context) {
	if c.page == nil {
		cxt.JSON(http.StatusServiceUnavailable, gin.H{"error": "data deletion is unavailable"})
		return
	}
	facebookUserID, err := c.service.VerifyDataDeletionRequest(cxt.PostForm("signed_request"))
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": "invalid data deletion request"})
		return
	}
	if err := c.page.DeleteFacebookUserData(cxt.Request.Context(), facebookUserID); err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user data"})
		return
	}
	confirmationCode, err := services.SecureToken()
	if err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "could not confirm data deletion"})
		return
	}
	statusURL, err := url.Parse(c.service.AppURL())
	if err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "invalid app URL"})
		return
	}
	statusURL.Path = "/data-deletion"
	query := statusURL.Query()
	query.Set("confirmation_code", confirmationCode)
	statusURL.RawQuery = query.Encode()
	cxt.JSON(http.StatusOK, gin.H{"url": statusURL.String(), "confirmation_code": confirmationCode})
}

func (c *FacebookController) redirectWithError(cxt *gin.Context, message string) {
	redirectURL, err := url.Parse(c.service.AppURL())
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": message})
		return
	}
	redirectURL.Path = "/connect-facebook"
	query := redirectURL.Query()
	query.Set("facebook_error", message)
	redirectURL.RawQuery = query.Encode()
	cxt.Redirect(http.StatusFound, redirectURL.String())
}
