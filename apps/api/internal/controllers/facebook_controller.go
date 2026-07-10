package controllers

import (
	"net/http"
	"net/url"

	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

const facebookStateCookie = "linora_facebook_oauth_state"

type FacebookController struct {
	service *services.FacebookService
}

func NewFacebookController(service *services.FacebookService) *FacebookController {
	return &FacebookController{service: service}
}

func (c *FacebookController) Begin(cxt *gin.Context) {
	if !c.service.Configured() {
		cxt.JSON(http.StatusServiceUnavailable, gin.H{"error": "Facebook Login ยังไม่ได้ตั้งค่าในระบบ"})
		return
	}

	state, err := services.SecureToken()
	if err != nil {
		cxt.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่ม Facebook Login ได้"})
		return
	}
	cxt.SetSameSite(http.SameSiteLaxMode)
	cxt.SetCookie(facebookStateCookie, state, int(services.FacebookHandoffLifetime.Seconds()), "/api/facebook", "", services.IsHTTPS(c.service.RedirectURI()), true)
	cxt.Redirect(http.StatusFound, c.service.AuthorizationURL(state))
}

func (c *FacebookController) Callback(cxt *gin.Context) {
	if cxt.Query("error") != "" {
		c.redirectWithError(cxt, "Facebook Login ถูกยกเลิกหรือไม่สำเร็จ")
		return
	}

	state, err := cxt.Cookie(facebookStateCookie)
	if err != nil || state == "" || state != cxt.Query("state") {
		c.redirectWithError(cxt, "ไม่สามารถยืนยัน Facebook Login ได้")
		return
	}
	cxt.SetSameSite(http.SameSiteLaxMode)
	cxt.SetCookie(facebookStateCookie, "", -1, "/api/facebook", "", services.IsHTTPS(c.service.RedirectURI()), true)

	code := cxt.Query("code")
	if code == "" {
		c.redirectWithError(cxt, "Facebook ไม่ส่งรหัสยืนยันกลับมา")
		return
	}
	handoff, err := c.service.CompleteLogin(cxt.Request.Context(), code)
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
	pages, err := c.service.RedeemHandoff(code)
	if err != nil {
		cxt.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	cxt.JSON(http.StatusOK, gin.H{"pages": pages})
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
