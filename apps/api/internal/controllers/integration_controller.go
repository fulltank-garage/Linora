package controllers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/fulltank-garage/linora/apps/api/internal/middleware"
	"github.com/fulltank-garage/linora/apps/api/internal/repositories"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

type IntegrationController struct {
	line       *services.LineService
	lineSecret string
	page       *services.PageService
}

func NewIntegrationController(page *services.PageService, line *services.LineService, lineSecret string) *IntegrationController {
	return &IntegrationController{line: line, lineSecret: lineSecret, page: page}
}

func (c *IntegrationController) ConnectPage(ctx *gin.Context) {
	var input struct {
		HandoffCode string `json:"handoffCode"`
		PageID      string `json:"pageId"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil || input.HandoffCode == "" || input.PageID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ handoffCode และ pageId"})
		return
	}
	result, err := c.page.Connect(ctx.Request.Context(), middleware.LineUserID(ctx), input.HandoffCode, input.PageID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถเชื่อมต่อและวิเคราะห์เพจได้: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, result)
}

func (c *IntegrationController) ListPages(ctx *gin.Context) {
	pages, err := c.page.List(ctx.Request.Context(), middleware.LineUserID(ctx))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดรายการเพจที่อนุญาตได้"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"pages": pages})
}

func (c *IntegrationController) SelectPage(ctx *gin.Context) {
	result, err := c.page.Select(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถเลือกและวิเคราะห์เพจได้ กรุณาเชื่อมต่อ Facebook ใหม่"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}

func (c *IntegrationController) SyncPage(ctx *gin.Context) {
	report, err := c.page.Sync(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถซิงก์ข้อมูลเพจได้: " + err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"report": report})
}

func (c *IntegrationController) LatestReport(ctx *gin.Context) {
	report, err := c.page.LatestReport(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID"))
	if err != nil {
		status := http.StatusInternalServerError
		if err == repositories.ErrNotFound {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{"error": "ยังไม่พบรายงานของเพจ"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"report": report})
}

func (c *IntegrationController) DeletePage(ctx *gin.Context) {
	if err := c.page.Delete(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลเพจได้"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IntegrationController) DisconnectPage(ctx *gin.Context) {
	if err := c.page.Disconnect(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID")); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถยกเลิกการเชื่อมต่อเพจได้"})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IntegrationController) CreateLineLinkCode(ctx *gin.Context) {
	code, err := c.page.CreateLineLinkCode(ctx.Request.Context(), middleware.LineUserID(ctx), ctx.Param("pageID"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ไม่สามารถสร้างรหัสเชื่อม LINE ได้"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"code": code, "expiresInSeconds": 600})
}

func (c *IntegrationController) ActivateDashboardRichMenu(ctx *gin.Context) {
	if err := c.line.LinkDashboardRichMenu(ctx.Request.Context(), middleware.LineUserID(ctx)); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "Unable to update the LINE menu."})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IntegrationController) ActivateConnectRichMenu(ctx *gin.Context) {
	if err := c.line.LinkConnectRichMenu(ctx.Request.Context(), middleware.LineUserID(ctx)); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "Unable to update the LINE menu."})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *IntegrationController) LocalLineChat(ctx *gin.Context) {
	var input struct {
		Message    string `json:"message"`
		LineUserID string `json:"lineUserId"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil || input.LineUserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ lineUserId และ message"})
		return
	}
	answer, err := c.line.Chat(ctx.Request.Context(), input.LineUserID, input.Message)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตอบข้อความได้"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"answer": answer})
}

func (c *IntegrationController) LinkLine(ctx *gin.Context) {
	var input struct {
		Code       string `json:"code"`
		LineUserID string `json:"lineUserId"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil || input.Code == "" || input.LineUserID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "ต้องระบุ lineUserId และ code"})
		return
	}
	pageID, err := c.line.Link(ctx.Request.Context(), input.LineUserID, input.Code)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "รหัสเชื่อม LINE ไม่ถูกต้องหรือหมดอายุแล้ว"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"pageId": pageID})
}

func (c *IntegrationController) LineWebhook(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil || !services.VerifyLineSignature(c.lineSecret, body, ctx.GetHeader("X-Line-Signature")) {
		ctx.Status(http.StatusUnauthorized)
		return
	}
	var payload struct {
		Events []struct {
			ReplyToken string `json:"replyToken"`
			Source     struct {
				UserID string `json:"userId"`
			} `json:"source"`
			Message struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"events"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	for _, event := range payload.Events {
		if event.Message.Type != "text" || event.Source.UserID == "" {
			continue
		}
		text := strings.TrimSpace(event.Message.Text)
		go c.replyToLineMessage(event.ReplyToken, event.Source.UserID, text)
		continue
		var answer string
		if strings.HasPrefix(strings.ToUpper(text), "LIN-") {
			if _, err := c.line.Link(ctx.Request.Context(), event.Source.UserID, text); err != nil {
				answer = "รหัสเชื่อมต่อไม่ถูกต้องหรือหมดอายุแล้วครับ"
			} else {
				answer = "เชื่อม LINE กับเพจเรียบร้อยแล้วครับ ถามผลวิเคราะห์ได้เลย"
			}
		} else {
			answer, _ = c.line.Chat(ctx.Request.Context(), event.Source.UserID, text)
		}
		_ = c.line.Reply(ctx.Request.Context(), event.ReplyToken, answer)
	}
	ctx.Status(http.StatusOK)
}

func (c *IntegrationController) replyToLineMessage(replyToken string, lineUserID string, text string) {
	requestCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	var answer string
	if strings.HasPrefix(strings.ToUpper(text), "LIN-") {
		if _, err := c.line.Link(requestCtx, lineUserID, text); err != nil {
			answer = "รหัสเชื่อมต่อไม่ถูกต้องหรือหมดอายุแล้ว กรุณาลองใหม่อีกครั้ง"
		} else {
			answer = "เชื่อมต่อ LINE กับเพจเรียบร้อยแล้ว คุณสามารถถามผลวิเคราะห์ได้เลย"
		}
	} else {
		var err error
		answer, err = c.line.Chat(requestCtx, lineUserID, text)
		if err != nil {
			log.Printf("LINE chat failed for %s: %v", lineUserID, err)
			answer = "ขออภัย ระบบยังตอบคำถามไม่ได้ในขณะนี้ กรุณาลองใหม่อีกครั้ง"
		}
	}

	if err := c.line.Reply(requestCtx, replyToken, answer); err != nil {
		log.Printf("LINE reply failed for %s: %v", lineUserID, err)
	}
}
