// Package iface exists to prevent circular dependencies when implementing the database interface.
package iface

import (
	"context"
	"io"

	"github.com/ethereum/go-ethereum/common"
	eth "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/db/filters"
	"github.com/prysmaticlabs/prysm/beacon-chain/state"
	"github.com/prysmaticlabs/prysm/proto/beacon/db"
	ethereum_beacon_p2p_v1 "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
)

// ReadOnlyDatabase -- See github.com/prysmaticlabs/prysm/beacon-chain/db.ReadOnlyDatabase
type ReadOnlyDatabase interface {
	// Attestation related methods.
	AttestationsByDataRoot(ctx context.Context, attDataRoot [32]byte) ([]*eth.Attestation, error)
	Attestations(ctx context.Context, f *filters.QueryFilter) ([]*eth.Attestation, error)
	HasAttestation(ctx context.Context, attDataRoot [32]byte) bool
	// Block related methods.
	Block(ctx context.Context, blockRoot [32]byte) (*eth.SignedBeaconBlock, error)
	Blocks(ctx context.Context, f *filters.QueryFilter) ([]*eth.SignedBeaconBlock, error)
	BlockRoots(ctx context.Context, f *filters.QueryFilter) ([][32]byte, error)
	HasBlock(ctx context.Context, blockRoot [32]byte) bool
	GenesisBlock(ctx context.Context) (*ethpb.SignedBeaconBlock, error)
	IsFinalizedBlock(ctx context.Context, blockRoot [32]byte) bool
	// Validator related methods.
	ValidatorIndex(ctx context.Context, publicKey []byte) (uint64, bool, error)
	HasValidatorIndex(ctx context.Context, publicKey []byte) bool
	// State related methods.
	State(ctx context.Context, blockRoot [32]byte) (*state.BeaconState, error)
	GenesisState(ctx context.Context) (*state.BeaconState, error)
	HasState(ctx context.Context, blockRoot [32]byte) bool
	// Slashing operations.
	ProposerSlashing(ctx context.Context, slashingRoot [32]byte) (*eth.ProposerSlashing, error)
	AttesterSlashing(ctx context.Context, slashingRoot [32]byte) (*eth.AttesterSlashing, error)
	HasProposerSlashing(ctx context.Context, slashingRoot [32]byte) bool
	HasAttesterSlashing(ctx context.Context, slashingRoot [32]byte) bool
	// Block operations.
	VoluntaryExit(ctx context.Context, exitRoot [32]byte) (*eth.VoluntaryExit, error)
	HasVoluntaryExit(ctx context.Context, exitRoot [32]byte) bool
	// Checkpoint operations.
	JustifiedCheckpoint(ctx context.Context) (*eth.Checkpoint, error)
	FinalizedCheckpoint(ctx context.Context) (*eth.Checkpoint, error)
	// Archival data handlers for storing/retrieving historical beacon node information.
	ArchivedActiveValidatorChanges(ctx context.Context, epoch uint64) (*ethereum_beacon_p2p_v1.ArchivedActiveSetChanges, error)
	ArchivedCommitteeInfo(ctx context.Context, epoch uint64) (*ethereum_beacon_p2p_v1.ArchivedCommitteeInfo, error)
	ArchivedBalances(ctx context.Context, epoch uint64) ([]uint64, error)
	ArchivedValidatorParticipation(ctx context.Context, epoch uint64) (*eth.ValidatorParticipation, error)
	// Deposit contract related handlers.
	DepositContractAddress(ctx context.Context) ([]byte, error)
	// Powchain operations.
	PowchainData(ctx context.Context) (*db.ETH1ChainData, error)
}

// NoHeadAccessDatabase -- See github.com/prysmaticlabs/prysm/beacon-chain/db.NoHeadAccessDatabase
type NoHeadAccessDatabase interface {
	ReadOnlyDatabase

	// Attestation related methods.
	DeleteAttestation(ctx context.Context, attDataRoot [32]byte) error
	DeleteAttestations(ctx context.Context, attDataRoots [][32]byte) error
	SaveAttestation(ctx context.Context, att *eth.Attestation) error
	SaveAttestations(ctx context.Context, atts []*eth.Attestation) error
	// Block related methods.
	DeleteBlock(ctx context.Context, blockRoot [32]byte) error
	DeleteBlocks(ctx context.Context, blockRoots [][32]byte) error
	SaveBlock(ctx context.Context, block *eth.SignedBeaconBlock) error
	SaveBlocks(ctx context.Context, blocks []*eth.SignedBeaconBlock) error
	SaveGenesisBlockRoot(ctx context.Context, blockRoot [32]byte) error
	// Validator related methods.
	DeleteValidatorIndex(ctx context.Context, publicKey []byte) error
	SaveValidatorIndex(ctx context.Context, publicKey []byte, validatorIdx uint64) error
	SaveValidatorIndices(ctx context.Context, publicKeys [][48]byte, validatorIndices []uint64) error
	// State related methods.
	SaveState(ctx context.Context, state *state.BeaconState, blockRoot [32]byte) error
	SaveStates(ctx context.Context, states []*state.BeaconState, blockRoots [][32]byte) error
	DeleteState(ctx context.Context, blockRoot [32]byte) error
	DeleteStates(ctx context.Context, blockRoots [][32]byte) error
	// Slashing operations.
	SaveProposerSlashing(ctx context.Context, slashing *eth.ProposerSlashing) error
	SaveAttesterSlashing(ctx context.Context, slashing *eth.AttesterSlashing) error
	DeleteProposerSlashing(ctx context.Context, slashingRoot [32]byte) error
	DeleteAttesterSlashing(ctx context.Context, slashingRoot [32]byte) error
	// Block operations.
	SaveVoluntaryExit(ctx context.Context, exit *eth.VoluntaryExit) error
	DeleteVoluntaryExit(ctx context.Context, exitRoot [32]byte) error
	// Checkpoint operations.
	SaveJustifiedCheckpoint(ctx context.Context, checkpoint *eth.Checkpoint) error
	SaveFinalizedCheckpoint(ctx context.Context, checkpoint *eth.Checkpoint) error
	// Archival data handlers for storing/retrieving historical beacon node information.
	SaveArchivedActiveValidatorChanges(ctx context.Context, epoch uint64, changes *ethereum_beacon_p2p_v1.ArchivedActiveSetChanges) error
	SaveArchivedCommitteeInfo(ctx context.Context, epoch uint64, info *ethereum_beacon_p2p_v1.ArchivedCommitteeInfo) error
	SaveArchivedBalances(ctx context.Context, epoch uint64, balances []uint64) error
	SaveArchivedValidatorParticipation(ctx context.Context, epoch uint64, part *eth.ValidatorParticipation) error
	// Deposit contract related handlers.
	SaveDepositContractAddress(ctx context.Context, addr common.Address) error
	// Powchain operations.
	SavePowchainData(ctx context.Context, data *db.ETH1ChainData) error
}

// HeadAccessDatabase -- See github.com/prysmaticlabs/prysm/beacon-chain/db.HeadAccessDatabase
type HeadAccessDatabase interface {
	NoHeadAccessDatabase

	// Block related methods.
	HeadBlock(ctx context.Context) (*eth.SignedBeaconBlock, error)
	SaveHeadBlockRoot(ctx context.Context, blockRoot [32]byte) error
	// State related methods.
	HeadState(ctx context.Context) (*state.BeaconState, error)
}

// Database -- See github.com/prysmaticlabs/prysm/beacon-chain/db.Database
type Database interface {
	io.Closer
	HeadAccessDatabase

	DatabasePath() string
	ClearDB() error

	// Backup and restore methods
	Backup(ctx context.Context) error
}
