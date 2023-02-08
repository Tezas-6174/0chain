package event

import (
	"github.com/0chain/common/core/logging"
	"go.uber.org/zap"
)

// swagger:model UserSnapshot
type UserSnapshot struct {
	UserID          string `json:"user_id" gorm:"uniqueIndex"`
	Round           int64  `json:"round"`
	CollectedReward int64  `json:"collected_reward"`
	TotalStake      int64  `json:"total_stake"`
	ReadPoolTotal   int64  `json:"read_pool_total"`
	WritePoolTotal  int64  `json:"write_pool_total"`
	PayedFees       int64  `json:"payed_fees"`
}

func (edb *EventDb) getUserSnapshots(limit, offset int64) (map[string]UserSnapshot, error) {
	var snapshots []UserSnapshot
	result := edb.Store.Get().
		Raw("SELECT * FROM user_snapshots WHERE user_id in (select id from temp_user_ids ORDER BY ID limit ? offset ?)", limit, offset).
		Scan(&snapshots)
	if result.Error != nil {
		return nil, result.Error
	}

	var mapSnapshots = make(map[string]UserSnapshot, len(snapshots))
	logging.Logger.Debug("get_user_snapshot", zap.Int("snapshots selected", len(snapshots)))
	logging.Logger.Debug("get_user_snapshot", zap.Int64("snapshots rows selected", result.RowsAffected))

	for _, snapshot := range snapshots {
		mapSnapshots[snapshot.UserID] = snapshot
	}

	result = edb.Store.Get().Where("user_id IN (select id from temp_user_ids ORDER BY ID limit ? offset ?)", limit, offset).Delete(&UserSnapshot{})
	logging.Logger.Debug("get_user_snapshot", zap.Int64("deleted rows", result.RowsAffected))
	return mapSnapshots, result.Error
}

func (edb *EventDb) addUserSnapshot(users []User) error {
	var snapshots []UserSnapshot
	for _, user := range users {
		snapshots = append(snapshots, UserSnapshot{
			UserID:          user.UserID,
			Round:           user.Round,
			CollectedReward: user.CollectedReward,
			TotalStake:      user.TotalStake,
			ReadPoolTotal:   user.ReadPoolTotal,
			WritePoolTotal:  user.WritePoolTotal,
			PayedFees:       user.PayedFees,
		})
	}

	return edb.Store.Get().Create(&snapshots).Error
}
