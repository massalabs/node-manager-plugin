package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	configPkg "github.com/massalabs/node-manager-plugin/int/config"
	stakingManagerPkg "github.com/massalabs/node-manager-plugin/int/staking-manager"
)

func HandlePostStakingAddresses(stakingManager stakingManagerPkg.StakingManager) func(operations.AddStakingAddressParams) middleware.Responder {
	return func(params operations.AddStakingAddressParams) middleware.Responder {
		// the first param is the pwd of the node, the second is the pwd to unlock the account file
		stakingAddress, err := stakingManager.AddStakingAddress(configPkg.GlobalPluginInfo.GetPwd(), params.Body.Password, params.Body.Nickname)
		if err != nil {
			return operations.NewAddStakingAddressInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewAddStakingAddressOK().WithPayload(&models.StakingAddress{
			Address:            stakingAddress.Address,
			TargetRolls:        int64(stakingAddress.TargetRolls),
			FinalRollCount:     int64(stakingAddress.FinalRolls),
			ActiveRollCount:    int64(stakingAddress.ActiveRolls),
			FinalBalance:       stakingAddress.FinalBalance,
			CandidateRollCount: int64(stakingAddress.CandidateRolls),
			CandidateBalance:   stakingAddress.CandidateBalance,
			Thread:             int64(stakingAddress.Thread),
		})
	}
}

func HandlePutStakingAddresses(stakingManager stakingManagerPkg.StakingManager) func(operations.UpdateStakingAddressParams) middleware.Responder {
	return func(params operations.UpdateStakingAddressParams) middleware.Responder {
		err := stakingManager.SetTargetRolls(params.Body.Address, uint64(params.Body.TargetRolls))
		if err != nil {
			return operations.NewUpdateStakingAddressInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewUpdateStakingAddressNoContent()
	}
}

func HandleDeleteStakingAddresses(stakingManager stakingManagerPkg.StakingManager) func(operations.RemoveStakingAddressParams) middleware.Responder {
	return func(params operations.RemoveStakingAddressParams) middleware.Responder {
		err := stakingManager.RemoveStakingAddress(configPkg.GlobalPluginInfo.GetPwd(), params.Body.Address)
		if err != nil {
			return operations.NewRemoveStakingAddressInternalServerError().WithPayload(&models.Error{
				Message: err.Error(),
			})
		}
		return operations.NewRemoveStakingAddressNoContent()
	}
}
