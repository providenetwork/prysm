package interop_cold_start

import (
	"context"
	"math/big"

	"github.com/pkg/errors"
	"github.com/prysmaticlabs/go-ssz"
	"github.com/prysmaticlabs/prysm/beacon-chain/cache/depositcache"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/blocks"
	"github.com/prysmaticlabs/prysm/beacon-chain/db"
	"github.com/prysmaticlabs/prysm/beacon-chain/powchain"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	ethpb "github.com/prysmaticlabs/prysm/proto/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared"
	"github.com/prysmaticlabs/prysm/shared/bytesutil"
	"github.com/prysmaticlabs/prysm/shared/interop"
)

var _ = shared.Service(&Service{})

type Service struct {
	ctx           context.Context
	cancel        context.CancelFunc
	genesisTime   uint64
	numValidators uint64
	beaconDB      db.Database
	powchain      powchain.Service
	depositCache  *depositcache.DepositCache
}

type Config struct {
	GenesisTime   uint64
	NumValidators uint64
	BeaconDB      db.Database
	DepositCache  *depositcache.DepositCache
}

// NewColdStartService is an interoperability testing service to inject a deterministically generated genesis state
// into the beacon chain database and running services at start up. This service should not be used in production
// as it does not have any value other than ease of use for testing purposes.
func NewColdStartService(ctx context.Context, cfg *Config) *Service {
	log.Warn("Saving generated genesis state in database for interop testing.")
	ctx, cancel := context.WithCancel(ctx)

	s := &Service{
		ctx:           ctx,
		cancel:        cancel,
		genesisTime:   cfg.GenesisTime,
		numValidators: cfg.NumValidators,
		beaconDB:      cfg.BeaconDB,
		depositCache:  cfg.DepositCache,
	}

	// Save genesis state in db
	genesisState, deposits, err := interop.GenerateGenesisState(s.genesisTime, s.numValidators)
	if err != nil {
		log.Fatalf("Could not generate interop genesis state: %v", err)
	}
	if err := s.saveGenesisState(ctx, genesisState, deposits); err != nil {
		log.Fatalf("Could not save interop genesis state %v", err)
	}

	return s
}

// Start initializes the genesis state from configured flags.
func (s *Service) Start() {
	// TODO: Does this need to be a service?
}

// Stop does nothing.
func (s *Service) Stop() error {
	return nil
}

// Status always returns nil.
func (s *Service) Status() error {
	return nil
}

func (s *Service) saveGenesisState(ctx context.Context, genesisState *pb.BeaconState, deposits []*ethpb.Deposit) error {
	stateRoot, err := ssz.HashTreeRoot(genesisState)
	if err != nil {
		return errors.Wrap(err, "could not tree hash genesis state")
	}
	genesisBlk := blocks.NewGenesisBlock(stateRoot[:])
	genesisBlkRoot, err := ssz.SigningRoot(genesisBlk)
	if err != nil {
		return errors.Wrap(err, "could not get genesis block root")
	}

	if err := s.beaconDB.SaveBlock(ctx, genesisBlk); err != nil {
		return errors.Wrap(err, "could not save genesis block")
	}
	if err := s.beaconDB.SaveHeadBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save head block root")
	}
	if err := s.beaconDB.SaveGenesisBlockRoot(ctx, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could save genesis block root")
	}
	if err := s.beaconDB.SaveState(ctx, genesisState, genesisBlkRoot); err != nil {
		return errors.Wrap(err, "could not save genesis state")
	}
	for i, v := range genesisState.Validators {
		if err := s.beaconDB.SaveValidatorIndex(ctx, bytesutil.ToBytes48(v.PublicKey), uint64(i)); err != nil {
			return errors.Wrapf(err, "could not save validator index: %d", i)
		}
		s.depositCache.MarkPubkeyForChainstart(ctx, string(v.PublicKey))
	}
	for i, dep := range deposits {
		s.depositCache.InsertDeposit(ctx, dep, big.NewInt(0), i, [32]byte{})
		s.depositCache.InsertChainStartDeposit(ctx, dep)
	}
	return nil
}
