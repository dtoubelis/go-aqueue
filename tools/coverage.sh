#!/bin/sh
#
# Code coverage generation
#
COVERAGE_DIR="${COVERAGE_DIR:-coverage}"
PKG_LIST=$(go list ./... | grep -v /vendor/)

# Create the coverage files directory
mkdir -p "$COVERAGE_DIR"

# Create a coverage file for each package
for package in ${PKG_LIST}; do
    mkdir -p "${COVERAGE_DIR}/$(dirname ${package})"
    go test -covermode=count -coverprofile "${COVERAGE_DIR}/${package}.cov" "$package"
done

# Merge the coverage profile files
echo 'mode: count' > coverage.cov
find "${COVERAGE_DIR}" -type f -name "*.cov" -exec tail -q -n +2 {} \; >> coverage.cov

# Display the global code coverage
go tool cover -func=coverage.cov

# If needed, generate HTML report
if [ "$1" = "html" ]; then
    go tool cover -html=coverage.cov -o coverage.html
fi

# Remove the coverage files directory
rm -rf "$COVERAGE_DIR"
