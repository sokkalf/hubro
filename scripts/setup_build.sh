#!/bin/sh

if ! test -f /buildtools/tailwindcss-musl; then
	mkdir -p /buildtools
	echo "Downloading Tailwind CLI (musl)"
	wget https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0/tailwindcss-linux-x64-musl -O /buildtools/tailwindcss-musl
	chmod +x /buildtools/tailwindcss-musl
fi

if ! test -f /buildtools/esbuild ; then
	mkdir -p /buildtools
	echo "Downloading esbuild"
	mkdir -p /tmp/bin/
	wget "https://registry.npmjs.org/@esbuild/linux-x64/-/linux-x64-0.24.1.tgz" -O /tmp/esbuild.tgz
	tar -xzf /tmp/esbuild.tgz -C /tmp/bin/
	mv /tmp/bin/package/bin/esbuild /buildtools/esbuild
	rm -rf /tmp/*
fi
