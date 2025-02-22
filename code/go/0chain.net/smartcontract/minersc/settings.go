package minersc

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/0chain/common/core/currency"

	"0chain.net/chaincore/smartcontractinterface"
	"0chain.net/smartcontract"

	cstate "0chain.net/chaincore/chain/state"
	"0chain.net/chaincore/transaction"
	"0chain.net/core/common"
)

const x10 float64 = 10 * 1000 * 1000 * 1000

type Setting int

const (
	MinStake Setting = iota
	MaxStake
	MaxN
	MinN
	TPercent
	KPercent
	XPercent
	MaxS
	MinS
	MaxDelegates
	RewardRoundFrequency
	RewardRate
	ShareRatio
	BlockReward
	MaxCharge
	Epoch
	RewardDeclineRate
	NumMinerDelegatesRewarded
	NumShardersRewarded
	NumSharderDelegatesRewarded
	MaxMint
	OwnerId
	CooldownPeriod
	CostAddMiner
	CostAddSharder
	CostDeleteMiner
	CostMinerHealthCheck
	CostSharderHealthCheck
	CostContributeMpk
	CostShareSignsOrShares
	CostWait
	CostUpdateGlobals
	CostUpdateSettings
	CostUpdateMinerSettings
	CostUpdateSharderSettings
	CostPayFees
	CostFeesPaid
	CostMintedTokens
	CostAddToDelegatePool
	CostDeleteFromDelegatePool
	CostSharderKeep
	CostKillMiner
	CostKillSharder
	NumberOfSettings
)

func (s Setting) String() string {
	if s >= NumberOfSettings { // should never happen
		return ""
	}
	return SettingName[s]
}

var (
	SettingName = make([]string, NumberOfSettings)
	Settings    map[string]struct {
		Setting    Setting
		ConfigType smartcontract.ConfigType
	}
)

func init() {
	initSettingName()
	initSettings()
}

func initSettingName() {
	SettingName[MinStake] = "min_stake"
	SettingName[MaxStake] = "max_stake"
	SettingName[MaxN] = "max_n"
	SettingName[MinN] = "min_n"
	SettingName[TPercent] = "t_percent"
	SettingName[KPercent] = "k_percent"
	SettingName[XPercent] = "x_percent"
	SettingName[MaxS] = "max_s"
	SettingName[MinS] = "min_s"
	SettingName[MaxDelegates] = "max_delegates"
	SettingName[RewardRoundFrequency] = "reward_round_frequency"
	SettingName[RewardRate] = "reward_rate"
	SettingName[ShareRatio] = "share_ratio"
	SettingName[BlockReward] = "block_reward"
	SettingName[MaxCharge] = "max_charge"
	SettingName[Epoch] = "epoch"
	SettingName[RewardDeclineRate] = "reward_decline_rate"
	SettingName[NumMinerDelegatesRewarded] = "num_miner_delegates_rewarded"
	SettingName[NumShardersRewarded] = "num_sharders_rewarded"
	SettingName[NumSharderDelegatesRewarded] = "num_sharder_delegates_rewarded"
	SettingName[MaxMint] = "max_mint"
	SettingName[OwnerId] = "owner_id"
	SettingName[CooldownPeriod] = "cooldown_period"
	SettingName[CostAddMiner] = "cost.add_miner"
	SettingName[CostAddSharder] = "cost.add_sharder"
	SettingName[CostDeleteMiner] = "cost.delete_miner"
	SettingName[CostMinerHealthCheck] = "cost.miner_health_check"
	SettingName[CostSharderHealthCheck] = "cost.sharder_health_check"
	SettingName[CostContributeMpk] = strings.ToLower("cost.contributeMpk")
	SettingName[CostShareSignsOrShares] = strings.ToLower("cost.shareSignsOrShares")
	SettingName[CostWait] = "cost.wait"
	SettingName[CostUpdateGlobals] = "cost.update_globals"
	SettingName[CostUpdateSettings] = "cost.update_settings"
	SettingName[CostUpdateMinerSettings] = "cost.update_miner_settings"
	SettingName[CostUpdateSharderSettings] = "cost.update_sharder_settings"
	SettingName[CostPayFees] = strings.ToLower("cost.payFees")
	SettingName[CostFeesPaid] = strings.ToLower("cost.feesPaid")
	SettingName[CostMintedTokens] = strings.ToLower("cost.mintedTokens")
	SettingName[CostAddToDelegatePool] = strings.ToLower("cost.addToDelegatePool")
	SettingName[CostDeleteFromDelegatePool] = strings.ToLower("cost.deleteFromDelegatePool")
	SettingName[CostSharderKeep] = "cost.sharder_keep"
	SettingName[CostKillMiner] = "cost.kill_miner"
	SettingName[CostKillSharder] = "cost.kill_sharder"
}

func initSettings() {
	Settings = map[string]struct {
		Setting    Setting
		ConfigType smartcontract.ConfigType
	}{
		MinStake.String():                    {MinStake, smartcontract.CurrencyCoin},
		MaxStake.String():                    {MaxStake, smartcontract.CurrencyCoin},
		MaxN.String():                        {MaxN, smartcontract.Int},
		MinN.String():                        {MinN, smartcontract.Int},
		TPercent.String():                    {TPercent, smartcontract.Float64},
		KPercent.String():                    {KPercent, smartcontract.Float64},
		XPercent.String():                    {XPercent, smartcontract.Float64},
		MaxS.String():                        {MaxS, smartcontract.Int},
		MinS.String():                        {MinS, smartcontract.Int},
		MaxDelegates.String():                {MaxDelegates, smartcontract.Int},
		RewardRoundFrequency.String():        {RewardRoundFrequency, smartcontract.Int64},
		RewardRate.String():                  {RewardRate, smartcontract.Float64},
		ShareRatio.String():                  {ShareRatio, smartcontract.Float64},
		BlockReward.String():                 {BlockReward, smartcontract.CurrencyCoin},
		MaxCharge.String():                   {MaxCharge, smartcontract.Float64},
		Epoch.String():                       {Epoch, smartcontract.Int64},
		RewardDeclineRate.String():           {RewardDeclineRate, smartcontract.Float64},
		NumMinerDelegatesRewarded.String():   {NumMinerDelegatesRewarded, smartcontract.Int},
		NumShardersRewarded.String():         {NumShardersRewarded, smartcontract.Int},
		NumSharderDelegatesRewarded.String(): {NumSharderDelegatesRewarded, smartcontract.Int},
		MaxMint.String():                     {MaxMint, smartcontract.CurrencyCoin},
		OwnerId.String():                     {OwnerId, smartcontract.Key},
		CooldownPeriod.String():              {CooldownPeriod, smartcontract.Int64},
		CostAddMiner.String():                {CostAddMiner, smartcontract.Cost},
		CostAddSharder.String():              {CostAddSharder, smartcontract.Cost},
		CostDeleteMiner.String():             {CostDeleteMiner, smartcontract.Cost},
		CostMinerHealthCheck.String():        {CostMinerHealthCheck, smartcontract.Cost},
		CostSharderHealthCheck.String():      {CostSharderHealthCheck, smartcontract.Cost},
		CostContributeMpk.String():           {CostContributeMpk, smartcontract.Cost},
		CostShareSignsOrShares.String():      {CostShareSignsOrShares, smartcontract.Cost},
		CostWait.String():                    {CostWait, smartcontract.Cost},
		CostUpdateGlobals.String():           {CostUpdateGlobals, smartcontract.Cost},
		CostUpdateSettings.String():          {CostUpdateSettings, smartcontract.Cost},
		CostUpdateMinerSettings.String():     {CostUpdateMinerSettings, smartcontract.Cost},
		CostUpdateSharderSettings.String():   {CostUpdateSharderSettings, smartcontract.Cost},
		CostPayFees.String():                 {CostPayFees, smartcontract.Cost},
		CostFeesPaid.String():                {CostFeesPaid, smartcontract.Cost},
		CostMintedTokens.String():            {CostMintedTokens, smartcontract.Cost},
		CostAddToDelegatePool.String():       {CostAddToDelegatePool, smartcontract.Cost},
		CostDeleteFromDelegatePool.String():  {CostDeleteFromDelegatePool, smartcontract.Cost},
		CostSharderKeep.String():             {CostSharderKeep, smartcontract.Cost},
		CostKillMiner.String():               {CostKillMiner, smartcontract.Cost},
		CostKillSharder.String():             {CostKillSharder, smartcontract.Cost},
	}
}

func (gn *GlobalNode) setInt(key string, change int) error {
	switch Settings[key].Setting {
	case MaxN:
		gn.MaxN = change
	case MinN:
		gn.MinN = change
	case MaxS:
		gn.MaxS = change
	case MinS:
		gn.MinS = change
	case MaxDelegates:
		gn.MaxDelegates = change
	case NumMinerDelegatesRewarded:
		gn.NumMinerDelegatesRewarded = change
	case NumShardersRewarded:
		gn.NumShardersRewarded = change
	case NumSharderDelegatesRewarded:
		gn.NumSharderDelegatesRewarded = change
	default:
		return fmt.Errorf("key: %v not implemented as int", key)
	}
	return nil
}

func (gn *GlobalNode) setBalance(key string, change currency.Coin) error {
	switch Settings[key].Setting {
	case MaxMint:
		gn.MaxMint = change
	case MinStake:
		gn.MinStake = change
	case MaxStake:
		gn.MaxStake = change
	case BlockReward:
		gn.BlockReward = change
	default:
		return fmt.Errorf("key: %v not implemented as balance", key)
	}
	return nil
}

func (gn *GlobalNode) setInt64(key string, change int64) error {
	switch Settings[key].Setting {
	case RewardRoundFrequency:
		gn.RewardRoundFrequency = change
	case Epoch:
		gn.Epoch = change
	case CooldownPeriod:
		gn.CooldownPeriod = change
	default:
		return fmt.Errorf("key: %v not implemented as int64", key)
	}
	return nil
}

func (gn *GlobalNode) setFloat64(key string, change float64) error {
	switch Settings[key].Setting {
	case TPercent:
		gn.TPercent = change
	case KPercent:
		gn.KPercent = change
	case XPercent:
		gn.XPercent = change
	case RewardRate:
		gn.RewardRate = change
	case ShareRatio:
		gn.ShareRatio = change
	case MaxCharge:
		gn.MaxCharge = change
	case RewardDeclineRate:
		gn.RewardDeclineRate = change
	default:
		return fmt.Errorf("key: %v not implemented as float64", key)
	}
	return nil
}

func (gn *GlobalNode) setKey(key string, change string) {
	switch Settings[key].Setting {
	case OwnerId:
		gn.OwnerId = change
	default:
		panic("key: " + key + "not implemented as key")
	}
}

const costPrefix = "cost."

func (gn *GlobalNode) setCost(key string, change int) error {
	if !isCost(key) {
		return fmt.Errorf("key: %v is not a cost", key)
	}
	if gn.Cost == nil {
		gn.Cost = make(map[string]int)
	}
	gn.Cost[strings.TrimPrefix(key, costPrefix)] = change
	return nil
}

func (gn *GlobalNode) getCost(key string) (int, error) {
	if !isCost(key) {
		return 0, fmt.Errorf("key: %v is not a cost", key)
	}
	if gn.Cost == nil {
		return 0, errors.New("cost object is nil")
	}
	value, ok := gn.Cost[strings.TrimPrefix(key, costPrefix)]
	if !ok {
		return 0, fmt.Errorf("cost %s not set", key)
	}
	return value, nil
}

func isCost(key string) bool {
	if len(key) <= len(costPrefix) {
		return false
	}
	return key[:len(costPrefix)] == costPrefix
}

func (gn *GlobalNode) set(key string, change string) error {
	settings, ok := Settings[key]
	if !ok {
		return fmt.Errorf("unsupported key %v", key)
	}

	switch settings.ConfigType {
	case smartcontract.Int:
		value, err := strconv.Atoi(change)
		if err != nil {
			return fmt.Errorf("cannot convert key %s value %v to int: %v", key, change, err)
		}
		if err := gn.setInt(key, value); err != nil {
			return err
		}
	case smartcontract.CurrencyCoin:
		value, err := strconv.ParseFloat(change, 64)
		if err != nil {
			return fmt.Errorf("cannot convert key %s value %v to state.balance: %v", key, change, err)
		}
		coinV, err := currency.ParseZCN(value)
		if err != nil {
			return err
		}
		if err := gn.setBalance(key, coinV); err != nil {
			return err
		}
	case smartcontract.Int64:
		value, err := strconv.ParseInt(change, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot convert key %s value %v to int64: %v", key, change, err)
		}
		if err := gn.setInt64(key, value); err != nil {
			return err
		}
	case smartcontract.Float64:
		value, err := strconv.ParseFloat(change, 64)
		if err != nil {
			return fmt.Errorf("cannot convert key %s value %v to float64: %v", key, change, err)
		}
		if err := gn.setFloat64(key, value); err != nil {
			return err
		}
	case smartcontract.Key:
		if _, err := hex.DecodeString(change); err != nil {
			return fmt.Errorf("%s must be a hex string: %v", key, err)
		}
		gn.setKey(key, change)
	case smartcontract.Cost:
		value, err := strconv.Atoi(change)
		if err != nil {
			return fmt.Errorf("cannot convert key %s value %v to int64: %v", key, change, err)
		}
		if err := gn.setCost(key, value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported type setting %v", smartcontract.ConfigTypeName[Settings[key].ConfigType])
	}

	return nil
}

func (gn *GlobalNode) update(changes smartcontract.StringMap) error {
	for key, value := range changes.Fields {
		if err := gn.set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func (msc *MinerSmartContract) updateSettings(
	t *transaction.Transaction,
	inputData []byte,
	gn *GlobalNode,
	balances cstate.StateContextI,
) (resp string, err error) {
	if err := smartcontractinterface.AuthorizeWithOwner("update_settings", func() bool {
		get, _ := gn.Get(OwnerId)
		return get == t.ClientID
	}); err != nil {
		return "", err
	}

	var changes smartcontract.StringMap
	if err = changes.Decode(inputData); err != nil {
		return "", common.NewError("update_settings", err.Error())
	}

	if err := gn.update(changes); err != nil {
		return "", common.NewError("update_settings", err.Error())
	}

	if err = gn.validate(); err != nil {
		return "", common.NewError("update_settings", err.Error())
	}

	if err := gn.save(balances); err != nil {
		return "", common.NewError("update_settings", err.Error())
	}

	return "", nil
}
