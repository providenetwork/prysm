/*
Package beaconclient defines a service that interacts with a beacon
node via a gRPC client to listen for streamed blocks, attestations, and to
submit proposer/attester slashings to the node in case they are detected.
*/
package beaconclient

import (
	"context"
	"errors"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/shared/event"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var log = logrus.WithField("prefix", "beaconclient")

// Service struct for the beaconclient service of the slasher.
type Service struct {
	context         context.Context
	cancel          context.CancelFunc
	cert            string
	conn            *grpc.ClientConn
	provider        string
	client          ethpb.BeaconChainClient
	blockFeed       *event.Feed
	attestationFeed *event.Feed
}

// BlockFeed returns a feed other services in slasher can subscribe to
// blocks received via the beacon node through gRPC.
func (bs *Service) BlockFeed() *event.Feed {
	return bs.blockFeed
}

// AttestationFeed returns a feed other services in slasher can subscribe to
// attestations received via the beacon node through gRPC.
func (bs *Service) AttestationFeed() *event.Feed {
	return bs.attestationFeed
}

// Stop the beacon client service by closing the gRPC connection.
func (bs *Service) Stop() error {
	bs.cancel()
	log.Info("Stopping service")
	if bs.conn != nil {
		return bs.conn.Close()
	}
	return nil
}

// Status returns an error if there exists a gRPC connection error
// in the service.
func (bs *Service) Status() error {
	if bs.conn == nil {
		return errors.New("no connection to beacon RPC")
	}
	return nil
}

// Start the main runtime of the beaconclient service, initializing
// a gRPC client connection with a beacon node, listening for
// streamed blocks/attestations, and submitting slashing operations
// after they are detected by other services in the slasher.
func (bs *Service) Start() {
	var dialOpt grpc.DialOption
	if bs.cert != "" {
		creds, err := credentials.NewClientTLSFromFile(bs.cert, "")
		if err != nil {
			log.Errorf("Could not get valid credentials: %v", err)
		}
		dialOpt = grpc.WithTransportCredentials(creds)
	} else {
		dialOpt = grpc.WithInsecure()
		log.Warn(
			"You are using an insecure gRPC connection to beacon chain! Please provide a certificate and key to use a secure connection",
		)
	}
	beaconOpts := []grpc.DialOption{
		dialOpt,
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithStreamInterceptor(middleware.ChainStreamClient(
			grpc_opentracing.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
		)),
		grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
			grpc_opentracing.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		)),
	}
	conn, err := grpc.DialContext(bs.context, bs.provider, beaconOpts...)
	if err != nil {
		log.Fatalf("Could not dial endpoint: %s, %v", bs.provider, err)
	}
	log.Info("Successfully started gRPC connection")
	bs.conn = conn
	bs.client = ethpb.NewBeaconChainClient(bs.conn)

	go bs.receiveBlocks(bs.context)
	go bs.receiveAttestations(bs.context)
}
