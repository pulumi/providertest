package grpclog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulumi/providertest/pulumitest/sanitize"
	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	jsonpb "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcLog struct {
	Entries []GrpcLogEntry `json:"entries"`
}

type GrpcLogEntry struct {
	// The gRPC method being called.
	// This is the full method name, e.g. "/pulumirpc.ResourceProvider/Check".
	Method   string          `json:"method"`
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
	Progress string          `json:"progress,omitempty"`
}

type Method string

const (
	Attach        Method = "/pulumirpc.ResourceProvider/Attach"
	Call          Method = "/pulumirpc.ResourceProvider/Call"
	Cancel        Method = "/pulumirpc.ResourceProvider/Cancel"
	Check         Method = "/pulumirpc.ResourceProvider/Check"
	CheckConfig   Method = "/pulumirpc.ResourceProvider/CheckConfig"
	Configure     Method = "/pulumirpc.ResourceProvider/Configure"
	Construct     Method = "/pulumirpc.ResourceProvider/Construct"
	Create        Method = "/pulumirpc.ResourceProvider/Create"
	Delete        Method = "/pulumirpc.ResourceProvider/Delete"
	Diff          Method = "/pulumirpc.ResourceProvider/Diff"
	DiffConfig    Method = "/pulumirpc.ResourceProvider/DiffConfig"
	GetMapping    Method = "/pulumirpc.ResourceProvider/GetMapping"
	GetMappings   Method = "/pulumirpc.ResourceProvider/GetMappings"
	GetPluginInfo Method = "/pulumirpc.ResourceProvider/GetPluginInfo"
	GetSchema     Method = "/pulumirpc.ResourceProvider/GetSchema"
	Invoke        Method = "/pulumirpc.ResourceProvider/Invoke"
	Read          Method = "/pulumirpc.ResourceProvider/Read"
	Update        Method = "/pulumirpc.ResourceProvider/Update"
)

type resourceRequest interface {
	rpc.CheckRequest | rpc.DiffRequest | rpc.ReadRequest | rpc.CreateRequest | rpc.UpdateRequest | rpc.DeleteRequest
}

type resourceResponse interface {
	rpc.CheckResponse | rpc.DiffResponse | rpc.ReadResponse | rpc.CreateResponse | rpc.UpdateResponse | emptypb.Empty
}

func (l *GrpcLog) Attaches() ([]TypedEntry[rpc.PluginAttach, emptypb.Empty], error) {
	return unmarshalTypedEntries[rpc.PluginAttach, emptypb.Empty](l.WhereMethod(Attach))
}

func (l *GrpcLog) Calls() ([]TypedEntry[rpc.CallRequest, rpc.CallResponse], error) {
	return unmarshalTypedEntries[rpc.CallRequest, rpc.CallResponse](l.WhereMethod(Call))
}

func (l *GrpcLog) Cancels() ([]TypedEntry[emptypb.Empty, emptypb.Empty], error) {
	return unmarshalTypedEntries[emptypb.Empty, emptypb.Empty](l.WhereMethod(Cancel))
}

func (l *GrpcLog) Checks() ([]TypedEntry[rpc.CheckRequest, rpc.CheckResponse], error) {
	return unmarshalTypedEntries[rpc.CheckRequest, rpc.CheckResponse](l.WhereMethod(Check))
}

func (l *GrpcLog) CheckConfigs() ([]TypedEntry[rpc.CheckRequest, rpc.CheckResponse], error) {
	return unmarshalTypedEntries[rpc.CheckRequest, rpc.CheckResponse](l.WhereMethod(CheckConfig))
}

func (l *GrpcLog) Configures() ([]TypedEntry[rpc.ConfigureRequest, rpc.ConfigureResponse], error) {
	return unmarshalTypedEntries[rpc.ConfigureRequest, rpc.ConfigureResponse](l.WhereMethod(Configure))
}

func (l *GrpcLog) Constructs() ([]TypedEntry[rpc.ConstructRequest, rpc.ConstructResponse], error) {
	return unmarshalTypedEntries[rpc.ConstructRequest, rpc.ConstructResponse](l.WhereMethod(Construct))
}

func (l *GrpcLog) Creates() ([]TypedEntry[rpc.CreateRequest, rpc.CreateResponse], error) {
	return unmarshalTypedEntries[rpc.CreateRequest, rpc.CreateResponse](l.WhereMethod(Create))
}

func (l *GrpcLog) Deletes() ([]TypedEntry[rpc.DeleteRequest, emptypb.Empty], error) {
	return unmarshalTypedEntries[rpc.DeleteRequest, emptypb.Empty](l.WhereMethod(Delete))
}

func (l *GrpcLog) Diffs() ([]TypedEntry[rpc.DiffRequest, rpc.DiffResponse], error) {
	return unmarshalTypedEntries[rpc.DiffRequest, rpc.DiffResponse](l.WhereMethod(Diff))
}

func (l *GrpcLog) DiffConfigs() ([]TypedEntry[rpc.DiffRequest, rpc.DiffResponse], error) {
	return unmarshalTypedEntries[rpc.DiffRequest, rpc.DiffResponse](l.WhereMethod(DiffConfig))
}

func (l *GrpcLog) GetMappings() ([]TypedEntry[rpc.GetMappingRequest, rpc.GetMappingResponse], error) {
	return unmarshalTypedEntries[rpc.GetMappingRequest, rpc.GetMappingResponse](l.WhereMethod(GetMapping))
}

func (l *GrpcLog) GetMappingsEntries() ([]TypedEntry[rpc.GetMappingsRequest, rpc.GetMappingsResponse], error) {
	return unmarshalTypedEntries[rpc.GetMappingsRequest, rpc.GetMappingsResponse](l.WhereMethod(GetMappings))
}

func (l *GrpcLog) GetPluginInfos() ([]TypedEntry[emptypb.Empty, rpc.PluginInfo], error) {
	return unmarshalTypedEntries[emptypb.Empty, rpc.PluginInfo](l.WhereMethod(GetPluginInfo))
}

func (l *GrpcLog) GetSchemas() ([]TypedEntry[emptypb.Empty, rpc.GetSchemaResponse], error) {
	return unmarshalTypedEntries[emptypb.Empty, rpc.GetSchemaResponse](l.WhereMethod(GetSchema))
}

func (l *GrpcLog) Invokes() ([]TypedEntry[rpc.InvokeRequest, rpc.InvokeResponse], error) {
	return unmarshalTypedEntries[rpc.InvokeRequest, rpc.InvokeResponse](l.WhereMethod(Invoke))
}

func (l *GrpcLog) Reads() ([]TypedEntry[rpc.ReadRequest, rpc.ReadResponse], error) {
	return unmarshalTypedEntries[rpc.ReadRequest, rpc.ReadResponse](l.WhereMethod(Read))
}

func (l *GrpcLog) Updates() ([]TypedEntry[rpc.UpdateRequest, rpc.UpdateResponse], error) {
	return unmarshalTypedEntries[rpc.UpdateRequest, rpc.UpdateResponse](l.WhereMethod(Update))
}

type TypedEntry[TRequest any, TResponse any] struct {
	Request  TRequest
	Response TResponse
}

func unmarshalTypedEntries[TRequest, TResponse any](entries []GrpcLogEntry) ([]TypedEntry[TRequest, TResponse], error) {
	var typedEntries []TypedEntry[TRequest, TResponse]
	for _, entry := range entries {
		typedEntry, err := unmarshalTypedEntry[TRequest, TResponse](entry)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, *typedEntry)
	}
	return typedEntries, nil
}

func unmarshalTypedEntry[TRequest, TResponse any](entry GrpcLogEntry) (*TypedEntry[TRequest, TResponse], error) {
	reqSlot := new(TRequest)
	resSlot := new(TResponse)
	jsonOpts := jsonpb.UnmarshalOptions{DiscardUnknown: true, AllowPartial: true}
	if err := jsonOpts.Unmarshal([]byte(entry.Request), any(reqSlot).(protoreflect.ProtoMessage)); err != nil {
		return nil, err
	}
	if err := jsonOpts.Unmarshal([]byte(entry.Response), any(resSlot).(protoreflect.ProtoMessage)); err != nil {
		return nil, err
	}
	typedEntry := TypedEntry[TRequest, TResponse]{
		Request:  *reqSlot,
		Response: *resSlot,
	}
	return &typedEntry, nil
}

func LoadLog(path string) (*GrpcLog, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseLog(file)
}

func ParseLog(log []byte) (*GrpcLog, error) {
	var parsed GrpcLog
	lines := strings.Split(string(log), "\n")
	for _, line := range lines {
		// Skip empty lines.
		if line == "" {
			continue
		}
		var entry GrpcLogEntry
		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			return nil, err
		}
		// Ignore incomplete entries.
		if entry.Progress == "request_started" {
			continue
		}
		parsed.Entries = append(parsed.Entries, entry)
	}
	return &parsed, nil
}

func (l *GrpcLog) WhereMethod(method Method) []GrpcLogEntry {
	var matching []GrpcLogEntry
	for _, entry := range l.Entries {
		if entry.Method == string(method) {
			matching = append(matching, entry)
		}
	}
	return matching
}

func (l *GrpcLog) SanitizeSecrets() {
	for i := range l.Entries {
		l.Entries[i].SanitizeSecrets()
	}
}

func (e *GrpcLogEntry) SanitizeSecrets() {
	e.Request = sanitize.SanitizeSecretsInGrpcLog(e.Request)
	e.Response = sanitize.SanitizeSecretsInGrpcLog(e.Response)
}

// WriteTo writes the log to the given path.
// Creates any directories needed.
func (l *GrpcLog) WriteTo(path string) error {
	bytes, err := l.Marshal()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path, bytes, 0644)
}

func (l *GrpcLog) Marshal() ([]byte, error) {
	var bytes []byte
	for i, entry := range l.Entries {
		entryBytes, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, entryBytes...)
		if i < len(l.Entries)-1 {
			bytes = append(bytes, '\n')
		}
	}
	return bytes, nil
}

type hasURN interface {
	GetUrn() string
}

// FindByURN finds the first entry with the given resource URN, or nil if none is found.
// Returns the index of the entry, if found.
func FindByURN[TRequest resourceRequest, TResponse resourceResponse](entries []TypedEntry[TRequest, TResponse],
	urn string) (*TypedEntry[TRequest, TResponse], int) {
	for i := range entries {
		var rI any = &entries[i].Request
		switch r := rI.(type) {
		case hasURN:
			if r.GetUrn() == urn {
				return &entries[i], i
			}
		}
	}
	return nil, -1
}
