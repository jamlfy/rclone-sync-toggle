BINARY_NAME=rclone-sync-toggle
INSTALL_DIR=$(HOME)/.config/rclone-sync-toggle
SYSTEMD_DIR=$(HOME)/.config/systemd/user
NAUTILUS_SCRIPTS_DIR=$(HOME)/.local/share/nautilus/scripts
NAUTILUS_EXT_DIR=$(HOME)/.local/share/nautilus-python/extensions
HICOLOR_ICONS_DIR=$(HOME)/.local/share/icons/hicolor
COSMIC_FILES_CONFIG_DIR=$(HOME)/.config/cosmic/com.system76.CosmicFiles/v1
APPS_DIR=$(HOME)/.local/share/applications
FILE_MANAGERS_DIR=file-managers
NAUTILUS_FILES_DIR=$(FILE_MANAGERS_DIR)/nautilus
COSMIC_FILES_DIR=$(FILE_MANAGERS_DIR)/cosmic-files
SERVICE_TEMPLATE=packaging/rclone-sync-toggle.service.in
SERVICE_DESCRIPTION?=Daemon de Sincronizacion Multiconector Rclone
SERVICE_AFTER?=network-online.target graphical-session.target
SERVICE_WANTS?=network-online.target
SERVICE_PASS_ENVIRONMENT?=DISPLAY DBUS_SESSION_BUS_ADDRESS XAUTHORITY LANG LC_ALL
SERVICE_RESTART?=on-failure
SERVICE_RESTART_SEC?=5
RCLONE_BIN_DIR?=$(HOME)/.local/bin
VERSION?=0.1.0
DEB_MAINTAINER?=jamlfy <jamlfy@users.noreply.github.com>
DEB_ARCH?=$(shell dpkg --print-architecture)
GOARCH?=$(shell go env GOARCH)
DEB_ROOT=build/deb-root
DEB_OUTPUT_DIR=dist

.PHONY: all build install install-service install-file-managers install-nautilus install-cosmic-files restart launch deb

all: build install restart

build:
	GOTOOLCHAIN=local go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/main.go

deb:
	@case "$(DEB_ARCH):$(GOARCH)" in amd64:amd64|arm64:arm64) ;; *) echo "Unsupported package architecture: $(DEB_ARCH) ($(GOARCH))"; exit 1 ;; esac
	mkdir -p "$(DEB_ROOT)/DEBIAN" \
		"$(DEB_ROOT)/usr/bin" \
		"$(DEB_ROOT)/usr/lib/systemd/user" \
		"$(DEB_ROOT)/usr/share/applications" \
		"$(DEB_ROOT)/usr/share/nautilus-python/extensions" \
		"$(DEB_ROOT)/usr/share/icons/hicolor/scalable/emblems" \
		"$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/file-managers/cosmic-files" \
		"$(DEB_OUTPUT_DIR)"
	GOTOOLCHAIN=local GOOS=linux GOARCH=$(GOARCH) go build -ldflags="-s -w" -o "$(DEB_ROOT)/usr/bin/$(BINARY_NAME)" ./cmd/main.go
	install -m 644 packaging/debian/rclone-sync-toggle.service "$(DEB_ROOT)/usr/lib/systemd/user/rclone-sync-toggle.service"
	install -m 644 packaging/debian/rclone-sync-toggle.desktop "$(DEB_ROOT)/usr/share/applications/rclone-sync-toggle.desktop"
	install -m 644 "$(NAUTILUS_FILES_DIR)/rclone-sync-toggle.py" "$(DEB_ROOT)/usr/share/nautilus-python/extensions/rclone-sync-toggle.py"
	install -m 644 "$(NAUTILUS_FILES_DIR)/icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg" "$(DEB_ROOT)/usr/share/icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg"
	install -m 644 "$(NAUTILUS_FILES_DIR)/icons/hicolor/scalable/emblems/emblem-rclone-local.svg" "$(DEB_ROOT)/usr/share/icons/hicolor/scalable/emblems/emblem-rclone-local.svg"
	install -m 644 "$(COSMIC_FILES_DIR)/context_actions" "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/file-managers/cosmic-files/context_actions"
	install -m 644 "$(COSMIC_FILES_DIR)/README.md" "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/file-managers/cosmic-files/README.md"
	install -m 644 "$(NAUTILUS_FILES_DIR)/README.md" "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/file-managers/nautilus-README.md"
	install -m 644 README.md "$(DEB_ROOT)/usr/share/doc/rclone-sync-toggle/README.md"
	printf 'Package: rclone-sync-toggle\nVersion: %s\nSection: utils\nPriority: optional\nArchitecture: %s\nMaintainer: %s\nDepends: rclone (>= 1.58.0), zenity, libgtk-3-0, libnotify-bin, systemd, fuse3 | fuse\nRecommends: python3-nautilus\nDescription: Local and virtual folder synchronization with Rclone\n Linux desktop utility for managing hybrid and virtual Rclone folder mappings.\n Includes Nautilus extension/emblems and COSMIC Files context action template.\n' "$(VERSION)" "$(DEB_ARCH)" "$(DEB_MAINTAINER)" > "$(DEB_ROOT)/DEBIAN/control"
	chmod -R go-w "$(DEB_ROOT)"
	dpkg-deb --build --root-owner-group "$(DEB_ROOT)" "$(DEB_OUTPUT_DIR)/$(BINARY_NAME)_$(VERSION)_$(DEB_ARCH).deb"

launch: build
	./$(BINARY_NAME) --config

install-service:
	mkdir -p $(SYSTEMD_DIR)
	@sed \
		-e 's|@SERVICE_DESCRIPTION@|$(SERVICE_DESCRIPTION)|g' \
		-e 's|@SERVICE_AFTER@|$(SERVICE_AFTER)|g' \
		-e 's|@SERVICE_WANTS@|$(SERVICE_WANTS)|g' \
		-e 's|@SERVICE_PASS_ENVIRONMENT@|$(SERVICE_PASS_ENVIRONMENT)|g' \
		-e 's|@RCLONE_BIN_DIR@|$(RCLONE_BIN_DIR)|g' \
		-e 's|@INSTALL_DIR@|$(INSTALL_DIR)|g' \
		-e 's|@BINARY_NAME@|$(BINARY_NAME)|g' \
		-e 's|@SERVICE_RESTART@|$(SERVICE_RESTART)|g' \
		-e 's|@SERVICE_RESTART_SEC@|$(SERVICE_RESTART_SEC)|g' \
		$(SERVICE_TEMPLATE) > "$(SYSTEMD_DIR)/rclone-sync-toggle.service"
	systemctl --user daemon-reload

install-nautilus:
	mkdir -p "$(NAUTILUS_EXT_DIR)" "$(HICOLOR_ICONS_DIR)/scalable/emblems"
	install -m 644 "$(NAUTILUS_FILES_DIR)/rclone-sync-toggle.py" "$(NAUTILUS_EXT_DIR)/rclone-sync-toggle.py"
	install -m 644 "$(NAUTILUS_FILES_DIR)/icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg" "$(HICOLOR_ICONS_DIR)/scalable/emblems/emblem-rclone-virtual.svg"
	install -m 644 "$(NAUTILUS_FILES_DIR)/icons/hicolor/scalable/emblems/emblem-rclone-local.svg" "$(HICOLOR_ICONS_DIR)/scalable/emblems/emblem-rclone-local.svg"
	gtk-update-icon-cache -f -t "$(HICOLOR_ICONS_DIR)" || true
	@echo "Nautilus integration installed. Restart Nautilus with: nautilus -q"

install-cosmic-files:
	mkdir -p "$(COSMIC_FILES_CONFIG_DIR)"
	install -m 644 "$(COSMIC_FILES_DIR)/context_actions" "$(COSMIC_FILES_CONFIG_DIR)/context_actions"
	@echo "COSMIC Files context action installed at $(COSMIC_FILES_CONFIG_DIR)/context_actions"

install-file-managers: install-nautilus install-cosmic-files

install: build install-service install-file-managers
	mkdir -p $(INSTALL_DIR) $(APPS_DIR)
	install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	chmod +x $(INSTALL_DIR)/$(BINARY_NAME)

	@printf '[Desktop Entry]\nName=Rclone Sync Toggle - Configurar\nComment=Gestionar mapeos Carpeta -> Conexion de Rclone\nExec=%s/%s --config\nIcon=preferences-system-network\nTerminal=false\nType=Application\nCategories=Settings;System;\n' "$(INSTALL_DIR)" "$(BINARY_NAME)" > $(APPS_DIR)/rclone-sync-toggle.desktop

restart:
	systemctl --user daemon-reload
	systemctl --user enable --now rclone-sync-toggle.service
	systemctl --user restart rclone-sync-toggle.service
