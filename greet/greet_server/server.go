package main

import (
	"context"
	"fmt"
	"github.com/tbuchaillot/grpc/greet/greetpb"
	"google.golang.org/grpc"
	"io"
	"log"
	"net"
	"time"
)

type server struct{}

func (srv *server) Greet(ctx context.Context,req *greetpb.GreetRequest) (*greetpb.GreetResponse, error){
	log.Printf("Greet function was invoked with %v",req)

	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()

	result := "Hello "+firstName+" "+ lastName
	res := &greetpb.GreetResponse{
		Result: result,
	}

	return res,nil
}

func (srv *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error{
	log.Printf("GreetManyTimes function was invoked with %v",req)

	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {

		result := "Hello " + firstName + " for the " + fmt.Sprint(i)+" time"
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}

		stream.Send(res)
		time.Sleep(1*time.Second)
	}
	return nil
}

func (srv *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error{
	log.Printf("LongGreet function was invoked ")
	result := ""

	for  {
		req, err := stream.Recv()
		if err == io.EOF{
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v",err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()

		result += "Hello "+firstName +"! "
	}

	return nil
}

func (srv *server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error{
	log.Printf("GreetEveryone function was invoked ")

	for {
		req, err := stream.Recv()
		if err == io.EOF{
			return nil
		}

		if err != nil{
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "!"
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		})
		if sendErr != nil {
			log.Fatalf("Error while sending data to client:%v", sendErr)
			return sendErr
		}
	}

	return nil
}

func main(){
	log.Println("Initializing the server...")

	lis, err := net.Listen("tcp","0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Fail to listen: %v", err)
	}

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s,&server{})

	log.Println("Serving at 50051")
	if err:= s.Serve(lis); err != nil {
		log.Fatalf("Fail to serve: %v",err)
	}
}