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
    PATCHLIST=`find -name '*.go' -exec grep -il '"os"' {} \;`

    # Patch'em
    while IFS= read -r line
    do
        echo "PATCH" $line
        sed -i 's/os\.Open/caviar\.Open/g' $line
        sed -i 's/\*os\.File/caviar\.File/g' $line
        sed -i 's/os\.FileInfo/ostmp\.FileInfo/g' $line
        sed -i 's/os\.File/caviar\.File/g' $line
        sed -i 's/ostmp\.FileInfo/os\.FileInfo/g' $line
        sed -i 's/"os"/"os"\n\t"github.com\/mvillalba\/caviar"/g' $line
        echo "" >> $line
        echo "// Bypass unused-imports problem (remove if this is your own code)" >> $line
        echo "var _ = os.Open" >> $line
        echo "var _ = caviar.Open" >> $line
    done <<< "${PATCHLIST}"
}

# Download and patch ($PKG and dependencies)
PKGLIST=`go get -v -d -u $PKG | awk '{ print $1 }'`
while IFS= read -r line
do
    echo $line
# FIXME: Broken, let's ignore dependencies
#    caviarpatch $line
    caviarpatch $PKG
done <<< "${PKGLIST}"

# Install $PKG and dependencies
go install $PKG
