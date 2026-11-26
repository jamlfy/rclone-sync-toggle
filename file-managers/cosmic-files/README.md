# COSMIC Files integration

Static user-defined context action for COSMIC Files (1.0.10+).

## Install (user)

```sh
mkdir -p ~/.config/cosmic/com.system76.CosmicFiles/v1
install -m 644 context_actions ~/.config/cosmic/com.system76.CosmicFiles/v1/context_actions
```

Restart COSMIC Files/session if needed.

## Limitation

COSMIC Files context actions currently have **static labels** and no dynamic emblem/status provider for this use case.

- Label cannot change depending on whether the selected item is virtual or local.
- Cloud/check emblems are **not implemented** for COSMIC Files in this integration.

This action uses a neutral label, "Alternar disponibilidad local", and calls the existing `rclone-sync-toggle --toggle` command, which performs the correct toggle based on current state.

## Notes

- The action uses `sh -c ... %F` because the current CLI handles one `--toggle` path at a time.
- If COSMIC Files adds dynamic menu/emblem extension support, Nautilus emblem logic can be used as the reference behavior.
