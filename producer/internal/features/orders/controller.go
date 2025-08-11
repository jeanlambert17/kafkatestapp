package orders

import (
	"net/http"

	"producer/internal/auth"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(router *gin.Engine) {
	router.POST("/orders", c.CreateOrder)
	router.GET("/orders", c.ListOrders)
	router.GET("/orders/recent", c.RecentOrders)
}

func (c *Controller) CreateOrder(ctx *gin.Context) {
	type createOrderItem struct {
		ID       string `json:"id"`
		Quantity int    `json:"quantity"`
	}
	type createOrderBody struct {
		Items []createOrderItem `json:"items"`
	}
	var body createOrderBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(body.Items) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "items are required"})
		return
	}

	_, org, err := auth.GetOrgID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// transform to event items with id and quantity
	items := make([]CreateOrderEventItem, 0, len(body.Items))
	for _, it := range body.Items {
		if it.ID == "" || it.Quantity <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "each item requires id and quantity > 0"})
			return
		}
		items = append(items, CreateOrderEventItem{ID: it.ID, Quantity: it.Quantity})
	}
	req := CreateOrderEvent{RestaurantID: org, Items: items}
	if err := c.service.PublishOrder(ctx, req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

func (c *Controller) ListOrders(ctx *gin.Context) {
	orgID, _, err := auth.GetOrgID(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := c.service.ListOrders(ctx, orgID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (c *Controller) RecentOrders(ctx *gin.Context) {
	org := ctx.GetHeader("x-org")
	if org == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "missing x-org header"})
		return
	}
	resp, err := c.service.RecentOrders(ctx, org)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
