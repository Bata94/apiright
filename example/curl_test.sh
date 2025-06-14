#! /usr/bin/env bash

echo "Starting curl tests..."

# Define a function to perform the curl operation
perform_curl_test() {
    local path="$1"
    local url="http://localhost:5500${path}"

    echo # Add a newline after the initial message
    echo "----------------------------------------" # Separator
    echo "Exec: curl ${url}"
    echo # Newline before curl output

    # Execute curl and capture its exit status
    curl "${url}"
    local curl_status=$? # Store the exit status of the last command (curl)

    echo # Newline after curl output

    if [ "${curl_status}" -ne 0 ]; then
        echo "!!! ERROR: Curl to ${url} failed with exit status ${curl_status} !!!" >&2
    fi
}

# Define an array of endpoints
endpoints=(
    "/"
    "/err"
    "/panic"
    "/test"
)

# Loop through each endpoint and call the function
for endpoint in "${endpoints[@]}"; do
    perform_curl_test "${endpoint}"
done

echo
echo "----------------------------------------" # Separator
echo "All curl tests finished."
