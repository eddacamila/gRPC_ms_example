package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	pb "gRPC_ms_example/transport"

	"google.golang.org/grpc"
)

var (
	bucketName = "data-lake-example-as"
	region     = "us-east-2" // ajusta según tu región
)

type server struct {
	pb.UnimplementedTransportServiceServer
	s3Client *s3.Client
}

func (s *server) SendOperationalData(ctx context.Context, data *pb.OperationalData) (*pb.Response, error) {
	log.Printf(":box: Datos recibidos:\n- Operador: %s\n- Ruta: %s\n- Ocupación: %d\n- Estado: %s\n- Tiempo: %s\n",
		data.OperatorId, data.RouteId, data.Occupancy, data.VehicleStatus, data.Timestamp)

	// Estructura del archivo
	content := fmt.Sprintf(
		"Operador: %s\nRuta: %s\nOcupación: %d\nEstado: %s\nTiempo: %s\n",
		data.OperatorId, data.RouteId, data.Occupancy, data.VehicleStatus, data.Timestamp,
	)

	key := fmt.Sprintf("datos_operacionales/%s_%s_%d.txt", data.OperatorId, data.RouteId, time.Now().Unix())

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &key,
		Body:        bytes.NewReader([]byte(content)),
		ContentType: awsString("text/plain"),
		ACL:         types.ObjectCannedACLPrivate,
	})
	if err != nil {
		log.Printf("❌ Error al guardar en S3: %v", err)
		return &pb.Response{Message: "❌ Fallo al almacenar datos"}, nil
	}

	return &pb.Response{Message: "✅ Datos recibidos y almacenados en S3"}, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("❌ No se pudo cargar la configuración AWS: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Error al escuchar: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTransportServiceServer(grpcServer, &server{s3Client: s3Client})

	log.Println("🚀 Servidor gRPC escuchando en puerto 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("❌ Error al servir: %v", err)
	}
}

func awsString(s string) *string {
	return &s
}
