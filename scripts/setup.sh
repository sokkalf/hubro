#!/bin/sh

# Download Tailwind CLI
#
if ! test -f bin/tailwindcss ; then
	mkdir -p bin
	echo "Downloading Tailwind CLI"
	wget https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0/tailwindcss-linux-x64 -O bin/tailwindcss
	chmod +x bin/tailwindcss
fi

if ! test -f bin/esbuild ; then
	mkdir -p bin
	echo "Downloading esbuild"
	mkdir -p bin/tmp/
	wget "https://registry.npmjs.org/@esbuild/linux-x64/-/linux-x64-0.24.1.tgz" -O bin/tmp/esbuild.tgz
	tar -xzf bin/tmp/esbuild.tgz -C bin/
	mv bin/package/bin/esbuild bin/esbuild
	rm -rf bin/tmp
	rm -rf bin/package
fi
