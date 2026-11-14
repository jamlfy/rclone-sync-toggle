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
		"ui.mapping_updated":         "Conexión actualizada",
		"ui.manager_title":           "Gestionar Carpetas en Nube",
		"ui.active":                  "Activo",
		"ui.connection":              "Conexión",
		"ui.local_folder":            "Carpeta Local",
		"ui.cloud_folder":            "Carpeta en Nube",
		"ui.mode":                    "Modo",
		"ui.edit":                    "Editar",
		"ui.edit_title":              "Editar conexión",
		"ui.save":                    "Guardar",
		"ui.cancel":                  "Cancelar",
		"ui.invalid_local_folder":    "Indica una carpeta local.",
		"ui.invalid_connection":      "Indica una conexión de Rclone.",
		"ui.resync_warning":          "Si el perfil queda en modo hybrid y cambias una ruta, se ejecutará una resincronización inicial. La versión local tiene prioridad en conflictos.",
		"ui.add_synced_folder":       "+ Agregar Carpeta Sincronizada",
		"ui.delete":                  "Borrar",
		"ui.save_apply":              "Guardar y Aplicar",
		"ui.no_active_folders":       "Sin carpetas activas. Servicio detenido.",
		"ui.applied":                 "Configuración aplicada y servicio reiniciado.",
		"ui.mapping_removed":         "Conexión eliminada",
		"ui.mapping_removed_details": "Se eliminaron las conexiones: %s. Los archivos locales y de la nube no se borraron.",
		"ui.save_failed":             "No se pudo guardar la configuración: %v",
		"ui.stop_failed":             "No se pudo detener el servicio: %v",
		"ui.restart_failed":          "No se pudo reiniciar el servicio: %v",
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
		"ui.mapping_updated":         "Mapping Updated",
		"ui.manager_title":           "Manage Cloud Folders",
		"ui.active":                  "Active",
		"ui.connection":              "Connection",
		"ui.local_folder":            "Local Folder",
		"ui.cloud_folder":            "Cloud Folder",
		"ui.mode":                    "Mode",
		"ui.edit":                    "Edit",
		"ui.edit_title":              "Edit Mapping",
		"ui.save":                    "Save",
		"ui.cancel":                  "Cancel",
		"ui.invalid_local_folder":    "Enter a local folder.",
		"ui.invalid_connection":      "Enter an Rclone connection.",
		"ui.resync_warning":          "If the profile is hybrid and you change a path, an initial resync will run. The local version takes precedence in conflicts.",
		"ui.add_synced_folder":       "+ Add Synced Folder",
		"ui.delete":                  "Delete",
		"ui.save_apply":              "Save and Apply",
		"ui.no_active_folders":       "No active folders. Service stopped.",
		"ui.applied":                 "Configuration applied and service restarted.",
		"ui.mapping_removed":         "Mapping Removed",
		"ui.mapping_removed_details": "Removed mappings: %s. Local and cloud files were not deleted.",
		"ui.save_failed":             "Could not save configuration: %v",
		"ui.stop_failed":             "Could not stop the service: %v",
		"ui.restart_failed":          "Could not restart the service: %v",
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
