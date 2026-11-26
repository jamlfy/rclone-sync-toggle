#!/usr/bin/env python3

import configparser
import os
import subprocess
from urllib.parse import unquote, urlparse

import gi

gi.require_version("Nautilus", "4.0")
from gi.repository import GObject, Nautilus

PROFILES_PATH = os.path.expanduser("~/.config/rclone/sync-profiles.ini")

LABELS = {
    "en": {
        "make_virtual": "Don't keep it on the hard drive",
        "make_local": "Keep it on the hard drive",
    },
    "es": {
        "make_virtual": "No mantenerlo en el disco duro",
        "make_local": "Mantenerlo en el disco duro",
    },
}

EMBLEM_VIRTUAL = "emblem-rclone-virtual"
EMBLEM_LOCAL = "emblem-rclone-local"

_profile_cache = {"mtime": None, "profiles": []}


def _lang():
    for key in ("LC_ALL", "LC_MESSAGES", "LANG"):
        value = os.environ.get(key, "")
        lower = value.lower()
        if lower.startswith("en"):
            return "en"
        if lower.startswith("es"):
            return "es"
    return "es"


def _labels():
    return LABELS[_lang()]


def _expand_tilde(path):
    if path.startswith("~/"):
        return os.path.join(os.path.expanduser("~"), path[2:])
    return path


def _load_profiles():
    try:
        mtime = os.path.getmtime(PROFILES_PATH)
    except OSError:
        _profile_cache["mtime"] = None
        _profile_cache["profiles"] = []
        return []

    if _profile_cache["mtime"] == mtime:
        return _profile_cache["profiles"]

    config = configparser.ConfigParser()
    config.optionxform = str
    if not config.read(PROFILES_PATH):
        _profile_cache["mtime"] = mtime
        _profile_cache["profiles"] = []
        return []

    profiles = []
    for section in config.sections():
        enabled = config.get(section, "enabled", fallback="true").strip().lower()
        if enabled == "false":
            continue
        mode = config.get(section, "mode", fallback="hybrid").strip().lower()
        local_dir = _expand_tilde(config.get(section, "local_dir", fallback="").strip())
        if not local_dir:
            continue
        profiles.append(
            {
                "name": section,
                "mode": mode,
                "local_dir": os.path.abspath(local_dir),
            }
        )

    _profile_cache["mtime"] = mtime
    _profile_cache["profiles"] = profiles
    return profiles


def _path_from_file(file_info):
    try:
        uri = file_info.get_uri()
    except Exception:
        return None
    if not uri or not uri.startswith("file://"):
        return None
    return os.path.abspath(unquote(urlparse(uri).path))


def _matching_profile(path, profiles):
    for profile in profiles:
        local_dir = profile["local_dir"]
        if path == local_dir or path.startswith(local_dir + os.sep):
            return profile
    return None


def _item_state(path, profiles):
    profile = _matching_profile(path, profiles)
    if profile is None:
        return None
    if profile["mode"] != "hybrid":
        return None
    if path == profile["local_dir"]:
        return None
    try:
        if os.path.islink(path):
            return "virtual"
        if os.path.exists(path):
            return "local"
    except OSError:
        return None
    return None


def _apply_emblem(file_info, state):
    emblem = EMBLEM_VIRTUAL if state == "virtual" else EMBLEM_LOCAL
    try:
        file_info.add_emblem(emblem)
    except Exception:
        pass


class RcloneSyncToggleMenuProvider(GObject.GObject, Nautilus.MenuProvider):
    def get_file_items(self, *args):
        # Nautilus 4 passes (files), older APIs may pass (window, files).
        files = args[-1] if args else []
        if not files:
            return []

        paths = []
        for file_info in files:
            path = _path_from_file(file_info)
            if not path:
                return []
            paths.append(path)

        profiles = _load_profiles()
        states = {_item_state(path, profiles) for path in paths}
        if len(states) != 1 or None in states:
            return []

        state = states.pop()
        labels = _labels()
        label = labels["make_local"] if state == "virtual" else labels["make_virtual"]

        item = Nautilus.MenuItem(
            name="RcloneSyncToggle::ToggleLocalAvailability",
            label=label,
            tip=label,
        )
        item.connect("activate", self._activate, paths)
        return [item]

    def _activate(self, _menu, paths):
        for path in paths:
            try:
                subprocess.Popen(
                    ["rclone-sync-toggle", "--toggle", path],
                    stdout=subprocess.DEVNULL,
                    stderr=subprocess.DEVNULL,
                )
            except Exception:
                pass


class RcloneSyncToggleInfoProvider(GObject.GObject, Nautilus.InfoProvider):
    def update_file_info(self, *args):
        # Nautilus 3: (file)
        # Nautilus 4 may vary; handle common signatures defensively.
        file_info = args[0] if args else None
        if file_info is None:
            return Nautilus.OperationResult.COMPLETE

        try:
            path = _path_from_file(file_info)
            if not path:
                return Nautilus.OperationResult.COMPLETE
            state = _item_state(path, _load_profiles())
            if state is None:
                return Nautilus.OperationResult.COMPLETE
            _apply_emblem(file_info, state)
        except Exception:
            pass

        # Nautilus 4 may pass a callback as last arg in async-style providers.
        if args and callable(args[-1]):
            try:
                args[-1]()
            except Exception:
                pass

        return Nautilus.OperationResult.COMPLETE
