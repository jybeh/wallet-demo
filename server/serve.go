package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"wallet/handler"
	"wallet/storage"
)

// Serve ...
func Serve() {
	dsn := "host=localhost dbname=wallet port=5432 sslmode=disable TimeZone=Asia/Kuala_Lumpur"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic("failed to connect database")
	}

	r := gin.Default()
	service := handler.NewWalletService(
		storage.NewAccountDAO(db),
		storage.NewTransactionDAO(db),
		storage.NewTransferDAO(db),
	)
	service.RegisterRoutes(r)
	r.Run()
}
