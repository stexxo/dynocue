#!/bin/bash

# Configuration: Define the license text
LICENSE_TEXT="This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/."

# Function to get the header based on comment style
# $1: style (xml, c, hash)
get_header() {
    case "$1" in
        xml)
            echo "<!--"
            echo "$LICENSE_TEXT" | sed 's/^/  /'
            echo "-->"
            ;;
        c)
            echo "$LICENSE_TEXT" | sed 's/^/\/\/ /'
            ;;
        hash)
            echo "$LICENSE_TEXT" | sed 's/^/# /'
            ;;
    esac
}

# Function to apply license to a file
apply_license() {
    local file="$1"
    local style="$2"
    
    # Generate the header for this style
    local header=$(get_header "$style")
    
    # Check if the license is already there
    # We search for a unique part of the license text
    if grep -q "Mozilla Public" "$file"; then
        return
    fi
    
    echo "Adding license to $file ($style style)"
    
    # Create a temporary file with the header and then the original content
    {
        echo "$header"
        echo "" # Add a blank line after the header
        cat "$file"
    } > "$file.tmp" && mv "$file.tmp" "$file"
}

# Main logic: Define mappings of extensions to comment styles
# Add or remove extensions here as needed
declare -A EXT_STYLE_MAP
EXT_STYLE_MAP["svelte"]="xml"
EXT_STYLE_MAP["go"]="c"
EXT_STYLE_MAP["ts"]="c"
EXT_STYLE_MAP["js"]="c"
EXT_STYLE_MAP["yml"]="hash"
EXT_STYLE_MAP["sh"]="hash"

# Exclude directories and generated files
EXCLUDE_DIRS="-not -path '*/node_modules/*' -not -path '*/.svelte-kit/*' -not -path '*/.git/*' -not -path '*/dist/*' -not -path '*/bindings/*' -not -path '*/dist/*"

# Process each extension
for ext in "${!EXT_STYLE_MAP[@]}"; do
    style="${EXT_STYLE_MAP[$ext]}"
    echo "Processing .$ext files..."
    
    # Find files and apply license
    # Using eval because of the EXCLUDE_DIRS variable containing wildcards
    eval "find . -name '*.$ext' -type f $EXCLUDE_DIRS" | while read -r file; do
        apply_license "$file" "$style"
    done
done

echo "Done!"
