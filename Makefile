BINARY_NAME=rclone-sync-toggle
INSTALL_DIR=$(HOME)/.config/rclone-sync-toggle
SYSTEMD_DIR=$(HOME)/.config/systemd/user
NAUTILUS_DIR=$(HOME)/.local/share/nautilus/scripts
APPS_DIR=$(HOME)/.local/share/applications
VERSION?=0.1.0
DEB_MAINTAINER?=jamlfy <jamlfy@users.noreply.github.com>
DEB_ARCH?=$(shell dpkg --print-architecture)
GOARCH?=$(shell go env GOARCH)
DEB_ROOT=build/deb-root
DEB_OUTPUT_DIR=dist

.PHONY: all build install restart launch deb

all: build install restart

build:
	GOTOOLCHAIN=local go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/main.go

deb:
	@case "$(DEB_ARCH):$(GOARCH)" in amd64:amd64|arm64:arm64) ;; *) echo "Unsupported package architecture: $(DEB_ARCH) ($(GOARCH))"; exit 1 ;; esac
	mkdir -p "$(DEB_ROOT)/DEBIAN" "$(DEB_ROOT)/usr/bin" "$(DEB_ROOT)/usr/lib/systemd/user" "$(DEB_ROOT)/usr/share/applications" "$(DEB_ROOT)/usr/share/nautilus/scripts" "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle" "$(DEB_OUTPUT_DIR)"
	GOTOOLCHAIN=local GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o "$(DEB_ROOT)/usr/bin/$(BINARY_NAME)" ./cmd/main.go
	install -m 644 packaging/debian/rclone-sync-toggle.service "$(DEB_ROOT)/usr/lib/systemd/user/rclone-sync-toggle.service"
	install -m 644 packaging/debian/rclone-sync-toggle.desktop "$(DEB_ROOT)/usr/share/applications/rclone-sync-toggle.desktop"
	install -m 755 "packaging/debian/No mantenerlo en el disco duro" "$(DEB_ROOT)/usr/share/nautilus/scripts/No mantenerlo en el disco duro"
	install -m 644 README.md "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/README.md"
	printf 'Package: rclone-sync-toggle\nVersion: %s\nSection: utils\nPriority: optional\nArchitecture: %s\nMaintainer: %s\nDepends: rclone (>= 1.58.0), zenity, libgtk-3-0, libnotify-bin, systemd, fuse3 | fuse\nDescription: Local and virtual folder synchronization with Rclone\n Linux desktop utility for managing hybrid and virtual Rclone folder mappings.\n' "$(VERSION)" "$(DEB_ARCH)" "$(DEB_MAINTAINER)" > "$(DEB_ROOT)/DEBIAN/control"
	chmod -R go-w "$(DEB_ROOT)"
	dpkg-deb --build --root-owner-group "$(DEB_ROOT)" "$(DEB_OUTPUT_DIR)/$(BINARY_NAME)_$(VERSION)_$(DEB_ARCH).deb"

launch: build
	./$(BINARY_NAME) --config

install: build
	mkdir -p $(INSTALL_DIR) $(SYSTEMD_DIR) "$(NAUTILUS_DIR)" $(APPS_DIR)
	install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

	@printf '[Unit]\nDescription=Daemon de Sincronizacion Multiconector Rclone\nAfter=network-online.target graphical-session.target\nWants=network-online.target\n\n[Service]\nType=simple\nPassEnvironment=DISPLAY DBUS_SESSION_BUS_ADDRESS XAUTHORITY LANG LC_ALL\nExecStart=%s/%s\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=default.target\n' "$(INSTALL_DIR)" "$(BINARY_NAME)" > $(SYSTEMD_DIR)/rclone-sync-toggle.service

	@printf '#!/bin/bash\nwhile IFS= read -r FILE_PATH; do\n    if [ -n "$$FILE_PATH" ]; then\n        "%s/%s" --toggle "$$FILE_PATH"\n    fi\ndone <<< "$$NAUTILUS_SCRIPT_SELECTED_FILE_PATHS"\n' "$(INSTALL_DIR)" "$(BINARY_NAME)" > "$(NAUTILUS_DIR)/No mantenerlo en el disco duro"
	chmod +x "$(NAUTILUS_DIR)/No mantenerlo en el disco duro"

	@printf '[Desktop Entry]\nName=Rclone Sync Toggle - Configurar\nComment=Gestionar mapeos Carpeta -> Conexion de Rclone\nExec=%s/%s --config\nIcon=preferences-system-network\nTerminal=false\nType=Application\nCategories=Settings;System;\n' "$(INSTALL_DIR)" "$(BINARY_NAME)" > $(APPS_DIR)/rclone-sync-toggle.desktop

restart:
	systemctl --user daemon-reload
	systemctl --user enable --now rclone-sync-toggle.service
	systemctl --user restart rclone-sync-toggle.service
