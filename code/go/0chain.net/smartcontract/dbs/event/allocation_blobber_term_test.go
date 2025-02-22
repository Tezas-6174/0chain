package event

import (
	"strconv"
	"testing"

	"0chain.net/core/encryption"
	common2 "0chain.net/smartcontract/common"
	"github.com/0chain/common/core/currency"
	"github.com/0chain/common/core/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	logging.Logger = zap.NewNop()
}

func TestAllocationBlobberTerms(t *testing.T) {
	t.Run("test edb.addOrOverwriteAllocationBlobberTerms", func(t *testing.T) {
		eventDb, clean := GetTestEventDB(t)
		defer clean()

		// Create owner and allocation
		err := eventDb.Get().Model(&User{}).Create(&User{
			UserID: OwnerId,
		}).Error
		require.NoError(t, err, "owner couldn't be created")

		allocId := createMockAllocations(t, eventDb, 1)[0]
		blobber1Id := encryption.Hash("mockBlobber_" + strconv.Itoa(0))
		blobber2Id := encryption.Hash("mockBlobber_" + strconv.Itoa(1))

		terms := []AllocationBlobberTerm{
			{
				AllocationID:  allocId,
				BlobberID:     blobber1Id,
				ReadPrice:     int64(currency.Coin(29)),
				WritePrice:    int64(currency.Coin(31)),
				MinLockDemand: 37.0,
			},
			{
				AllocationID:  allocId,
				BlobberID:     blobber2Id,
				ReadPrice:     int64(currency.Coin(41)),
				WritePrice:    int64(currency.Coin(43)),
				MinLockDemand: 47.0,
			},
		}

		err = eventDb.addOrOverwriteAllocationBlobberTerms(terms)
		require.NoError(t, err, "Error while inserting Allocation's Blobber's AllocationBlobberTerm to event database")

		var term *AllocationBlobberTerm
		var res []AllocationBlobberTerm
		limit := common2.Pagination{
			Offset:       0,
			Limit:        20,
			IsDescending: true,
		}
		res, err = eventDb.GetAllocationBlobberTerms(terms[0].AllocationID, limit)
		require.Equal(t, 2, len(res), "AllocationBlobberTerm not getting inserted")

		terms[1].MinLockDemand = 70.0
		err = eventDb.addOrOverwriteAllocationBlobberTerms(terms)
		require.NoError(t, err, "Error while inserting Allocation's Blobber's AllocationBlobberTerm to event database")

		term, err = eventDb.GetAllocationBlobberTerm(terms[1].AllocationID, terms[1].BlobberID)
		require.Equal(t, terms[1].MinLockDemand, term.MinLockDemand, "Error while overriding AllocationBlobberTerm in event Database")
	})

	t.Run("test edb.updateAllocationBlobberTerms", func(t *testing.T) {
		eventDb, clean := GetTestEventDB(t)
		defer clean()

		// Create owner and allocation
		err := eventDb.Get().Model(&User{}).Create(&User{
			UserID: OwnerId,
		}).Error
		require.NoError(t, err, "owner couldn't be created")

		allocId := createMockAllocations(t, eventDb, 1)[0]
		blobber1Id := encryption.Hash("mockBlobber_" + strconv.Itoa(0))
		blobber2Id := encryption.Hash("mockBlobber_" + strconv.Itoa(1))

		terms := []AllocationBlobberTerm{
			{
				AllocationID:  allocId,
				BlobberID:     blobber1Id,
				ReadPrice:     int64(currency.Coin(29)),
				WritePrice:    int64(currency.Coin(31)),
				MinLockDemand: 37.0,
			},
			{
				AllocationID:  allocId,
				BlobberID:     blobber2Id,
				ReadPrice:     int64(currency.Coin(41)),
				WritePrice:    int64(currency.Coin(43)),
				MinLockDemand: 47.0,
			},
		}

		err = eventDb.addOrOverwriteAllocationBlobberTerms(terms)
		require.NoError(t, err, "Error while inserting Allocation's Blobber's AllocationBlobberTerm to event database")

		err = eventDb.updateAllocationBlobberTerms([]AllocationBlobberTerm{
			{
				AllocationID:  allocId,
				BlobberID:     blobber1Id,
				ReadPrice:     int64(currency.Coin(59)),
				WritePrice:    int64(currency.Coin(61)),
				MinLockDemand: 57.0,
			}, {
				AllocationID:  allocId,
				BlobberID:     blobber2Id,
				ReadPrice:     int64(currency.Coin(61)),
				WritePrice:    int64(currency.Coin(63)),
				MinLockDemand: 67.0,
			},
		})
		require.NoError(t, err, "Error while updating Allocation's Blobber's AllocationBlobberTerm to event database")

		term, err := eventDb.GetAllocationBlobberTerm(allocId, blobber1Id)
		require.NoError(t, err, "Error while reading Allocation Blobber Terms")

		require.Equal(t, int64(currency.Coin(59)), term.ReadPrice)
		require.Equal(t, int64(currency.Coin(61)), term.WritePrice)
		require.Equal(t, float64(57.0), term.MinLockDemand)

		term, err = eventDb.GetAllocationBlobberTerm(allocId, blobber2Id)
		require.NoError(t, err, "Error while reading Allocation Blobber Terms")

		require.Equal(t, int64(currency.Coin(61)), term.ReadPrice)
		require.Equal(t, int64(currency.Coin(63)), term.WritePrice)
		require.Equal(t, float64(67.0), term.MinLockDemand)
	})
}
