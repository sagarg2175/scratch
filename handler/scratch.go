package handler

import (
	"net/http"
	"scratch/core/domain"
	repo "scratch/repo"
	"scratch/tocken"
	"scratch/utils"
	"time"

	"github.com/gin-gonic/gin"
)

type ScratchHandler struct {
	svc         repo.ScratchRepoInterface
	TockenMaker tocken.Maker
}

func NewScratchHandler(svc repo.ScratchRepoInterface, TockenMaker tocken.Maker) *ScratchHandler {
	return &ScratchHandler{
		svc,
		TockenMaker,
	}
}

type CreateRequest struct {
	Name     string `json:"name" validate:"required,customName"`
	Password string `json:"password" validate:"required,min=8"`
}

func (sh *ScratchHandler) CreateScratchHandler(ctx *gin.Context) {

	var req CreateRequest

	err := ctx.ShouldBind(&req)
	if err != nil {
		return
	}

	if ok := handleValidation(ctx, &req); !ok {
		return
	}

	Hashedpassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return
	}

	batch := domain.Scratch{
		Name:     &req.Name,
		Password: &Hashedpassword,
	}

	err = sh.svc.CreateScratch(ctx, &batch)
	if err != nil {
		return
	}

	// tocken logic

	tocken, err := sh.TockenMaker.CreateTocken(req.Name, time.Hour)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": []string{"Failed to generate token"},
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": []string{"User created successfully"},
		"token":   tocken,
	})

}

type FetchRequest struct {
	Name string `uri:"name" validate:"required,customName"`
}

func (sh *ScratchHandler) FetchScratchHandler(ctx *gin.Context) {
	var req FetchRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": []string{"Invalid request"},
			"error":   err.Error(),
		})
		return
	}

	if ok := handleValidation(ctx, &req); !ok {
		return
	}

	data, err := sh.svc.FetchScratch(ctx, req.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": []string{"Failed to fetch data"},
			"error":   err.Error(),
		})
		return
	}

	// ✅ Get token payload (from middleware)
	authPayload, ok := ctx.Get(authorizationPayloadKey)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": []string{"Missing auth payload"},
		})
		return
	}

	payload := authPayload.(*tocken.Payload)

	// ✅ Check if user is authorized for that name
	if req.Name != payload.Username {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": []string{"Account is unauthorized for you"},
		})
		return
	}

	// ✅ Success
	handleSuccess(ctx, data)
}
