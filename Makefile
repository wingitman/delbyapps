BINARY := delbyapps
INSTALL_DIR ?= $(HOME)/.local/bin
WORK_DIR ?= $(abspath ..)
LDFLAGS := -X 'main.defaultWorkDir=$(WORK_DIR)'

.PHONY: build install uninstall clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build
	mkdir -p "$(INSTALL_DIR)"
	cp "$(BINARY)" "$(INSTALL_DIR)/$(BINARY)"
	@case ":$$PATH:" in *:":$(INSTALL_DIR)":*) ;; *) echo "Add $(INSTALL_DIR) to your PATH if delbyapps is not found." ;; esac

uninstall:
	rm -f "$(INSTALL_DIR)/$(BINARY)"

clean:
	rm -f "$(BINARY)" "$(BINARY).exe"
