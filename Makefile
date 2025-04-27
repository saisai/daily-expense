.PHONY: build-cli
build-cli:
	bash build.sh

package:
	#rm -rf dist
	#mkdir dist
	mkdir -p dist/Dailyexpense
	cp -rv bin dist/Dailyexpense/
	cp -rv .env dist/Dailyexpense/bin/
	bash scripts/package/package-all.sh