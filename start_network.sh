#!/bin/bash

# ==============================================================================
# Script: start_network.sh
# Description: Orchestrates the Minichord distributed system by launching 
#              the registry and messenger nodes in separate terminal windows.
# ==============================================================================

set -e

# ------------------------------------------------------------------------------
# Function: assert_equal
# Description: Provides unit testing assertion capabilities for validating logic.
# ------------------------------------------------------------------------------
assert_equal() {
    local expected="$1"
    local actual="$2"
    local test_name="$3"
    
    if [ "$expected" == "$actual" ]; then
        echo "[PASS] $test_name"
    else
        echo "[FAIL] $test_name: expected '$expected', got '$actual'"
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# Function: validate_integer
# Description: Validates if the provided string is a strictly positive integer.
# Returns: "valid" or "invalid"
# ------------------------------------------------------------------------------
validate_integer() {
    local input="$1"
    if [[ "$input" =~ ^[0-9]+$ ]] && [ "$input" -gt 0 ]; then
        echo "valid"
    else
        echo "invalid"
    fi
}

# ------------------------------------------------------------------------------
# Function: test_suite
# Description: Executes unit tests for the core logical components.
# ------------------------------------------------------------------------------
test_suite() {
    echo "Running unit tests..."
    
    # Unit Tests for Input Validation
    assert_equal "valid" "$(validate_integer 15)" "Validate valid positive integer"
    assert_equal "invalid" "$(validate_integer -5)" "Validate negative integer rejection"
    assert_equal "invalid" "$(validate_integer 0)" "Validate zero integer rejection"
    assert_equal "invalid" "$(validate_integer 'abc')" "Validate string rejection"
    
    echo "All unit tests passed successfully."
}

# ------------------------------------------------------------------------------
# Function: spawn_terminal
# Description: Opens a new terminal emulator and executes the specified command
#              within the current working directory.
# ------------------------------------------------------------------------------
spawn_terminal() {
    local cmd="$1"
    local title="$2"
    local current_dir="$(pwd)"
    
    if command -v osascript &> /dev/null; then
        # macOS Terminal implementation
        osascript -e "tell application \"Terminal\"" \
                  -e "do script \"cd '$current_dir' && $cmd\"" \
                  -e "end tell" > /dev/null
    elif command -v gnome-terminal &> /dev/null; then
        # Linux GNOME Terminal implementation
        gnome-terminal --title="$title" -- bash -c "cd '$current_dir' && $cmd; exec bash" &
    elif command -v xterm &> /dev/null; then
        # Linux X11 fallback implementation
        xterm -T "$title" -e bash -c "cd '$current_dir' && $cmd; exec bash" &
    else
        echo "Error: No supported terminal emulator found (osascript, gnome-terminal, xterm)."
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# Function: main
# Description: The primary execution routine for deploying the network.
# ------------------------------------------------------------------------------
main() {
    if [ "$#" -ne 2 ]; then
        echo "Usage: $0 <number_of_messengers> <number_of_messages>"
        echo "To run unit tests: $0 --test"
        exit 1
    fi

    local num_messengers="$1"
    local num_messages="$2"

    if [ "$(validate_integer "$num_messengers")" != "valid" ]; then
        echo "Error: <number_of_messengers> must be a positive integer."
        exit 1
    fi

    if [ "$(validate_integer "$num_messages")" != "valid" ]; then
        echo "Error: <number_of_messages> must be a positive integer."
        exit 1
    fi

    echo "Starting Registry Process..."
    spawn_terminal "go run ./registry" "Minichord Registry"

    # Allow the registry sufficient time to initialize its listener loop 
    # and bind to the designated TCP port before launching clients.
    sleep 2

    echo "Spawning $num_messengers messenger nodes..."
    for (( i=1; i<=num_messengers; i++ ))
    do
        echo "Bootstrapping Messenger $i..."
        spawn_terminal "go run ./messenger" "Messenger Node $i"
        
        # Micro-sleep to prevent TCP handshake bottlenecks
        sleep 0.2
    done

    echo "============================================================"
    echo "Overlay network initialization complete."
    echo "Total Messengers Spawned : $num_messengers"
    echo "Configured Task Messages : $num_messages"
    echo ""
    echo "Note: The network is running. Use your Registry terminal"
    echo "to issue the command to begin routing the $num_messages messages."
    echo "============================================================"
}

# ------------------------------------------------------------------------------
# Entry Point
# ------------------------------------------------------------------------------
if [ "$1" == "--test" ]; then
    test_suite
    exit 0
else
    # Prerequisite verification
    if ! command -v go &> /dev/null; then
        echo "Error: The 'go' compiler is not installed or not in the system PATH."
        exit 1
    fi
    main "$@"
fi