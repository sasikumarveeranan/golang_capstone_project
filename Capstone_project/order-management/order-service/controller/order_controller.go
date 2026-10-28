package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"order-management/order-service/model"
	"order-management/order-service/service"
	"strconv"
)

type OrderController struct{ s service.OrderService }

func NewOrderController(s service.OrderService) *OrderController { return &OrderController{s} }

// Create godoc
// @Summary Create order
// @Description Creates an order after reserving product stock and obtaining the current product price.
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body model.CreateOrderRequest true "Order details"
// @Success 201 {object} model.Order
// @Failure 400 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /api/v1/orders [post]
func (cn *OrderController) Create(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	o, err := cn.s.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, o)
}

// Get godoc
// @Summary Get order
// @Description Returns an order by ID.
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} model.Order
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id} [get]
func (cn *OrderController) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	o, err := cn.s.Get(c, id)
	if err != nil {
		c.JSON(404, gin.H{"error": "order not found"})
		return
	}
	c.JSON(200, o)
}

// List godoc
// @Summary List orders
// @Description Returns all orders sorted by creation time.
// @Tags Orders
// @Produce json
// @Success 200 {array} model.Order
// @Failure 500 {object} map[string]string
// @Router /api/v1/orders [get]
func (cn *OrderController) List(c *gin.Context) {
	o, err := cn.s.List(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "database error"})
		return
	}
	c.JSON(200, o)
}

// Status godoc
// @Summary Update order status
// @Description Updates an order status. Cancelling an order releases its reserved stock.
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body model.UpdateStatusRequest true "New status"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/orders/{id}/status [put]
func (cn *OrderController) Status(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}
	var req model.UpdateStatusRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid status"})
		return
	}
	if err = cn.s.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		if err.Error() == "order not found" {
			c.JSON(404, gin.H{"error": err.Error()})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "status updated"})
}
