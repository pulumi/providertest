package grpclog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	rpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
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

// BUG: will currently fail to unmarshal the response due to a enums not having JSON unmashalling implemented.
func (l *GrpcLog) Diffs() ([]TypedEntry[rpc.DiffRequest, rpc.DiffResponse], error) {
	return unmarshalTypedEntries[rpc.DiffRequest, rpc.DiffResponse](l.WhereMethod(Diff))
}

// BUG: will currently fail to unmarshal the response due to a enums not having JSON unmashalling implemented.
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

func unmarshalTypedEntries[Request any, Response any](entries []GrpcLogEntry) ([]TypedEntry[Request, Response], error) {
	var typedEntries []TypedEntry[Request, Response]
	for _, entry := range entries {
		typedEntry, err := unmarshalTypedEntry[Request, Response](entry)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, *typedEntry)
	}
	return typedEntries, nil
}

func unmarshalTypedEntry[Request any, Response any](entry GrpcLogEntry) (*TypedEntry[Request, Response], error) {
	var typedEntry TypedEntry[Request, Response]
	err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
	if err != nil {
		return nil, err
	}
	return &typedEntry, nil
}

func unmarshalEntry(entry GrpcLogEntry, req any, res any) error {
	err := json.Unmarshal(entry.Request, req)
	if err != nil {
		return err
	}
	return json.Unmarshal(entry.Response, res)
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

// FindByUrn finds the first entry with the given resource URN, or nil if none is found.
func FindByUrn[TRequest resourceRequest, TResponse resourceResponse](entries []TypedEntry[TRequest, TResponse],
	urn string) *TypedEntry[TRequest, TResponse] {
	// nolint:copylocks
	for _, e := range entries {
		var eI any = &e.Request
		switch r := eI.(type) {
		case hasURN:
			if r.GetUrn() == urn {
				return &e
			}
		}
	}
	return nil
}
