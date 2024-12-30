#!/bin/sh

# Download Tailwind CLI
#
if ! test -f bin/tailwindcss ; then
	mkdir -p bin
	echo "Downloading Tailwind CLI"
	wget https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64 -O bin/tailwindcss
	chmod +x bin/tailwindcss
fi
