package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/slipros/roamer"
	"github.com/slipros/roamer/examples/custom_parser/middleware"
	"github.com/slipros/roamer/examples/custom_parser/model"
	"github.com/slipros/roamer/parser"
)

const TagMember = "member"

type MemberParser struct{}

func NewMemberParser() *MemberParser {
	return &MemberParser{}
}

// Parse extracts member data from the request context based on the struct tag.
// It implements the roamer.Parser interface.
func (p *MemberParser) Parse(r *http.Request, tag reflect.StructTag, _ parser.Cache) (any, bool) {
	// Get the tag value for "member"
	tagValue, ok := tag.Lookup(TagMember)
	if !ok {
		return nil, false
	}

	m, ok := model.MemberFromContext(r.Context())
	if !ok {
		return nil, false
	}

	switch tagValue {
	case "organization_id":
		return m.OrganizationID, true
	case "id":
		return m.ID, true
	case "member":
		return m, true
	default:
		return nil, false
	}
}

// Tag returns the name of the struct tag that this parser handles.
func (p *MemberParser) Tag() string {
	return TagMember
}

// AppRequest demonstrates using custom struct tags for extracting member data from context.
type AppRequest struct {
	MemberID       string        `member:"id"`
	MemberIDAsUUID uuid.UUID     `member:"id"`
	Member         *model.Member `member:"member"`
	MemberAsStruct model.Member  `member:"member"`

	// Standard query parameter
	Action string `query:"action"`
}

func main() {
	// Create a roamer instance with our custom parser and query parser
	r := roamer.NewRoamer(
		roamer.WithParsers(
			NewMemberParser(),
			parser.NewQuery(), // For the action parameter
		),
	)

	action := middleware.Member(func(w http.ResponseWriter, req *http.Request) {
		var appReq AppRequest

		// Parse the request using our custom parser
		if err := r.Parse(req, &appReq); err != nil {
			http.Error(w, fmt.Sprintf("Parse error: %v", err), http.StatusBadRequest)
			return
		}

		// Log the parsed values
		log.Printf("Received request:")
		log.Printf("  MemberID: %s", appReq.MemberID)
		log.Printf("  MemberIDAsUUID: %v", appReq.MemberIDAsUUID)
		log.Printf("  Member: %v", appReq.Member)
		log.Printf("  MemberAsStruct: %v", appReq.MemberAsStruct)

		// Respond with the parsed data
		fmt.Fprintf(w, "Successfully parsed member data from context:\n")
		fmt.Fprintf(w, "MemberID: %s\n", appReq.MemberID)
		fmt.Fprintf(w, "MemberID (UUID): %v\n", appReq.MemberIDAsUUID)
		fmt.Fprintf(w, "Member Pointer: %+v\n", appReq.Member)
		fmt.Fprintf(w, "Member Struct: %+v\n", appReq.MemberAsStruct)
		fmt.Fprintf(w, "Action: %s\n", appReq.Action)
	})

	// HTTP handler that uses the custom parser
	http.HandleFunc("/api/action", action)

	// Example endpoint to demonstrate the parser
	http.HandleFunc("/example", func(w http.ResponseWriter, req *http.Request) {
		examples := []string{
			"Test the custom member parser with curl:",
			"",
			"curl -X GET 'http://localhost:8080/api/action?action=view'",
			"",
			"Expected output:",
			"  MemberID: <generated UUID>",
			"  MemberID (UUID): <generated UUID>",
			"  Member Pointer: &{ID:<uuid> OrganizationID:<uuid> Name:Jack Role:Admin}",
			"  Member Struct: {ID:<uuid> OrganizationID:<uuid> Name:Jack Role:Admin}",
			"  Action: view",
			"",
			"Note: The Member middleware automatically injects member data into the request context.",
			"The custom MemberParser extracts this data and populates the struct fields.",
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, strings.Join(examples, "\n"))
	})

	// Start the server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("Visit http://localhost:8080/example for usage instructions")
	log.Printf("\nTest with: curl -X GET 'http://localhost:8080/api/action?action=view'")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
