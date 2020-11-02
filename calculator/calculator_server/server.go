package main

import (
	"context"
	"fmt"
	"math"

	"github.com/tbuchaillot/grpc/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/reflection"

	"io"
	"log"
	"net"
)

type server struct{}

func (srv *server) Sum(ctx context.Context,req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error){
	log.Printf("Sum function was invoked with %v",req)

	firstNumber := req.GetFirstNumber()
	secondNumber := req.GetSecondNumber()

	result := firstNumber+ secondNumber
	res := &calculatorpb.SumResponse{
		SumResult: result,
	}

	return res,nil
}

func (srv *server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error{
	log.Printf("PrimeNumberDecomposition function was invoked with %v",req)

	number := req.GetNumber()
	divisor := int64(2)
	for number > 1{
		if number %divisor == 0{
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			number= number/divisor
		}else{
			divisor++
			log.Printf("Divisor has increased to %v",divisor)
		}
	}

	return nil
}

func (srv *server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error{
	log.Println("ComputeAverage function was invoked")

	sum := int32(0)
	count := 0

	for  {
		req, err := stream.Recv()
		if err == io.EOF{
			average := float64(sum) / float64(count)
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: average,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v",err)
		}

		sum +=  req.GetNumber()
		count ++
	}

	return nil
}

func (srv *server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error{
	log.Println("FindMaximum function was invoked")

	maximum := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF{
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v",err)
			return nil
		}

		number := req.GetNumber()
		if number > maximum{
			maximum = number
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maximum,
			})
			if sendErr != nil{
				log.Fatalf("Error while sending data to client: %v",sendErr)
				return nil
			}
		}

	}

	return nil
}

func (srv *server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error){
	log.Println("SquareRoot function was invoked")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
				codes.InvalidArgument,
				fmt.Sprintf("Received a negative number: %v", number),
			)
	}
	response := &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}
	return response,nil
}

func main(){
	log.Println("Initializing calculator server...")

	lis, err := net.Listen("tcp","0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Fail to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s,&server{})
	reflection.Register(s)

	log.Println("Serving at 50051")
	if err:= s.Serve(lis); err != nil {
		log.Fatalf("Fail to serve: %v",err)
	}
}