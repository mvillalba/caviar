#!/usr/bin/env bash

set -e

# Check args
# TODO: check arg count (1 or 2). If 2, the first arg must be '-l' to mean this
# is a local package already present in the $GOPATH tree, don't download it.
PKG=$1

# Patching function
function caviarpatch {
    cd $GOPATH/src/$1

    # Who needs patching?
    PATCHLIST=`find -name '*.go' -exec grep -il 'os.Open' {} \;`

    # Patch'em
    while IFS= read -r line
    do
        sed -i 's/os\.Open/caviar\.Open/g' $line
        sed -i 's/"os"/"os"\n\t"github.com/mvillalba/caviar"/g' $line
        echo "" >> $line
        echo "// Bypass unused-imports problem (remove if this is your own code)" >> $line
        echo "var _ = os.Open" >> $line
    done <<< "${PATCHLIST}"
}

# Download and patch ($PKG and dependencies)
PKGLIST=`go get -v -d -u $PKG | awk '{ print $1 }'`
while IFS= read -r line
do
    caviarpatch $line
done <<< "${PKGLIST}"

# Install $PKG and dependencies
go install $PKG
