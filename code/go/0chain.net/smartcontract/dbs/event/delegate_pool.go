package event

import (
	"fmt"

	"0chain.net/core/common"
	common2 "0chain.net/smartcontract/common"

	"0chain.net/smartcontract/dbs/model"
	"github.com/0chain/common/core/currency"
	"gorm.io/gorm/clause"

	"0chain.net/smartcontract/stakepool/spenum"

	"0chain.net/smartcontract/dbs"
)

type DelegatePool struct {
	model.UpdatableModel
	PoolID       string          `json:"pool_id" gorm:"uniqueIndex:ppp;index:idx_ddel_active"`
	ProviderType spenum.Provider `json:"provider_type" gorm:"uniqueIndex:ppp;index:idx_dprov_active,priority:2;index:idx_ddel_active,priority:2" `
	ProviderID   string          `json:"provider_id" gorm:"uniqueIndex:ppp;index:idx_dprov_active,priority:1;index:idx_ddel_active,priority:2"`
	DelegateID   string          `json:"delegate_id" gorm:"index:idx_ddel_active,priority:2;index:idx_del_id;index:idx_dp_total_staked,priority:1"` //todo think of changing priority for idx_ddel_active

	Balance              currency.Coin     `json:"balance"`
	Reward               currency.Coin     `json:"reward"`       // unclaimed reward
	TotalReward          currency.Coin     `json:"total_reward"` // total reward paid to pool
	TotalPenalty         currency.Coin     `json:"total_penalty"`
	Status               spenum.PoolStatus `json:"status" gorm:"index:idx_dprov_active,priority:3;index:idx_ddel_active,priority:3;index:idx_dp_total_staked,priority:2"`
	RoundCreated         int64             `json:"round_created"`
	RoundPoolLastUpdated int64             `json:"round_pool_last_updated"`
	StakedAt             common.Timestamp  `json:"staked_at"`
}

func (edb *EventDb) GetDelegatePools(id string) ([]DelegatePool, error) {
	var dps []DelegatePool
	result := edb.Store.Get().
		Model(&DelegatePool{}).
		Where(&DelegatePool{
			ProviderID: id,
		}).
		Not(&DelegatePool{Status: spenum.Deleted}).
		Find(&dps)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting delegate pools, %v", result.Error)
	}
	return dps, nil
}

func (edb *EventDb) GetDelegatePool(poolID, pID string) (*DelegatePool, error) {
	var dp DelegatePool
	err := edb.Store.Get().Model(&DelegatePool{}).
		Where("pool_id = ? and provider_id = ? AND status != ?", poolID, pID, spenum.Deleted).
		First(&dp).Error
	if err != nil {
		return nil, fmt.Errorf("error getting delegate pool, %v", err)
	}

	return &dp, nil
}

func (edb *EventDb) GetUserTotalLocked(id string) (int64, error) {
	res := int64(0)
	err := edb.Store.Get().Table("delegate_pools").Select("coalesce(sum(balance),0)").
		Where("delegate_id = ? AND status in (?, ?)", id, spenum.Active, spenum.Pending).Row().Scan(&res)
	return res, err
}

func (edb *EventDb) GetUserDelegatePools(userId string, pType spenum.Provider, pagination common2.Pagination) ([]DelegatePool, error) {
	var dps []DelegatePool
	result := edb.Store.Get().
		Model(&DelegatePool{}).
		Where(&DelegatePool{
			ProviderType: pType,
			DelegateID:   userId,
		}).
		Not(&DelegatePool{Status: spenum.Deleted}).
		Offset(pagination.Offset).Limit(pagination.Limit).
		Order(clause.OrderByColumn{
			Column: clause.Column{Name: "id"},
		}).
		Find(&dps)
	if result.Error != nil {
		return nil, fmt.Errorf("error getting delegate pools, %v", result.Error)
	}
	return dps, nil
}

func (edb *EventDb) updateDelegatePool(updates dbs.DelegatePoolUpdate) error {
	var dp = DelegatePool{
		ProviderID:   updates.ID,
		ProviderType: updates.Type,
		PoolID:       updates.PoolId,
	}

	result := edb.Store.Get().
		Model(&DelegatePool{}).
		Where(&DelegatePool{
			ProviderType: dp.ProviderType,
			ProviderID:   dp.ProviderID,
			PoolID:       dp.PoolID,
		}).
		Updates(updates.Updates)
	return result.Error
}

func mergeAddDelegatePoolsEvents() *eventsMergerImpl[DelegatePool] {
	return newEventsMerger[DelegatePool](TagAddDelegatePool, withUniqueEventOverwrite())
}

func (edb *EventDb) addDelegatePools(dps []DelegatePool) error {
	return edb.Store.Get().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "provider_id"}, {Name: "provider_type"}, {Name: "pool_id"}},
		UpdateAll: true,
	}).Create(&dps).Error
}
