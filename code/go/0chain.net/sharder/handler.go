package sharder

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"0chain.net/chaincore/block"
	"0chain.net/chaincore/chain"
	"0chain.net/chaincore/config"
	"0chain.net/chaincore/diagnostics"
	"0chain.net/chaincore/node"
	"0chain.net/core/build"
	"0chain.net/core/common"
)

func handlersMap() map[string]func(http.ResponseWriter, *http.Request) {
	reqRespHandlers := map[string]common.ReqRespHandlerf{
		"/v1/block/get":                    common.ToJSONResponse(BlockHandler),
		"/v1/block/magic/get":              common.ToJSONResponse(MagicBlockHandler),
		"/v1/transaction/get/confirmation": common.ToJSONResponse(TransactionConfirmationHandler),
		"/v1/healthcheck":                  common.ToJSONResponse(HealthcheckHandler),
		"/v1/chain/get/stats":              common.ToJSONResponse(ChainStatsHandler),
		"/_chain_stats":                    ChainStatsWriter,
		"/_healthcheck":                    HealthCheckWriter,
		"/v1/sharder/get/stats":            common.ToJSONResponse(SharderStatsHandler),
		"/v1/state/nodes":                  common.ToJSONResponse(chain.StateNodesHandler),
		"/v1/block/state_change":           common.ToJSONResponse(BlockStateChangeHandler),
		"/_transaction_errors":             TransactionErrorWriter,
	}

	handlers := make(map[string]func(http.ResponseWriter, *http.Request))
	for pattern, handler := range reqRespHandlers {
		handlers[pattern] = common.WithCORS(common.UserRateLimit(handler))
	}
	return handlers
}

func BlockStateChangeHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	c := chain.GetServerChain()
	return c.BlockStateChangeHandler(ctx, r)
}

type ChainInfo struct {
	LatestFinalizedBlock *block.BlockSummary `json:"latest_finalized_block"`
}

func HealthcheckHandler(ctx context.Context, r *http.Request) (interface{}, error) {

	return struct {
		//Version  string `json:"version"`
		BuildTag string        `json:"build_tag"`
		Uptime   time.Duration `json:"uptime"`
		NodeType string        `json:"node_type"`

		Chain ChainInfo `json:"chain"`
	}{
		BuildTag: build.BuildTag,
		Uptime:   time.Since(chain.StartTime),
		NodeType: node.Self.Underlying().Type.String(),
		Chain: ChainInfo{
			LatestFinalizedBlock: chain.GetServerChain().GetLatestFinalizedBlockSummary(),
		},
	}, nil
}

/*BlockHandler - a handler to respond to block queries */
func BlockHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	roundData := r.FormValue("round")
	hash := r.FormValue("block")
	content := r.FormValue("content")
	if content == "" {
		content = "header"
	}
	parts := strings.Split(content, ",")
	sc := GetSharderChain()
	lfb := sc.GetLatestFinalizedBlock()
	if roundData != "" {
		roundNumber, err := strconv.ParseInt(roundData, 10, 64)
		if err != nil {
			return nil, err
		}
		if roundNumber > lfb.Round {
			return nil, common.InvalidRequest("Block not available")
		}
		roundEntity := sc.GetSharderRound(roundNumber)
		if roundEntity == nil {
			_, err = sc.GetRoundFromStore(ctx, roundNumber)
			if err != nil {
				return nil, err
			}
		}

		hash, err = sc.GetBlockHash(ctx, roundNumber)
		if err != nil {
			return nil, err
		}
	}
	var err error
	var b *block.Block
	if hash == "" {
		return nil, common.InvalidRequest("Block hash or round number is required")
	}
	b, err = chain.GetServerChain().GetBlock(ctx, hash)
	if err == nil {
		return chain.GetBlockResponse(b, parts)
	}
	/*NOTE: We store chain.RoundRange number of blocks in the same directory and that's a large number (10M).
	So, as long as people query the last 10M blocks most of the time, we only end up with 1 or 2 iterations.
	Anything older than that, there is a cost to query the database and get the round information anyway.
	*/
	for roundEntity := lfb.Round; roundEntity > 0; roundEntity -= sc.RoundRange() {
		b, err = sc.GetBlockFromStore(hash, roundEntity)
		if err != nil {
			return nil, err
		}
	}
	return chain.GetBlockResponse(b, parts)
}

/*MagicBlockHandler - a handler to respond to magic block queries */
func MagicBlockHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	magicBlockNumber := r.FormValue("magic_block_number")
	sc := GetSharderChain()
	mbm, err := sc.GetMagicBlockMap(ctx, magicBlockNumber)
	if err != nil {
		return nil, err
	}
	b, err := chain.GetServerChain().GetBlock(ctx, mbm.Hash)
	if err != nil {
		lfb := sc.GetLatestFinalizedBlock()
		for roundEntity := lfb.Round; roundEntity > 0; roundEntity -= sc.RoundRange() {
			b, err = sc.GetBlockFromStore(mbm.Hash, roundEntity)
			if err != nil {
				return nil, err
			}
		}
	}
	return b, nil
}

func ChainStatsHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	c := GetSharderChain().Chain
	return diagnostics.GetStatistics(c, chain.SteadyStateFinalizationTimer, 1000000.0), nil
}

func ChainStatsWriter(w http.ResponseWriter, r *http.Request) {
	sc := GetSharderChain()
	c := sc.Chain

	w.Header().Set("Content-Type", "text/html")
	chain.PrintCSS(w)
	diagnostics.WriteStatisticsCSS(w)

	self := node.Self.Underlying()
	fmt.Fprintf(w, "<h2>%v - %v</h2>", self.GetPseudoName(), self.Description)
	fmt.Fprintf(w, "<br>")

	fmt.Fprintf(w, "<table>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>Configuration <a href='v1/config/get'>...</a></h3>")
	diagnostics.WriteConfiguration(w, c)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>Current Status</h3>")
	diagnostics.WriteCurrentStatus(w, c)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>Summary</h3>")
	fmt.Fprintf(w, "<table width='100%%'>")
	fmt.Fprintf(w, "<tr><td>Sharded Blocks</td><td class='number'>%v</td></tr>", sc.SharderStats.ShardedBlocksCount)
	fmt.Fprintf(w, "<tr><td>QOS Round</td><td class='number'>%v</td></tr>", sc.SharderStats.QOSRound)
	fmt.Fprintf(w, "</table>")
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>Block Finalization Statistics (Steady State)</h3>")
	diagnostics.WriteTimerStatistics(w, c, chain.SteadyStateFinalizationTimer, 1000000.0)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>Block Finalization Statistics (Start to Finish)</h3>")
	diagnostics.WriteTimerStatistics(w, c, chain.StartToFinalizeTimer, 1000000.0)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td colspan='2'>")
	fmt.Fprintf(w, "<p>Steady state block finalization time = block generation + block processing + network time (1*large message + 2*small message)</p>")
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>Txn Finalization Statistics (Start to Finish)</h3>")
	if config.Development() {
		diagnostics.WriteTimerStatistics(w, c, chain.StartToFinalizeTxnTimer, 1000000.0)
	} else {
		fmt.Fprintf(w, "Available only in development mode")
	}
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>Finalization Lag Statistics</h3>")
	diagnostics.WriteHistogramStatistics(w, c, chain.FinalizationLagMetric)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>Transactions Save Statistics</h3>")
	diagnostics.WriteTimerStatistics(w, c, txnSaveTimer, 1000000.0)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>Block Save Statistics</h3>")
	diagnostics.WriteTimerStatistics(w, c, blockSaveTimer, 1000000.0)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>State Save Statistics</h3>")
	diagnostics.WriteTimerStatistics(w, c, block.StateSaveTimer, 1000000.0)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>State Change Statistics</h3>")
	diagnostics.WriteHistogramStatistics(w, c, block.StateChangeSizeMetric)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>State Prune Update Statistics</h3>")
	diagnostics.WriteTimerStatistics(w, c, chain.StatePruneUpdateTimer, 1000000.0)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>State Prune Delete Statistics</h3>")
	diagnostics.WriteTimerStatistics(w, c, chain.StatePruneDeleteTimer, 1000000.0)
	fmt.Fprintf(w, "</td></tr>")

	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<h3>State Computation Statistics</h3>")
	diagnostics.WriteHistogramStatistics(w, c, chain.StateComputationTimer)
	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "<h3>Events Computation Statistics</h3>")
	diagnostics.WriteHistogramStatistics(w, c, chain.EventsComputationTimer)
	fmt.Fprintf(w, "</td></tr>")

	if c.GetPruneStats() != nil {
		fmt.Fprintf(w, "<tr><td>")
		fmt.Fprintf(w, "<h3>Prune Stats</h3>")
		diagnostics.WritePruneStats(w, c.GetPruneStats())
		fmt.Fprintf(w, "</td><td valign='top'>")
		fmt.Fprintf(w, "<h3>Sync catchup time Statistics</h3>")
		diagnostics.WriteHistogramStatistics(w, c, syncCatchupTime)
		fmt.Fprintf(w, "</td></tr>")
	} else {
		fmt.Fprintf(w, "<tr><td>")
		fmt.Fprintf(w, "<h3>Sync catchup time Statistics</h3>")
		diagnostics.WriteHistogramStatistics(w, c, syncCatchupTime)
		fmt.Fprintf(w, "</td></tr>")
	}

	fmt.Fprintf(w, "</table>")
}

//
// swagger:route GET /v1/sharder/get/stats sharderstats
// a handler to get sharder stats
//
// responses:
//  200: ExplorerStats
//  404:

func SharderStatsHandler(ctx context.Context, r *http.Request) (interface{}, error) {
	sc := GetSharderChain()
	bss := sc.BlockSyncStats
	cc := bss.getCycleControl(ProximityScan)
	previous := &cc.counters.previous
	var previousElapsed string
	if previous.CycleStart.IsZero() {
		previousElapsed = "n/a"
	} else {
		previousElapsed = previous.CycleDuration.Round(time.Second).String()
	}
	selfNodeInfo := node.Self.Underlying().Info
	return ExplorerStats{LastFinalizedRound: sc.Chain.GetLatestFinalizedBlock().Round,
		StateHealth:            selfNodeInfo.StateMissingNodes,
		AverageBlockSize:       selfNodeInfo.AvgBlockTxns,
		PrevInvocationCount:    previous.HealthCheckInvocations,
		PrevInvocationScanTime: previousElapsed,
		MeanScanBlockStatsTime: cc.BlockSyncTimer.Mean() / 1000000.0,
	}, nil
}

func TransactionErrorWriter(w http.ResponseWriter, r *http.Request) {

	transactionErrors, err := GetSharderChain().Chain.GetEventDb().GetTransactionErrors()
	if err != nil {
		fmt.Fprintf(w, "Error getting transaction errors: %v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	chain.PrintCSS(w)
	diagnostics.WriteStatisticsCSS(w)

	fmt.Fprintf(w, "<h2>Transaction Output - Count</h2>")

	fmt.Fprintf(w, "<br>")
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr><td>")
	fmt.Fprintf(w, "<table width='100%%'>")

	for _, transactionError := range transactionErrors {
		fmt.Fprintf(w, "<tr><td class='tname'>%s</td><td>%d</td></tr>", transactionError.TransactionOutput, transactionError.Count)
	}

	fmt.Fprintf(w, "</td><td valign='top'>")
	fmt.Fprintf(w, "</table>")
}
