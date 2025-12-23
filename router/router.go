package router

import (
	"os"
	"shogi-rakuen/controller"

	"github.com/go-playground/validator/v10"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomValidator Echo 用のバリデータ
// main と同じロジックを使い回せます
type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewRouter コントローラを渡して Echo を返す
func NewRouter(
	userController controller.IUserController,
	gameController controller.IGameController,
) *echo.Echo {
	// Echo インスタンス生成
	e := echo.New()

	// ログとパニック復旧ミドルウェア
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS 設定
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3000", os.Getenv("FE_URL")},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXCSRFToken},
		AllowCredentials: true,
	}))

	// バリデータ設定
	e.Validator = &CustomValidator{validator: validator.New()}

	// 認証不要ルート
	e.POST("/signup", userController.SignUp)
	e.POST("/login", userController.Login)

	// JWT ミドルウェア設定
	// middleware パッケージには JWT が含まれないため外部 echo-jwt パッケージを使う
	echoJWTConfig := echojwt.Config{
		SigningKey:  []byte(os.Getenv("JWT_SECRET")),
		TokenLookup: "cookie:access_token", // クッキー名と一致

	}

	auth := e.Group("")
	auth.Use(echojwt.WithConfig(echoJWTConfig))

	// 認証必須ルート
	auth.GET("/me", userController.Me)

	game := e.Group("/game")
	game.POST("/start", gameController.StartGame)
	game.GET("/:id", gameController.GetGame)
	game.POST("/:id/move", gameController.ApplyMove)
	game.GET("/:id/moves", gameController.ListMoves)

	return e
}
