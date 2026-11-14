package i18n

import (
	"fmt"
	"os"
	"strings"
)

var language = "es"

var messages = map[string]map[string]string{
	"es": {
		"config.read_error":          "Error leyendo configuración: %v",
		"config.empty":               "No hay carpetas configuradas en %s. Ejecuta con --config para agregar.",
		"daemon.started":             "Daemon iniciado con %d mapeo(s) activo(s).",
		"daemon.stopping":            "Apagando daemon y desmontando unidades virtuales...",
		"ui.sync_new":                "Sincronizando carpeta nueva (--resync)...",
		"ui.mapping_added":           "Mapeo Agregado",
		"ui.manager_title":           "Gestor Carpeta -> Conexión (Rclone)",
		"ui.manager_text":            "Mapeos activos leídos de sync-profiles.ini.\n• Desmarca una fila y pulsa 'Guardar y Aplicar' para quitarla.\n• Pulsa '+ Agregar Carpeta Sincronizada' para añadir otra.",
		"ui.active":                  "Activo",
		"ui.id":                      "ID",
		"ui.local_folder":            "Carpeta Local",
		"ui.rclone_remote":           "Conexión Rclone",
		"ui.mode":                    "Modo",
		"ui.add_synced_folder":       "+ Agregar Carpeta Sincronizada",
		"ui.save_apply":              "Guardar y Aplicar",
		"ui.close":                   "Cerrar",
		"ui.no_active_folders":       "Sin carpetas activas. Servicio detenido.",
		"ui.applied":                 "Configuración aplicada y servicio reiniciado.",
		"ui.add_title":               "Agregar Carpeta -> Conexión",
		"ui.add_text":                "Asocia una carpeta de tu PC con una conexión de Rclone:\n\n• Modo 'hybrid': Descarga en disco + carpetas virtuales selectivas.\n• Modo 'virtual': Solo acceso bajo demanda (sin ocupar disco).",
		"ui.remote":                  "Conexión Rclone",
		"ui.local_path_hint":         "Carpeta Local (ej. ~/Proyectos)",
		"ui.cloud_path":              "Ruta en la nube (ej. / o /Proyectos)",
		"ui.bisync_flags":            "Flags Bisync (vacío = por defecto)",
		"ui.mount_flags":             "Flags Mount (vacío = por defecto)",
		"ui.list_remotes_error":      "No se pudieron leer las conexiones de rclone.",
		"ui.no_remotes":              "No hay conexiones en rclone. Ejecuta 'rclone config' primero.",
		"ui.toggle_unmatched":        "La ruta no pertenece a ninguna carpeta configurada.",
		"ui.already_virtual":         "Esta carpeta completa ya es 100% virtual.",
		"ui.item_missing":            "No se encontró el elemento seleccionado.",
		"ui.cloud_item_missing":      "El archivo aún no está en la nube. Espera a que termine de subir.",
		"ui.downloading":             "Descargando '%s' al disco duro...",
		"ui.now_virtual":             "'%s' ahora es solo virtual.",
		"worker.virtual_mode":        "Modo 100%% virtual: %s -> %s",
		"worker.watcher_start_error": "Error iniciando watcher: %v",
		"worker.scan_error":          "Error escaneando directorios: %v",
		"worker.hybrid_mode":         "Modo híbrido activo: %s <-> %s",
		"worker.watcher_error":       "Error en watcher: %v",
		"worker.filter_error":        "Error generando filtros: %v",
		"worker.bisync_error":        "Error en bisync: %v\n%s",
		"worker.sync_failed_title":   " - Error",
		"worker.sync_failed":         "Falló la sincronización. Revisa journalctl.",
		"worker.files_updated":       "Se actualizaron %d archivo(s).",
		"worker.synced":              " Sincronizado",
		"worker.mounting":            "Montando %s en %s",
		"worker.mount_stopped":       "Montaje detenido (%v), reintentando en 5s...",
	},
	"en": {
		"config.read_error":          "Error reading configuration: %v",
		"config.empty":               "No folders are configured in %s. Run with --config to add one.",
		"daemon.started":             "Daemon started with %d active mapping(s).",
		"daemon.stopping":            "Stopping daemon and unmounting virtual drives...",
		"ui.sync_new":                "Synchronizing new folder (--resync)...",
		"ui.mapping_added":           "Mapping Added",
		"ui.manager_title":           "Folder -> Remote Manager (Rclone)",
		"ui.manager_text":            "Active mappings loaded from sync-profiles.ini.\n• Uncheck a row and click 'Save and Apply' to remove it.\n• Click '+ Add Synced Folder' to add another.",
		"ui.active":                  "Active",
		"ui.id":                      "ID",
		"ui.local_folder":            "Local Folder",
		"ui.rclone_remote":           "Rclone Remote",
		"ui.mode":                    "Mode",
		"ui.add_synced_folder":       "+ Add Synced Folder",
		"ui.save_apply":              "Save and Apply",
		"ui.close":                   "Close",
		"ui.no_active_folders":       "No active folders. Service stopped.",
		"ui.applied":                 "Configuration applied and service restarted.",
		"ui.add_title":               "Add Folder -> Remote",
		"ui.add_text":                "Link a folder on your computer to an Rclone remote:\n\n• 'hybrid' mode: Files are downloaded to disk; selected folders can be virtual.\n• 'virtual' mode: On-demand access only (no disk space used).",
		"ui.remote":                  "Rclone Remote",
		"ui.local_path_hint":         "Local folder (e.g. ~/Projects)",
		"ui.cloud_path":              "Cloud path (e.g. / or /Projects)",
		"ui.bisync_flags":            "Bisync flags (blank = defaults)",
		"ui.mount_flags":             "Mount flags (blank = defaults)",
		"ui.list_remotes_error":      "Could not read the configured Rclone remotes.",
		"ui.no_remotes":              "No Rclone remotes found. Run 'rclone config' first.",
		"ui.toggle_unmatched":        "This path does not belong to any configured folder.",
		"ui.already_virtual":         "This entire folder is already 100% virtual.",
		"ui.item_missing":            "The selected item was not found.",
		"ui.cloud_item_missing":      "The file is not in the cloud yet. Wait for the upload to finish.",
		"ui.downloading":             "Downloading '%s' to disk...",
		"ui.now_virtual":             "'%s' is now virtual-only.",
		"worker.virtual_mode":        "100%% virtual mode: %s -> %s",
		"worker.watcher_start_error": "Error starting watcher: %v",
		"worker.scan_error":          "Error scanning directories: %v",
		"worker.hybrid_mode":         "Hybrid mode active: %s <-> %s",
		"worker.watcher_error":       "Watcher error: %v",
		"worker.filter_error":        "Error generating filters: %v",
		"worker.bisync_error":        "Bisync error: %v\n%s",
		"worker.sync_failed_title":   " - Error",
		"worker.sync_failed":         "Synchronization failed. Check journalctl.",
		"worker.files_updated":       "Updated %d file(s).",
		"worker.synced":              " Synced",
		"worker.mounting":            "Mounting %s at %s",
		"worker.mount_stopped":       "Mount stopped (%v), retrying in 5s...",
	},
}

func Init(preferred string) error {
	if preferred == "" {
		preferred = detectLanguage()
		if normalizeLanguage(preferred) == "" {
			preferred = "es"
		}
	}
	normalized := normalizeLanguage(preferred)
	if normalized == "" {
		return fmt.Errorf("unsupported language %q (use es or en)", preferred)
	}
	language = normalized
	return nil
}

func T(key string) string {
	if value, ok := messages[language][key]; ok {
		return value
	}
	return key
}

func detectLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(key); locale != "" {
			return locale
		}
	}
	return "es"
}

func normalizeLanguage(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if len(value) >= 2 {
		switch value[:2] {
		case "es", "en":
			return value[:2]
		}
	}
	return ""
}
