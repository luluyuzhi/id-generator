#!/bin/bash
PROJECTROOT=$PWD

mkdir build && cd build

cmake -D CMAKE_INSTALL_PREFIX=${PROJECTROOT} ..

make install

cd ${PROJECTROOT}

# rm -rf build