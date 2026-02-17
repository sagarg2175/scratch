 package handler

// import (
// 	"bytes"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"scratch/repo"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang/mock/gomock"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCreateScratchHandler(t *testing.T) {

// 	gin.SetMode(gin.TestMode)

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	// Step 2a: Create a mock repository
// 	mockRepo := repo.NewMockScratchRepoInterface(ctrl)

// 	// Step 2b: Create the handler with mock repo
// 	sh := NewScratchHandler(mockRepo)

// 	// Step 2c: Prepare test request
// 	body := `{"name":"TestScratch"}`
// 	req := httptest.NewRequest("POST", "/api", bytes.NewBufferString(body))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()

// 	// Step 2d: Set expectations on the mock
// 	mockRepo.EXPECT().
// 		CreateScratch(gomock.Any(), gomock.Any()).
// 		Return(nil) // simulate successful creation

// 	// Step 2e: Create a gin context
// 	c, _ := gin.CreateTestContext(w)
// 	c.Request = req

// 	// Step 2f: Call handler
// 	sh.CreateScratchHandler(c)

// 	// Step 2g: Assert response
// 	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
// 	assert.Contains(t, w.Body.String(), `"success":true`)
// }
