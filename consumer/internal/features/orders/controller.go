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
	orders.GET("", c.ListOrders)
	orders.GET("/:id", c.GetOrder)
	orders.POST("", c.CreateOrder)
}

func (c *Controller) ListOrders(ctx *gin.Context) {
	data, err := c.service.ListOrders(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, data)
}

func (c *Controller) GetOrder(ctx *gin.Context) {
	idParam := ctx.Param("id")
	oid, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	order, err := c.service.GetOrderByID(ctx.Request.Context(), oid)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	ctx.JSON(http.StatusOK, order)
}

type createOrderBody struct {
	Items []string `form:"items" json:"items"`
}

func (c *Controller) CreateOrder(ctx *gin.Context) {
	var body createOrderBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if len(body.Items) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "items required"})
		return
	}

	// Convert item hex IDs to ObjectIDs
	itemIDs := make([]primitive.ObjectID, 0, len(body.Items))
	for _, idStr := range body.Items {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id: " + idStr})
			return
		}
		itemIDs = append(itemIDs, oid)
	}

	order := models.Order{
		Items: itemIDs,
	}

	id, err := c.service.CreateOrder(ctx, &order)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	order.ID = id
	ctx.JSON(http.StatusCreated, order)
}
