package orders

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"consumer/internal/models"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(router *gin.Engine) {
	orders := router.Group("/orders")
	orders.POST("", c.CreateOrder)
}

type createOrderItem struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}

type createOrderBody struct {
	Items []createOrderItem `json:"items"`
}

func (c *Controller) CreateOrder(ctx *gin.Context) {
	var body createOrderBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if len(body.Items) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "items required"})
		return
	}

	// Convert to order items
	orderItems := make([]models.OrderItem, 0, len(body.Items))
	for _, it := range body.Items {
		oid, err := primitive.ObjectIDFromHex(it.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id: " + it.ID})
			return
		}
		if it.Quantity <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be > 0 for item: " + it.ID})
			return
		}
		orderItems = append(orderItems, models.OrderItem{ItemID: oid, Quantity: it.Quantity})
	}

	order := models.Order{Items: orderItems}

	id, err := c.service.CreateOrder(ctx, &order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	order.ID = id
	ctx.JSON(http.StatusCreated, order)
}
