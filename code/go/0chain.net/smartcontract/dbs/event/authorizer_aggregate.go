package event

import (
	"0chain.net/chaincore/config"
	"0chain.net/smartcontract/dbs/model"
	"github.com/0chain/common/core/currency"
	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

type AuthorizerAggregate struct {
	model.ImmutableModel

	AuthorizerID string `json:"authorizer_id" gorm:"index:idx_authorizer_aggregate,unique"`
	Round        int64  `json:"round" gorm:"index:idx_authorizer_aggregate,unique"`
	BucketID     int64  `json:"bucket_id"`

	Fee           currency.Coin `json:"fee"`
	TotalStake    currency.Coin `json:"total_stake"`
	TotalRewards  currency.Coin `json:"total_rewards"`
	TotalMint     currency.Coin `json:"total_mint"`
	TotalBurn     currency.Coin `json:"total_burn"`
	ServiceCharge float64       `json:"service_charge"`
}

func (a *AuthorizerAggregate) GetTotalStake() currency.Coin {
	return a.TotalStake
}

func (a *AuthorizerAggregate) GetServiceCharge() float64 {
	return a.ServiceCharge
}

func (a *AuthorizerAggregate) GetTotalRewards() currency.Coin {
	return a.TotalRewards
}

func (a *AuthorizerAggregate) SetTotalStake(value currency.Coin) {
	a.TotalStake = value
}

func (a *AuthorizerAggregate) SetServiceCharge(value float64) {
	a.ServiceCharge = value
}

func (a *AuthorizerAggregate) SetTotalRewards(value currency.Coin) {
	a.TotalRewards = value
}

func (edb *EventDb) updateAuthorizerAggregate(round, pageAmount int64, gs *Snapshot) {
	currentBucket := round % config.Configuration().ChainConfig.DbSettings().AggregatePeriod

	exec := edb.Store.Get().Exec("CREATE TEMP TABLE IF NOT EXISTS authorizer_temp_ids "+
		"ON COMMIT DROP AS SELECT id as id FROM authorizers where bucket_id = ?",
		currentBucket)
	if exec.Error != nil {
		logging.Logger.Error("error creating temp table", zap.Error(exec.Error))
		return
	}

	exec = edb.Store.Get().Exec("CREATE TEMP TABLE IF NOT EXISTS authorizer_old_temp_ids "+
		"ON COMMIT DROP AS SELECT authorizer_id as id FROM authorizer_snapshots where bucket_id = ?",
		currentBucket)
	if exec.Error != nil {
		logging.Logger.Error("error creating old temp table", zap.Error(exec.Error))
		return
	}

	var count int64
	r := edb.Store.Get().Raw("SELECT count(*) FROM authorizer_temp_ids").Scan(&count)
	if r.Error != nil {
		logging.Logger.Error("getting ids count", zap.Error(r.Error))
		return
	}
	if count == 0 {
		return
	}
	pageCount := count / edb.PageLimit()

	logging.Logger.Debug("authorizer aggregate/snapshot started", zap.Int64("round", round), zap.Int64("bucket_id", currentBucket), zap.Int64("page_limit", edb.PageLimit()))
	for i := int64(0); i <= pageCount; i++ {
		edb.calculateAuthorizerAggregate(gs, round, edb.PageLimit(), i*edb.PageLimit())
	}
}

func (edb *EventDb) calculateAuthorizerAggregate(gs *Snapshot, round, limit, offset int64) {

	var ids []string
	r := edb.Store.Get().
		Raw("select id from authorizer_temp_ids ORDER BY ID limit ? offset ?", limit, offset).Scan(&ids)
	if r.Error != nil {
		logging.Logger.Error("getting ids", zap.Error(r.Error))
		return
	}

	var currentAuthorizers []Authorizer
	result := edb.Store.Get().Model(&Authorizer{}).
		Where("authorizers.id in (select id from authorizer_temp_ids ORDER BY ID limit ? offset ?)", limit, offset).
		Joins("Rewards").
		Find(&currentAuthorizers)
	if result.Error != nil {
		logging.Logger.Error("getting current Authorizers", zap.Error(result.Error))
		return
	}

	oldAuthorizers, err := edb.getAuthorizerSnapshots(limit, offset)
	if err != nil {
		logging.Logger.Error("getting Authorizer snapshots", zap.Error(err))
		return
	}

	var (
		oldAuthorizersProcessingMap = MakeProcessingMap(oldAuthorizers)
		aggregates                  []AuthorizerAggregate
		gsDiff                      Snapshot
		old                         AuthorizerSnapshot
		ok                          bool
	)

	for _, current := range currentAuthorizers {
		processingEntity, found := oldAuthorizersProcessingMap[current.ID]
		if !found {
			old = AuthorizerSnapshot{ /* zero values */ }
			gsDiff.AuthorizerCount += 1
		} else {
			old, ok = processingEntity.Entity.(AuthorizerSnapshot)
			if !ok {
				logging.Logger.Error("error converting processable entity to authorizer snapshot")
				continue
			}
		}

		// Case: authorizer becomes killed/shutdown
		if current.IsOffline() && !old.IsOffline() {
			handleOfflineAuthorizer(&gsDiff, old)
			continue
		}

		aggregate := AuthorizerAggregate{
			Round:        round,
			AuthorizerID: current.ID,
			BucketID:     current.BucketId,
		}

		recalculateProviderFields(&old, &current, &aggregate)

		aggregate.TotalMint = (old.TotalMint + current.TotalMint) / 2
		aggregate.TotalBurn = (old.TotalBurn + current.TotalBurn) / 2
		aggregate.Fee = (old.Fee + current.Fee) / 2
		aggregates = append(aggregates, aggregate)

		gsDiff.TotalRewards += int64(current.Rewards.TotalRewards - old.TotalRewards)
		gsDiff.TotalStaked += int64(current.TotalStake - old.TotalStake)
		gsDiff.TotalMint += int64(current.TotalMint - old.TotalMint)

		oldAuthorizersProcessingMap[current.ID] = processingEntity
	}

	gs.ApplyDiff(&gsDiff)
	if len(aggregates) > 0 {
		if result := edb.Store.Get().Create(&aggregates); result.Error != nil {
			logging.Logger.Error("saving aggregates", zap.Error(result.Error))
		}
	}

	if len(currentAuthorizers) > 0 {
		if err := edb.addAuthorizerSnapshot(currentAuthorizers, round); err != nil {
			logging.Logger.Error("saving Authorizer snapshots", zap.Error(err))
		}
	}

	logging.Logger.Debug("authorizer aggregate/snapshots finished successfully",
		zap.Int("current_authorizers", len(currentAuthorizers)),
		zap.Int("old_authorizers", len(oldAuthorizers)),
		zap.Int("aggregates", len(aggregates)),
		zap.Any("global_snapshot_after", gs),
	)
}

func handleOfflineAuthorizer(gs *Snapshot, old AuthorizerSnapshot) {
	gs.AuthorizerCount -= 1
	gs.TotalRewards -= int64(old.TotalRewards)
	gs.TotalStaked -= int64(old.TotalStake)
}
