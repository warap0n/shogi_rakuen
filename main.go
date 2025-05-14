package main

import (
	"log"
	"os"
	"shogi-rakuen/controller"
	"shogi-rakuen/model"
	"shogi-rakuen/repository"
	"shogi-rakuen/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	// .env ロード
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	e := echo.New()

	// ログとパニック復旧
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", os.Getenv("FE_URL")},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXCSRFToken,
		},
		AllowCredentials: true,
	}))

	// バリデータ
	e.Validator = &CustomValidator{validator: validator.New()}

	// DB
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&model.User{})

	// DI
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo)
	userController := controller.NewUserController(userUsecase)

	// ルーティング
	e.POST("/signup", userController.SignUp)
	e.POST("/login", userController.Login)

	u := e.Group("")
	u.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:  []byte(os.Getenv("JWT_SECRET")),
		TokenLookup: "cookie:access_token",
	}))

	u.GET("/users/:id", userController.GetUserByEmail)

	// サーバ起動（必要なら Graceful Shutdown 化）
	e.Logger.Fatal(e.Start(":8080"))
}
