# .NET Support Implementation TODO

This document tracks the implementation of .NET/C# support for the pulumitest library.

## High Priority (Required for basic support)

### ✅ 1. Installation Support
- Status: Already works via `pulumi install`
- The current implementation delegates to `dotnet restore` automatically

### ✅ 2. Test Data with .NET Examples
- [x] Create `testdata/csharp_simple/` directory
  - [x] Add `Pulumi.yaml` with `runtime: dotnet`
  - [x] Add `.csproj` file with Pulumi package references
  - [x] Add `Program.cs` with simple resource definitions
- [x] Create `testdata/csharp_aws/` directory for AWS provider tests
- [ ] Create `testdata/fsharp_simple/` directory for F# support validation (deferred)

### ✅ 3. .NET Package Reference Option
- [x] Add `DotNetReferences` field to `Options` struct in `opttest/opttest.go`
- [x] Implement `DotNetReference()` option function
- [x] Implement .csproj manipulation logic in `newStack.go`
  - [x] Parse existing .csproj XML (text-based manipulation)
  - [x] Add `<ProjectReference>` elements
  - [x] Handle absolute path resolution
  - [x] Write modified .csproj back to disk
- [x] Update `Defaults()` function to initialize .NET fields
- [x] Created `csproj.go` helper file with functions:
  - `findCsprojFile()` - finds .csproj in a directory
  - `addProjectReferences()` - adds ProjectReference elements to .csproj

### ✅ 4. Basic .NET Integration Tests
- [x] Add `TestDotNetDeploy` - Basic deployment test with preview, up, and second preview
- [x] Add `TestDotNetSkipInstall` - Test with manual install and stack creation
- [x] Tests verified passing with .NET 8.0
- [x] Add `TestDotNetWithLocalReference` - Test local package references
- [x] Add `TestDotNetAwsDeploy` - Real AWS S3 bucket deployment and verification

## Medium Priority (Nice to have)

### ✅ 5. Build Configuration Options
- [x] Add `DotNetBuildConfig` field to Options struct
- [x] Implement `DotNetBuildConfiguration()` option (Debug/Release)
- [x] Pass configuration via `DOTNET_BUILD_CONFIGURATION` environment variable

### ✅ 6. Target Framework Specification
- [x] Add `DotNetTargetFramework` field to Options struct
- [x] Implement `DotNetTargetFramework()` option (net6.0, net7.0, net8.0, etc.)
- [x] Handle framework selection in .csproj manipulation via `setTargetFramework()`
- [x] Added to Options struct and integrated in newStack.go

### 7. NuGet.config Handling
- [ ] Support for custom NuGet feeds
- [ ] Option to specify NuGet.config path
- [ ] Auto-detection of NuGet.config in project directory

## Low Priority (Optional enhancements)

### 8. MSBuild Verbosity Controls
- [ ] Option to control build output verbosity
- [ ] Environment variable configuration

### 9. Custom dotnet CLI Arguments
- [ ] Generic option to pass additional arguments to dotnet commands
- [ ] Support for dotnet build arguments
- [ ] Support for dotnet restore arguments

### 10. Multi-targeting Support
- [ ] Handle projects with multiple target frameworks
- [ ] Option to select specific framework for testing

## Documentation

### ✅ 11. Update Documentation
- [x] Update `pulumitest/CLAUDE.md` with .NET support details
  - [x] Added all .NET options to key options list (`DotNetReference`, `DotNetBuildConfiguration`, `DotNetTargetFramework`)
  - [x] Added comprehensive .NET/C# SDK configuration section with detailed usage
  - [x] Added `csproj.go` to file organization
  - [x] Added troubleshooting section for common .NET issues
  - [x] Added code examples section with 3 usage patterns
- [x] Checked root README.md (no updates needed - remains language-agnostic)

### ✅ 12. Code Examples
- [x] Added examples in `pulumitest/CLAUDE.md` showing:
  - Basic .NET test
  - Test with local SDK reference using `DotNetReference`
  - Test with specific framework and configuration
- [x] Created comprehensive test data:
  - `testdata/csharp_simple/` - Simple Random provider example
  - `testdata/csharp_aws/` - AWS S3 bucket example
  - `testdata/csharp_with_ref/` - Program using local project reference
  - `testdata/mock_sdk/` - Mock SDK for reference testing

## Testing Strategy

### Unit Tests
- [ ] Test .csproj XML parsing and modification
- [ ] Test option parsing and application
- [ ] Test path resolution for .NET references

### Integration Tests
- [ ] Test with .NET 6, 7, and 8
- [ ] Test with C# and F# projects
- [ ] Test with various Pulumi providers (AWS, Azure, GCP)
- [ ] Test local SDK development workflow

## Code Locations

### Files to Modify
1. `/pulumitest/opttest/opttest.go` - Add .NET options (~lines 100-246)
2. `/pulumitest/newStack.go` - Implement .csproj manipulation (after line 172)
3. `/pulumitest/testdata/` - Add .NET test programs
4. `/pulumitest/pulumiTest_test.go` - Add .NET integration tests
5. `/pulumitest/CLAUDE.md` - Document .NET support
6. `/CLAUDE.md` - Update root documentation if needed

### New Files to Create
1. `/pulumitest/testdata/csharp_simple/Pulumi.yaml`
2. `/pulumitest/testdata/csharp_simple/Program.cs`
3. `/pulumitest/testdata/csharp_simple/{ProjectName}.csproj`
4. Additional test data directories as needed

## Summary of Completed Work

### ✅ High Priority - COMPLETE
All basic .NET support features have been implemented and tested:
- Installation works via `pulumi install` (delegates to `dotnet restore`)
- Test data with C# examples (csharp_simple, csharp_aws)
- `DotNetReference()` option for local package/project references
- `.csproj` manipulation helpers: `addProjectReferences()`, `setTargetFramework()`
- Comprehensive integration tests:
  - `TestDotNetDeploy` - Full deployment workflow
  - `TestDotNetSkipInstall` - Manual install flow
  - `TestDotNetWithLocalReference` - Local reference verification
  - `TestDotNetAwsDeploy` - Real AWS S3 deployment

### ✅ Medium Priority - COMPLETE
Enhanced .NET support features:
- `DotNetBuildConfiguration()` for Debug/Release builds (via environment variable)
- `DotNetTargetFramework()` for framework selection (.csproj modification)
- Full documentation with examples and troubleshooting

### 🔲 Low Priority (Future Enhancements)
Optional features not yet implemented:
- Item 7: NuGet.config handling (custom feeds, auto-detection)
- Item 8: MSBuild verbosity controls
- Item 9: Custom dotnet CLI arguments
- Item 10: Multi-targeting support

### 🔲 Additional Testing (Future)
- Test with .NET 6 and 7 (currently only testing with 8)
- F# project support validation
- Azure and GCP provider examples
- Multi-project .NET solution testing

## Implementation Summary

**Files Modified:**
- `/pulumitest/opttest/opttest.go` - Added 3 .NET options and 2 fields to Options struct
- `/pulumitest/newStack.go` - Integrated .csproj manipulation during stack creation
- `/pulumitest/pulumiTest_test.go` - Added 3 .NET integration tests
- `/pulumitest/CLAUDE.md` - Comprehensive documentation, examples, and troubleshooting

**Files Created:**
- `/pulumitest/csproj.go` - Helper functions for .csproj manipulation
- `/pulumitest/dotnet_aws_test.go` - Real AWS deployment integration test
- `/pulumitest/testdata/csharp_simple/*` - Simple Random provider test program
- `/pulumitest/testdata/csharp_aws/*` - AWS S3 bucket test program
- `/pulumitest/testdata/csharp_with_ref/*` - Local reference test program
- `/pulumitest/testdata/mock_sdk/*` - Mock SDK for testing references

**Package Versions Used:**
- Pulumi: 3.90.0 (latest stable .NET SDK)
- Pulumi.Random: 4.18.4
- Pulumi.Aws: 7.8.0
- Target Framework: net8.0

**Status:** ✅ **Production-ready for basic and intermediate .NET testing workflows**

## Original Estimated Effort vs. Actual

- **Original Estimate**: 1-2 weeks for production-ready support
- **High Priority**: Completed
- **Medium Priority**: Completed
- **Low Priority**: Deferred to future enhancements
- **Documentation**: Completed with examples and troubleshooting

## Notes

- The architecture is well-designed and extensible - following existing patterns from `YarnLink` and `GoModReplacement` worked excellently
- Most Automation API functionality already supported .NET - main work was language-specific package management
- The `.csproj` manipulation uses text-based approach for reliability and simplicity
- All tests pass including end-to-end AWS deployment with automatic cleanup
