package benchutil

import (
	"fmt"
	"io/ioutil"

	"github.com/bazelbuild/rules_go/go/tools/bazel"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/go-ssz"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
)

// ValidatorCount is for declaring how many validators the benchmarks will be
// performed with. Default is 16384 or 524K ETH staked.
var ValidatorCount = uint64(16384)

// AttestationsPerEpoch represents the requested amount attestations in an epoch.
// This affects the amount of attestations in a fully attested for block and the amount
// of attestations in the state per epoch, so a full 2 epochs should result in twice
// this amount of attestations in the state. Default is 128.
var AttestationsPerEpoch = uint64(128)

// GenesisFileName is the generated genesis beacon state file name.
var GenesisFileName = fmt.Sprintf("bStateGenesis-%dAtts-%dVals.ssz", AttestationsPerEpoch, ValidatorCount)

// BState1EpochFileName is the generated beacon state after 1 skipped epoch file name.
var BState1EpochFileName = fmt.Sprintf("bState1Epoch-%dAtts-%dVals.ssz", AttestationsPerEpoch, ValidatorCount)

// BState2EpochFileName is the generated beacon state after 2 full epochs file name.
var BState2EpochFileName = fmt.Sprintf("bState2Epochs-%dAtts-%dVals.ssz", AttestationsPerEpoch, ValidatorCount)

// FullBlockFileName is the generated full block file name.
var FullBlockFileName = fmt.Sprintf("fullBlock-%dAtts-%dVals.ssz", AttestationsPerEpoch, ValidatorCount)

func filePath(path string) string {
	return fmt.Sprintf("shared/benchutil/benchmark_files/%s", path)
}

// PreGenState1Epoch unmarshals the pre-generated beacon state after 1 epoch of block processing and returns it.
func PreGenState1Epoch() (*pb.BeaconState, error) {
	path, err := bazel.Runfile(filePath(BState1EpochFileName))
	if err != nil {
		return nil, err
	}
	beaconBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	beaconState := &pb.BeaconState{}
	if err := ssz.Unmarshal(beaconBytes, beaconState); err != nil {
		return nil, err
	}
	return beaconState, nil
}

// PreGenState2FullEpochs unmarshals the pre-generated beacon state after 2 epoch of full block processing and returns it.
func PreGenState2FullEpochs() (*pb.BeaconState, error) {
	path, err := bazel.Runfile(filePath(BState2EpochFileName))
	if err != nil {
		return nil, err
	}
	beaconBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	beaconState := &pb.BeaconState{}
	if err := ssz.Unmarshal(beaconBytes, beaconState); err != nil {
		return nil, err
	}
	return beaconState, nil
}

// PreGenFullBlock unmarshals the pre-generated signed beacon block containing an epochs worth of attestations and returns it.
func PreGenFullBlock() (*ethpb.SignedBeaconBlock, error) {
	path, err := bazel.Runfile(filePath(FullBlockFileName))
	if err != nil {
		return nil, err
	}
	blockBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	beaconBlock := &ethpb.SignedBeaconBlock{}
	if err := ssz.Unmarshal(blockBytes, beaconBlock); err != nil {
		return nil, err
	}
	return beaconBlock, nil
}

// SetBenchmarkConfig changes the beacon config to match the requested amount of
// attestations set to AttestationsPerEpoch.
func SetBenchmarkConfig() {
	maxAtts := AttestationsPerEpoch
	slotsPerEpoch := params.BeaconConfig().SlotsPerEpoch
	committeeSize := (ValidatorCount / slotsPerEpoch) / (maxAtts / slotsPerEpoch)
	c := params.BeaconConfig()
	c.PersistentCommitteePeriod = 0
	c.MinValidatorWithdrawabilityDelay = 0
	c.TargetCommitteeSize = committeeSize
	c.MaxAttestations = maxAtts
	params.OverrideBeaconConfig(c)
}