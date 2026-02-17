package repo

import (
	"scratch/core/domain"

	"github.com/gin-gonic/gin"
)

type ScratchRepoInterface interface {
	CreateScratch(ctx *gin.Context, req *domain.Scratch) error
	FetchScratch(ctx *gin.Context, name string) (*domain.Scratch, error)
}
