#
# Copyright © 2025 pegagio
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
# THE SOFTWARE.
#

# envy - A Bash proxy for the envy Go layer
#
# This script acts as a transparent proxy to the Go-based `envy` CLI.
# It:
# - Passes the current shell environment state to the Go layer.
# - Updates the shell environment based on the output of `bin/envy`.
# - Tracks loaded environments dynamically using the ENVYLOADED variable.
#
# Usage:
#   envy load <profile>    # Load an environment
#   envy unload <profile>  # Unload an environment
#   envy list              # List all environments
#
# The `envy` function automatically modifies the environment by:
# - Setting variables, modifying PATH, and adding aliases when loading.
# - Unsetting variables, restoring PATH, and removing aliases when unloading.
#
# ENVYLOADED tracks active environments as a comma-separated list.
#
# This script should be sourced in your shell configuration file:
#   source /path/to/envy.sh
#
# Dependencies:
# - Requires `bin/envy` (Go CLI) in the PATH or specified manually.

envy() {
    local cmd="$1"
    shift

    envy_debug "ENVYLOADED (pre): $ENVYLOADED"
    local loaded_profiles="${ENVYLOADED:-}"

    # Fetch the current list of loaded environments from the tracked variable
    local cmd_str
    if [[ -n "$loaded_profiles" ]]; then
      cmd_str="envy-go -p $loaded_profiles"
    else
      cmd_str="envy-go"
    fi

    # Call the Go binary and capture its output
    args=("$@")
    cmd_str="$cmd_str $cmd ${args[*]}"
    envy_debug "Executing: $cmd_str" >&2
    output=$(eval "$cmd_str") || return $?

    # If the command is `load`, update ENVYLOADED and apply environment changes
    if [[ "$cmd" == "load" ]]; then
        local profile="$1"

        # Create or append to the list of loaded profiles
        if [[ -z "$ENVYLOADED" ]]; then
          envy_debug "Creating ENVYLOADED from $profile"
          ENVYLOADED="$profile"
        else
          envy_debug "Appending profile $profile to ENVYLOADED"
          ENVYLOADED+=",${profile}"
        fi

        # Evaluate the output to set environment variables, aliases, etc.
        envy_debug "Sourcing $output"
        # shellcheck disable=SC1090
        source "$output"

    # If the command is `unload`, update ENVYLOADED and revert environment changes
    elif [[ "$cmd" == "unload" ]]; then
        local profile="$1"
        if [[ -z "$profile" ]]; then
            echo "Usage: envy unload <profile>"
            return 1
        fi

        # Remove the profile from ENVYLOADED
        envy_debug "Removing profile $profile from ENVYLOADED"
        ENVYLOADED=$(echo "$ENVYLOADED" | awk -v profile="$profile" -F, '
        {
            first=1;
            for (i=1; i<=NF; i++) {
                if ($i != profile) {
                    if (!first) printf ",";
                    printf "%s", $i;
                    first=0;
                }
            }
            print "";
        }')

        # Evaluate the output to unset environment variables, aliases, etc.
        envy_debug "Sourcing $output"
        # shellcheck disable=SC1090
        source "$output"

    # Just pass it through to `envy`
    else
      envy_debug ">>> Command Output >>>"
      echo "$output"
      envy_debug "<<< Command Output <<<"
    fi

    envy_debug "ENVYLOADED (post): $ENVYLOADED"
}

envy_debug() {
    if [[ -n "$ENVYDEBUG" && "$ENVYDEBUG" =~ ^(1|true|yes|on)$ ]]; then
        echo "[DEBUG] $*" >&2
    fi
}

