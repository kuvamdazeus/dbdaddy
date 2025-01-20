package serverHandlers

import (
	"github.com/fossmedaddy/dbdaddy/constants"
	"github.com/fossmedaddy/dbdaddy/db"
	"github.com/fossmedaddy/dbdaddy/db/db_int"
	"github.com/fossmedaddy/dbdaddy/globals"
	"github.com/fossmedaddy/dbdaddy/lib"
	"github.com/fossmedaddy/dbdaddy/types"

	"github.com/gofiber/fiber/v2"
)

func HandleGetBranches(c *fiber.Ctx) error {
	dbs, err := db_int.GetExistingDbs(false)
	if err != nil {
		return err
	}
	dbs = append(dbs, constants.SelfDbName)

	return c.JSON(types.Response{
		Message: "success",
		Data:    dbs,
	})
}

func HandleGetCurrentBranch(c *fiber.Ctx) error {
	return c.JSON(types.Response{
		Message: "success",
		Data:    globals.CliConfig.State.CurrentBranch,
	})
}

func HandlePutCurrentBranch(c *fiber.Ctx) error {
	type reqSchema struct {
		BranchName string
	}

	var reqBody reqSchema
	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.Response{
			Message: err.Error(),
			Data:    nil,
		})
	}

	if err := lib.SetCurrentBranch(reqBody.BranchName); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(types.Response{
			Message: err.Error(),
			Data:    nil,
		})
	}

	connConfig := globals.CliConfig.MainConnConfig
	connConfig.Database = reqBody.BranchName

	_, err := db.ConnectDb(connConfig)
	if err != nil {
		return err
	}

	return c.JSON(types.Response{
		Message: "success",
		Data:    nil,
	})
}
