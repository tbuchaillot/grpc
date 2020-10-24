package main

import (
	"context"
	"github.com/tbuchaillot/grpc/calculator/calculatorpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"time"
)

func main(){
	log.Println("Initializing calculator client...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v",err)
	}
	defer conn.Close()

	c := calculatorpb.NewCalculatorServiceClient(conn)

	log.Println("-------------------------------")
	doUnary(c)
	log.Println("-------------------------------")
	doServerStreaming(c)
	log.Println("-------------------------------")
	doClientStreaming(c)
	log.Println("-------------------------------")
	doBiDiStreaming(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient){
	log.Println("Starting to do a Unary RPC..")
	req := &calculatorpb.SumRequest{
		FirstNumber: 1,
		SecondNumber: 2,
	}
	res, err := c.Sum(context.Background(),req)
	if err != nil {
		log.Fatalf("Error while calling Calculator RPC: %v",err)
	}

	log.Printf("Response from Calculator: %v",res.SumResult)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient){
	log.Println("Starting to do a Server Streaming RPC..")

	req:= &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 12,
	}

	resStream, err := c.PrimeNumberDecomposition(context.Background(),req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v",err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF{
			//we've reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("Error while call reading the stream: %v",err)
		}
		log.Printf("Response from PrimeNumberDecomposition: %v",msg.GetPrimeFactor())
	}

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient){
	log.Println("Starting to do a Client Streaming RPC..")

	numbers := []int32{3,5,9,54,23}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage RPC: %v", err)
	}

	//we iterate over the slice and send each msg
	for _, number:= range numbers{
		log.Printf("Sending number:%v \n" ,number)
		req := &calculatorpb.ComputeAverageRequest{
			Number:number,
		}
		stream.Send(req)
		time.Sleep(100*time.Millisecond)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while closing and receiving ComputeAverage RPC: %v", err)
	}
	log.Printf("ComputeAverage response: %v", response.GetAverage())
}

func doBiDiStreaming( c calculatorpb.CalculatorServiceClient){
	log.Println("Starting to do a FindMaximum BiDi Streaming RPC..")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while opening stream and calling FindMaximum:%v",err)
	}

	waitCh := make(chan struct{})

	//send go routine
	go func(){
			numbers := []int32{4,7,2,19,4,6,32}
			for _, number := range numbers{
				log.Printf("Sending number:%v\n",number)
				stream.Send(&calculatorpb.FindMaximumRequest{
					Number: number,
				})
				time.Sleep(1 * time.Second)
			}
			stream.CloseSend()
	}()

	//receive go routine
	go func(){
		for {
			res, err := stream.Recv()
			if err == io.EOF{
				break
			}
			if err != nil{
				log.Fatalf("Problem while reading server stream: %v",err)
				break
			}

			maximum := res.GetMaximum()
			log.Printf("Received a new maximum of ... :%v\n",maximum)
		}
		close(waitCh)
	}()

	//waiting for close
	<- waitCh
}