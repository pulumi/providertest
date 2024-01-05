package grpclog

import (
	"encoding/json"
	"os"
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
	GetMapping    Method = "/pulumirpc.ResourceProvider/GetMapping"
	GetMappings   Method = "/pulumirpc.ResourceProvider/GetMappings"
	GetPluginInfo Method = "/pulumirpc.ResourceProvider/GetPluginInfo"
	GetSchema     Method = "/pulumirpc.ResourceProvider/GetSchema"
	Invoke        Method = "/pulumirpc.ResourceProvider/Invoke"
	Read          Method = "/pulumirpc.ResourceProvider/Read"
	Update        Method = "/pulumirpc.ResourceProvider/Update"
)

type AttachLog struct {
	Request  rpc.PluginAttach
	Response emptypb.Empty
}

func (l *GrpcLog) Attaches() ([]AttachLog, error) {
	var typedEntries []AttachLog
	for _, entry := range l.WhereMethod(Attach) {
		var typedEntry AttachLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type CallLog struct {
	Request  rpc.CallRequest
	Response rpc.CallResponse
}

func (l *GrpcLog) Calls() ([]CallLog, error) {
	var typedEntries []CallLog
	for _, entry := range l.WhereMethod(Call) {
		var typedEntry CallLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type CancelLog struct {
	Request  emptypb.Empty
	Response emptypb.Empty
}

func (l *GrpcLog) Cancels() ([]CancelLog, error) {
	var typedEntries []CancelLog
	for _, entry := range l.WhereMethod(Cancel) {
		var typedEntry CancelLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type CheckLog struct {
	Request  rpc.CheckRequest
	Response rpc.CheckResponse
}

func (l *GrpcLog) Checks() ([]CheckLog, error) {
	var typedEntries []CheckLog
	for _, entry := range l.WhereMethod(Check) {
		var typedEntry CheckLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type CheckConfigLog struct {
	Request  rpc.CheckRequest
	Response rpc.CheckResponse
}

func (l *GrpcLog) CheckConfigs() ([]CheckConfigLog, error) {
	var typedEntries []CheckConfigLog
	for _, entry := range l.WhereMethod(CheckConfig) {
		var typedEntry CheckConfigLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type ConfigureLog struct {
	Request  rpc.ConfigureRequest
	Response rpc.ConfigureResponse
}

func (l *GrpcLog) Configures() ([]ConfigureLog, error) {
	var typedEntries []ConfigureLog
	for _, entry := range l.WhereMethod(Configure) {
		var typedEntry ConfigureLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type ConstructLog struct {
	Request  rpc.ConstructRequest
	Response rpc.ConstructResponse
}

func (l *GrpcLog) Constructs() ([]ConstructLog, error) {
	var typedEntries []ConstructLog
	for _, entry := range l.WhereMethod(Construct) {
		var typedEntry ConstructLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type CreateLog struct {
	Request  rpc.CreateRequest
	Response rpc.CreateResponse
}

func (l *GrpcLog) Creates() ([]CreateLog, error) {
	var typedEntries []CreateLog
	for _, entry := range l.WhereMethod(Create) {
		var typedEntry CreateLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type DeleteLog struct {
	Request  rpc.DeleteRequest
	Response emptypb.Empty
}

func (l *GrpcLog) Deletes() ([]DeleteLog, error) {
	var typedEntries []DeleteLog
	for _, entry := range l.WhereMethod(Delete) {
		var typedEntry DeleteLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type DiffLog struct {
	Request  rpc.DiffRequest
	Response rpc.DiffResponse
}

func (l *GrpcLog) Diffs() ([]DiffLog, error) {
	var typedEntries []DiffLog
	for _, entry := range l.WhereMethod(Diff) {
		var typedEntry DiffLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
}

type GetMappingLog struct {
	Request  rpc.GetMappingRequest
	Response rpc.GetMappingResponse
}

func (l *GrpcLog) GetMappings() ([]GetMappingLog, error) {
	var typedEntries []GetMappingLog
	for _, entry := range l.WhereMethod(GetMapping) {
		var typedEntry GetMappingLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
	}
	return typedEntries, nil
}

type GetMappingsLog struct {
	Request  rpc.GetMappingsRequest
	Response rpc.GetMappingsResponse
}

func (l *GrpcLog) GetMappingsEntries() ([]GetMappingLog, error) {
	var typedEntries []GetMappingLog
	for _, entry := range l.WhereMethod(GetMappings) {
		var typedEntry GetMappingLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
	}
	return typedEntries, nil
}

type GetPluginInfoLog struct {
	Request  emptypb.Empty
	Response rpc.PluginInfo
}

func (l *GrpcLog) GetPluginInfos() ([]GetPluginInfoLog, error) {
	var typedEntries []GetPluginInfoLog
	for _, entry := range l.WhereMethod(GetPluginInfo) {
		var typedEntry GetPluginInfoLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
	}
	return typedEntries, nil
}

type GetSchemaLog struct {
	Request  emptypb.Empty
	Response rpc.GetSchemaResponse
}

func (l *GrpcLog) GetSchemas() ([]GetSchemaLog, error) {
	var typedEntries []GetSchemaLog
	for _, entry := range l.WhereMethod(GetSchema) {
		var typedEntry GetSchemaLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
	}
	return typedEntries, nil
}

type InvokeLog struct {
	Request  rpc.InvokeRequest
	Response rpc.InvokeResponse
}

func (l *GrpcLog) Invokes() ([]InvokeLog, error) {
	var typedEntries []InvokeLog
	for _, entry := range l.WhereMethod(Invoke) {
		var typedEntry InvokeLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
	}
	return typedEntries, nil
}

type ReadLog struct {
	Request  rpc.ReadRequest
	Response rpc.ReadResponse
}

func (l *GrpcLog) Reads() ([]ReadLog, error) {
	var typedEntries []ReadLog
	for _, entry := range l.WhereMethod(Read) {
		var typedEntry ReadLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}

	}
	return typedEntries, nil
}

type UpdateLog struct {
	Request  rpc.UpdateRequest
	Response rpc.UpdateResponse
}

func (l *GrpcLog) Updates() ([]UpdateLog, error) {
	var typedEntries []UpdateLog
	for _, entry := range l.WhereMethod(Update) {
		var typedEntry UpdateLog
		err := unmarshalEntry(entry, &typedEntry.Request, &typedEntry.Response)
		if err != nil {
			return nil, err
		}
		typedEntries = append(typedEntries, typedEntry)
	}
	return typedEntries, nil
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
