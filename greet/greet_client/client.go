package main

import (
	"context"
	"github.com/tbuchaillot/grpc/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"io"
	"log"
	"time"
)

func main(){
	log.Println("Initializing the client...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v",err)
	}
	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)

	doUnary(c)
	doServerStreaming(c)
	doClientStreaming(c)
	doBiDiStreaming(c)
	//this will fail
	doUnaryWithDeadline(c, 1* time.Second)
	//this will be ok
	doUnaryWithDeadline(c, 5*time.Second)
}

func doUnary(c greetpb.GreetServiceClient){
	log.Println("Starting to do a Unary RPC..")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Tom",
			LastName: "Buchaillot",
		},
	}
	res, err := c.Greet(context.Background(),req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v",err)
	}

	log.Printf("Response from Greet: %v",res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient){
	log.Println("Starting to do a Server Streaming RPC..")

	req:= &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Tom",
			LastName: "Buchaillot",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(),req)
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
		log.Printf("Response from Greet: %v",msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient){
	log.Println("Starting to do a Client Streaming RPC..")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tomas",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Chris",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nana",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tomas again",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet RPC: %v", err)
	}

	//we iterate over the slice and send each msg
	for _, req:= range requests{
		log.Printf("Sending req:%v \n" ,req.Greeting.FirstName)
		stream.Send(req)
		time.Sleep(100*time.Millisecond)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while closing and receiving LongGreet RPC: %v", err)
	}
	log.Printf("LongGreet response: %v", response.GetResult())
}

func doBiDiStreaming(c greetpb.GreetServiceClient){
	log.Println("Starting to do a BiDi Streaming RPC..")

	// we create a stream by invoking the client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tomas",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Chris",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Nana",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tomas again",
			},
		},
	}

	waitCh := make(chan struct{})
	// we send a buch of msgs to the client (go routine)
	go func(){
		//func to send a bunch of msgs
		for _, req := range requests{
			log.Printf("Sending msg: %v \n",req)
			stream.Send(req)
			time.Sleep(1 * time.Second)
		}
		stream.CloseSend()
	}()

	// we receive a bunch of msgs from the client (go routine)
	go func(){
		//func to receive a bunch of msgs
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v",err)
				break
			}

			log.Printf("Received: %v \n", res.GetResult())
		}
		close(waitCh)

	}()
	// block until everything is done
	<- waitCh
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration){
	log.Println("Starting to do a UnaryWithDeadline RPC..")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Tom",
			LastName: "Buchaillot",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx,req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				log.Println("Timeout was hit! Deadline was execeded")
			}else{
				log.Fatalf("Unexpected error: %v", statusErr)
			}
		}else{
			log.Fatalf("Error while calling UnaryWithDeadline RPC: %v",err)
		}
		return
	}

	log.Printf("Response from UnaryWithDeadline: %v",res.Result)
}