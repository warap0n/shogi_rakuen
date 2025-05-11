package main

import (
	"log"
	"shogi-rakuen/controller"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Echoで使うカスタムバリデータ
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	e := echo.New()

	// バリデータをセット（validator:"required" などに対応）
	e.Validator = &CustomValidator{validator: validator.New()}

	// DB接続（例: SQLite）
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&model.User{})

	// DI: repository → usecase → controller
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userController := controller.NewUserController(userUsecase)

	// ルーティング
	e.POST("/signup", userController.SignUp)
	// e.POST("/login", userController.Login)

	// 起動
	e.Logger.Fatal(e.Start(":8080"))
}
