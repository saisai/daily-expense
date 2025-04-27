#!/bin/bash

set -e

echoerr() { echo "$@" 1>&2; }

function get_platform() {
    # Will return "linux" for GNU/Linux
    #   I'd just like to interject for a moment...
    #   https://wiki.installgentoo.com/index.php/Interjection
    # Will return "macos" for macOS/OS X
    # Will return "windows" for Windows/MinGW/msys

    _platform=$(uname | tr '[:upper:]' '[:lower:]')
    if [[ $_platform == "darwin" ]]; then
        _platform="macos";
    elif [[ $_platform == "msys"* ]]; then
        _platform="windows";
    elif [[ $_platform == "mingw"* ]]; then
        _platform="windows";
    elif [[ $_platform == "linux" ]]; then
        # Nothing to do
        true;
    else
        echoerr "ERROR: $_platform is not a valid platform";
        exit 1;
    fi

    echo $_platform;
}


function get_arch() {
    _arch="$(uname -m)"
    echo $_arch;
}

platform=$(get_platform)
version="0.0.1"
arch=$(get_arch)
echo "Platform: $platform, arch: $arch, version: $version"

function build_zip() {
    echo "Zipping executables..."
    pushd dist;
    filename="Dailyexpense-${version}-${platform}-${arch}.zip"
    echo "Name of package will be: $filename"

    if [[ $platform == "windows"* ]]; then
        7z a $filename Dailyexpense;
    else
        zip -r $filename Dailyexpense;
    fi
    popd;
    echo "Zip built!"
}

function build_setup() {
    filename="Dailyexpense-${version}-${platform}-${arch}-setup.exe"
    echo "Name of package will be: $filename"

    innosetupdir="/c/Program Files (x86)/Inno Setup 6"
    if [ ! -d "$innosetupdir" ]; then
        echo "ERROR: Couldn't find innosetup which is needed to build the installer. We suggest you install it using chocolatey. Exiting."
        exit 1
    fi

    # Windows installer version should not include 'v' prefix, see: https://github.com/microsoft/winget-pkgs/pull/17564
    version_no_prefix="$(echo $version | sed -e 's/^v//')"
    env EXPENSE_VERSION=$version_no_prefix
    "$innosetupdir/iscc.exe" scripts/package/dailyexpense-setup.iss
    mv dist/Dailyexpense-setup.exe dist/$filename
    echo "Setup built!"
}

build_zip
if [[ $platform == "windows"* ]]; then
    build_setup
fi

echo
echo "-------------------------------------"
echo "Contents of ./dist"
ls -l dist
echo "-------------------------------------"
