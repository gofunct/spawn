// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.pie.apple.com/privatecloud/spawn/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)



var (
	// Create a metrics registry.
	reg = prometheus.NewRegistry()

	// Create some standard server metrics.
	grpcMetrics = grpc_prometheus.NewServerMetrics()

	// Create a customized counter metric.
	customizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "demo_server_say_hello_method_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
// Listen an actual port.
lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 3100))
if err != nil {
	log.Fatalf("failed to listen: %v", err)
}
defer lis.Close()

reg.MustRegister(grpcMetrics, customizedCounterMetric)
	customizedCounterMetric.WithLabelValues("Test")
// Create a HTTP server for prometheus.
httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf(":3101")}

// Create a gRPC Server with gRPC interceptor.
grpcServer := grpc.NewServer(
	grpc.StreamInterceptor(grpcMetrics.StreamServerInterceptor()),
	grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
)

// Create a new api server.
demoServer := newDemoServer()

// Register your service.
pb.RegisterDemoServiceServer(grpcServer, demoServer)

// Initialize all metrics.
grpcMetrics.InitializeMetrics(grpcServer)

// Start your http server for prometheus.
go func() {
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal("Unable to start a http server.")
	}
}()

// Start your gRPC server.
log.Fatal(grpcServer.Serve(lis))	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	// Register standard server metrics and customized metrics to registry.
	

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// DemoServiceServer defines a Server.
type DemoServiceServer struct{}

func newDemoServer() *DemoServiceServer {
	return &DemoServiceServer{}
}

// SayHello implements a interface defined by protobuf.
func (s *DemoServiceServer) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloResponse, error) {
	customizedCounterMetric.WithLabelValues(request.Name).Inc()
	return &pb.HelloResponse{Message: fmt.Sprintf("Hello %s", request.Name)}, nil
}