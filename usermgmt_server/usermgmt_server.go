package main

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"

	pb "example.com/go-usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	port = ":50051"
)

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{}
}

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
}

func (server *UserManagementServer) Run() error {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserManagementServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	return s.Serve(lis)
}

func writeJsonFile(users_list *pb.UserList) {
	jsonBytes, err := protojson.Marshal(users_list)
	if err != nil {
		log.Fatalf("JSON Marshaling failed: %v", err)
	}
	if err := ioutil.WriteFile("users.json", jsonBytes, 0664); err != nil {
		log.Fatalf("Failed write to file")
	}
}

func (s *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())
	readBytes, err := ioutil.ReadFile("users.json")
	var users_list *pb.UserList = &pb.UserList{}
	var user_id int32 = int32(rand.Intn(1000))
	created_user := &pb.User{Name: in.GetName(), Age: in.GetAge(), Id: user_id}
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("File not found. Creting a new file")
			users_list.Users = append(users_list.Users, created_user)
			writeJsonFile(users_list)
			return created_user, nil
		} else {
			log.Fatalf("Error reading file: %v", err)
		}
	}
	if err := protojson.Unmarshal(readBytes, users_list); err != nil {
		log.Fatalf("Failed to parse user list: %v", err)
	}
	users_list.Users = append(users_list.Users, created_user)
	writeJsonFile(users_list)
	return created_user, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UserList, error) {
	jsonBytes, err := ioutil.ReadFile("users.json")
	if err != nil {
		log.Fatalf("Failed read from file: %v", err)
	}
	users_list := &pb.UserList{}
	if err := protojson.Unmarshal(jsonBytes, users_list); err != nil {
		log.Fatalf("Unmarshaling failed: %v", err)
	}
	return users_list, nil
}

func main() {
	user_mgmt_server := NewUserManagementServer()
	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("failed to server: %v", err)
	}
}
