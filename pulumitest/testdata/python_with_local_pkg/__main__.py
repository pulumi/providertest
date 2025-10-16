"""Pulumi program that uses the local test package."""

import pulumi_test_pkg

# Import the test package and verify version
version = pulumi_test_pkg.get_version()
message = pulumi_test_pkg.get_message()

# Export the version as a stack output
pulumi.export("package_version", version)
pulumi.export("package_message", message)
