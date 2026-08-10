BINARY  := fixxl
PKG     := github.com/krishnasureshcpa/fixxl/cmd/fixxl
OUT     := dist

.PHONY: build test vet demo demo-gif dist clean

build:
	go build -trimpath -o bin/$(BINARY) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

demo:
	go run $(PKG) demo --plain

demo-gif:
	vhs assets/type.tap

# Cross-platform release binaries (pure Go, no cgo).
dist:
	rm -rf $(OUT)
	mkdir -p $(OUT)
	for t in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64; do \
		goos=$${t%/*}; goarch=$${t#*/}; \
		out=$(OUT)/$(BINARY)-$$goos-$$goarch; \
		[ $$goos = windows ] && out=$$out.exe; \
		CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags "-s -w" \
			-o $$out $(PKG); \
	done

checksums: dist
	cd $(OUT) && shasum -a 256 * | tee SHA256SUMS

clean:
	rm -rf $(OUT) bin