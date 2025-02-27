package dbs

import (
	"0chain.net/core/common"
	"0chain.net/smartcontract/stakepool/spenum"
	"github.com/0chain/common/core/currency"
)

type DbHealthCheck struct {
	ID              string           `json:"id"`
	LastHealthCheck common.Timestamp `json:"last_health_check"`
	Downtime        uint64           `json:"downtime"`
}

type DbUpdates struct {
	Id      string                 `json:"id"`
	Updates map[string]interface{} `json:"updates"`
}

type DbUpdateProvider struct {
	DbUpdates
	Type spenum.Provider `json:"type"`
}

func NewDbUpdateProvider(id string, typ spenum.Provider) *DbUpdateProvider {
	return &DbUpdateProvider{
		DbUpdates: DbUpdates{
			Id:      id,
			Updates: make(map[string]interface{}),
		},
		Type: typ,
	}

}

type ProviderID struct {
	ID   string          `json:"provider_id"`
	Type spenum.Provider `json:"provider_type"`
}

type StakePoolReward struct {
	ProviderID
	Reward     currency.Coin `json:"reward"`
	RewardType spenum.Reward `json:"reward_type"`
	// rewards delegate pools
	DelegateRewards map[string]currency.Coin `json:"delegate_rewards"`
	// penalties delegate pools
	DelegatePenalties map[string]currency.Coin `json:"delegate_penalties"`
	// challenge id
	ChallengeID string `json:"challenge_id"`
}

type DelegatePoolId struct {
	ProviderID
	PoolId string `json:"pool_id"`
}

type DelegatePoolUpdate struct {
	DelegatePoolId
	Updates map[string]interface{} `json:"updates"`
}

func NewDelegatePoolUpdate(pool, provider string, pType spenum.Provider) *DelegatePoolUpdate {
	var dpu DelegatePoolUpdate
	dpu.PoolId = pool
	dpu.ID = provider
	dpu.Type = pType
	dpu.Updates = make(map[string]interface{})
	return &dpu
}

type SpBalance struct {
	ProviderID
	Balance         int64            `json:"sp_reward"`
	DelegateBalance map[string]int64 `json:"delegate_reward"`
}

type SpReward struct {
	ProviderID
	SpReward       int64            `json:"sp_reward"`
	DelegateReward map[string]int64 `json:"delegate_reward"`
}

type ChallengeResult struct {
	BlobberId string `json:"blobberId"`
	Passed    bool   `json:"passed"`
}
