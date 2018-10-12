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
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	pb "github.pie.apple.com/privatecloud/spawn/protobuf"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		

		// Create a metrics registry.
	reg := prometheus.NewRegistry()
	// Create some standard client metrics.
	grpcMetrics := grpc_prometheus.NewClientMetrics()
	// Register client metrics to registry.
	reg.MustRegister(grpcMetrics)
	// Create a insecure gRPC channel to communicate with the server.
	conn, err := grpc.Dial(
		fmt.Sprintf(":3100"),
		grpc.WithUnaryInterceptor(grpcMetrics.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpcMetrics.StreamClientInterceptor()),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	
		// Create a HTTP server for prometheus.
		httpServer := &http.Server{Handler: promhttp.HandlerFor(reg, promhttp.HandlerOpts{}), Addr: fmt.Sprintf(":3102")}
	
		// Start your http server for prometheus.
		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				log.Fatal("Unable to start a http server.")
			}
		}()
	
		// Create a gRPC server client.
		client := pb.NewDemoServiceClient(conn)
		fmt.Println("Start to call the method called SayHello every 3 seconds")
		go func() {
			for {
				// Call “SayHello” method and wait for response from gRPC Server.
				_, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "Test"})
				if err != nil {
					log.Printf("Calling the SayHello method unsuccessfully. ErrorInfo: %+v", err)
					log.Printf("You should to stop the process")
					return
				}
				time.Sleep(3 * time.Second)
			}
		}()
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("You can press n or N to stop the process of client")
		for scanner.Scan() {
			if strings.ToLower(scanner.Text()) == "n" {
				os.Exit(0)
			}
		}	
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
	
	reg := prometheus.NewRegistry()

	// Create some standard client metrics.
	grpcMetrics := grpc_prometheus.NewClientMetrics()
	// Register client metrics to registry.
	reg.MustRegister(grpcMetrics)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	
}
