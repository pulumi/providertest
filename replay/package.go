// Copyright 2016-2023, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Replay tests allow to quickly add regression tests that exercise one or a small number of gRPC
// methods.
//
// How this works: once a problem reproduces, run Pulumi with PULUMI_DEBUG_GRPC="$PWD/logs.json"
// flag, find the offending method record under "logs.json" and turn it into a test using Replay
// from this package.
//
// Note: this package used to be exposed under pulumi/pulumi-terraform-bridge/testing Go module but
// is moving here as it is not specific to bridged providers.
package replay
