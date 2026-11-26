# Nautilus integration

Dynamic context menu action and status emblems for Nautilus using `python3-nautilus`.

## Install (user)

```sh
mkdir -p ~/.local/share/nautilus-python/extensions
install -m 644 rclone-sync-toggle.py ~/.local/share/nautilus-python/extensions/rclone-sync-toggle.py

mkdir -p ~/.local/share/icons/hicolor/scalable/emblems
install -m 644 icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg ~/.local/share/icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg
install -m 644 icons/hicolor/scalable/emblems/emblem-rclone-local.svg ~/.local/share/icons/hicolor/scalable/emblems/emblem-rclone-local.svg

gtk-update-icon-cache -f -t ~/.local/share/icons/hicolor || true
nautilus -q
```

## Install (system / package)

```text
/usr/share/nautilus-python/extensions/rclone-sync-toggle.py
/usr/share/icons/hicolor/scalable/emblems/emblem-rclone-virtual.svg
/usr/share/icons/hicolor/scalable/emblems/emblem-rclone-local.svg
```

## Dependency

- `python3-nautilus` / `nautilus-python`
- `rclone-sync-toggle` available in `PATH`

## Behavior

- Shows action only for items inside an enabled `hybrid` profile.
- Dynamic menu label:
  - virtual/symlink item -> "Keep it on the hard drive" / "Mantenerlo en el disco duro"
  - local item -> "Don't keep it on the hard drive" / "No mantenerlo en el disco duro"
- Adds emblems:
  - virtual item -> cloud emblem (`emblem-rclone-virtual`)
  - local item in hybrid profile -> check emblem (`emblem-rclone-local`)
- Hides menu action for virtual profiles, unmanaged paths, profile root, or mixed selections.
- Calls the existing `rclone-sync-toggle --toggle PATH` command, so toggle logic stays in the Go binary.

## Limitation

The extension reads `~/.config/rclone/sync-profiles.ini` directly to decide visibility/label/emblem. It does not reimplement toggle behavior; execution is delegated to the CLI.
