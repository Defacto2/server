#!/bin/bash

# The following script is used to run the server with environment variables.
# The environment variables are loaded from a file named ".env.local" but 
# this can be changed by modifying the FILENAME variable below.
#
# The df2-server binary should be in the same directory as this script.

# Filename containing the environment variables
FILENAME=.env.local

# Load environment variables from .env.local
echo -e "Loading environment variables from $FILENAME\n"
export $(grep -E -v '^#' $FILENAME | xargs)

# Run the server
./df2-server

# Unset environment variables from .env.local
echo -e "\nUnset environment variables from $FILENAME\n"
unset $(grep -E -v '^#' $FILENAME | sed -E 's/(.*)=.*/\1/' | xargs)
