# Rclone Sync Toggle

A Linux desktop utility that keeps local folders synchronized with configured Rclone remotes. Hybrid profiles keep files on disk while allowing selected folders to be virtual; virtual profiles expose a remote on demand without storing its contents locally.

## Requirements

- Go 1.22 or later
- GTK 3 runtime library
- [Rclone](https://rclone.org/) with at least one remote configured using `rclone config`
- Zenity, `notify-send`, and FUSE userspace tools (`fusermount` or `fusermount3`)
- A Linux desktop session with a user-level systemd instance

## Build and install

```sh
make install
```

This builds and installs the executable under `~/.config/rclone-sync-toggle`, creates the `rclone-sync-toggle.service` user unit, adds a desktop launcher, and installs a Nautilus script for toggling selected files or folders. To build without installing, run `make build`; run `make restart` to enable and start the user service after configuring profiles.

## Debian package

Build a native `.deb` package on Debian, Ubuntu, or Pop!_OS:

```sh
make deb
```

The package is written to `dist/rclone-sync-toggle_0.1.0_<architecture>.deb`. Override the version with `make deb VERSION=1.0.0`. Install it with:

```sh
sudo apt install ./dist/rclone-sync-toggle_0.1.0_amd64.deb
```

Open the desktop launcher or run `rclone-sync-toggle --config` to create a profile, then enable the user service:

```sh
systemctl --user enable --now rclone-sync-toggle.service
```

Packages are built for the native `amd64` or `arm64` architecture. A package includes the application, desktop launcher, user service unit, Nautilus script, and README; it does not include personal Rclone credentials or `sync-profiles.ini`. Install Rclone and configure each remote separately on every machine.

Open the configuration window from the desktop launcher or run:

```sh
rclone-sync-toggle --config
```

The first run asks you to create a mapping. Choose a local folder, an existing Rclone remote, a path on that remote, and either `hybrid` or `virtual` mode. Optional Bisync and mount flags can be left blank to use the built-in defaults. Profiles are stored in `~/.config/rclone/sync-profiles.ini`.

## Language

The interface, desktop notifications, and daemon messages are available in English and Spanish. By default, the application uses the system locale (`LC_ALL`, `LC_MESSAGES`, or `LANG`) and falls back to Spanish. Override it for a run with:

```sh
rclone-sync-toggle --lang en --config
rclone-sync-toggle --lang es --config
```

Locale variants such as `en_US.UTF-8` and `es_MX.UTF-8` are recognized.

## How it works

- **Hybrid mode:** Rclone mounts the remote in a cache directory, and Rclone Bisync synchronizes that remote with the local folder. Filesystem changes trigger a sync after a short debounce; a periodic sync also runs every three minutes. Globally ignored directories include `.git`, `node_modules`, and Python virtual environments.
- **Virtual mode:** Rclone mounts the remote directly at the configured local folder. Files are accessed on demand rather than mirrored as local files.
- **Toggle selected items:** The Nautilus script runs `rclone-sync-toggle --toggle PATH`. In a hybrid profile, selecting a local item checks that its remote-mounted counterpart exists, replaces the local item with a symlink to the mount, and excludes it from Bisync so it is accessed virtually. Selecting that symlink again removes the virtual exclusion and triggers a resync to restore the local copy. The whole folder in a virtual profile is already virtual.
- **Manage mappings:** The Go/GTK window lists each profile's active checkbox, Rclone connection, local folder, cloud folder, mode, and per-row edit and trash icons. The pencil edits the connection, paths, mode, and Bisync/mount flags; changing a hybrid profile's sync paths triggers an initial resync. The checkbox enables or pauses a profile; the trash icon removes its configuration. **Save and Apply** persists changes, and **Add Synced Folder** creates a mapping. Removing a mapping leaves local and cloud files untouched. Existing profiles without an `enabled` setting remain enabled.
- **Configuration changes:** The graphical manager reads and writes the profile file, then restarts or stops the user service to apply changes. The service supervises Rclone mounts and starts a Bisync worker for each configured profile.

You can inspect service logs with:

```sh
journalctl --user -u rclone-sync-toggle.service
```

## Commands

```text
rclone-sync-toggle --config           Open the graphical profile manager
rclone-sync-toggle --toggle PATH      Toggle a file/folder between local and virtual
rclone-sync-toggle --lang en          Use English (also accepts es)
```
