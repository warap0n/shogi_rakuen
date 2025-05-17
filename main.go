package main

import (
	"log"
	"shogi-rakuen/controller"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/router"
	"shogi-rakuen/usecase"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// .env ロード
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// DB接続
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&model.User{})

	// DI: repository → usecase → controller
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userController := controller.NewUserController(userUsecase)

	// Echo インスタンスとルータ設定
	e := router.NewRouter(userController)

	// サーバ起動
	e.Logger.Fatal(e.Start(":8080"))
}
