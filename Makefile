PLATFORMS := linux/amd64 linux/arm64 windows/amd64 darwin/arm64 darwin/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS):
	mkdir -p bin/$(os)-$(arch)
	GOOS=$(os) GOARCH=$(arch) go build -o 'bin/$(os)-$(arch)/magnetron' cmd/magnetron/main.go
	mkdir -p bin/dist
	tar -zcvf 'bin/dist/magnetron-$(os)-$(arch).tar.gz' 'bin/$(os)-$(arch)'

clean:
	rm -rf bin

all: clean release $(PLATFORMS)

.PHONY: all