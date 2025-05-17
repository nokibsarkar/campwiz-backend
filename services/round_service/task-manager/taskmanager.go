package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"nokib/campwiz/models"
	distributionstrategy "nokib/campwiz/services/round_service/task-manager/distribution-strategy"
	importsources "nokib/campwiz/services/round_service/task-manager/import-sources"
	statisticsupdater "nokib/campwiz/services/round_service/task-manager/statistics-updater"

	"google.golang.org/grpc"
)

func main() {
	var (
		// tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
		// certFile = flag.String("cert_file", "", "The TLS cert file")
		// keyFile  = flag.String("key_file", "", "The TLS key file")
		port = flag.Int("rpcport", 50051, "The server port")
		host = flag.String("rpchost", "0.0.0.0", "The server host")
	)
	flag.Parse()
	var opts []grpc.ServerOption
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// if *tls {
	// 	if *certFile == "" {
	// 		*certFile = data.Path("x509/server_cert.pem")
	// 	}
	// 	if *keyFile == "" {
	// 		*keyFile = data.Path("x509/server_key.pem")
	// 	}
	// 	creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
	// 	if err != nil {
	// 		log.Fatalf("Failed to generate credentials: %v", err)
	// 	}
	// 	opts = []grpc.ServerOption{grpc.Creds(creds)}
	// }
	grpcServer := grpc.NewServer(opts...)
	models.RegisterImporterServer(grpcServer, importsources.NewImporterServer())
	models.RegisterDistributorServer(grpcServer, distributionstrategy.NewDistributorServer())
	models.RegisterStatisticsUpdaterServer(grpcServer, statisticsupdater.NewStatisticsUpdaterServer())

	log.Printf("Task Manager Server listening at %v", lis.Addr())
	grpcServer.Serve(lis)
}
