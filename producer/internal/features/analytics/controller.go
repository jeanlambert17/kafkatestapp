package analytics

import (
	"net/http"
	"time"

	"producer/internal/auth"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller { return &Controller{service: service} }

func (c *Controller) RegisterRoutes(r *gin.Engine) {
	g := r.Group("/analytics")
	g.GET("/daily-aggregates", c.GetDailyAggregates)
	g.GET("/popular-items", c.GetPopularItems)
}

func parseDate(ctx *gin.Context, key string) (time.Time, bool) {
	s := ctx.Query(key)
	if s == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": key + " is required (MM/DD/YYYY)"})
		return time.Time{}, false
	}
	// MM/DD/YYYY
	t, err := time.Parse("01/02/2006", s)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": key + " must be MM/DD/YYYY"})
		return time.Time{}, false
	}
	return t, true
}

func (c *Controller) GetDailyAggregates(ctx *gin.Context) {
	_, ridHex, err := auth.GetOrgID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	from, ok := parseDate(ctx, "from")
	if !ok {
		return
	}
	to, ok := parseDate(ctx, "to")
	if !ok {
		return
	}
	// Build params map for service
	params := ctx.Request.URL.Query()
	params.Set("restaurantId", ridHex)
	params.Set("from", from.Format("01/02/2006"))
	params.Set("to", to.Format("01/02/2006"))
	data, err := c.service.DailyAggregates(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *Controller) GetPopularItems(ctx *gin.Context) {
	from, ok := parseDate(ctx, "from")
	if !ok {
		return
	}
	to, ok := parseDate(ctx, "to")
	if !ok {
		return
	}
	params := ctx.Request.URL.Query()
	params.Set("from", from.Format("01/02/2006"))
	params.Set("to", to.Format("01/02/2006"))
	data, err := c.service.MostPopularItems(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data)
}
