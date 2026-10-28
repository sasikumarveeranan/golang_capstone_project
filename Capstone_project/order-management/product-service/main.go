package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Product struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Stock     int     `json:"stock"`
	Available bool    `json:"available"`
}

func main() {
	// Load .env from current folder or project root
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	pool, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	r := gin.Default()

	// Health - Live
	r.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	// Health - Ready
	r.GET("/health/ready", func(c *gin.Context) {
		if err := pool.Ping(c); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "DOWN",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
		})
	})

	// Get Product
	r.GET("/products/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid product id",
			})
			return
		}

		var p Product

		err = pool.QueryRow(
			c,
			`SELECT id, name, price::float8, stock
			 FROM products
			 WHERE id = $1`,
			id,
		).Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.Stock,
		)

		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "database error",
			})
			return
		}

		p.Available = p.Stock > 0

		c.JSON(http.StatusOK, p)
	})

	// Reduce Product Stock
	r.PUT("/products/:id/stock/reduce", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid product id",
			})
			return
		}

		quantity, err := strconv.Atoi(c.Query("quantity"))
		if err != nil || quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid quantity",
			})
			return
		}

		var stock int

		err = pool.QueryRow(
			c,
			`UPDATE products
			 SET stock = stock - $1
			 WHERE id = $2
			   AND stock >= $1
			 RETURNING stock`,
			quantity,
			id,
		).Scan(&stock)

		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "product not found or insufficient stock",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to reduce stock",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "stock reduced successfully",
			"stock":   stock,
		})
	})

	// Increase Product Stock
	r.PUT("/products/:id/stock/increase", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid product id",
			})
			return
		}

		quantity, err := strconv.Atoi(c.Query("quantity"))
		if err != nil || quantity <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid quantity",
			})
			return
		}

		var stock int

		err = pool.QueryRow(
			c,
			`UPDATE products
			 SET stock = stock + $1
			 WHERE id = $2
			 RETURNING stock`,
			quantity,
			id,
		).Scan(&stock)

		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "product not found",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to increase stock",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "stock increased successfully",
			"stock":   stock,
		})
	})

	log.Println("product service listening on :8081")

	if err := r.Run(":8081"); err != nil {
		log.Fatal(err)
	}
}

func connectDB() (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MinConns = 1

	return pgxpool.NewWithConfig(context.Background(), cfg)
}
