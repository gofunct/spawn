package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.pie.apple.com/privatecloud/spawn/cmd"

)



func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(tamird): point to merged gRPC code rather than a PR.
		// This is a partial recreation of gRPC's internal checks https://github.com/grpc/grpc-go/pull/514/files#diff-95e9a25b738459a2d3030e1e6fa2a718R61
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

// ListenAndServe starts grpc server
func ListenAndServe(grpcServer *grpc.Server, otherHandler http.Handler) error {
	lis, err := net.Listen("tcp", *ListenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	if *serverCert != "" {
		serverCertKeypair, err := tls.LoadX509KeyPair(*serverCert, *serverKey)
		if err != nil {
			return fmt.Errorf("failed to load server tls cert/key: %v", err)
		}

		var clientCertPool *x509.CertPool
		if *clientCA != "" {
			caCert, err := ioutil.ReadFile(*clientCA)
			if err != nil {
				return fmt.Errorf("failed to load client ca: %v", err)
			}
			clientCertPool = x509.NewCertPool()
			clientCertPool.AppendCertsFromPEM(caCert)
		}

		var h http.Handler
		if otherHandler == nil {
			h = grpcServer
		} else {
			h = grpcHandlerFunc(grpcServer, otherHandler)
		}

		httpsServer := &http.Server{
			Handler: h,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{serverCertKeypair},
				NextProtos:   []string{"h2"},
			},
		}

		if clientCertPool != nil {
			httpsServer.TLSConfig.ClientCAs = clientCertPool
			httpsServer.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			glog.Warningf("no client ca provided for grpc server")
		}

		glog.Infof("serving on %v", *ListenAddress)
		err = httpsServer.Serve(tls.NewListener(lis, httpsServer.TLSConfig))
		return fmt.Errorf("failed to serve: %v", err)
	}

	glog.Warningf("serving INSECURE on %v", *ListenAddress)
	err = grpcServer.Serve(lis)
	return fmt.Errorf("failed to serve: %v", err)
}

// NewServer creates a new GRPC server stub with credstore auth (if requested).
func NewServer() (*grpc.Server, *client.CredstoreClient, error) {
	var grpcServer *grpc.Server
	var cc *client.CredstoreClient

	if *credStoreAddress != "" {
		var err error
		cc, err = client.NewCredstoreClient(context.Background(), *credStoreAddress, *credStoreCA)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to init credstore: %v", err)
		}

		glog.Infof("enabled credstore auth")
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer()),
				grpc_prometheus.UnaryServerInterceptor,
				client.CredStoreTokenInterceptor(cc.SigningKey()),
				client.CredStoreMethodAuthInterceptor(),
			)))
	} else {
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(
				otgrpc.OpenTracingServerInterceptor(opentracing.GlobalTracer())))
	}

	reflection.Register(grpcServer)
	grpc_prometheus.Register(grpcServer)

	return grpcServer, cc, nil
}

// NewGRPCConn is a helper wrapper around grpc.Dial.
func NewGRPCConn(
	address string,
	serverCAFileName string,
	clientCertFileName string,
	clientKeyFileName string,
) (*grpc.ClientConn, error) {
	if serverCAFileName == "" {
		return grpc.Dial(address,
			grpc.WithInsecure(),
			grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())))
	}

	caCert, err := ioutil.ReadFile(serverCAFileName)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cfg := &tls.Config{
		RootCAs: caCertPool,
	}

	if clientCertFileName != "" && clientKeyFileName != "" {
		peerCert, err := tls.LoadX509KeyPair(clientCertFileName, clientKeyFileName)
		if err != nil {
			return nil, err
		}
		cfg.Certificates = []tls.Certificate{peerCert}
	}

	return grpc.Dial(address,
		grpc.WithTransportCredentials(credentials.NewTLS(cfg)),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(opentracing.GlobalTracer())),
	)
}