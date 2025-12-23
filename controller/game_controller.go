// controller/game_controller.go

package controller

import (
	"errors"
	"net/http"

	"shogi-rakuen/controller/dto"
	"shogi-rakuen/usecase"
	"shogi-rakuen/usecase/input"

	"github.com/labstack/echo/v4"
)

type IGameController interface {
	StartGame(c echo.Context) error
	GetGame(c echo.Context) error
	ApplyMove(c echo.Context) error
	ListMoves(c echo.Context) error
}

type GameController struct {
	gu usecase.IGameUsecase
}

func NewGameController(gu usecase.IGameUsecase) IGameController {
	return &GameController{gu: gu}
}

// ─── POST /start ───────────────────────────────────────────────────────────────
func (gc *GameController) StartGame(c echo.Context) error {
	var req dto.StartGameRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "invalid request format",
			"details": err.Error(),
		})
	}
	if err := c.Validate(&req); err != nil {
		// エラーの詳細を出したい場合は、カスタムバリデータで detail を返すようにする
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "validation failed",
			"details": err.Error(),
		})
	}

	in := input.StartGameInput{
		BlackID: req.BlackID,
		WhiteID: req.WhiteID,
	}
	game, err := gc.gu.StartGame(in)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidPlayerID):
			return c.JSON(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "invalid player ID",
			})
		case errors.Is(err, usecase.ErrSamePlayer):
			return c.JSON(http.StatusConflict, echo.Map{
				"status":  "error",
				"message": "black and white must differ",
			})
		default:
			c.Logger().Errorf("unexpected error in StartGame: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "error",
				"message": "internal server error",
			})
		}
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"status": "success",
		"data":   dto.NewGameResponse(game),
	})
}

// ─── GET /:id ────────────────────────────────────────────────────────────
func (gc *GameController) GetGame(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "missing game ID",
		})
	}

	game, err := gc.gu.GetGameByID(id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "game not found",
			})
		default:
			c.Logger().Errorf("unexpected error in GetGame: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "error",
				"message": "internal server error",
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   dto.NewGameResponse(game),
	})
}

// ─── POST /:id/move ───────────────────────────────────────────────────────
func (gc *GameController) ApplyMove(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "missing game ID",
		})
	}

	var req dto.ApplyMoveRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "invalid request format",
			"details": err.Error(),
		})
	}
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "validation failed",
			"details": err.Error(),
		})
	}

	in := input.ApplyMoveInput{
		GameID:    id,
		From:      req.From,
		To:        req.To,
		Promote:   req.Promote,
		Drop:      req.Drop,
		DropPiece: req.DropPiece,
	}
	game, err := gc.gu.ApplyMove(in)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "game not found",
			})
		case errors.Is(err, usecase.ErrInvalidMove):
			return c.JSON(http.StatusBadRequest, echo.Map{
				"status":  "error",
				"message": "invalid move",
			})
		default:
			c.Logger().Errorf("unexpected error in ApplyMove: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "error",
				"message": "internal server error",
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   dto.NewGameResponse(game),
	})
}

// ─── GET /:id/moves ───────────────────────────────────────────────────────
func (gc *GameController) ListMoves(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"status":  "error",
			"message": "missing game ID",
		})
	}

	moves, err := gc.gu.ListMoves(id)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGameNotFound):
			return c.JSON(http.StatusNotFound, echo.Map{
				"status":  "error",
				"message": "game not found",
			})
		default:
			c.Logger().Errorf("unexpected error in ListMoves: %v", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "error",
				"message": "internal server error",
			})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   dto.NewMovesResponse(moves),
	})
}
