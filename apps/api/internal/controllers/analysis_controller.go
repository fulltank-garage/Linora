package controllers

import (
	"net/http"

	"github.com/fulltank-garage/linora/apps/api/internal/models"
	"github.com/fulltank-garage/linora/apps/api/internal/services"
	"github.com/gin-gonic/gin"
)

type AnalysisController struct {
	service *services.AnalysisService
}

func NewAnalysisController(service *services.AnalysisService) *AnalysisController {
	return &AnalysisController{service: service}
}

func (c *AnalysisController) Manual(cxt *gin.Context) {
	var input models.ManualAnalysisInput
	if err := cxt.ShouldBindJSON(&input); err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": "รูปแบบข้อมูลไม่ถูกต้อง"})
		return
	}

	report, err := c.service.AnalyzeManualInput(input)
	if err != nil {
		cxt.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cxt.JSON(http.StatusOK, gin.H{"report": report})
}
