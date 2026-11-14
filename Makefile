BINARY_NAME=rclone-sync-toggle
INSTALL_DIR=$(HOME)/.config/rclone-sync-toggle
SYSTEMD_DIR=$(HOME)/.config/systemd/user
NAUTILUS_DIR=$(HOME)/.local/share/nautilus/scripts
APPS_DIR=$(HOME)/.local/share/applications

.PHONY: all build install restart launch

all: build install restart

build:
	GOTOOLCHAIN=local go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/main.go

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
