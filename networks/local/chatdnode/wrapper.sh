#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/chatd/${BINARY:-chatd}
ID=${ID:-0}
LOG=${LOG:-chatd.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'chatd' E.g.: -e BINARY=chatd_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export CHATDHOME="/chatd/node${ID}/chatd"

if [ -d "`dirname ${CHATDHOME}/${LOG}`" ]; then
  "${BINARY}" --home "${CHATDHOME}" "$@" | tee "${CHATDHOME}/${LOG}"
else
  "${BINARY}" --home "${CHATDHOME}" "$@"
fi

chmod 777 -R /chatd

