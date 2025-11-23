"""Test package for PythonLink integration tests."""

__version__ = "0.0.2"


def get_version():
    """Return the package version."""
    return __version__


def get_message():
    """Return a version-specific message."""
    return f"pulumi-test-pkg version {__version__}"
