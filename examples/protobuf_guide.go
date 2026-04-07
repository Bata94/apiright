package main

import (
	"fmt"
	"log"

	"github.com/bata94/apiright/pkg/core"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	fmt.Println("=== APIRight Protobuf Implementation Guide ===")

	// Create content negotiator
	cn := core.NewContentNegotiator()

	// Example 1: Using structpb for dynamic data (useful for generic APIs)
	fmt.Println("\n1. Dynamic protobuf with structpb:")
	dynamicData := map[string]any{
		"user_id":  123,
		"username": "johndoe",
		"active":   true,
		"tags":     []string{"golang", "protobuf", "api"},
		"metadata": map[string]string{"role": "admin", "department": "engineering"},
	}

	// Convert to protobuf Struct message
	protoStruct, err := structpb.NewStruct(dynamicData)
	if err != nil {
		log.Fatalf("Failed to create structpb: %v", err)
	}

	// Serialize to protobuf
	protobufData, err := cn.SerializeResponse(protoStruct, "application/protobuf")
	if err != nil {
		log.Fatalf("Protobuf serialization failed: %v", err)
	}

	fmt.Printf("   Original: %+v\n", dynamicData)
	fmt.Printf("   Protobuf size: %d bytes\n", len(protobufData))

	// Deserialize back
	var restoredStruct structpb.Struct
	err = cn.DeserializeRequest(protobufData, "application/protobuf", &restoredStruct)
	if err != nil {
		log.Fatalf("Protobuf deserialization failed: %v", err)
	}

	fmt.Printf("   Restored user_id: %v\n", restoredStruct.Fields["user_id"].GetNumberValue())
	fmt.Printf("   Restored username: %s\n", restoredStruct.Fields["username"].GetStringValue())
	fmt.Printf("   Restored active: %v\n", restoredStruct.Fields["active"].GetBoolValue())

	// Example 2: Comparing protobuf vs JSON serialization
	fmt.Println("\n2. Comparing Protobuf vs JSON:")

	// JSON serialization
	jsonData, err := cn.SerializeResponse(dynamicData, "application/json")
	if err != nil {
		log.Fatalf("JSON serialization failed: %v", err)
	}

	// Protobuf serialization
	protobufData, err = cn.SerializeResponse(protoStruct, "application/protobuf")
	if err != nil {
		log.Fatalf("Protobuf serialization failed: %v", err)
	}

	fmt.Printf("   JSON size:     %d bytes\n", len(jsonData))
	fmt.Printf("   Protobuf size: %d bytes\n", len(protobufData))
	fmt.Printf("   Size reduction: %.1f%%\n", float64(len(jsonData)-len(protobufData))/float64(len(jsonData))*100)

	// Example 3: Content negotiation with protobuf
	fmt.Println("\n3. Content Negotiation:")

	testCases := []string{
		"application/protobuf",
		"application/json",
		"application/protobuf, application/json;q=0.9",
		"application/json, application/protobuf;q=0.8",
		"*/*",
	}

	for _, acceptHeader := range testCases {
		negotiated := cn.DetectContentType(acceptHeader)
		fmt.Printf("   Accept: %-45s -> %s\n", acceptHeader, negotiated)
	}

	// Example 4: Error handling
	fmt.Println("\n4. Error Handling:")

	// Try to serialize non-protobuf data as protobuf
	_, err = cn.SerializeResponse(map[string]string{"test": "data"}, "application/protobuf")
	if err != nil {
		fmt.Printf("   Non-protobuf error: %v\n", err)
	}

	// Try to deserialize to wrong type
	err = cn.DeserializeRequest([]byte("invalid"), "application/protobuf", "string")
	if err != nil {
		fmt.Printf("   Wrong target error: %v\n", err)
	}

	fmt.Println("\n=== Integration with Generated Protobuf Messages ===")
	fmt.Println("To use with your generated protobuf messages:")
	fmt.Println("1. Run: protoc --go_out=. gen/proto/*.proto")
	fmt.Println("2. Import: \"your-project/gen/go/db\"")
	fmt.Println("3. Use generated types directly:")
	fmt.Println("   user := &db.User{Id: 1, Name: \"John\"}")
	fmt.Println("   data, err := cn.SerializeResponse(user, \"application/protobuf\")")
	fmt.Println("4. Generated types automatically implement proto.Message")
}
