.PHONY: build serve clean

build:
	@echo "Building site binary..."
	@templ generate
	@go build -o ./site
	@echo "✓ Binary built: ./site"

serve: build
	@./site serve

deploy: build
	@./site build

clean:
	@rm -rf public/ site
	@find . -name '*_templ.go' -delete
	@echo "✓ Cleaned"
